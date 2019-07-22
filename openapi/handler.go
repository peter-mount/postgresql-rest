package openapi

import (
	"database/sql"
	"fmt"
	"github.com/peter-mount/golib/rest"
	"io/ioutil"
	"log"
	"strings"
)

// Method represents the handler of a method
type Method struct {
	Tags        []string    `yaml:"tags,omitempty"`
	Summary     string      `yaml:"summary,omitempty"`
	Description string      `yaml:"description,omitempty"`
	Parameters  []Parameter `yaml:"parameters,omitempty"`
	Handler     *Handler    `yaml:"handler,omitempty"`
}

type Parameter struct {
	Name     string `yaml:"name"`
	In       string `yaml:"in"`
	Required bool   `yaml:"required"`
	Schema   struct {
		Type    string      `yaml:"type"`
		Format  interface{} `yaml:"format,omitempty"`
		Minimum interface{} `yaml:"minimum,omitempty"`
		Maximum interface{} `yaml:"maximum,omitempty"`
		Enum    interface{} `yaml:"enum,omitempty"`
		Default interface{} `yaml:"default,omitempty"`
	} `yaml:"schema"`
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
	}
}

func (m *Method) start(path, method string, server *rest.Server) {
	if m.Handler == nil {
		return
	}

	var params []string
	for i := range m.Parameters {
		params = append(params, fmt.Sprintf("$%d", i+1))
	}

	m.Handler.sql = "SELECT " + m.Handler.Function + "(" + strings.Join(params, ",") + ")"

	log.Println("ADD", method, path, m.Handler.sql)
	server.Handle(path, m.handleRequest).Methods(strings.ToUpper(method))
}

func (m *Method) extractArgs(r *rest.Rest) ([]interface{}, error) {
	var args []interface{}

	for i, param := range m.Parameters {
		var val interface{}

		switch param.In {
		// Non OpenAPI standard, allow body content to be used
		case "body":
			br, err := r.BodyReader()
			if err != nil {
				return nil, err
			}

			b, err := ioutil.ReadAll(br)
			if err != nil {
				return nil, err
			}
			val = string(b)

		case "header":
			val = r.GetHeader(param.Name)

		case "path":
			val = r.Var(param.Name)

			// Query params are handed the same as path in mux
		case "query":
			val = r.Var(param.Name)

		default:
			return nil, fmt.Errorf("no in for \"%s\" arg %d", param.Name, i)
		}

		args = append(args, val)
	}

	log.Println(args)
	return args, nil
}

func (m *Method) handleRequest(r *rest.Rest) error {
	args, err := m.extractArgs(r)
	if err != nil {
		return err
	}

	log.Println(m.Handler.sql)

	var result sql.NullString
	err = m.Handler.DB.QueryRow(m.Handler.sql, args...).Scan(&result)
	if err != nil {
		return err
	}

	if result.Valid {
		// As we are returning a single value then write that to the response as-is.
		// If we use Value(result) then it will get escaped
		r.Reader(strings.NewReader(result.String))
	} else {
		r.Value(nil)
	}

	if m.Handler.ContentType != "" {
		r.ContentType(m.Handler.ContentType)
	}

	if m.Handler.MaxAge < 0 {
		r.CacheNoCache()
	} else if m.Handler.MaxAge > 0 {
		r.CacheMaxAge(m.Handler.MaxAge)
	}

	return nil
}
