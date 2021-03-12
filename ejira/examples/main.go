package main

import (
	"fmt"

	"github.com/gotomicro/ego"
	"github.com/gotomicro/ego-component/ejira"
	"github.com/gotomicro/ego/core/elog"
)

// export EGO_DEBUG=true && go run main.go --config=config.toml
func main() {
	err := ego.New().Invoker(
		invokerJira,
	).Run()
	if err != nil {
		elog.Error("startup", elog.Any("err", err))
	}
}

func invokerJira() error {
	comp := ejira.Load("jira").Build()
	userInfo, err := comp.GetUserInfoByUsername("admin")
	fmt.Println(userInfo, err)

	// priority
	priority, err := comp.GetPriorities()
	fmt.Println("GetPriorities", priority, err)

	// resolution
	resolutions, err := comp.GetResolutions()
	fmt.Println("GetResolutions", resolutions, err)

	// user
	userList, err := comp.FindUsers(&ejira.UserSearchOption{})
	fmt.Println("FindUsers", userList, err)

	// project
	projects, err := comp.GetAllProjects(&ejira.ProjectGetQueryOption{
		Expand: "description,lead,url,projectKeys",
	})
	fmt.Println("GetAllProjects", projects, err)

	projectKey := "DEVOPS"
	project, err := comp.GetProject(projectKey, &ejira.ProjectGetQueryOption{
		Expand: "description,lead,url,projectKeys",
	})
	fmt.Println("GetProject", project, err)

	// issue
	issues, err := comp.FindIssues("", &ejira.SearchOptions{
		MaxResults: 10,
	})
	fmt.Println("FindIssues", issues, err)

	// issueLinkTypes
	issueLinkTypes, err := comp.GetIssueLinkTypes()
	fmt.Println("GetIssueLinkTypes", issueLinkTypes, err)

	issue, err := comp.CreateIssue(&ejira.Issue{
		Fields: &ejira.IssueFields{
			Description: "example bug report",
			Project: ejira.Project{
				Key: projectKey,
			},
			Type: ejira.IssueType{
				ID: "10001",
			},
		},
	})
	fmt.Println("CreateIssue", issue, err)

	// version
	versions, err := comp.GetVersions(projectKey)
	fmt.Println("GetVersions", versions, err)

	version, err := comp.CreateVersion(&ejira.Version{
		Name:        "v1.1.0",
		Description: "example version",
		Archived:    ejira.Bool(true),
		Released:    ejira.Bool(true),
		ReleaseDate: "2021-3-20",
		Project:     projectKey,
		ProjectID:   10002,
		StartDate:   "2021-3-12",
	})
	fmt.Println("CreateVersion", version, err)

	updateVersion := ejira.NewVersionUpdateReq(version.ID)
	updateVersion.SetDescription("代码关联修复test")
	_, err = comp.UpdateVersion(updateVersion)
	fmt.Println("UpdateVersion", err)

	err = comp.DeleteVersion(version.ID)
	fmt.Println("DeleteVersion", err)

	// components
	components, err := comp.GetComponents(projectKey)
	fmt.Println("GetComponents", components, err)

	component, err := comp.CreateComponent(&ejira.JiraComponent{
		Name:         "avatar",
		Description:  "avatar component",
		LeadUserName: "xxxx",
		AssigneeType: "PROJECT_LEAD",
		Project:      projectKey,
		ProjectID:    10002,
	})
	fmt.Println("CreateComponent", component, err)

	updateComponent := ejira.NewJiraComponentUpdateReq(component.ID)
	updateComponent.SetDescription("avatar component modify")
	_, err = comp.UpdateComponent(updateComponent)
	fmt.Println("UpdateComponent", err)

	err = comp.DeleteComponent(component.ID)
	fmt.Println("DeleteComponent", err)
	return nil
}
