package ejenkins

import "encoding/json"

func getJsonString(data interface{}) string {
	str, err := json.Marshal(data)
	if err != nil {
		return ""
	}
	return string(json.RawMessage(str))
}
