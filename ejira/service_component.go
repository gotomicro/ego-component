package ejira

import (
	"encoding/json"
	"fmt"
)

// JiraComponent represents a single component of a project
type JiraComponent struct {
	Self             string `json:"self,omitempty"`
	ID               string `json:"id,omitempty"`
	Name             string `json:"name,omitempty"`
	Description      string `json:"description,omitempty"`
	Lead             *User  `json:"lead,omitempty"`
	LeadUserName     string `json:"leadUserName,omitempty"`
	AssigneeType     string `json:"assigneeType,omitempty"`
	Assignee         *User  `json:"assignee,omitempty"`
	RealAssigneeType string `json:"realAssigneeType,omitempty"`
	RealAssignee     *User  `json:"realAssignee,omitempty"`
	Project          string `json:"project,omitempty"`
	ProjectID        int    `json:"projectId,omitempty"`
}

// JiraComponentUpdateReq ...
type JiraComponentUpdateReq struct {
	ID           string  `json:"id,omitempty"`
	Name         *string `json:"name,omitempty"`
	Description  *string `json:"description,omitempty"`
	LeadUserName *string `json:"leadUserName,omitempty"`
	AssigneeType *string `json:"assigneeType,omitempty"`
}

// NewJiraComponentUpdateReq ...
func NewJiraComponentUpdateReq(componentID string) *JiraComponentUpdateReq {
	return &JiraComponentUpdateReq{
		ID: componentID,
	}
}

// SetName ...
func (j *JiraComponentUpdateReq) SetName(name string) *JiraComponentUpdateReq {
	j.Name = &name
	return j
}

// SetDescription ...
func (j *JiraComponentUpdateReq) SetDescription(description string) *JiraComponentUpdateReq {
	j.Description = &description
	return j
}

// SetLeadUserName ...
func (j *JiraComponentUpdateReq) SetLeadUserName(leadUserName string) *JiraComponentUpdateReq {
	j.LeadUserName = &leadUserName
	return j
}

// SetAssigneeType ...
func (j *JiraComponentUpdateReq) SetAssigneeType(assigneeType string) *JiraComponentUpdateReq {
	j.AssigneeType = &assigneeType
	return j
}

// GetComponents get project all components
func (c *Component) GetComponents(projectID string) (*[]JiraComponent, error) {
	var components []JiraComponent
	resp, err := c.ehttp.R().SetBasicAuth(c.config.Username, c.config.Password).SetResult(&components).Get(fmt.Sprintf(APIGetComponents, projectID))
	if err != nil {
		return nil, fmt.Errorf("components get request fail, %w", err)
	}

	var respError Error
	_ = json.Unmarshal(resp.Body(), &respError)
	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("components get fail, %s", respError.LongError())
	}
	return &components, err
}

// CreateComponent create project component
func (c *Component) CreateComponent(component *JiraComponent) (*JiraComponent, error) {
	var respComponent JiraComponent
	resp, err := c.ehttp.R().SetBasicAuth(c.config.Username, c.config.Password).SetBody(component).SetResult(&respComponent).Post(fmt.Sprintf(APICreateComponent))
	if err != nil {
		return nil, fmt.Errorf("create component request fail, %w", err)
	}

	var respError Error
	_ = json.Unmarshal(resp.Body(), &respError)
	if resp.StatusCode() != 201 {
		return nil, fmt.Errorf("create component fail, %s", respError.LongError())
	}
	return &respComponent, err
}

// UpdateComponent update project component
func (c *Component) UpdateComponent(component *JiraComponentUpdateReq) (*JiraComponent, error) {
	var respComponent JiraComponent
	resp, err := c.ehttp.R().SetBasicAuth(c.config.Username, c.config.Password).SetBody(component).SetResult(&respComponent).Put(fmt.Sprintf(APIComponent, component.ID))
	if err != nil {
		return nil, fmt.Errorf("update component request fail, %w", err)
	}

	var respError Error
	_ = json.Unmarshal(resp.Body(), &respError)
	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("update component fail, %s", respError.LongError())
	}
	return &respComponent, err
}

// DeleteComponent delete project component
func (c *Component) DeleteComponent(componentID string) error {
	resp, err := c.ehttp.R().SetBasicAuth(c.config.Username, c.config.Password).Delete(fmt.Sprintf(APIComponent, componentID))
	if err != nil {
		return fmt.Errorf("delete version request fail, %w", err)
	}

	var respError Error
	_ = json.Unmarshal(resp.Body(), &respError)
	if resp.StatusCode() != 204 {
		return fmt.Errorf("delete version fail, %s", respError.LongError())
	}
	return err
}
