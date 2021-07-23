package main

import (
	"fmt"
	"io/ioutil"
	"sync"
	"time"

	"github.com/gotomicro/ego"
	"github.com/gotomicro/ego/core/elog"

	"github.com/gotomicro/ego-component/ejenkins"
)

const (
	JobCreateXmlFile = "_testFiles/job_buildWithDockerfile.xml"
	JobUpdateXmlFile = "_testFiles/job_update.xml"
	TestDockerfile   = "_testFiles/Dockerfile4test"

	jobNameUnderDeepFolder = "job_1"
	folderGrandpa          = "folder-A"
	folderParent           = "folder-B"
	folderChild            = "folder_c"
)

// export EGO_DEBUG=true && go run main.go --config=config.toml
func main() {
	err := ego.New().Invoker(
		invokerJenkins,
	).Run()
	if err != nil {
		elog.Error("startup", elog.Any("err", err))
	}
}

func invokerJenkins() error {
	// 0. get load config from config.toml && build the component
	c := ejenkins.Load("jenkins").Build()

	// 1. get jenkins server basic info:
	resp, err := c.Info()
	if err != nil {
		fmt.Println("Get jenkins info error.", err)
	}
	fmt.Printf("Jenkins info: %+v\n", *resp)

	// 2. create folder:
	_, err = c.CreateFolder(folderGrandpa)
	if err != nil {
		fmt.Printf("create folder (%s) failed.\n", folderGrandpa)
		return err
	}
	time.Sleep(1 * time.Second)
	_, err = c.CreateFolder(folderParent, folderGrandpa)
	if err != nil {
		fmt.Printf("create folder (%s) under folder (%s) failed\n", folderParent, folderGrandpa)
		return err
	}
	time.Sleep(1 * time.Second)
	_, err = c.CreateFolder(folderChild, folderGrandpa, folderParent)
	if err != nil {
		fmt.Printf("create folder (%s) under folder (%s/%s) failed\n", folderChild, folderGrandpa, folderParent)
		return err
	}

	// 3 create job by xml directly
	newXmlBuf, err := ioutil.ReadFile(JobCreateXmlFile)
	if err != nil {
		panic(err)
	}
	_, err = c.CreateJob(string(newXmlBuf), "demo-new-Job")
	if err != nil {
		fmt.Println("Create job directly failed", err)
		return err
	}

	// 4. update job by xml
	updateXmlBuf, err := ioutil.ReadFile(JobUpdateXmlFile)
	if err != nil {
		panic(err)
	}
	_, err = c.UpdateJob("demo-new-Job", string(updateXmlBuf))
	if err != nil {
		fmt.Println("Update job directly failed", err)
		return err
	}

	// 5. create job in deep folder:
	jobUnderFolder, err := c.CreateJobInFolder(string(newXmlBuf), jobNameUnderDeepFolder, folderGrandpa, folderParent, folderChild)
	if err != nil {
		fmt.Println("Create job under folder failed.", err)
		return err
	}
	fmt.Printf("Create job under folder success\n")
	time.Sleep(3 * time.Second)

	// 6. invoke parameterized pipeline(has file parameter) which created in strep 5, directly
	buildParams := ejenkins.BuildParameters{Parameter: []ejenkins.ParameterItem{
		{Name: "AppName", Value: "mock-app-name"},
		{Name: "ImageFullName", Value: "mock-image-name"},
		{Name: "UUID", Value: "a-mock-uuid-string"},
		{Name: "DockerfileInLocal", File: TestDockerfile},
	}}
	_, err = jobUnderFolder.Invoke(nil, &buildParams, "")
	if err != nil {
		fmt.Println("Invoke pipeline error.", err)
		return err
	}

	// 7 concurrent invoke job, and get invoked buildInfo.
	wg := sync.WaitGroup{}
	wg.Add(3)
	for i := 0; i < 3; i++ {
		go func(i int) {
			defer wg.Done()
			jobUnderDeepFolder, err := c.GetJob(jobNameUnderDeepFolder, folderGrandpa, folderParent, folderChild)
			if err != nil {
				fmt.Println("get job under folder error.", err)
				return
			}
			buildParams := ejenkins.BuildParameters{Parameter: []ejenkins.ParameterItem{
				{Name: "AppName", Value: fmt.Sprintf("mock-app-name-%d", i)},
				{Name: "ImageFullName", Value: "mock-image-name"},
				{Name: "UUID", Value: fmt.Sprintf("a-mock-uuid-string-%d", i)},
				{Name: "DockerfileInLocal", File: TestDockerfile},
			}}
			invokedBuild, err := jobUnderDeepFolder.Invoke(nil, &buildParams, "")

			if err != nil {
				fmt.Printf("=> Index: %d. Invoke job error. %v\n", i, err)
				return
			}
			fmt.Printf("=> Index: %d. BuildNum: %v\n", i, invokedBuild.GetBuildNumber())
			// invokedBuild.OtherFunc() to get other buildInfo
			// e.g. you can using invokedBuild.GetConsoleOutputFromIndex() to tailing the running build output.

		}(i)
	}
	wg.Wait()

	// create file credential under folders
	fileCred := ejenkins.FileCredentials{
		ID:          "testFileCredential",
		Scope:       "GLOBAL",
		Description: "SomeDesc",
		Filename:    "testFile.json",
		SecretBytes: "VGhpcyBpcyBhIHRlc3Qu\n",
	}
	if err := c.AddCredential("_", fileCred, folderGrandpa, folderParent); err != nil {
		fmt.Println("Add new fileCredential underFolder failed.", err)
		return err
	}
	fmt.Println("Add new fileCredential underFolder successfully.")

	// create userPass credential directly:
	newUserPassCred := ejenkins.UsernameCredentials{
		ID:          "testUserPass",
		Scope:       "GLOBAL",
		Description: "SomeDesc",
		Username:    "UserNameTest",
		Password:    "pass",
	}
	if err := c.AddCredential("_", newUserPassCred); err != nil {
		fmt.Println("Add new userPasswordCredential failed.", err)
		return err
	}
	fmt.Println("Add new userPasswordCredential successfully.")

	// update userPass credential
	updateUserPassCred := ejenkins.UsernameCredentials{
		ID:          "testUserPass2",
		Scope:       "GLOBAL",
		Description: "UpdatedDesc",
		Username:    "UserNameTest2",
		Password:    "pass2",
	}
	if err := c.UpdateCredential("_", "testUserPass", updateUserPassCred); err != nil {
		fmt.Println("Update userPassCred failed.", err)
		return err
	}
	fmt.Println("Update userPassCred successfully.")

	// list credentials:
	globalDomainCreds, err := c.ListCredential("_")
	if err != nil {
		fmt.Println("List Global credentials error.", err)
		return err
	}
	fmt.Printf("ListCredentials: %+v\n", globalDomainCreds)
	folderCreds, err := c.ListCredential("_", folderGrandpa, folderParent)
	if err != nil {
		fmt.Println("List credentials under folder error.", err)
		return err
	}
	fmt.Printf("ListCredentialsUnderFolder: %+v\n", folderCreds)

	// delete credential
	if err := c.DeleteCredential("_", "testUserPass2"); err != nil {
		fmt.Println("Delete global credential failed.", err)
		return err
	}
	if err := c.DeleteCredential("_", "testFileCredential", folderGrandpa, folderParent); err != nil {
		fmt.Println("Delete credential under folder err.", err)
		return err
	}
	fmt.Println("Delete userPassCred successfully.")

	return nil
}
