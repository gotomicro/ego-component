package main

import (
	"fmt"
	"io/ioutil"

	"github.com/gotomicro/ego"
	"github.com/gotomicro/ego/core/elog"

	"github.com/gotomicro/ego-component/ejenkins"
)

const (
	xmlFile4JobCreate = "_testFiles/job_buildWithDockerfile.xml"
	xmlFile4JobUpdate = "_testFiles/job_update.xml"
	testDockerfile    = "_testFiles/Dockerfile4test"

	jobName      = "job_1"
	folderParent = "folder-A"
	folderChild  = "folder_b"
)

var (
	yourAppLogger *elog.Component
	jenkinsC      *ejenkins.Component
)

// export EGO_DEBUG=true && go run main.go --config=config.toml
func main() {
	if err := ego.New().Invoker(
		initComponents,
	).Run(); err != nil {
		panic(err)
	}

	runTest(jenkinsC)
}

func initComponents() error {
	yourAppLogger = elog.Load("yourAppLogger").Build()
	jenkinsC = ejenkins.Load("jenkins").Build(ejenkins.WithLogger(yourAppLogger))
	return nil
}

func runTest(c *ejenkins.Component) {
	// 1. get jenkins server basic info:
	_, err := c.Info()
	if err != nil {
		c.Logger().Error("Failed to get jenkins info", elog.Any("error", err))
	} else {
		c.Logger().Info("Got jenkins info successfully.")
	}

	// 2. create folder:
	_, err = c.CreateFolder(folderParent)
	if err != nil {
		c.Logger().Error("Failed to create folder.", elog.Any("folder", folderParent),
			elog.Any("error", err))
	} else {
		c.Logger().Info("Created folder successfully.", elog.Any("folder", folderParent))
	}
	_, err = c.CreateFolder(folderChild, folderParent)
	if err != nil {
		c.Logger().Error("Failed to create folder.", elog.Any("folder", folderParent+"/"+folderChild),
			elog.Any("error", err))
	} else {
		c.Logger().Info("Created folder successfully.", elog.Any("folder", folderParent+"/"+folderChild))
	}

	// 3. create job in folder:
	newXmlBuf, err := ioutil.ReadFile(xmlFile4JobCreate)
	if err != nil {
		panic(err)
	}
	jobUnderFolder, err := c.CreateJobInFolder(string(newXmlBuf), jobName, folderParent, folderChild)
	if err != nil {
		c.Logger().Panic("Create job under folder failed.", elog.Any("error", err))
	}

	// 4. update job
	if updateXmlBuf, err := ioutil.ReadFile(xmlFile4JobUpdate); err == nil {
		if _, err := c.UpdateJobInFolder(jobName, string(updateXmlBuf), folderParent, folderChild); err != nil {
			c.Logger().Error("update job under folder failed.", elog.Any("error", err))
		}
	} else {
		c.Logger().Error("failed to read xml file for job updating", elog.Any("error", err))
	}

	// 5. invoke parameterized pipeline(has file parameter) which created in above, directly
	buildParams := ejenkins.BuildParameters{Parameter: []ejenkins.ParameterItem{
		{Name: "AppName", Value: "mock-app-name"},
		{Name: "ImageFullName", Value: "mock-image-name"},
		{Name: "UUID", Value: "a-mock-uuid-string"},
		{Name: "DockerfileInLocal", File: testDockerfile},
	}}
	invokedBuild, err := jobUnderFolder.Invoke(nil, &buildParams, "")
	if err != nil {
		c.Logger().Error("failed to invoke pipeline.", elog.Any("error", err))
	}

	// 6 get  invoked build number and output
	if invokedBuild != nil {
		c.Logger().Info("Tailing console output of pipeline", elog.Any("buildNum", invokedBuild.GetBuildNumber()))
		var fromIdx int64 = 0
		for {
			resp, err := invokedBuild.GetConsoleOutputFromIndex(fromIdx)
			if err != nil {
				c.Logger().Error("failed to get the logs of pipeline.", elog.Any("error", err))
				break
			} else {
				fmt.Printf("%s", resp.Content)
				fromIdx = resp.Offset
				if !resp.HasMoreText {
					break
				}
			}
		}
	}
	// finally, clean up
	// delete the job created above
	if _, err := c.DeleteJob(jobName, folderParent, folderChild); err != nil {
		c.Logger().Error("Failed to delete job", elog.Any("job", jobName), elog.Any("error", err))
	} else {
		c.Logger().Info("Deleted job", elog.Any("job", jobName))
	}
	// delete folder(s)
	if _, err := c.DeleteJob(folderParent); err != nil {
		c.Logger().Error("Failed to delete folder.", elog.Any("folder", folderParent), elog.Any("error", err))
	} else {
		c.Logger().Info("Delete folder succeed.", elog.Any("folder", folderParent))
	}

}
