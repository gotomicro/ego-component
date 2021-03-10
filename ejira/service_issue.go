package ejira

// import "github.com/trivago/tgo/tcontainer"

// IssueType is a type of a Jira issue.
// Typical types are "Bug", "Story", ...
type IssueType struct {
	Self        string `json:"self,omitempty" structs:"self,omitempty"`
	ID          string `json:"id,omitempty" structs:"id,omitempty"`
	Description string `json:"description,omitempty" structs:"description,omitempty"`
	IconURL     string `json:"iconUrl,omitempty" structs:"iconUrl,omitempty"`
	Name        string `json:"name,omitempty" structs:"name,omitempty"`
	Subtask     bool   `json:"subtask,omitempty" structs:"subtask,omitempty"`
	AvatarID    int    `json:"avatarId,omitempty" structs:"avatarId,omitempty"`
}

// Issue represents a Jira issue.
type Issue struct {
	Expand      string            `json:"expand,omitempty" structs:"expand,omitempty"`
	ID          string            `json:"id,omitempty" structs:"id,omitempty"`
	Self        string            `json:"self,omitempty" structs:"self,omitempty"`
	Key         string            `json:"key,omitempty" structs:"key,omitempty"`
	Changelog   *Changelog        `json:"changelog,omitempty" structs:"changelog,omitempty"`
	Transitions []Transition      `json:"transitions,omitempty" structs:"transitions,omitempty"`
	Names       map[string]string `json:"names,omitempty" structs:"names,omitempty"`
}

// ChangelogItems is one single changelog item of a history item
type ChangelogItems struct {
	Field      string      `json:"field" structs:"field"`
	FieldType  string      `json:"fieldtype" structs:"fieldtype"`
	From       interface{} `json:"from" structs:"from"`
	FromString string      `json:"fromString" structs:"fromString"`
	To         interface{} `json:"to" structs:"to"`
	ToString   string      `json:"toString" structs:"toString"`
}

// ChangelogHistory is one single changelog history entry
type ChangelogHistory struct {
	ID      string           `json:"id" structs:"id"`
	Author  User             `json:"author" structs:"author"`
	Created string           `json:"created" structs:"created"`
	Items   []ChangelogItems `json:"items" structs:"items"`
}

// Changelog is the change log of an issue
type Changelog struct {
	Histories []ChangelogHistory `json:"histories,omitempty"`
}

// Transition represents an issue transition in Jira
type Transition struct {
	ID   string `json:"id" structs:"id"`
	Name string `json:"name" structs:"name"`
	To   Status `json:"to" structs:"status"`
}
