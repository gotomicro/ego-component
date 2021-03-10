package ejira

// Status represents the current status of a Jira issue.
// Typical status are "Open", "In Progress", "Closed", ...
// Status can be user defined in every Jira instance.
type Status struct {
	Self           string         `json:"self" structs:"self"`
	Description    string         `json:"description" structs:"description"`
	IconURL        string         `json:"iconUrl" structs:"iconUrl"`
	Name           string         `json:"name" structs:"name"`
	ID             string         `json:"id" structs:"id"`
	StatusCategory StatusCategory `json:"statusCategory" structs:"statusCategory"`
}

// StatusCategory represents the category a status belongs to.
// Those categories can be user defined in every Jira instance.
type StatusCategory struct {
	Self      string `json:"self" structs:"self"`
	ID        int    `json:"id" structs:"id"`
	Name      string `json:"name" structs:"name"`
	Key       string `json:"key" structs:"key"`
	ColorName string `json:"colorName" structs:"colorName"`
}
