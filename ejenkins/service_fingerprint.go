package ejenkins

import (
	"errors"
	"fmt"
)

type FingerPrint struct {
	Jenkins *Jenkins
	Base    string
	Id      string
	Raw     *FingerPrintResponse
}

type FingerPrintResponse struct {
	FileName string `json:"fileName"`
	Hash     string `json:"hash"`
	Original struct {
		Name   string
		Number int64
	} `json:"original"`
	Timestamp int64 `json:"timestamp"`
	Usage     []struct {
		Name   string `json:"name"`
		Ranges struct {
			Ranges []struct {
				End   int64 `json:"end"`
				Start int64 `json:"start"`
			} `json:"ranges"`
		} `json:"ranges"`
	} `json:"usage"`
}

func (f FingerPrint) Valid() (bool, error) {
	status, err := f.Poll()

	if err != nil {
		return false, err
	}

	if status != 200 || f.Raw.Hash != f.Id {
		return false, fmt.Errorf("jenkins says %s is Invalid or the Status is unknown", f.Id)
	}
	return true, nil
}

func (f FingerPrint) ValidateForBuild(filename string, build *Build) (bool, error) {
	valid, err := f.Valid()
	if err != nil {
		return false, err
	}

	if valid {
		return true, nil
	}

	if f.Raw.FileName != filename {
		return false, errors.New("filename does not Match")
	}
	if build != nil && f.Raw.Original.Name == build.Job.GetName() &&
		f.Raw.Original.Number == build.GetBuildNumber() {
		return true, nil
	}
	return false, nil
}

func (f FingerPrint) GetInfo() (*FingerPrintResponse, error) {
	_, err := f.Poll()
	if err != nil {
		return nil, err
	}
	return f.Raw, nil
}

func (f FingerPrint) Poll() (int, error) {
	response, err := f.Jenkins.Requester.GetJSON(f.Base+f.Id, nil, f.Raw, nil)
	if err != nil {
		return 0, err
	}
	return response.StatusCode(), nil
}
