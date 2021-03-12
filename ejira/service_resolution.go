package ejira

import (
	"encoding/json"
	"fmt"
)

// Resolution is a resolution of a Jira issue.
// Typical types are "Fixed", "Suspended", "Won't Fix", ...
type Resolution struct {
	Self        string `json:"self"`
	ID          string `json:"id"`
	Description string `json:"description"`
	Name        string `json:"name"`
}

// GetResolutions get project all resolutions
func (c *Component) GetResolutions() (*[]Resolution, error) {
	var resolution []Resolution
	resp, err := c.ehttp.R().SetBasicAuth(c.config.Username, c.config.Password).SetResult(&resolution).Get(fmt.Sprintf(APIGetResolutions))
	if err != nil {
		return nil, fmt.Errorf("priorities get request fail, %w", err)
	}

	var respError Error
	_ = json.Unmarshal(resp.Body(), &respError)
	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("priorities get fail, %s", respError.LongError())
	}
	return &resolution, err
}
