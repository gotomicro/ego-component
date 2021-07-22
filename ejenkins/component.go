package ejenkins

import (
	"github.com/gotomicro/ego/client/ehttp"
	"github.com/gotomicro/ego/core/elog"
	"strings"
)

const PackageName = "component.ejenkins"


type Component struct {
	name 	string
	config 	*Config
	jenkins *Jenkins
	logger 	*elog.Component
}

// New ...
func newComponent(compName string, config *Config, logger *elog.Component) *Component {
	if strings.HasSuffix(config.Addr, "/") {
		config.Addr = config.Addr[:len(config.Addr)-1]
	}
	ehttpClient := ehttp.DefaultContainer().Build(
		ehttp.WithDebug(config.Debug),
		ehttp.WithRawDebug(config.RawDebug),
		ehttp.WithAddr(config.Addr),
		ehttp.WithReadTimeout(config.ReadTimeout),
		ehttp.WithEnableAccessInterceptor(config.EnableAccessInterceptor),
		ehttp.WithEnableAccessInterceptorRes(config.EnableAccessInterceptorRes),
	)

	// create Jenkins
	j := &Jenkins{
		Server:    config.Addr,
		Version:   "",
		Raw:       nil,
		Requester: &Requester{
			Base:      config.Addr,
			BasicAuth: &BasicAuth{
				Username: config.Username,
				Password: config.Credential,
			},
			Client:    ehttpClient,
			SslVerify: true,
			logger:    logger,
		},
		logger:    logger,
	}
	// init(check) jenkins
	_, err := j.Init()
	if err != nil {
		logger.Panic("new ejenkins component err", elog.FieldErr(err))
	}
	return &Component{
		name: 	 compName,
		config:  config,
		logger:  logger,
		jenkins: j,

	}
}

// exposes the jenkins REST api
func (c *Component) Info() (*ExecutorResponse, error) {
	return c.jenkins.Info()
}

func (c *Component) SafeRestart() error {
	return c.jenkins.SafeRestart()
}

func (c *Component) CreateFolder(name string, parents ...string) (*Folder, error) {
	return c.jenkins.CreateFolder(name, parents...)
}

func (c *Component) CreateJobInFolder(config string, jobName string, parentIDs ...string) (*Job, error) {
	return c.jenkins.CreateJobInFolder(config, jobName, parentIDs...)
}

func (c *Component) CreateJob(config string, options ...interface{}) (*Job, error) {
	return c.jenkins.CreateJob(config, options...)
}

func (c *Component) UpdateJob(job string, config string) (*Job, error) {
	return c.jenkins.UpdateJob(job, config)
}

func (c *Component) UpdateJobInFolder(jobName string, config string, parentIDs ...string) (*Job, error) {
	return c.jenkins.UpdateJobInFolder(jobName, config, parentIDs...)
}

func (c *Component) RenameJob(job string, name string) (*Job, error) {
	return c.jenkins.RenameJob(job, name)
}

func (c *Component) CopyJob(copyFrom string, newName string) (*Job, error) {
	return c.jenkins.CopyJob(copyFrom, newName)
}

func (c *Component) DeleteJob(name string) (bool, error) {
	return c.jenkins.DeleteJob(name)
}

func (c *Component) BuildJob(JobName string, payload map[string]string) (int64, error) {
	return c.jenkins.BuildJob(JobName, payload)
}

func (c *Component) GetBuildFromQueueID(job *Job, queueId int64) (*Build, error) {
	return c.jenkins.GetBuildFromQueueID(job, queueId)
}

func (c *Component) GetLabel(name string) (*Label, error) {
	return c.jenkins.GetLabel(name)
}

func (c *Component) GetBuild(jobName string, number int64) (*Build, error) {
	return c.jenkins.GetBuild(jobName, number)
}

func (c *Component) GetJob(id string, parentIDs ...string) (*Job, error) {
	return c.jenkins.GetJob(id, parentIDs...)
}

func (c *Component) GetSubJob(parentId string, childId string) (*Job, error) {
	return c.jenkins.GetSubJob(parentId, childId)
}

func (c *Component) GetFolder(id string, parents ...string) (*Folder, error) {
	return c.jenkins.GetFolder(id, parents...)
}

func (c *Component) GetAllBuildIds(job string) ([]JobBuild, error) {
	return c.jenkins.GetAllBuildIds(job)
}

func (c *Component) GetAllJobNames() ([]InnerJob, error) {
	return c.jenkins.GetAllJobNames()
}

func (c *Component) GetAllJobs() ([]*Job, error) {
	return c.jenkins.GetAllJobs()
}

func (c *Component) GetQueue() (*Queue, error) {
	return c.jenkins.GetQueue()
}

func (c *Component) GetQueueItem(id int64) (*Task, error) {
	return c.jenkins.GetQueueItem(id)
}

func (c *Component) GetArtifactData(id string) (*FingerPrintResponse, error) {
	return c.jenkins.GetArtifactData(id)
}

func (c *Component) GetPlugins(depth int) (*Plugins, error) {
	return c.jenkins.GetPlugins(depth)
}

func (c *Component) UninstallPlugin(name string) error {
	return c.jenkins.UninstallPlugin(name)
}

func (c *Component) HasPlugin(name string) (*Plugin, error) {
	return c.jenkins.HasPlugin(name)
}

func (c *Component) InstallPlugin(name string, version string) error {
	return c.jenkins.InstallPlugin(name, version)
}

func (c *Component) ValidateFingerPrint(id string) (bool, error) {
	return c.jenkins.ValidateFingerPrint(id)
}

func (c *Component) GetView(name string) (*View, error) {
	return c.jenkins.GetView(name)
}

func (c *Component) GetAllViews() ([]*View, error) {
	return c.jenkins.GetAllViews()
}

func (c *Component) CreateView(name string, viewType string) (*View, error) {
	return c.jenkins.CreateView(name, viewType)
}

func (c *Component) GetCredentialManager(folders ...string) *CredentialsManager {
	return c.jenkins.NewCredentialsManager(folders...)
}

func (c *Component) ListCredential(domain string, folders ...string) ([]string, error) {
	return c.jenkins.NewCredentialsManager(folders...).List(domain)
}

func (c *Component) GetSingleCredential(domain string, id string, cred interface{}, folders ...string) error {
	return c.jenkins.NewCredentialsManager(folders...).GetSingle(domain, id, cred)
}

func (c *Component) AddCredential(domain string, cred interface{}, folders ...string) error {
	return c.jenkins.NewCredentialsManager(folders...).Add(domain, cred)
}
func (c *Component) DeleteCredential(domain string, id string, folders ...string) error {
	return c.jenkins.NewCredentialsManager(folders...).Delete(domain, id)
}

func (c *Component) UpdateCredential(domain string, id string, cred interface{}, folders ...string) error {
	return c.jenkins.NewCredentialsManager(folders...).Update(domain, id, cred)
}



