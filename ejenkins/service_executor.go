package ejenkins

type Executor struct {
	Raw     *ExecutorResponse
	Jenkins *Jenkins
}
type ViewData struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}
type ExecutorResponse struct {
	AssignedLabels  []map[string]string `json:"assignedLabels"`
	Description     interface{}         `json:"description"`
	Jobs            []InnerJob          `json:"jobs"`
	Mode            string              `json:"mode"`
	NodeDescription string              `json:"nodeDescription"`
	NodeName        string              `json:"nodeName"`
	NumExecutors    int64               `json:"numExecutors"`
	OverallLoad     struct{}            `json:"overallLoad"`
	PrimaryView     struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"primaryView"`
	QuietingDown   bool       `json:"quietingDown"`
	SlaveAgentPort int64      `json:"slaveAgentPort"`
	UnlabeledLoad  struct{}   `json:"unlabeledLoad"`
	UseCrumbs      bool       `json:"useCrumbs"`
	UseSecurity    bool       `json:"useSecurity"`
	Views          []ViewData `json:"views"`
}
