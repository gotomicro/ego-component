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
	ProjectID       int    `json:"projectId,omitempty"`
	StartDate       string `json:"startDate,omitempty"`
}

// VersionUpdateReq version update request
type VersionUpdateReq struct {
	ID          string  `json:"id,omitempty"`
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	Archived    *bool   `json:"archived,omitempty"`
	Released    *bool   `json:"released,omitempty"`
	ReleaseDate *string `json:"releaseDate,omitempty"`
	ProjectID   *int    `json:"projectId,omitempty"`
	StartDate   *string `json:"startDate,omitempty"`
}

// NewVersionUpdateReq ...
func NewVersionUpdateReq(versionID string) *VersionUpdateReq {
	return &VersionUpdateReq{
		ID: versionID,
	}
}

// SetName ...
func (v *VersionUpdateReq) SetName(name string) *VersionUpdateReq {
	v.Name = &name
	return v
}

// SetDescription ...
func (v *VersionUpdateReq) SetDescription(description string) *VersionUpdateReq {
	v.Description = &description
	return v
}

// SetArchived ...
func (v *VersionUpdateReq) SetArchived(archived bool) *VersionUpdateReq {
	v.Archived = &archived
	return v
}

// SetReleased ...
func (v *VersionUpdateReq) SetReleased(released bool) *VersionUpdateReq {
	v.Released = &released
	return v
}

// SetProjectID ...
func (v *VersionUpdateReq) SetProjectID(projectID int) *VersionUpdateReq {
	v.ProjectID = &projectID
	return v
}

// SetStartDate ...
func (v *VersionUpdateReq) SetStartDate(startDate string) *VersionUpdateReq {
	v.StartDate = &startDate
	return v
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
	resp, err := c.ehttp.R().SetBasicAuth(c.config.Username, c.config.Password).SetBody(version).SetResult(&respVersion).Post(fmt.Sprintf(APICreateVersion))
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

// UpdateVersion update project version
func (c *Component) UpdateVersion(version *VersionUpdateReq) (*Version, error) {
	var respVersion Version
	resp, err := c.ehttp.R().SetBasicAuth(c.config.Username, c.config.Password).SetBody(version).SetResult(&respVersion).Put(fmt.Sprintf(APIVersion, version.ID))
	if err != nil {
		return nil, fmt.Errorf("update version request fail, %w", err)
	}

	var respError Error
	_ = json.Unmarshal(resp.Body(), &respError)
	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("update version fail, %s", respError.LongError())
	}
	return &respVersion, err
}

// DeleteVersion delete project version
func (c *Component) DeleteVersion(versionID string) error {
	resp, err := c.ehttp.R().SetBasicAuth(c.config.Username, c.config.Password).Delete(fmt.Sprintf(APIVersion, versionID))
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
