package openapi

import (
	"database/sql"
	"fmt"
	"github.com/peter-mount/golib/rest"
	"log"
	"strings"
)

// Method represents the handler of a method
type Method struct {
	Tags          []string            `yaml:"tags,omitempty"`
	Summary       string              `yaml:"summary,omitempty"`
	Description   string              `yaml:"description,omitempty"`
	Parameters    []*Parameter        `yaml:"parameters,omitempty"`
	Handler       *Handler            `yaml:"handler,omitempty"`
	Responses     map[string]Response `yaml:"responses,omitempty"`
	handler       rest.RestHandler    `yaml:"-"`
	paramHandlers []paramHandler      `yaml:"-"`
}

type Parameter struct {
	Reference     `yaml:",inline"`
	ParameterImpl `yaml:",inline"`
}

func (p *Parameter) MarshalYAML() (interface{}, error) {
	if p.Reference.Ref == "" {
		return p.ParameterImpl, nil
	}
	return p.Reference, nil
}

type ParameterImpl struct {
	Name            string      `yaml:"name"`
	In              string      `yaml:"in"`
	Required        bool        `yaml:"required"`
	Schema          Schema      `yaml:"schema"`
	Description     string      `yaml:"description,omitempty"`
	AllowReserved   bool        `yaml:"allowReserved,omitempty"`
	AllowEmptyValue bool        `yaml:"allowEmptyValue,omitempty"`
	Deprecated      bool        `yaml:"deprecated,omitempty"`
	Example         interface{} `yaml:"example,omitempty"`
	Examples        interface{} `yaml:"examples,omitempty"`
}

// Handler contains non-OpenAPI hander config used by dbrest to implement the API
type Handler struct {
	Function    string `yaml:"function"`
	MaxAge      int    `yaml:"maxAge"`
	ContentType string `yaml:"content-type"`
	DB          *DB    `yaml:"-"`
	sql         string
}

func (m *Method) Publish() *Method {
	if m == nil {
		return nil
	}

	return &Method{
		Description: m.Description,
		Parameters:  m.Parameters,
		Summary:     m.Summary,
		Tags:        m.Tags,
		Responses:   m.Responses,
	}
}

func (m *Method) start(path, method string, server *rest.Server) error {
	if m.Handler == nil {
		return nil
	}

	err := m.compile()
	if err != nil {
		return err
	}

	var params []string
	for i := range m.paramHandlers {
		params = append(params, fmt.Sprintf("$%d", i+1))
	}

	m.Handler.sql = "SELECT " + m.Handler.Function + "(" + strings.Join(params, ",") + ")"

	server.Handle(path, m.handler).Methods(strings.ToUpper(method))

	return nil
}

func (m *Method) defaultHandler(r *rest.Rest) error {
	log.Println(m.paramHandlers)
	args, err := m.extractArgs(r)
	if err != nil {
		return err
	}

	var result sql.NullString
	err = m.Handler.DB.QueryRow(m.Handler.sql, args...).Scan(&result)
	if err != nil {
		return WrapError(err)
	}

	if result.Valid {
		// As we are returning a single value then write that to the response as-is.
		// If we use Value(result) then it will get escaped
		r.Status(200).
			Reader(strings.NewReader(result.String))
	} else {
		// No value so return a 404
		return Error404("")
	}

	if m.Handler.ContentType != "" {
		// Forced in handler definition
		r.ContentType(m.Handler.ContentType)
	} else {
		// Look for first content type in the methods responses for "200"
		if resp, ok := m.Responses["200"]; ok {
			ct := ""
			for c, _ := range resp.Content {
				if ct == "" {
					ct = c
				}
			}
			if ct != "" {
				r.ContentType(ct)
			}
		}
	}

	if m.Handler.MaxAge < 0 {
		r.CacheNoCache()
	} else if m.Handler.MaxAge > 0 {
		r.CacheMaxAge(m.Handler.MaxAge)
	}

	return nil
}
