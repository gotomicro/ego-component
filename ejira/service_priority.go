package ejira

import (
	"encoding/json"
	"fmt"
)

// Priority is a priority of a Jira issue.
// Typical types are "Normal", "Moderate", "Urgent", ...
type Priority struct {
	Self        string `json:"self,omitempty"`
	IconURL     string `json:"iconUrl,omitempty"`
	Name        string `json:"name,omitempty"`
	ID          string `json:"id,omitempty"`
	StatusColor string `json:"statusColor,omitempty"`
	Description string `json:"description,omitempty"`
}

// GetPriorities get project all priorities
func (c *Component) GetPriorities() (*[]Priority, error) {
	var priorities []Priority
	resp, err := c.ehttp.R().SetBasicAuth(c.config.Username, c.config.Password).SetResult(&priorities).Get(fmt.Sprintf(APIGetPriorities))
	if err != nil {
		return nil, fmt.Errorf("priorities get request fail, %w", err)
	}

	var respError Error
	_ = json.Unmarshal(resp.Body(), &respError)
	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("priorities get fail, %s", respError.LongError())
	}
	return &priorities, err
}
