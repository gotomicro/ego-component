package ejira

const (
	// APIGetUserInfo https://docs.atlassian.com/software/jira/docs/api/REST/8.8.0/#api/2/user-getUser
	APIGetUserInfo = "/rest/api/2/user?username=%s"
	// APIFindUsers https://docs.atlassian.com/software/jira/docs/api/REST/8.8.0/#api/2/user-findUsers
	APIFindUsers = "/rest/api/2/user/search"

	// APIGetVersions https://docs.atlassian.com/software/jira/docs/api/REST/8.8.0/#api/2/version-getVersion
	APIGetVersions = "/rest/api/2/project/%s/versions"

	// APICreateVersion https://docs.atlassian.com/software/jira/docs/api/REST/8.8.0/#api/2/version-createVersion
	APICreateVersion = "/rest/api/2/version"

	// APIVersion https://docs.atlassian.com/software/jira/docs/api/REST/8.8.0/#api/2/version
	APIVersion = "/rest/api/2/version/%s"
)
