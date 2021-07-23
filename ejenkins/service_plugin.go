package ejenkins

import (
	"strconv"
)

type Plugins struct {
	Jenkins *Jenkins
	Raw     *PluginResponse
	Base    string
	Depth   int
}

type PluginResponse struct {
	Plugins []Plugin `json:"plugins"`
}

type Plugin struct {
	Active        bool        `json:"active"`
	BackupVersion interface{} `json:"backupVersion"`
	Bundled       bool        `json:"bundled"`
	Deleted       bool        `json:"deleted"`
	Dependencies  []struct {
		Optional  string `json:"optional"`
		ShortName string `json:"shortname"`
		Version   string `json:"version"`
	} `json:"dependencies"`
	Downgradable        bool   `json:"downgradable"`
	Enabled             bool   `json:"enabled"`
	HasUpdate           bool   `json:"hasUpdate"`
	LongName            string `json:"longName"`
	Pinned              bool   `json:"pinned"`
	ShortName           string `json:"shortName"`
	SupportsDynamicLoad string `json:"supportsDynamicLoad"`
	URL                 string `json:"url"`
	Version             string `json:"version"`
}

func (p *Plugins) Count() int {
	return len(p.Raw.Plugins)
}

func (p *Plugins) Contains(name string) *Plugin {
	for _, p := range p.Raw.Plugins {
		if p.LongName == name || p.ShortName == name {
			return &p
		}
	}
	return nil
}

func (p *Plugins) Poll() (int, error) {
	qr := map[string]string{
		"depth": strconv.Itoa(p.Depth),
	}
	response, err := p.Jenkins.Requester.GetJSON(p.Base, nil, p.Raw, qr)
	if err != nil {
		return 0, err
	}
	return response.StatusCode(), nil
}
