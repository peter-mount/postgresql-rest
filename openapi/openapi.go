package openapi

import (
	"errors"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"log"
	"path/filepath"
)

type OpenAPI struct {
	OpenAPI string   `yaml:"openapi,omitempty"`
	Info    *Info    `yaml:"info,omitempty"`
	Servers []Server `yaml:"servers,omitempty"`
	Paths   Paths    `yaml:"paths"`
	//Security     []SecurityRequirement   `yaml:"security,omitempty"`
	//Tags         []Tag                   `yaml:"tags,omitempty"`
	//ExternalDocs []ExternalDocumentation `yaml:"externalDocs,omitempty"`

	// Non standard, specific to dbrest
	Prefix    string            `yaml:"-"`
	Webserver *Webserver        `yaml:"webserver,omitempty"`
	DB        *DB               `yaml:"db,omitempty"`
	Imports   map[string]string `yaml:"import,omitempty"`
	children  []*OpenAPI
}

func NewOpenAPI() *OpenAPI {
	return newChildOpenAPI("")
}

func newChildOpenAPI(prefix string) *OpenAPI {
	api := &OpenAPI{Prefix: prefix}
	return api
}

func (c *OpenAPI) Unmarshal(filename string) error {
	temp := NewOpenAPI()

	err := temp.unmarshal(nil, filename)
	if err != nil {
		return err
	}

	c.Info = temp.Info
	c.Servers = temp.Servers
	c.DB = temp.DB
	c.Webserver = temp.Webserver

	// Now flatten it using ourselves as the destination
	return temp.flatten(c)
}

func (c *OpenAPI) unmarshal(parent *OpenAPI, filename string) error {
	filename, err := filepath.Abs(filename)
	if err != nil {
		return err
	}

	log.Println("Loading", filename)

	in, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(in, c)
	if err != nil {
		return err
	}

	if c.DB == nil {
		if parent == nil {
			return errors.New("Database is mandatory for the root config.yaml")
		}
		c.DB = parent.DB
	}

	if len(c.Imports) > 0 {

		base := filepath.Dir(filename)

		for prefix, path := range c.Imports {
			importFileName := filepath.Join(base, path)
			importFileName, err := filepath.Abs(importFileName)
			if err != nil {
				return err
			}

			ci := newChildOpenAPI(AddPrefix(c.Prefix, prefix))

			err = ci.unmarshal(c, importFileName)
			if err != nil {
				return err
			}

			c.children = append(c.children, ci)
		}
	}

	return nil
}

// Publish creates a clean valid OpenAPI based on our config.
func (c *OpenAPI) Publish() *OpenAPI {
	d := NewOpenAPI()
	d.OpenAPI = "3.0.0"
	d.Info = c.Info
	d.Servers = c.Servers

	// Copy the paths without our extensions
	for _, e := range c.Paths.paths {
		methods := e.path
		d.Paths.Set(
			e.key,
			&Path{
				Summary:     methods.Summary,
				Description: methods.Description,
				Get:         methods.Get.Publish(),
				Post:        methods.Post.Publish(),
				Put:         methods.Put.Publish(),
				Patch:       methods.Patch.Publish(),
				Delete:      methods.Delete.Publish(),
				Head:        methods.Head.Publish(),
				Options:     methods.Options.Publish(),
				Trace:       methods.Trace.Publish(),
			},
		)
	}

	return d
}

func (c *OpenAPI) flatten(d *OpenAPI) error {

	err := c.DB.Start()
	if err != nil {
		return err
	}

	// Ensure head Method has it's DB attached
	_ = c.ForEachPath(func(path, method string, handler *Method) error {
		if handler.Handler != nil {
			handler.Handler.DB = c.DB
		}
		return nil
	})

	// Import the paths
	for _, e := range c.Paths.paths {
		d.Paths.Set(AddPrefix(c.Prefix, e.key), e.path)
	}

	// Now any children
	for _, child := range c.children {
		err = child.flatten(d)
		if err != nil {
			return err
		}
	}

	return nil
}
