package ejira

const (
	// APIGetUserInfo https://docs.atlassian.com/software/jira/docs/api/REST/8.8.0/#api/2/user-getUser
	APIGetUserInfo = "/rest/api/2/user?username=%s"

	// APIFindUsers https://docs.atlassian.com/software/jira/docs/api/REST/8.8.0/#api/2/user-findUsers
	APIFindUsers = "/rest/api/2/user/search"
)
