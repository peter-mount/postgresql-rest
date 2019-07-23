package openapi

import (
	"fmt"
	"github.com/peter-mount/golib/rest"
	"io/ioutil"
	"log"
)

// ForEachPath calls a function for each path and each method within it.
// Note this will not call a path with the special summary and description keys
// used in the OpenAPI spec
func (api *OpenAPI) Start(server *rest.Server) error {
	return api.ForEachPath(func(path, method string, m *Method) error {
		return m.start(path, method, server)
	})
}

// paramHandler is a function that extracts a parameter or fails if invalid
type paramHandler func(r *rest.Rest) (interface{}, error)

// extractArgs runs through the required parameters, validating them and returns a slice or an error
func (m *Method) extractArgs(r *rest.Rest) ([]interface{}, error) {
	var args []interface{}

	for _, h := range m.paramHandlers {
		arg, err := h(r)
		if err != nil {
			return nil, err
		}
		args = append(args, arg)
	}

	return args, nil
}

// compile compiles the method by compiling all parameters
func (m *Method) compile() error {

	for _, param := range m.Parameters {
		h, err := param.compile()
		if err != nil {
			return err
		}
		if h != nil {
			m.paramHandlers = append(m.paramHandlers, h)
		}
	}

	// Finally compile the rest handler
	return m.compileHandler()
}

// compile the parameter by compiling it and applying any validations defined in the schema
func (param *Parameter) compile() (paramHandler, error) {
	h, err := param.compileParam()
	if err != nil {
		return nil, err
	}

	h, err = param.Schema.compile(param.Name, h)

	return h, nil
}

// compileParam returns a paramHandler that extracts the value.
// This is used prior to any other paramHandler's to validate the value
func (param *Parameter) compileParam() (paramHandler, error) {

	switch param.In {
	// Non OpenAPI standard, allow body content to be used
	case "body":
		return func(r *rest.Rest) (interface{}, error) {
			br, err := r.BodyReader()
			if err != nil {
				return nil, err
			}

			b, err := ioutil.ReadAll(br)
			if err != nil {
				return nil, err
			}
			return string(b), nil
		}, nil

	case "header":
		return func(r *rest.Rest) (interface{}, error) {
			val := r.GetHeader(param.Name)
			if val == "" {
				return nil, fmt.Errorf("missing header %s", param.Name)
			}
			return val, nil
		}, nil

	case "path":
		return func(r *rest.Rest) (interface{}, error) {
			val := r.Var(param.Name)
			log.Println("path", param.Name, val)
			if val == "" {
				return nil, fmt.Errorf("missing path %s", param.Name)
			}
			return val, nil
		}, nil

		// Query params are handed the same as path in mux
	case "query":
		return func(r *rest.Rest) (interface{}, error) {
			val := r.Var(param.Name)
			if val == "" {
				return nil, fmt.Errorf("missing query %s", param.Name)
			}
			return val, nil
		}, nil

	default:
		return nil, fmt.Errorf("no in for \"%s\"", param.Name)
	}

}
