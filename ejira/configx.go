package ejira

const (
	// APIGetUserInfo https://docs.atlassian.com/software/jira/docs/api/REST/8.8.0/#api/2/user-getUser
	APIGetUserInfo = "/rest/api/2/user?username=%s"
	// APIFindUsers https://docs.atlassian.com/software/jira/docs/api/REST/8.8.0/#api/2/user-findUsers
	APIFindUsers = "/rest/api/2/user/search"

	// APIGetAllProjects https://docs.atlassian.com/software/jira/docs/api/REST/8.8.0/#api/2/project-getAllProjects
	APIGetAllProjects = "/rest/api/2/project"
	// APIProject https://docs.atlassian.com/software/jira/docs/api/REST/8.8.0/#api/2/project
	APIProject = "/rest/api/2/project/%s"

	// APIGetVersions https://docs.atlassian.com/software/jira/docs/api/REST/8.8.0/#api/2/version-getVersion
	APIGetVersions = "/rest/api/2/project/%s/versions"
	// APICreateVersion https://docs.atlassian.com/software/jira/docs/api/REST/8.8.0/#api/2/version-createVersion
	APICreateVersion = "/rest/api/2/version"
	// APIVersion https://docs.atlassian.com/software/jira/docs/api/REST/8.8.0/#api/2/version
	APIVersion = "/rest/api/2/version/%s"

	// APIGetComponents https://docs.atlassian.com/software/jira/docs/api/REST/8.8.0/#api/2/component
	APIGetComponents = "/rest/api/2/project/%s/components"
	// APICreateComponent https://docs.atlassian.com/software/jira/docs/api/REST/8.8.0/#api/2/version-createComponent
	APICreateComponent = "/rest/api/2/component"
	// APIComponent https://docs.atlassian.com/software/jira/docs/api/REST/8.8.0/#api/2/component
	APIComponent = "/rest/api/2/component/%s"

	// APIGetPriorities https://docs.atlassian.com/software/jira/docs/api/REST/8.8.0/#api/2/priority
	APIGetPriorities = "/rest/api/2/priority"

	// APIGetResolutions https://docs.atlassian.com/software/jira/docs/api/REST/8.8.0/#api/2/resolution
	APIGetResolutions = "/rest/api/2/resolution"

	// APISearch https://docs.atlassian.com/software/jira/docs/api/REST/8.8.0/#api/2/search-search
	APISearch = "/rest/api/2/search"
	// APICreateIssue https://docs.atlassian.com/software/jira/docs/api/REST/8.8.0/#api/2/issue-createIssue
	APICreateIssue = "/rest/api/2/issue"

	// APIGetIssueLinkTypes https://docs.atlassian.com/software/jira/docs/api/REST/8.8.0/#api/2/issueLinkType
	APIGetIssueLinkTypes = "/rest/api/2/issueLinkType"

	// APIGetIssueTypes https://docs.atlassian.com/software/jira/docs/api/REST/8.8.0/#api/2/issuetype-getIssueAllTypes
	APIGetIssueTypes = "/rest/api/2/issuetype"
)
