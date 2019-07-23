package openapi

import (
	"gopkg.in/yaml.v3"
)

// Paths is a MapSlice of methods which are themselves a MapSlice of Handlers
type Paths struct {
	paths []pathEntry
}

type pathEntry struct {
	key  string
	path *Path
}

func (p *Paths) UnmarshalYAML(node *yaml.Node) error {
	for i := 0; i < len(node.Content); i = i + 2 {

		var key string
		err := node.Content[i].Decode(&key)
		if err != nil {
			return err
		}

		path := &Path{}
		err = node.Content[i+1].Decode(path)
		if err != nil {
			return err
		}

		p.Set(key, path)
	}

	return nil
}

func (p Paths) MarshalYAML() (interface{}, error) {
	n := &yaml.Node{
		Kind: yaml.MappingNode,
		Tag:  "!!map",
	}

	for _, e := range p.paths {

		b, err := yaml.Marshal(e.path)
		if err != nil {
			return nil, err
		}
		c := &yaml.Node{}
		err = yaml.Unmarshal(b, c)
		if err != nil {
			return nil, err
		}

		n.Content = append(n.Content,
			&yaml.Node{
				Kind:  yaml.ScalarNode,
				Tag:   "!!str",
				Value: e.key,
			},
			c.Content[0],
		)
	}

	return n, nil
}

func (p *Paths) Set(key string, path *Path) {
	for _, e := range p.paths {
		if e.key == key {
			e.path = path
			return
		}
	}
	p.paths = append(p.paths, pathEntry{key: key, path: path})
}

func (p *Paths) Get(key string) *Path {
	for _, e := range p.paths {
		if e.key == key {
			return e.path
		}
	}
	return nil
}

type Path struct {
	Summary     string  `yaml:"summary,omitempty"`
	Description string  `yaml:"description,omitempty"`
	Get         *Method `yaml:"get,omitempty"`
	Post        *Method `yaml:"post,omitempty"`
	Put         *Method `yaml:"put,omitempty"`
	Patch       *Method `yaml:"patch,omitempty"`
	Delete      *Method `yaml:"delete,omitempty"`
	Head        *Method `yaml:"head,omitempty"`
	Options     *Method `yaml:"options,omitempty"`
	Trace       *Method `yaml:"trace,omitempty"`
}

// AddHandler adds a handler to this OpenAPI using the specified path and method.
// OpenAPI supports the following methods: get, post, put, patch, delete, head, options and trace.
func (api *OpenAPI) AddHandler(path, method string, handler *Method) {
	p := api.Paths.Get(path)
	if p == nil {
		p = &Path{}
		api.Paths.Set(path, p)
	}

	switch method {
	case "get":
		p.Get = handler
	case "post":
		p.Post = handler
	case "put":
		p.Put = handler
	case "patch":
		p.Patch = handler
	case "delete":
		p.Delete = handler
	case "head":
		p.Head = handler
	case "options":
		p.Options = handler
	case "trace":
		p.Trace = handler
	}
}

// ForEachPath calls a function for each path and each method within it.
// Note this will not call a path with the special summary and description keys
// used in the OpenAPI spec
func (api *OpenAPI) ForEachPath(f func(path, method string, handler *Method) error) error {
	for _, e := range api.Paths.paths {
		path := e.key
		methods := e.path
		var err error

		if methods.Get != nil {
			err = f(path, "get", methods.Get)
		}

		if err == nil && methods.Post != nil {
			err = f(path, "post", methods.Post)
		}

		if err == nil && methods.Put != nil {
			err = f(path, "put", methods.Put)
		}

		if err == nil && methods.Patch != nil {
			err = f(path, "patch", methods.Patch)
		}

		if err == nil && methods.Delete != nil {
			err = f(path, "delete", methods.Delete)
		}

		if err == nil && methods.Head != nil {
			err = f(path, "head", methods.Head)
		}

		if err == nil && methods.Options != nil {
			err = f(path, "options", methods.Options)
		}

		if err == nil && methods.Trace != nil {
			err = f(path, "trace", methods.Trace)
		}

		if err != nil {
			return err
		}
	}
	return nil
}
