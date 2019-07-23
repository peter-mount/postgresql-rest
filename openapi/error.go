package openapi

import (
	"fmt"
	"github.com/peter-mount/golib/rest"
	"log"
	"strconv"
)

// A common error object
type restError struct {
	Status  int    `json:"status,omitempty" xml:"status,attr,omitempty" yaml:"status,omitempty"`
	Message string `json:"message,omitempty" xml:"message,attr,omitempty" yaml:"message,omitempty"`
}

func (e *restError) Error() string {
	return e.Message
}

func (e *restError) Apply(r *rest.Rest) {
	r.Status(e.Status).
		Value(e)
}

func NewError(s int, m string, a ...interface{}) error {
	if a != nil && len(a) > 0 {
		m = fmt.Sprintf(m, a...)
	}
	return &restError{Status: s, Message: m}
}

func Error400(m string, a ...interface{}) error {
	return NewError(400, m, a...)
}

func Error404(m string, a ...interface{}) error {
	if m == "" {
		m = "Not found"
	}
	return NewError(404, m, a...)
}

func Error500(m string, a ...interface{}) error {
	return NewError(500, m, a...)
}

func WrapError(e error) error {
	if a, ok := e.(*restError); ok {
		return a
	}
	return Error500(e.Error())
}

// IsErrorSchema returns true if the schema defines our Error struct
func (s *Schema) IsErrorSchema() bool {
	if !(s != nil && s.Type == "object" && len(s.Properties) == 2) {
		return false
	}

	ps, ok := s.Properties["status"]
	if !ok || ps.Type != "integer" {
		return false
	}

	ps, ok = s.Properties["message"]
	if !ok || ps.Type != "string" {
		return false
	}

	return true
}

func (m *Method) isValidContentType(c string) bool {
	for _, v := range []string{
		"application/json",
		"application/xml",
		"application/yaml",
		"text/json",
		"text/xml",
		"text/yaml",
	} {
		if v == c {
			return true
		}
	}
	return false
}

// compileHandler takes the default handler and wraps it with handlers to handle custom errors
// if those response codes are defined in the schema
func (m *Method) compileHandler() error {
	// Start with the default handler
	m.handler = m.defaultHandler

	for status, content := range m.Responses {
		statusCode, err := strconv.Atoi(status)
		if err != nil {
			return err
		}

		for contentType, response := range content.Content {
			if m.isValidContentType(contentType) && response.Schema.IsErrorSchema() {
				m.handler = wrapError(statusCode, contentType, m.handler)
			}
		}
	}

	m.handler = wrapAnyErrors(m.handler)
	return nil
}

// wrapError wraps the request to catch specific errors
func wrapError(status int, contentType string, h rest.RestHandler) rest.RestHandler {
	return func(r *rest.Rest) error {
		err := h(r)

		if err != nil {
			// Not our error or for the wrong status then ignore
			ourError, ok := err.(*restError)
			if !ok || ourError.Status != status {
				return err
			}

			// Replace with our new response and follow through as if successful
			ourError.Apply(r.ContentType(contentType))
		}
		return nil
	}
}

// Catches all untrapped errors and responds with a 500 & no body unless it's one of our errors so then use it's
// status. The error message is logged to the console
func wrapAnyErrors(h rest.RestHandler) rest.RestHandler {
	return func(r *rest.Rest) error {
		err := h(r)
		if err != nil {
			status := 500
			if a, ok := err.(*restError); ok {
				status = a.Status
			}

			log.Println(status, err)

			r.Status(status).
				Value(nil)
		}
		return nil
	}
}
