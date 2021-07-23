package ejenkins

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gotomicro/ego/core/elog"
)

// Basic Authentication
type BasicAuth struct {
	Username string
	Password string
}

type Jenkins struct {
	Server    string
	Version   string
	Raw       *ExecutorResponse
	Requester *Requester
	logger    *elog.Component
}

func (j *Jenkins) Init() (*Jenkins, error) {
	// Check Connection
	j.Raw = new(ExecutorResponse)
	rsp, err := j.Requester.GetJSON("/", nil, j.Raw, nil)
	if err != nil {
		return nil, err
	}

	j.Version = rsp.Header().Get("X-Jenkins")
	if j.Raw == nil || rsp.StatusCode() != http.StatusOK {
		return nil, errors.New("connect to jenkins Failed, Please verify that the host and credentials are correct")
	}
	return j, nil
}

// poll jenkins, update jenkins.raw
func (j *Jenkins) Poll() (int, error) {
	resp, err := j.Requester.GetJSON("/", nil, j.Raw, nil)
	if err != nil {
		return 0, err
	}
	return resp.StatusCode(), nil
}

// Get Basic Information About Jenkins
func (j *Jenkins) Info() (*ExecutorResponse, error) {
	rsp, err := j.Requester.GetJSON("/", nil, j.Raw, nil)
	if err != nil {
		return nil, err
	}
	j.Version = rsp.Header().Get("X-Jenkins")
	return j.Raw, nil
}

// SafeRestart jenkins, restart will be done when there are no jobs running
func (j *Jenkins) SafeRestart() error {
	_, err := j.Requester.Post("/safeRestart", nil, struct{}{}, map[string]string{})
	return err
}

// Create a new folder
// This folder can be nested in other parent folders
// Example: CreateFolder("newFolder", "grandparentFolder", "parentFolder")
func (j *Jenkins) CreateFolder(name string, parents ...string) (*Folder, error) {
	folderObj := &Folder{Jenkins: j, Raw: new(FolderResponse), Base: "/job/" + strings.Join(append(parents, name), "/job/")}
	folder, err := folderObj.Create(name)
	if err != nil {
		return nil, err
	}
	return folder, nil
}

/*
Create a new job in the folder
	Example: CreateJobInFolder("<config></config>", "newJobName", "folder1", "folder2");
		if create successfully, the url of the new job will be {JenkinsHost}/job/folder1/job/folder2/job/newJobName
*/
func (j *Jenkins) CreateJobInFolder(config string, jobName string, parentIDs ...string) (*Job, error) {
	jobObj := Job{Jenkins: j, Raw: new(JobResponse), Base: "/job/" + strings.Join(append(parentIDs, jobName), "/job/")}
	qr := map[string]string{
		"name": jobName,
	}
	job, err := jobObj.Create(config, qr)
	if err != nil {
		return nil, err
	}
	return job, nil
}

// Create a new job from config(xml) File
// Method takes XML string as first parameter, and if the name is not specified in the config file
// takes name as string as second parameter
// e.g CreateJob("<config></config>","newJobName")
func (j *Jenkins) CreateJob(config string, options ...interface{}) (*Job, error) {
	qr := make(map[string]string)
	if len(options) > 0 {
		qr["name"] = options[0].(string)
	} else {
		return nil, errors.New("creating Job failed, job name is missing")
	}
	jobObj := Job{Jenkins: j, Raw: new(JobResponse), Base: "/job/" + qr["name"]}
	job, err := jobObj.Create(config, qr)
	if err != nil {
		return nil, err
	}
	return job, nil
}

// Update a job.
// If a job is exist, update its config
func (j *Jenkins) UpdateJob(job string, config string) (*Job, error) {
	jobObj := Job{Jenkins: j, Raw: new(JobResponse), Base: "/job/" + job}
	if err := jobObj.UpdateConfig(config); err != nil {
		return nil, err
	}
	return &jobObj, nil
}

// Update job in folder
// if parents is empty, this function equal to UpdateJob func.
// Example: UpdateJobInFolder("<xml></xml>", "jobName", "myFolder", "parentFolder", "grandparentFolder",...)
func (j *Jenkins) UpdateJobInFolder(jobName string, config string, parentIDs ...string) (*Job, error) {
	jobObj := Job{Jenkins: j, Raw: new(JobResponse), Base: "/job/" + strings.Join(append(parentIDs, jobName), "/job/")}
	if err := jobObj.UpdateConfig(config); err != nil {
		return nil, err
	}
	return &jobObj, nil
}

// Rename a job.
// First parameter job old name, Second parameter job new name.
func (j *Jenkins) RenameJob(job string, name string) (*Job, error) {
	jobObj := Job{Jenkins: j, Raw: new(JobResponse), Base: "/job/" + job}
	if _, err := jobObj.Rename(name); err != nil {
		return nil, err
	}
	return &jobObj, nil
}

// Create a copy of a job.
// First parameter Name of the job to copy from, Second parameter new job name.
func (j *Jenkins) CopyJob(copyFrom string, newName string) (*Job, error) {
	job := Job{Jenkins: j, Raw: new(JobResponse), Base: "/job/" + copyFrom}
	_, err := job.Poll()
	if err != nil {
		return nil, err
	}
	return job.Copy(newName)
}

// Delete a job.
func (j *Jenkins) DeleteJob(name string) (bool, error) {
	job := Job{Jenkins: j, Raw: new(JobResponse), Base: "/job/" + name}
	return job.Delete()
}

// Get a temp job object
func (j *Jenkins) GetJobObj(name string) *Job {
	return &Job{Jenkins: j, Raw: new(JobResponse), Base: "/job/" + name}
}

// Invoke a job.
// First parameter job name, second parameter is optional Build parameters.
// Returns queue id
// Node, this function cannot post file(s)
func (j *Jenkins) BuildJob(JobName string, payload map[string]string) (int64, error) {
	job := j.GetJobObj(JobName)
	return job.InvokeSimple(payload)
}

// A task in queue will be assigned a build number in a job after a few seconds.
// this function will return the build object.
func (j *Jenkins) GetBuildFromQueueID(job *Job, queueId int64) (*Build, error) {
	task, err := j.GetQueueItem(queueId)
	if err != nil {
		return nil, err
	}
	// Jenkins queue API has about 4.7second quiet period
	for task.Raw.Executable.Number == 0 {
		time.Sleep(1000 * time.Millisecond)
		_, err = task.Poll()
		if err != nil {
			return nil, err
		}
	}

	build, err := job.GetBuild(task.Raw.Executable.Number)
	if err != nil {
		return nil, err
	}
	return build, nil
}

func (j *Jenkins) GetLabel(name string) (*Label, error) {
	label := Label{Jenkins: j, Raw: new(LabelResponse), Base: "/label/" + name}
	status, err := label.Poll()
	if err != nil {
		return nil, err
	}
	if status == 200 {
		return &label, nil
	}
	return nil, errors.New("no label found")
}

func (j *Jenkins) GetBuild(jobName string, number int64) (*Build, error) {
	job, err := j.GetJob(jobName)
	if err != nil {
		return nil, err
	}
	build, err := job.GetBuild(number)

	if err != nil {
		return nil, err
	}
	return build, nil
}

func (j *Jenkins) GetJob(id string, parentIDs ...string) (*Job, error) {
	job := Job{Jenkins: j, Raw: new(JobResponse), Base: "/job/" + strings.Join(append(parentIDs, id), "/job/")}
	status, err := job.Poll()
	if err != nil {
		return nil, err
	}
	if status == 200 {
		return &job, nil
	}
	return nil, errors.New(strconv.Itoa(status))
}

func (j *Jenkins) GetSubJob(parentId string, childId string) (*Job, error) {
	job := Job{Jenkins: j, Raw: new(JobResponse), Base: "/job/" + parentId + "/job/" + childId}
	status, err := job.Poll()
	if err != nil {
		return nil, fmt.Errorf("trouble polling job: %v", err)
	}
	if status == 200 {
		return &job, nil
	}
	return nil, errors.New(strconv.Itoa(status))
}

func (j *Jenkins) GetFolder(id string, parents ...string) (*Folder, error) {
	folder := Folder{Jenkins: j, Raw: new(FolderResponse), Base: "/job/" + strings.Join(append(parents, id), "/job/")}
	status, err := folder.Poll()
	if err != nil {
		return nil, fmt.Errorf("trouble polling folder: %v", err)
	}
	if status == 200 {
		return &folder, nil
	}
	return nil, errors.New(strconv.Itoa(status))
}

// Get all builds Numbers and URLS for a specific job.
// There are only build IDs here,
// To get all the other info of the build use GetBuild(job,buildNumber)
// or job.GetBuild(buildNumber)
func (j *Jenkins) GetAllBuildIds(job string) ([]JobBuild, error) {
	jobObj, err := j.GetJob(job)
	if err != nil {
		return nil, err
	}
	return jobObj.GetAllBuildIds()
}

// Get Only Array of Job Names, Color, URL
// Does not query each single Job.
func (j *Jenkins) GetAllJobNames() ([]InnerJob, error) {
	exec := Executor{Raw: new(ExecutorResponse), Jenkins: j}
	_, err := j.Requester.GetJSON("/", nil, exec.Raw, nil)

	if err != nil {
		return nil, err
	}

	return exec.Raw.Jobs, nil
}

// Get All Possible Job Objects.
// Each job will be queried.
func (j *Jenkins) GetAllJobs() ([]*Job, error) {
	exec := Executor{Raw: new(ExecutorResponse), Jenkins: j}
	_, err := j.Requester.GetJSON("/", nil, exec.Raw, nil)

	if err != nil {
		return nil, err
	}

	jobs := make([]*Job, len(exec.Raw.Jobs))
	for i, job := range exec.Raw.Jobs {
		ji, err := j.GetJob(job.Name)
		if err != nil {
			return nil, err
		}
		jobs[i] = ji
	}
	return jobs, nil
}

// Returns a Queue
func (j *Jenkins) GetQueue() (*Queue, error) {
	q := &Queue{Jenkins: j, Raw: new(queueResponse), Base: j.GetQueueUrl()}
	_, err := q.Poll()
	if err != nil {
		return nil, err
	}
	return q, nil
}

func (j *Jenkins) GetQueueUrl() string {
	return "/queue"
}

// GetQueueItem returns a single queue Task
func (j *Jenkins) GetQueueItem(id int64) (*Task, error) {
	t := &Task{Raw: new(taskResponse), Jenkins: j, Base: j.getQueueItemURL(id)}
	_, err := t.Poll()
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (j *Jenkins) getQueueItemURL(id int64) string {
	return fmt.Sprintf("/queue/item/%d", id)
}

// Get Artifact data by Hash
func (j *Jenkins) GetArtifactData(id string) (*FingerPrintResponse, error) {
	fp := FingerPrint{Jenkins: j, Base: "/fingerprint/", Id: id, Raw: new(FingerPrintResponse)}
	return fp.GetInfo()
}

// Returns the list of all plugins installed on the Jenkins server.
// You can supply depth parameter, to limit how much data is returned.
func (j *Jenkins) GetPlugins(depth int) (*Plugins, error) {
	p := Plugins{Jenkins: j, Raw: new(PluginResponse), Base: "/pluginManager", Depth: depth}
	_, err := p.Poll()
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// UninstallPlugin plugin otherwise returns error
func (j *Jenkins) UninstallPlugin(name string) error {
	url := fmt.Sprintf("/pluginManager/plugin/%s/doUninstall", name)
	resp, err := j.Requester.Post(url, nil, struct{}{}, map[string]string{})
	if err != nil {
		return err
	}
	if resp.StatusCode() != 200 {
		return fmt.Errorf("invalid status code returned: %d", resp.StatusCode())
	}
	return err
}

// Check if the plugin is installed on the server.
// Depth level 1 is used. If you need to go deeper, you can use GetPlugins, and iterate through them.
func (j *Jenkins) HasPlugin(name string) (*Plugin, error) {
	p, err := j.GetPlugins(1)

	if err != nil {
		return nil, err
	}
	return p.Contains(name), nil
}

// InstallPlugin with given version and name
func (j *Jenkins) InstallPlugin(name string, version string) error {
	xml := fmt.Sprintf(`<jenkins><install plugin="%s@%s" /></jenkins>`, name, version)
	resp, err := j.Requester.PostXML("/pluginManager/installNecessaryPlugins", nil, xml,
		j.Raw, map[string]string{})
	if err != nil {
		return err
	}
	if resp.StatusCode() != 200 {
		return fmt.Errorf("invalid status code returned: %d", resp.StatusCode())
	}
	return err
}

// Verify FingerPrint
func (j *Jenkins) ValidateFingerPrint(id string) (bool, error) {
	fp := FingerPrint{Jenkins: j, Base: "/fingerprint/", Id: id, Raw: new(FingerPrintResponse)}
	valid, err := fp.Valid()
	if err != nil {
		return false, err
	}
	if valid {
		return true, nil
	}
	return false, nil
}

func (j *Jenkins) GetView(name string) (*View, error) {
	url := "/view/" + name
	view := View{Jenkins: j, Raw: new(ViewResponse), Base: url}
	_, err := view.Poll()
	if err != nil {
		return nil, err
	}
	return &view, nil
}

func (j *Jenkins) GetAllViews() ([]*View, error) {
	_, err := j.Poll()
	if err != nil {
		return nil, err
	}
	views := make([]*View, len(j.Raw.Views))
	for i, v := range j.Raw.Views {
		views[i], _ = j.GetView(v.Name)
	}
	return views, nil
}

// Create View
// First Parameter - name of the View
// Second parameter - Type
// Possible Types(const string of this pkg mops-be/pkg/service/jenkins/api):
// 		ListView
// 		NestedView
// 		MyView
// 		DashboardView
// 		PipelineView
// Example: CreateView("newView", LIST_VIEW)
func (j *Jenkins) CreateView(name string, viewType string) (*View, error) {
	view := &View{Jenkins: j, Raw: new(ViewResponse), Base: "/view/" + name}
	endpoint := "/createView"
	data := map[string]string{
		"name":   name,
		"mode":   viewType,
		"Submit": "OK",
		"json": getJsonString(map[string]string{
			"name": name,
			"mode": viewType,
		}),
	}
	r, err := j.Requester.Post(endpoint, nil, view.Raw, data)

	if err != nil {
		return nil, err
	}

	if r.StatusCode() == 200 {
		return j.GetView(name)
	}
	return nil, errors.New(strconv.Itoa(r.StatusCode()))
}
