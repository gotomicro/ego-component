package ejenkins

import (
	"fmt"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"
)

type Job struct {
	Raw     *JobResponse
	Jenkins *Jenkins
	Base    string
}

type JobBuild struct {
	Number int64
	URL    string
}

type InnerJob struct {
	Class string `json:"_class"`
	Name  string `json:"name"`
	Url   string `json:"url"`
	Color string `json:"color"`
}

type ParameterDefinition struct {
	DefaultParameterValue struct {
		Name  string      `json:"name"`
		Value interface{} `json:"value"`
	} `json:"defaultParameterValue"`
	Description string `json:"description"`
	Name        string `json:"name"`
	Type        string `json:"type"`
}

type JobResponse struct {
	Class              string `json:"_class"`
	Actions            []generalObj
	Buildable          bool `json:"buildable"`
	Builds             []JobBuild
	Color              string      `json:"color"`
	ConcurrentBuild    bool        `json:"concurrentBuild"`
	Description        string      `json:"description"`
	DisplayName        string      `json:"displayName"`
	DisplayNameOrNull  interface{} `json:"displayNameOrNull"`
	DownstreamProjects []InnerJob  `json:"downstreamProjects"`
	FirstBuild         JobBuild
	FullName           string `json:"fullName"`
	FullDisplayName    string `json:"fullDisplayName"`
	HealthReport       []struct {
		Description   string `json:"description"`
		IconClassName string `json:"iconClassName"`
		IconUrl       string `json:"iconUrl"`
		Score         int64  `json:"score"`
	} `json:"healthReport"`
	InQueue               bool     `json:"inQueue"`
	KeepDependencies      bool     `json:"keepDependencies"`
	LastBuild             JobBuild `json:"lastBuild"`
	LastCompletedBuild    JobBuild `json:"lastCompletedBuild"`
	LastFailedBuild       JobBuild `json:"lastFailedBuild"`
	LastStableBuild       JobBuild `json:"lastStableBuild"`
	LastSuccessfulBuild   JobBuild `json:"lastSuccessfulBuild"`
	LastUnstableBuild     JobBuild `json:"lastUnstableBuild"`
	LastUnsuccessfulBuild JobBuild `json:"lastUnsuccessfulBuild"`
	Name                  string   `json:"name"`
	NextBuildNumber       int64    `json:"nextBuildNumber"`
	Property              []struct {
		ParameterDefinitions []ParameterDefinition `json:"parameterDefinitions"`
	} `json:"property"`
	QueueItem        interface{} `json:"queueItem"`
	Scm              struct{}    `json:"scm"`
	UpstreamProjects []InnerJob  `json:"upstreamProjects"`
	URL              string      `json:"url"`
	Jobs             []InnerJob  `json:"jobs"`
	PrimaryView      *ViewData   `json:"primaryView"`
	Views            []ViewData  `json:"views"`
}


type BuildFilter struct {
	EarlySecondsTimestamp 	int64
	FilterParams 			map[string]string
}

func (j *Job) parentBase() string {
	return j.Base[:strings.LastIndex(j.Base, "/job/")]
}

type History struct {
	BuildDisplayName string
	BuildNumber      int
	BuildStatus      string
	BuildTimestamp   int64
}

func (j *Job) GetName() string {
	return j.Raw.Name
}

func (j *Job) GetDescription() string {
	return j.Raw.Description
}

func (j *Job) GetDetails() *JobResponse {
	return j.Raw
}

func (j *Job) GetBuild(id int64) (*Build, error) {
	build := Build{Jenkins: j.Jenkins, Job: j, Raw: new(BuildResponse), Depth: 1, Base: j.Base + "/" + strconv.FormatInt(id, 10)}
	status, err := build.Poll()
	if err != nil {
		return nil, err
	}
	if status == 200 {
		return &build, nil
	}
	return nil, errors.New(strconv.Itoa(status))
}

func (j *Job) getBuildByType(buildType string) (*Build, error) {
	allowed := map[string]JobBuild{
		"lastStableBuild":     j.Raw.LastStableBuild,
		"lastSuccessfulBuild": j.Raw.LastSuccessfulBuild,
		"lastBuild":           j.Raw.LastBuild,
		"lastCompletedBuild":  j.Raw.LastCompletedBuild,
		"firstBuild":          j.Raw.FirstBuild,
		"lastFailedBuild":     j.Raw.LastFailedBuild,
	}
	number := ""
	if val, ok := allowed[buildType]; ok {
		number = strconv.FormatInt(val.Number, 10)
	} else {
		return nil, errors.New("no such build")
	}
	if number == "0" {
		return nil, errors.New("this job do not have any build history currently")
	}
	build := Build{
		Jenkins: j.Jenkins,
		Depth:   1,
		Job:     j,
		Raw:     new(BuildResponse),
		Base:    j.Base + "/" + number}
	status, err := build.Poll()
	if err != nil {
		return nil, err
	}
	if status == 200 {
		return &build, nil
	}
	return nil, errors.New(strconv.Itoa(status))
}

func (j *Job) GetLastSuccessfulBuild() (*Build, error) {
	return j.getBuildByType("lastSuccessfulBuild")
}

func (j *Job) GetFirstBuild() (*Build, error) {
	return j.getBuildByType("firstBuild")
}

func (j *Job) GetLastBuild() (*Build, error) {
	return j.getBuildByType("lastBuild")
}

func (j *Job) GetLastStableBuild() (*Build, error) {
	return j.getBuildByType("lastStableBuild")
}

func (j *Job) GetLastFailedBuild() (*Build, error) {
	return j.getBuildByType("lastFailedBuild")
}

func (j *Job) GetLastCompletedBuild() (*Build, error) {
	return j.getBuildByType("lastCompletedBuild")
}

func (j *Job) GetBuildsFields(fields []string, custom interface{}) error {
	if fields == nil || len(fields) == 0 {
		return fmt.Errorf("one or more field value needs to be specified")
	}
	// limit overhead using builds instead of allBuilds, which returns the last 100 build
	_, err := j.Jenkins.Requester.GetJSON(j.Base, nil, &custom, map[string]string{"tree": "builds[" + strings.Join(fields, ",") + "]"})
	if err != nil {
		return err
	}
	return nil
}

// Returns All Builds with Number and URL
func (j *Job) GetAllBuildIds() ([]JobBuild, error) {
	var buildsResp struct {
		Builds []JobBuild `json:"allBuilds"`
	}
	_, err := j.Jenkins.Requester.GetJSON(j.Base, nil, &buildsResp, map[string]string{"tree": "allBuilds[number,url]"})
	if err != nil {
		return nil, err
	}
	return buildsResp.Builds, nil
}

func (j *Job) GetUpstreamJobsMetadata() []InnerJob {
	return j.Raw.UpstreamProjects
}

func (j *Job) GetDownstreamJobsMetadata() []InnerJob {
	return j.Raw.DownstreamProjects
}

func (j *Job) GetInnerJobsMetadata() []InnerJob {
	return j.Raw.Jobs
}

func (j *Job) GetUpstreamJobs() ([]*Job, error) {
	jobs := make([]*Job, len(j.Raw.UpstreamProjects))
	for i, job := range j.Raw.UpstreamProjects {
		ji, err := j.Jenkins.GetJob(job.Name)
		if err != nil {
			return nil, err
		}
		jobs[i] = ji
	}
	return jobs, nil
}

func (j *Job) GetDownstreamJobs() ([]*Job, error) {
	jobs := make([]*Job, len(j.Raw.DownstreamProjects))
	for i, job := range j.Raw.DownstreamProjects {
		ji, err := j.Jenkins.GetJob(job.Name)
		if err != nil {
			return nil, err
		}
		jobs[i] = ji
	}
	return jobs, nil
}

func (j *Job) GetInnerJob(id string) (*Job, error) {
	job := Job{Jenkins: j.Jenkins, Raw: new(JobResponse), Base: j.Base + "/job/" + id}
	status, err := job.Poll()
	if err != nil {
		return nil, err
	}
	if status == 200 {
		return &job, nil
	}
	return nil, errors.New(strconv.Itoa(status))
}

func (j *Job) GetInnerJobs() ([]*Job, error) {
	jobs := make([]*Job, len(j.Raw.Jobs))
	for i, job := range j.Raw.Jobs {
		ji, err := j.GetInnerJob(job.Name)
		if err != nil {
			return nil, err
		}
		jobs[i] = ji
	}
	return jobs, nil
}

func (j *Job) Enable() (bool, error) {
	resp, err := j.Jenkins.Requester.Post(j.Base+"/enable", nil, nil, nil)
	if err != nil {
		return false, err
	}
	if resp.StatusCode() != 200 {
		return false, errors.New(strconv.Itoa(resp.StatusCode()))
	}
	return true, nil
}

func (j *Job) Disable() (bool, error) {
	resp, err := j.Jenkins.Requester.Post(j.Base+"/disable", nil, nil, nil)
	if err != nil {
		return false, err
	}
	if resp.StatusCode() != 200 {
		return false, errors.New(strconv.Itoa(resp.StatusCode()))
	}
	return true, nil
}

func (j *Job) Delete() (bool, error) {
	resp, err := j.Jenkins.Requester.Post(j.Base+"/doDelete", nil, nil, nil)
	if err != nil {
		return false, err
	}
	if resp.StatusCode() != 200 {
		return false, errors.New(strconv.Itoa(resp.StatusCode()))
	}
	return true, nil
}

func (j *Job) Rename(name string) (bool, error) {
	_, err := j.Jenkins.Requester.Post(j.Base+"/doRename", map[string]string{"newName": name}, nil, nil)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (j *Job) Create(config string, qr ...interface{}) (*Job, error) {
	var querystring map[string]string
	if len(qr) > 0 {
		querystring = qr[0].(map[string]string)
	}
	resp, err := j.Jenkins.Requester.PostXML(j.parentBase()+"/createItem", nil, config, j.Raw, querystring)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() == 200 {
		j.Poll()
		return j, nil
	}
	return nil, errors.New(strconv.Itoa(resp.StatusCode()))
}

func (j *Job) Copy(destinationName string) (*Job, error) {
	qr := map[string]string{"name": destinationName, "from": j.GetName(), "mode": "copy"}
	resp, err := j.Jenkins.Requester.Post(j.parentBase()+"/createItem", nil, nil, qr)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() == 200 {
		newJob := &Job{Jenkins: j.Jenkins, Raw: new(JobResponse), Base: "/job/" + destinationName}
		_, err := newJob.Poll()
		if err != nil {
			return nil, err
		}
		return newJob, nil
	}
	return nil, errors.New(strconv.Itoa(resp.StatusCode()))
}

func (j *Job) UpdateConfig(config string) error {

	var querystring map[string]string

	resp, err := j.Jenkins.Requester.PostXML(j.Base+"/config.xml", nil, config, nil, querystring)
	if err != nil {
		return err
	}
	if resp.StatusCode() == 200 {
		if _, err := j.Poll(); err != nil {
			return err
		}
		return nil
	}
	return errors.New(strconv.Itoa(resp.StatusCode()))

}

func (j *Job) GetConfig() (string, error) {
	var data string
	_, err := j.Jenkins.Requester.GetXML(j.Base+"/config.xml", nil, &data, nil)
	if err != nil {
		return "", err
	}
	return data, nil
}

func (j *Job) GetParameters() ([]ParameterDefinition, error) {
	_, err := j.Poll()
	if err != nil {
		return nil, err
	}
	var parameters []ParameterDefinition
	for _, property := range j.Raw.Property {
		parameters = append(parameters, property.ParameterDefinitions...)
	}
	return parameters, nil
}

func (j *Job) IsQueued() (bool, error) {
	if _, err := j.Poll(); err != nil {
		return false, err
	}
	return j.Raw.InQueue, nil
}

func (j *Job) IsRunning() (bool, error) {
	if _, err := j.Poll(); err != nil {
		return false, err
	}
	lastBuild, err := j.GetLastBuild()
	if err != nil {
		return false, err
	}
	return lastBuild.IsRunning(), nil
}

func (j *Job) IsEnabled() (bool, error) {
	if _, err := j.Poll(); err != nil {
		return false, err
	}
	return j.Raw.Color != "disabled", nil
}

//func (j *Job) HasQueuedBuild() {
//	//TODO: "Not Implemented yet"
//}

func (j *Job) InvokeSimple(payload map[string]string) (int64, error) {
	isQueued, err := j.IsQueued()
	if err != nil {
		return 0, err
	}
	if isQueued {
		j.Jenkins.logger.Error("Job is already running, do not need to be invoked", zap.String("job", j.GetName()))
		return 0, nil
	}

	endpoint := "/build"
	parameters, err := j.GetParameters()
	if err != nil {
		return 0, err
	}
	if len(parameters) > 0 {
		endpoint = "/buildWithParameters"
	}
	resp, err := j.Jenkins.Requester.Post(j.Base+endpoint, payload, nil, nil)
	if err != nil {
		return 0, err
	}

	if resp.StatusCode() != 200 && resp.StatusCode() != 201 {
		return 0, fmt.Errorf("could not invoke job %q: %s", j.GetName(), resp.Status())
	}

	location := resp.Header().Get("Location")
	if location == "" {
		return 0, errors.New("don't have key \"Location\" in response of header")
	}

	u, err := url.Parse(location)
	if err != nil {
		return 0, err
	}

	number, err := strconv.ParseInt(path.Base(u.Path), 10, 64)
	if err != nil {
		return 0, err
	}

	return number, nil
}

/*
Job.Invoke function will invoke a new build of the Job based on the "payload"(formData) and "buildParams"(support files)
return:
	error:  nil or the error encountered in invoking
	*build: the pointer which point to the invoked build; nil if error != nil
Note:
	1. the value of the same key of payload will be replaced by corresponding value of the same key which in buildParams
	2. this function support post(upload) file(s), please set file(s) in buildParams
	3. about the return. the *Build can be nil, even if the error is nil!
		the error is nil, just means the job has been invoked successfully, but the invoked build may not be found currently.
		so, if using the result *Build which this function returned, please check it is not nil before using it,
		even the error which this function returned is nil
 */
func (j *Job) Invoke(payload map[string]string, buildParams *BuildParameters, securityToken string) (*Build, error) {
	isQueued, err := j.IsQueued()
	if err != nil {
		return nil, err
	}
	if isQueued {
		j.Jenkins.logger.Warnf("Job %s is already running(queued)", j.GetName())
		return nil, nil
	}

	endpoint := "/build"
	parameters, err := j.GetParameters()
	if err != nil {
		return nil, errors.Wrap(err, "get job parameters failed.")
	}
	if len(parameters) > 0 {
		endpoint = "/buildWithParameters"
	}
	queryParams := map[string]string{}
	if securityToken != "" {
		queryParams["token"] = securityToken
	}

	resp, err := j.Jenkins.Requester.PostFiles(j.Base+endpoint, payload, nil, queryParams, buildParams)
	if err != nil {
		return nil, errors.Wrap(err, "Post invoking request failed.")
	}
	if resp.StatusCode() != 200 && resp.StatusCode() != 201 {
		return nil, errors.Errorf("could not invoke job %s, status: %s", j.GetName(), resp.Status())
	}

	location := resp.Header().Get("Location")
	if location == "" {
		j.Jenkins.logger.Warn("Job has been invoked successfully, but the response do not have key Location in header.")
		return nil, nil
	}

	u, err := url.Parse(location)
	if err != nil {
		j.Jenkins.logger.Error("parse Location of Invoked job response header failed.", zap.Error(err))
		return nil, nil
	}

	queueId, err := strconv.ParseInt(path.Base(u.Path), 10, 64)
	if err != nil {
		j.Jenkins.logger.Error("parse InvokedJobQueueNum from response failed.", zap.Error(err))
		return nil, nil
	}
	return j.Jenkins.GetBuildFromQueueID(j, queueId)
}

func (j *Job) GetBuildByFilter(filter *BuildFilter, intervalSeconds int, maxTryTimes uint) (targetBuild *Build, err error) {
	interval := 6
	if intervalSeconds > 0 && intervalSeconds < interval{
		interval = intervalSeconds
	}
	maxTry := uint(10)
	if maxTryTimes < maxTry && maxTryTimes > 0{
		maxTry = maxTryTimes
	}
	if len(filter.FilterParams) <= 0 {
		return nil, errors.New("filter params cannot be empty")
	}
	if filter.EarlySecondsTimestamp <= 0 {
		firstBuild, err := j.GetFirstBuild()
		if err == nil {
			filter.EarlySecondsTimestamp = firstBuild.GetSecondsTimestamp()
		}
	}
	tryTimes := uint(0)
	for {
		targetBuild, err = j.getBuildByFilterOnce(filter)
		// found targetBuild just return
		if targetBuild != nil {
			return targetBuild, nil
		}
		time.Sleep(time.Duration(interval)*time.Second)
		tryTimes ++
		if tryTimes > maxTry {
			break
		}
	}
	return nil, errors.New("not found target build based on filter")
}

// Note: if invoke this function immediately after job.Invoke, may be failed to get the invoked build
// 		so sleep several seconds after job.Invoke to Invoke this func, and try several times.
func (j *Job) getBuildByFilterOnce(filter *BuildFilter) (targetBuild *Build, err error) {
	_, err = j.Poll()
	if err != nil {
		return
	}

	var currentBuild *Build
	currentBuild, err = j.GetLastBuild()
	if err != nil {
		return nil, err
	}
	for {
		if currentBuild.GetSecondsTimestamp() < filter.EarlySecondsTimestamp {
			break
		}
		currentBuildParams := currentBuild.GetParameters()
		if currentBuildParams == nil || len(currentBuildParams) == 0{
			continue
		}
		var tempFilterMap = make(map[string]string)
		for k, v := range filter.FilterParams {
			tempFilterMap[k] = v
		}
		for _, buildParam := range currentBuildParams {
			if len(tempFilterMap) == 0 {break}
			if wantValue, exist := tempFilterMap[buildParam.Name]; exist {
				if wantValue == buildParam.Value {
					delete(tempFilterMap, buildParam.Name)
				}
			}
		}
		// if len(tempFilterMap) == 0  means found the target build
		if len(tempFilterMap) == 0 {
			targetBuild = currentBuild
			break
		}
		// check whether currentBuild has previous build or not. if its the first build of the job, then break
		if currentBuild.Raw.PreviousBuild.Number  == 0 {
			break
		}
		// set currentBuild to previous build of the job
		currentBuild, err = j.GetBuild(currentBuild.Raw.PreviousBuild.Number)
		if err != nil {
			break
		}

	}

	if targetBuild == nil {
		return nil, errors.New("not found target build based on filter currently")
	}
	return targetBuild, nil
}

func (j *Job) Poll() (int, error) {
	response, err := j.Jenkins.Requester.GetJSON(j.Base, nil, j.Raw, nil)
	if err != nil {
		return 0, err
	}
	return response.StatusCode(), nil
}

//func (j *Job) History() ([]*History, error) {
//	var s string
//	_, err := j.Jenkins.Requester.Get(j.Base+"/buildHistory/ajax", &s, nil)
//	if err != nil {
//		return nil, err
//	}
//
//	return parseBuildHistory(strings.NewReader(s)), nil
//}

func (pr *PipelineRun) ProceedInput() (bool, error) {
	actions, _ := pr.GetPendingInputActions()
	params := make(map[string]string)
	payload := map[string]string{"inputId": actions[0].ID, "json": getJsonString(params)}

	href := pr.Base + "/wfapi/inputSubmit"

	resp, err := pr.Job.Jenkins.Requester.Post(href, payload, nil, nil)
	if err != nil {
		return false, err
	}
	if resp.StatusCode() != 200 {
		return false, errors.New(strconv.Itoa(resp.StatusCode()))
	}
	return true, nil
}

func (pr *PipelineRun) AbortInput() (bool, error) {
	actions, _ := pr.GetPendingInputActions()
	params := make(map[string]string)
	payload := map[string]string{"json": getJsonString(params)}

	href := pr.Base + "/input/" + actions[0].ID + "/abort"

	resp, err := pr.Job.Jenkins.Requester.Post(href, payload, nil, nil)
	if err != nil {
		return false, err
	}
	if resp.StatusCode() != 200 {
		return false, errors.New(strconv.Itoa(resp.StatusCode()))
	}
	return true, nil
}
