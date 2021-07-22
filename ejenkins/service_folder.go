package ejenkins

import (
	"errors"
	"strconv"
	"strings"
)

type Folder struct {
	Raw     *FolderResponse
	Jenkins *Jenkins
	Base    string
}

type FolderResponse struct {
	Actions     []generalObj
	Description string     `json:"description"`
	DisplayName string     `json:"displayName"`
	Name        string     `json:"name"`
	URL         string     `json:"url"`
	Jobs        []InnerJob `json:"jobs"`
	PrimaryView *ViewData  `json:"primaryView"`
	Views       []ViewData `json:"views"`
}

func (f *Folder) parentBase() string {
	return f.Base[:strings.LastIndex(f.Base, "/job")]
}

func (f *Folder) GetName() string {
	return f.Raw.Name
}

func (f *Folder) Create(name string) (*Folder, error) {
	mode := "com.cloudbees.hudson.plugins.folder.Folder"
	data := map[string]string{
		"name":   name,
		"mode":   mode,
		"Submit": "OK",
		"json": getJsonString(map[string]string{
			"name": name,
			"mode": mode,
		}),
	}
	r, err := f.Jenkins.Requester.Post(f.parentBase()+"/createItem", nil, f.Raw, data)
	if err != nil {
		return nil, err
	}
	if r.StatusCode() == 200 {
		f.Poll()
		return f, nil
	}
	return nil, errors.New(strconv.Itoa(r.StatusCode()))
}

func (f *Folder) Poll() (int, error) {
	response, err := f.Jenkins.Requester.GetJSON(f.Base, nil, f.Raw, nil)
	if err != nil {
		return 0, err
	}
	return response.StatusCode(), nil
}
