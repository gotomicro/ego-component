package ejenkins

import (
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/gotomicro/ego/client/ehttp"
	"github.com/gotomicro/ego/core/elog"
	"github.com/pkg/errors"
	"net/url"
	"strings"
)


type APIRequest struct {
	RequestInst *resty.Request  // request instance, init in Requester.NewAPIRequest func
	Method   	string
	Endpoint 	string
	Payload  	map[string]string
	Suffix   	string
}

func (ar *APIRequest) SetHeader(key string, value string) *APIRequest {
	//ar.Headers.Set(key, value)
	ar.RequestInst.SetHeader(key, value)
	return ar
}

func (r *Requester) NewAPIRequest(method string, endpoint string, payload map[string]string) *APIRequest {
	ar := &APIRequest{
		RequestInst:r.Client.R(),
		Method:     method,
		Endpoint:   endpoint,
		Payload:    payload,
		Suffix:     "",
	}
	if r.BasicAuth != nil {
		ar.RequestInst.SetBasicAuth(r.BasicAuth.Username, r.BasicAuth.Password)
	}
	return ar
}

type Requester struct {
	Base      string
	BasicAuth *BasicAuth
	Client    *ehttp.Component
	SslVerify bool
	logger    *elog.Component
}

func (r *Requester) SetCrumb(ar *APIRequest) error {
	crumbData := map[string]string{}
	response, _ := r.GetJSON("/crumbIssuer/api/json", nil, &crumbData, nil)

	if response.StatusCode() == 200 && crumbData["crumbRequestField"] != "" {
		ar.SetHeader(crumbData["crumbRequestField"], crumbData["crumb"])
		ar.SetHeader("Cookie", response.Header().Get("set-cookie"))
	}
	return nil
}

func (r *Requester) PostJSON(endpoint string, payload map[string]string, responseStruct interface{}, querystring map[string]string) (*resty.Response, error) {
	ar := r.NewAPIRequest("POST", endpoint, payload)
	if err := r.SetCrumb(ar); err != nil {
		return nil, err
	}
	ar.SetHeader("Content-Type", "application/x-www-form-urlencoded")
	ar.Suffix = "api/json"
	return r.Do(ar, &responseStruct, querystring)
}

func (r *Requester) Post(endpoint string, payload map[string]string, responseStruct interface{}, querystring map[string]string) (*resty.Response, error) {
	ar := r.NewAPIRequest("POST", endpoint, payload)
	if err := r.SetCrumb(ar); err != nil {
		return nil, err
	}
	ar.SetHeader("Content-Type", "application/x-www-form-urlencoded")
	ar.Suffix = ""
	return r.Do(ar, &responseStruct, querystring)
}

// Note, files in BuildParameters
func (r *Requester) PostFiles(endpoint string, payload map[string]string, responseStruct interface{},
querystring map[string]string, params *BuildParameters) (*resty.Response, error) {
	ar := r.NewAPIRequest("POST", endpoint, payload)
	if err := r.SetCrumb(ar); err != nil {
		return nil, err
	}
	return r.Do(ar, &responseStruct, querystring, params)
}

func (r *Requester) PostXML(endpoint string, payload map[string]string, xml string, responseStruct interface{}, querystring map[string]string) (*resty.Response, error) {
	ar := r.NewAPIRequest("POST", endpoint, payload)
	if err := r.SetCrumb(ar); err != nil {
		return nil, err
	}
	ar.SetHeader("Content-Type", "application/xml")
	ar.Suffix = ""
	ar.RequestInst.SetBody(xml)
	return r.Do(ar, &responseStruct, querystring)
}

func (r *Requester) GetJSON(endpoint string, payload map[string]string, responseStruct interface{}, query map[string]string) (*resty.Response, error) {
	ar := r.NewAPIRequest("GET", endpoint, payload)
	ar.SetHeader("Content-Type", "application/json")
	ar.Suffix = "api/json"
	return r.Do(ar, responseStruct, query)
}

func (r *Requester) GetXML(endpoint string, payload map[string]string, responseStruct interface{}, query map[string]string) (*resty.Response, error) {
	ar := r.NewAPIRequest("GET", endpoint, payload)
	ar.SetHeader("Content-Type", "application/xml")
	ar.Suffix = ""
	return r.Do(ar, responseStruct, query)
}

func (r *Requester) Get(endpoint string, payload map[string]string, responseStruct interface{}, querystring map[string]string) (*resty.Response, error) {
	ar := r.NewAPIRequest("GET", endpoint, payload)
	ar.Suffix = ""
	return r.Do(ar, responseStruct, querystring)
}


type BuildParameters struct {
	Parameter []ParameterItem `json:"parameter"`
}
type ParameterItem struct {
	Name  string 		`json:"name"`
	Value string 		`json:"value,omitempty"`
	File  string 		`json:"file,omitempty"`  // file with path
}

func (r *Requester) Do(ar *APIRequest, responseStruct interface{}, options ...interface{}) (resp *resty.Response, err error) {
	if !strings.HasSuffix(ar.Endpoint, "/") && ar.Method != "POST" {
		ar.Endpoint += "/"
	}

	URL, err := url.Parse(r.Base + ar.Endpoint + ar.Suffix)
	if err != nil {
		return nil, err
	}
	files := map[string]string{}

	for _, o := range options {
		switch v := o.(type) {
		case map[string]string:
			ar.RequestInst.SetQueryParams(v)
		case *BuildParameters:
			if v == nil {
				continue
			}
			if ar.Payload == nil {
				ar.Payload = map[string]string{}
			}
			//ar.Payload["json"] = getJsonString(*v)
			for _, buildParam := range v.Parameter {
				if buildParam.File != "" {
					files[buildParam.Name] = buildParam.File
					continue
				}
				ar.Payload[buildParam.Name] = buildParam.Value
			}
		}
	}
	if len(files) > 0 {
		ar.RequestInst.SetFiles(files)
	}
	if len(ar.Payload) > 0 {
		ar.RequestInst.SetFormData(ar.Payload)
	}
	ar.RequestInst.SetResult(responseStruct)
	urlStr := URL.String()
	switch strings.ToUpper(ar.Method) {
	case "GET":
		resp, err = ar.RequestInst.Get(urlStr)
	case "POST":
		resp, err = ar.RequestInst.Post(urlStr)
	case "PUT":
		resp, err = ar.RequestInst.Put(urlStr)
	case "PATCH":
		resp, err = ar.RequestInst.Patch(urlStr)
	case "DELETE":
		resp, err = ar.RequestInst.Delete(urlStr)
	case "HEAD":
		resp, err = ar.RequestInst.Head(urlStr)
	case "OPTIONS":
		resp, err = ar.RequestInst.Options(urlStr)
	default:
		return nil, errors.New(ar.Method + " method is not supported")
	}

	if err != nil {
		return nil, err
	} else {
		errorText := resp.Header().Get("X-Error")
		if errorText != "" {
			return nil, errors.New(errorText)
		}
		return resp, nil
	}

}

func (r *Requester) ReadRawResponse(response *resty.Response, responseStruct interface{}) (*resty.Response, error) {
	if str, ok := responseStruct.(*string); ok {
		*str = string(response.Body())
	} else {
		return nil, fmt.Errorf("could not cast responseStruct to *string")
	}

	return response, nil
}

func (r *Requester) ReadJSONResponse(response *resty.Response, responseStruct interface{}) (*resty.Response, error) {
	response.Body()
	err := json.Unmarshal(response.Body(), responseStruct)
	return response, err
}
