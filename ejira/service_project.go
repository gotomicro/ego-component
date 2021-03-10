package ejira

import (
	"encoding/json"
	"fmt"

	"github.com/google/go-querystring/query"
)

// ProjectList represent a list of Projects
type ProjectList []struct {
	Expand          string          `json:"expand" structs:"expand"`
	Self            string          `json:"self" structs:"self"`
	ID              string          `json:"id" structs:"id"`
	Key             string          `json:"key" structs:"key"`
	Name            string          `json:"name" structs:"name"`
	AvatarUrls      AvatarUrls      `json:"avatarUrls" structs:"avatarUrls"`
	ProjectTypeKey  string          `json:"projectTypeKey" structs:"projectTypeKey"`
	ProjectCategory ProjectCategory `json:"projectCategory,omitempty" structs:"projectsCategory,omitempty"`
	IssueTypes      []IssueType     `json:"issueTypes,omitempty" structs:"issueTypes,omitempty"`
}

// Project jira project info
type Project struct {
	Expand          string             `json:"expand,omitempty"`
	Self            string             `json:"self,omitempty"`
	ID              string             `json:"id,omitempty"`
	Key             string             `json:"key,omitempty"`
	Description     string             `json:"description,omitempty"`
	Lead            User               `json:"lead,omitempty"`
	Components      []ProjectComponent `json:"components,omitempty"`
	IssueTypes      []IssueType        `json:"issueTypes,omitempty"`
	URL             string             `json:"url,omitempty"`
	Email           string             `json:"email,omitempty"`
	AssigneeType    string             `json:"assigneeType,omitempty"`
	Versions        []Version          `json:"versions,omitempty"`
	Name            string             `json:"name,omitempty"`
	Roles           map[string]string  `json:"roles,omitempty"`
	AvatarUrls      AvatarUrls         `json:"avatarUrls,omitempty"`
	ProjectCategory ProjectCategory    `json:"projectCategory,omitempty"`
}

// ProjectComponent represents a single component of a project
type ProjectComponent struct {
	Self                string `json:"self" structs:"self,omitempty"`
	ID                  string `json:"id" structs:"id,omitempty"`
	Name                string `json:"name" structs:"name,omitempty"`
	Description         string `json:"description" structs:"description,omitempty"`
	Lead                User   `json:"lead,omitempty" structs:"lead,omitempty"`
	AssigneeType        string `json:"assigneeType" structs:"assigneeType,omitempty"`
	Assignee            User   `json:"assignee" structs:"assignee,omitempty"`
	RealAssigneeType    string `json:"realAssigneeType" structs:"realAssigneeType,omitempty"`
	RealAssignee        User   `json:"realAssignee" structs:"realAssignee,omitempty"`
	IsAssigneeTypeValid bool   `json:"isAssigneeTypeValid" structs:"isAssigneeTypeValid,omitempty"`
	Project             string `json:"project" structs:"project,omitempty"`
	ProjectID           int    `json:"projectId" structs:"projectId,omitempty"`
}

// ProjectCategory represents a single project category
type ProjectCategory struct {
	Self        string `json:"self" structs:"self,omitempty"`
	ID          string `json:"id" structs:"id,omitempty"`
	Name        string `json:"name" structs:"name,omitempty"`
	Description string `json:"description" structs:"description,omitempty"`
}

// ProjectGetQueryOption ...
type ProjectGetQueryOption struct {
	Expand          string `url:"expand,omitempty"` // 可选参数：description,lead,url,projectKeys
	Recent          int    `url:"recent,omitempty"`
	IncludeArchived string `url:"includeArchived,omitempty"`
	BrowseArchive   string `url:"browseArchive,omitempty"`
}

// GetAllProjects get all projects
func (c *Component) GetAllProjects(options *ProjectGetQueryOption) (*ProjectList, error) {
	request := c.ehttp.R().SetBasicAuth(c.config.Username, c.config.Password)
	if options != nil {
		paramValues, err := query.Values(options)
		if err != nil {
			return nil, err
		}

		request.SetQueryParamsFromValues(paramValues)
	}

	var projects ProjectList
	resp, err := request.SetResult(&projects).Get(fmt.Sprintf(APIGetAllProjects))
	if err != nil {
		return nil, fmt.Errorf("components get request fail, %w", err)
	}

	var respError Error
	_ = json.Unmarshal(resp.Body(), &respError)
	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("components get fail, %s", respError.LongError())
	}
	return &projects, err
}

// GetProject get single project info
func (c *Component) GetProject(projectKey string, options *ProjectGetQueryOption) (*Project, error) {
	request := c.ehttp.R().SetBasicAuth(c.config.Username, c.config.Password)
	if options != nil {
		paramValues, err := query.Values(options)
		if err != nil {
			return nil, err
		}

		request.SetQueryParamsFromValues(paramValues)
	}

	var project Project
	resp, err := request.SetResult(&project).Get(fmt.Sprintf(APIProject, projectKey))
	if err != nil {
		return nil, fmt.Errorf("components get request fail, %w", err)
	}

	var respError Error
	_ = json.Unmarshal(resp.Body(), &respError)
	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("components get fail, %s", respError.LongError())
	}
	return &project, err
}
