package ejira

import (
	"encoding/json"
	"fmt"
)

// IssueLinkTypes represents
type IssueLinkTypes struct {
	IssueLinkTypes []IssueLinkType `json:"issueLinkTypes,omitempty"`
}

// IssueLinkType represents a type of a link between to issues in Jira.
// Typical issue link types are "Related to", "Duplicate", "Is blocked by", etc.
type IssueLinkType struct {
	ID      string `json:"id,omitempty"`
	Self    string `json:"self,omitempty"`
	Name    string `json:"name"`
	Inward  string `json:"inward"`
	Outward string `json:"outward"`
}

// GetIssueLinkTypes get project all issue link types
func (c *Component) GetIssueLinkTypes() (*[]IssueLinkType, error) {
	var issueLinkTypes IssueLinkTypes
	resp, err := c.ehttp.R().SetBasicAuth(c.config.Username, c.config.Password).SetResult(&issueLinkTypes).Get(fmt.Sprintf(APIGetIssueLinkTypes))
	if err != nil {
		return nil, fmt.Errorf("priorities get request fail, %w", err)
	}

	var respError Error
	_ = json.Unmarshal(resp.Body(), &respError)
	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("priorities get fail, %s", respError.LongError())
	}
	return &issueLinkTypes.IssueLinkTypes, err
}
