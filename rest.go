package dbrest

import (
	"database/sql"
	"fmt"
	"github.com/peter-mount/golib/rest"
	"github.com/peter-mount/postgresql-rest/openapi"
	"io/ioutil"
	"log"
	"strings"
)

// Handler represents a REST endpoint backed by a postgresql function.
// Note that (currently) no authentication is done unless you add it to your postgres code.
type RestHandler struct {
	// The method, defaults to GET
	Method string `yaml:"method"`
	// The path of the endpoint including any parameter definitions
	Path string `yaml:"path"`
	// The parameter names in the order to be provided to Postgres
	Params []string `yaml:"params"`
	// The postgres function to call
	Function string `yaml:"function"`
	// If true then the response will have a JSON ContentType
	JSON bool `yaml:"json"`
	// If true then the response will have a XML ContentType
	XML bool `yaml:"xml"`
	// Cache control, <0 for no cache, >0 for max age in seconds
	Cache int `yaml:"maxAge"`
	// Headers to set in response
	Headers map[string]string `yaml:"headers"`
	// Prepared statement
	sql string
	db  *openapi.DB
}

func (handler *RestHandler) init(prefix string, db *openapi.DB, server *rest.Server) {
	if handler.Method == "" {
		handler.Method = "GET"
	} else {
		handler.Method = strings.ToUpper(handler.Method)
	}

	handler.db = db

	var params []string
	for i := range handler.Params {
		params = append(params, fmt.Sprintf("$%d", i+1))
	}

	handler.sql = "SELECT " + handler.Function + "(" + strings.Join(params, ",") + ")"

	path := handler.Path
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	if prefix != "" {
		path = prefix + path
		if !strings.HasPrefix(path, "/") {
			path = "/" + path
		}
	}

	log.Println(path, handler.sql)
	server.Handle(path, handler.handleRequest).Methods(handler.Method)
}

const (
	restBody         = "body"
	restNull         = "null"
	restHeaderPrefix = "header:"
	restHeaderLen    = len(restHeaderPrefix)
	restVarPrefix    = "var:"
	restVarLen       = len(restVarPrefix)
)

func (handler *RestHandler) extractArgs(r *rest.Rest) ([]interface{}, error) {
	var args []interface{}

	for _, param := range handler.Params {
		var val interface{}

		if param == restBody {
			br, err := r.BodyReader()
			if err != nil {
				return nil, err
			}

			b, err := ioutil.ReadAll(br)
			if err != nil {
				return nil, err
			}
			val = string(b)
		} else if param == restNull {
			val = nil
		} else if strings.HasPrefix(param, restHeaderPrefix) {
			val = r.GetHeader(param[restHeaderLen:])
		} else if strings.HasPrefix(param, restVarPrefix) {
			val = r.Var(param[restVarLen:])
		} else {
			val = r.Var(param)
		}

		args = append(args, val)
	}

	return args, nil
}

func (handler *RestHandler) handleRequest(r *rest.Rest) error {
	args, err := handler.extractArgs(r)
	if err != nil {
		return err
	}

	log.Println(handler.sql)

	var result sql.NullString
	err = handler.db.QueryRow(handler.sql, args...).Scan(&result)
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

	if handler.JSON {
		r.JSON()
	} else if handler.XML {
		r.XML()
	}

	if handler.Cache < 0 {
		r.CacheNoCache()
	} else if handler.Cache > 0 {
		r.CacheMaxAge(handler.Cache)
	}

	for k, v := range handler.Headers {
		r.AddHeader(k, v)
	}

	return nil
}
