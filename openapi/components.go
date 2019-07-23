package openapi

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"log"
	"strings"
)

type Components struct {
	Schemas         map[string]Schema     `yaml:"schemas,omitempty"`
	Parameters      map[string]Parameter  `yaml:"parameters,omitempty"`
	SecuritySchemes map[string]*yaml.Node `yaml:"securitySchemes,omitempty"`
	RequestBodies   map[string]*yaml.Node `yaml:"requestBodies,omitempty"`
	Responses       map[string]Response   `yaml:"responses,omitempty"`
	Headers         map[string]*yaml.Node `yaml:"headers,omitempty"`
	Examples        map[string]*yaml.Node `yaml:"examples,omitempty"`
	Links           map[string]*yaml.Node `yaml:"links,omitempty"`
	Callbacks       map[string]*yaml.Node `yaml:"callbacks,omitempty"`
}

func (c *Components) init() {
	c.Schemas = make(map[string]Schema)
	c.Parameters = make(map[string]Parameter)
	c.Responses = make(map[string]Response)

	c.SecuritySchemes = make(map[string]*yaml.Node)
	c.RequestBodies = make(map[string]*yaml.Node)
	c.Headers = make(map[string]*yaml.Node)
	c.Examples = make(map[string]*yaml.Node)
	c.Links = make(map[string]*yaml.Node)
	c.Callbacks = make(map[string]*yaml.Node)
}

func (c *Components) flatten(d *Components) error {
	if c == nil {
		return nil
	}

	err := flattenSchema(c.Schemas, &d.Schemas)

	if err == nil {
		err = flattenParameter(c.Parameters, &d.Parameters)
	}

	if err == nil {
		err = flattenNode(c.SecuritySchemes, &d.SecuritySchemes)
	}

	if err == nil {
		err = flattenNode(c.RequestBodies, &d.RequestBodies)
	}

	if err == nil {
		err = flattenResponses(c.Responses, &d.Responses)
	}

	if err == nil {
		err = flattenNode(c.Headers, &d.Headers)
	}

	if err == nil {
		err = flattenNode(c.Examples, &d.Examples)
	}

	if err == nil {
		err = flattenNode(c.Links, &d.Links)
	}

	if err == nil {
		err = flattenNode(c.Callbacks, &d.Callbacks)
	}

	return err
}

func flattenParameter(s map[string]Parameter, d *map[string]Parameter) error {
	for k, v := range s {
		_, exists := (*d)[k]
		if exists {
			return fmt.Errorf("parameter \"%s\" already exists", k)
		}
		(*d)[k] = v
	}

	return nil
}

func flattenSchema(s map[string]Schema, d *map[string]Schema) error {
	for k, v := range s {
		_, exists := (*d)[k]
		if exists {
			return fmt.Errorf("schema \"%s\" already exists", k)
		}
		(*d)[k] = v
	}

	return nil
}

func flattenResponses(s map[string]Response, d *map[string]Response) error {
	for k, v := range s {
		_, exists := (*d)[k]
		if exists {
			return fmt.Errorf("response \"%s\" already exists", k)
		}
		(*d)[k] = v
	}

	return nil
}

func flattenNode(s map[string]*yaml.Node, d *map[string]*yaml.Node) error {
	//log.Println("flattenNode")
	//log.Println(s)
	//log.Println(*d)
	for k, v := range s {
		e, exists := (*d)[k]
		if exists {
			return fmt.Errorf(
				"Entry \"%s\" already exists at [%d:%d] 2nd ref at [%d:%d]",
				k,
				e.Line, e.Column,
				v.Line, v.Column,
			)
		}
		(*d)[k] = v
		debug(k, 0, v)
	}

	return nil
}

func debug(k string, i int, n *yaml.Node) {
	s := k + " %02d " + strings.Repeat(" ", i*2) + "%02d %s = %s\n"
	log.Printf(s, i,
		n.Kind,
		n.Tag,
		n.Value,
	)
	for _, c := range n.Content {
		debug(k, i+1, c)
	}
}
