package ejira

import (
	"encoding/json"
	"fmt"
)

// Version represents a single release version of a project
type Version struct {
	Self            string `json:"self,omitempty"`
	ID              string `json:"id,omitempty"`
	Name            string `json:"name,omitempty"`
	Description     string `json:"description,omitempty"`
	Archived        *bool  `json:"archived,omitempty"`
	Released        *bool  `json:"released,omitempty"`
	ReleaseDate     string `json:"releaseDate,omitempty"`
	UserReleaseDate string `json:"userReleaseDate,omitempty"`
	Project         string `json:"project,omitempty"`
	ProjectID       int    `json:"projectId,omitempty"` // Unlike other IDs, this is returned as a number
	StartDate       string `json:"startDate,omitempty"`
}

// GetVersions get project all versions
func (c *Component) GetVersions(projectID string) (*[]Version, error) {
	var versions []Version
	resp, err := c.ehttp.R().SetBasicAuth(c.config.Username, c.config.Password).SetResult(&versions).Get(fmt.Sprintf(APIGetVersions, projectID))
	if err != nil {
		return nil, fmt.Errorf("versions get request fail, %w", err)
	}

	var respError Error
	_ = json.Unmarshal(resp.Body(), &respError)
	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("versions get fail, %s", respError.LongError())
	}
	return &versions, err
}

// CreateVersion create project version
func (c *Component) CreateVersion(version *Version) (*Version, error) {
	var respVersion Version
	resp, err := c.ehttp.R().SetBasicAuth(c.config.Username, c.config.Password).SetHeader("Content-Type", "application/json").SetBody(version).SetResult(&respVersion).Post(fmt.Sprintf(APICreateVersion))
	if err != nil {
		return nil, fmt.Errorf("create version request fail, %w", err)
	}

	var respError Error
	_ = json.Unmarshal(resp.Body(), &respError)
	if resp.StatusCode() != 201 {
		return nil, fmt.Errorf("create version fail, %s", respError.LongError())
	}
	return &respVersion, err
}

// DeleteVersion delete project version
func (c *Component) DeleteVersion(versionID string) error {
	resp, err := c.ehttp.R().SetBasicAuth(c.config.Username, c.config.Password).Delete(fmt.Sprintf(APIVersion, versionID))
	if err != nil {
		return fmt.Errorf("create version request fail, %w", err)
	}

	var respError Error
	_ = json.Unmarshal(resp.Body(), &respError)
	if resp.StatusCode() != 204 {
		return fmt.Errorf("create version fail, %s", respError.LongError())
	}
	return err
}
