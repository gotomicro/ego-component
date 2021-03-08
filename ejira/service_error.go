package ejira

import "bytes"

// Error message from Jira
// See https://developer.atlassian.com/cloud/jira/platform/rest/v2/intro/#status-codes
type Error struct {
	ErrorMessages []string          `json:"errorMessages"`
	Errors        map[string]string `json:"errors"`
}

// LongError is a full representation of the error as a string
func (e *Error) LongError() string {
	var msg bytes.Buffer
	if len(e.ErrorMessages) > 0 {
		msg.WriteString("\nMessages:\n")
		for _, v := range e.ErrorMessages {
			msg.WriteString(" - ")
			msg.WriteString(v)
			msg.WriteString("\n")
		}
	}
	if len(e.Errors) > 0 {
		for key, value := range e.Errors {
			msg.WriteString(" - ")
			msg.WriteString(key)
			msg.WriteString(" - ")
			msg.WriteString(value)
			msg.WriteString("\n")
		}
	}
	return msg.String()
}
