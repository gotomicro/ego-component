package ejira

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/trivago/tgo/tcontainer"
)

// Issue represents a Jira issue.
type Issue struct {
	Expand         string               `json:"expand,omitempty"`
	ID             string               `json:"id,omitempty"`
	Self           string               `json:"self,omitempty"`
	Key            string               `json:"key,omitempty"`
	Fields         *IssueFields         `json:"fields,omitempty"`
	RenderedFields *IssueRenderedFields `json:"renderedFields,omitempty"`
	Changelog      *Changelog           `json:"changelog,omitempty"`
	Transitions    []Transition         `json:"transitions,omitempty"`
	Names          map[string]string    `json:"names,omitempty"`
}

// ChangelogItems is one single changelog item of a history item
type ChangelogItems struct {
	Field      string      `json:"field"`
	FieldType  string      `json:"fieldtype"`
	From       interface{} `json:"from"`
	FromString string      `json:"fromString"`
	To         interface{} `json:"to"`
	ToString   string      `json:"toString"`
}

// ChangelogHistory is one single changelog history entry
type ChangelogHistory struct {
	ID      string           `json:"id"`
	Author  User             `json:"author"`
	Created string           `json:"created"`
	Items   []ChangelogItems `json:"items"`
}

// Changelog is the change log of an issue
type Changelog struct {
	Histories []ChangelogHistory `json:"histories,omitempty"`
}

// Attachment represents a Jira attachment
type Attachment struct {
	Self      string `json:"self,omitempty"`
	ID        string `json:"id,omitempty"`
	Filename  string `json:"filename,omitempty"`
	Author    *User  `json:"author,omitempty"`
	Created   string `json:"created,omitempty"`
	Size      int    `json:"size,omitempty"`
	MimeType  string `json:"mimeType,omitempty"`
	Content   string `json:"content,omitempty"`
	Thumbnail string `json:"thumbnail,omitempty"`
}

// Epic represents the epic to which an issue is associated
// Not that this struct does not process the returned "color" value
type Epic struct {
	ID      int    `json:"id"`
	Key     string `json:"key"`
	Self    string `json:"self"`
	Name    string `json:"name"`
	Summary string `json:"summary"`
	Done    bool   `json:"done"`
}

// IssueFields represents single fields of a Jira issue.
// Every Jira issue has several fields attached.
type IssueFields struct {
	Expand                        string            `json:"expand,omitempty"`
	Type                          IssueType         `json:"issuetype,omitempty"`
	Project                       Project           `json:"project,omitempty"`
	Resolution                    *Resolution       `json:"resolution,omitempty"`
	Priority                      *Priority         `json:"priority,omitempty"`
	Resolutiondate                Time              `json:"resolutiondate,omitempty"`
	Created                       Time              `json:"created,omitempty"`
	Duedate                       Date              `json:"duedate,omitempty"`
	Watches                       *Watches          `json:"watches,omitempty"`
	Assignee                      *User             `json:"assignee,omitempty"`
	Updated                       Time              `json:"updated,omitempty"`
	Description                   string            `json:"description,omitempty"`
	Summary                       string            `json:"summary,omitempty"`
	Creator                       *User             `json:"Creator,omitempty"`
	Reporter                      *User             `json:"reporter,omitempty"`
	Components                    []*Component      `json:"components,omitempty"`
	Status                        *Status           `json:"status,omitempty"`
	Progress                      *Progress         `json:"progress,omitempty"`
	AggregateProgress             *Progress         `json:"aggregateprogress,omitempty"`
	TimeTracking                  *TimeTracking     `json:"timetracking,omitempty"`
	TimeSpent                     int               `json:"timespent,omitempty"`
	TimeEstimate                  int               `json:"timeestimate,omitempty"`
	TimeOriginalEstimate          int               `json:"timeoriginalestimate,omitempty"`
	Worklog                       *Worklog          `json:"worklog,omitempty"`
	IssueLinks                    []*IssueLink      `json:"issuelinks,omitempty"`
	Comments                      *Comments         `json:"comment,omitempty"`
	FixVersions                   []*FixVersion     `json:"fixVersions,omitempty"`
	AffectsVersions               []*AffectsVersion `json:"versions,omitempty"`
	Labels                        []string          `json:"labels,omitempty"`
	Subtasks                      []*Subtasks       `json:"subtasks,omitempty"`
	Attachments                   []*Attachment     `json:"attachment,omitempty"`
	Epic                          *Epic             `json:"epic,omitempty"`
	Parent                        *Parent           `json:"parent,omitempty"`
	AggregateTimeOriginalEstimate int               `json:"aggregatetimeoriginalestimate,omitempty"`
	AggregateTimeSpent            int               `json:"aggregatetimespent,omitempty"`
	AggregateTimeEstimate         int               `json:"aggregatetimeestimate,omitempty"`
	Unknowns                      tcontainer.MarshalMap
}

// IssueRenderedFields represents rendered fields of a Jira issue.
// Not all IssueFields are rendered.
type IssueRenderedFields struct {
	Resolutiondate string    `json:"resolutiondate,omitempty"`
	Created        string    `json:"created,omitempty"`
	Duedate        string    `json:"duedate,omitempty"`
	Updated        string    `json:"updated,omitempty"`
	Comments       *Comments `json:"comment,omitempty"`
	Description    string    `json:"description,omitempty"`
}

// IssueType is a type of a Jira issue.
// Typical types are "Bug", "Story", ...
type IssueType struct {
	Self        string `json:"self,omitempty"`
	ID          string `json:"id,omitempty"`
	Description string `json:"description,omitempty"`
	IconURL     string `json:"iconUrl,omitempty"`
	Name        string `json:"name,omitempty"`
	Subtask     bool   `json:"subtask,omitempty"`
	AvatarID    int    `json:"avatarId,omitempty"`
}

// Transition represents an issue transition in Jira
type Transition struct {
	ID   string `json:"id" structs:"id"`
	Name string `json:"name" structs:"name"`
	To   Status `json:"to" structs:"status"`
}

// Watches represents a type of how many and which user are "observing" a Jira issue to track the status / updates.
type Watches struct {
	Self       string     `json:"self,omitempty"`
	WatchCount int        `json:"watchCount,omitempty"`
	IsWatching bool       `json:"isWatching,omitempty"`
	Watchers   []*Watcher `json:"watchers,omitempty"`
}

// Watcher represents a simplified user that "observes" the issue
type Watcher struct {
	Self        string `json:"self,omitempty"`
	Name        string `json:"name,omitempty"`
	AccountID   string `json:"accountId,omitempty"`
	DisplayName string `json:"displayName,omitempty"`
	Active      bool   `json:"active,omitempty"`
}

// Progress represents the progress of a Jira issue.
type Progress struct {
	Progress int `json:"progress"`
	Total    int `json:"total"`
	Percent  int `json:"percent"`
}

// Parent represents the parent of a Jira issue, to be used with subtask issue types.
type Parent struct {
	ID  string `json:"id,omitempty"`
	Key string `json:"key,omitempty"`
}

// TimeTracking represents the timetracking fields of a Jira issue.
type TimeTracking struct {
	OriginalEstimate         string `json:"originalEstimate,omitempty"`
	RemainingEstimate        string `json:"remainingEstimate,omitempty"`
	TimeSpent                string `json:"timeSpent,omitempty"`
	OriginalEstimateSeconds  int    `json:"originalEstimateSeconds,omitempty"`
	RemainingEstimateSeconds int    `json:"remainingEstimateSeconds,omitempty"`
	TimeSpentSeconds         int    `json:"timeSpentSeconds,omitempty"`
}

// Time represents the Time definition of Jira as a time.Time of go
type Time time.Time

// Equal ...
func (t Time) Equal(u Time) bool {
	return time.Time(t).Equal(time.Time(u))
}

// UnmarshalJSON will transform the Jira time into a time.Time
// during the transformation of the Jira JSON response
func (t *Time) UnmarshalJSON(b []byte) error {
	// Ignore null, like in the main JSON package.
	if string(b) == "null" {
		return nil
	}
	ti, err := time.Parse("\"2006-01-02T15:04:05.999-0700\"", string(b))
	if err != nil {
		return err
	}
	*t = Time(ti)
	return nil
}

// MarshalJSON will transform the time.Time into a Jira time
// during the creation of a Jira request
func (t Time) MarshalJSON() ([]byte, error) {
	return []byte(time.Time(t).Format("\"2006-01-02T15:04:05.000-0700\"")), nil
}

// Date represents the Date definition of Jira as a time.Time of go
type Date time.Time

// UnmarshalJSON will transform the Jira date into a time.Time
// during the transformation of the Jira JSON response
func (t *Date) UnmarshalJSON(b []byte) error {
	// Ignore null, like in the main JSON package.
	if string(b) == "null" {
		return nil
	}
	ti, err := time.Parse("\"2006-01-02\"", string(b))
	if err != nil {
		return err
	}
	*t = Date(ti)
	return nil
}

// MarshalJSON will transform the Date object into a short
// date string as Jira expects during the creation of a
// Jira request
func (t Date) MarshalJSON() ([]byte, error) {
	time := time.Time(t)
	return []byte(time.Format("\"2006-01-02\"")), nil
}

// Worklog represents the work log of a Jira issue.
// One Worklog contains zero or n WorklogRecords
// Jira Wiki: https://confluence.atlassian.com/jira/logging-work-on-an-issue-185729605.html
type Worklog struct {
	StartAt    int             `json:"startAt"`
	MaxResults int             `json:"maxResults"`
	Total      int             `json:"total"`
	Worklogs   []WorklogRecord `json:"worklogs"`
}

// WorklogRecord represents one entry of a Worklog
type WorklogRecord struct {
	Self             string           `json:"self,omitempty"`
	Author           *User            `json:"author,omitempty"`
	UpdateAuthor     *User            `json:"updateAuthor,omitempty"`
	Comment          string           `json:"comment,omitempty"`
	Created          *Time            `json:"created,omitempty"`
	Updated          *Time            `json:"updated,omitempty"`
	Started          *Time            `json:"started,omitempty"`
	TimeSpent        string           `json:"timeSpent,omitempty"`
	TimeSpentSeconds int              `json:"timeSpentSeconds,omitempty"`
	ID               string           `json:"id,omitempty"`
	IssueID          string           `json:"issueId,omitempty"`
	Properties       []EntityProperty `json:"properties,omitempty"`
}

// EntityProperty represents one key-value entity
type EntityProperty struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
}

// Subtasks represents all issues of a parent issue.
type Subtasks struct {
	ID     string      `json:"id"`
	Key    string      `json:"key"`
	Self   string      `json:"self"`
	Fields IssueFields `json:"fields"`
}

// IssueLink represents a link between two issues in Jira.
type IssueLink struct {
	ID           string        `json:"id,omitempty"`
	Self         string        `json:"self,omitempty"`
	Type         IssueLinkType `json:"type"`
	OutwardIssue *Issue        `json:"outwardIssue"`
	InwardIssue  *Issue        `json:"inwardIssue"`
	Comment      *Comment      `json:"comment,omitempty"`
}

// Comments represents a list of Comment.
type Comments struct {
	Comments []*Comment `json:"comments,omitempty"`
}

// Comment represents a comment by a person to an issue in Jira.
type Comment struct {
	ID           string            `json:"id,omitempty"`
	Self         string            `json:"self,omitempty"`
	Name         string            `json:"name,omitempty"`
	Author       User              `json:"author,omitempty"`
	Body         string            `json:"body,omitempty"`
	UpdateAuthor User              `json:"updateAuthor,omitempty"`
	Updated      string            `json:"updated,omitempty"`
	Created      string            `json:"created,omitempty"`
	Visibility   CommentVisibility `json:"visibility,omitempty"`
}

// FixVersion represents a software release in which an issue is fixed.
type FixVersion struct {
	Self            string `json:"self,omitempty"`
	ID              string `json:"id,omitempty"`
	Name            string `json:"name,omitempty"`
	Description     string `json:"description,omitempty"`
	Archived        *bool  `json:"archived,omitempty"`
	Released        *bool  `json:"released,omitempty"`
	ReleaseDate     string `json:"releaseDate,omitempty"`
	UserReleaseDate string `json:"userReleaseDate,omitempty"`
	ProjectID       int    `json:"projectId,omitempty"` // Unlike other IDs, this is returned as a number
	StartDate       string `json:"startDate,omitempty"`
}

// AffectsVersion represents a software release which is affected by an issue.
type AffectsVersion Version

// CommentVisibility represents he visibility of a comment.
// E.g. Type could be "role" and Value "Administrators"
type CommentVisibility struct {
	Type  string `json:"type,omitempty"`
	Value string `json:"value,omitempty"`
}

// SearchOptions ...
type SearchOptions struct {
	// StartAt: The starting index of the returned projects. Base index: 0.
	StartAt int `url:"startAt,omitempty"`
	// MaxResults: The maximum number of projects to return per page. Default: 50.
	MaxResults int `url:"maxResults,omitempty"`
	// Expand: Expand specific sections in the returned issues
	Expand string `url:"expand,omitempty"`
	Fields []string
	// ValidateQuery: The validateQuery param offers control over whether to validate and how strictly to treat the validation. Default: strict.
	ValidateQuery string `url:"validateQuery,omitempty"`
}

// searchResult is only a small wrapper around the Search (with JQL) method
// to be able to parse the results
type SearchResult struct {
	Issues     []Issue `json:"issues"`
	StartAt    int     `json:"startAt"`
	MaxResults int     `json:"maxResults"`
	Total      int     `json:"total"`
}

// FindIssues 查找issues
// Jira API docs: https://docs.atlassian.com/software/jira/docs/api/REST/8.8.0/#api/2/search
func (c *Component) FindIssues(jql string, options *SearchOptions) (*SearchResult, error) {
	uv := url.Values{}
	if jql != "" {
		uv.Add("jql", jql)
	}
	if options != nil {
		if options.StartAt != 0 {
			uv.Add("startAt", strconv.Itoa(options.StartAt))
		}
		if options.MaxResults != 0 {
			uv.Add("maxResults", strconv.Itoa(options.MaxResults))
		}
		if options.Expand != "" {
			uv.Add("expand", options.Expand)
		}
		if strings.Join(options.Fields, ",") != "" {
			uv.Add("fields", strings.Join(options.Fields, ","))
		}
		if options.ValidateQuery != "" {
			uv.Add("validateQuery", options.ValidateQuery)
		}
	}

	var result SearchResult
	resp, err := c.ehttp.R().SetBasicAuth(c.config.Username, c.config.Password).SetQueryParamsFromValues(uv).SetResult(&result).Get(fmt.Sprintf(APISearch))
	if err != nil {
		return nil, fmt.Errorf("issues get request fail, %w", err)
	}

	var respError Error
	_ = json.Unmarshal(resp.Body(), &respError)
	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("issues get fail, %s", respError.LongError())
	}

	return &result, err
}

// CreateIssue create issue
func (c *Component) CreateIssue(issue *Issue) (*Issue, error) {
	var respIssue Issue
	resp, err := c.ehttp.R().SetBasicAuth(c.config.Username, c.config.Password).SetBody(issue).SetResult(&respIssue).Post(fmt.Sprintf(APICreateIssue))
	if err != nil {
		return nil, fmt.Errorf("create component request fail, %w", err)
	}

	var respError Error
	_ = json.Unmarshal(resp.Body(), &respError)
	if resp.StatusCode() != 201 {
		return nil, fmt.Errorf("create component fail, %s", respError.LongError())
	}
	return &respIssue, err
}

type IssueTypes []*IssueType

// GetAllIssueTypes 获取所有issue类型
// Jira API docs: https://docs.atlassian.com/software/jira/docs/api/REST/8.8.0/#api/2/issuetype-getIssueAllTypes
func (c *Component) GetAllIssueTypes() (*IssueTypes, error) {
	var result IssueTypes
	resp, err := c.ehttp.R().SetBasicAuth(c.config.Username, c.config.Password).SetResult(&result).Get(fmt.Sprintf(APIGetIssueTypes))
	if err != nil {
		return nil, fmt.Errorf("issueTypes get request fail, %w", err)
	}
	var respError Error
	err = json.Unmarshal(resp.Body(), &respError)
	if err != nil {
		return nil, fmt.Errorf("issueTypes unmarshal error, %w", err)
	}
	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("issueTypes get fail, %s", respError.LongError())
	}

	return &result, err
}
