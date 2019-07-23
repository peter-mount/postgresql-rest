package openapi

import (
	"fmt"
	"github.com/peter-mount/golib/rest"
	"regexp"
	"strconv"
)

type Schema struct {
	Reference  `yaml:",inline"`
	SchemaImpl `yaml:",inline"`
}

type SchemaImpl struct {
	Type                 string             `yaml:"type"`
	Format               interface{}        `yaml:"format,omitempty"`
	Pattern              string             `yaml:"pattern,omitempty"`
	Minimum              *int               `yaml:"minimum,omitempty"`
	Maximum              *int               `yaml:"maximum,omitempty"`
	ExclusiveMinimum     bool               `yaml:"exclusiveMinimum,omitempty"`
	ExclusiveMaximum     bool               `yaml:"exclusiveMaximum,omitempty"`
	Enum                 interface{}        `yaml:"enum,omitempty"`
	Default              interface{}        `yaml:"default,omitempty"`
	Properties           map[string]*Schema `yaml:"properties,omitempty"`
	Items                *Schema            `yaml:"items,omitempty"`
	AdditionalProperties *Schema            `yaml:"additionalProperties,omitempty"`
}

func (p *Schema) MarshalYAML() (interface{}, error) {
	if p.Reference.Ref == "" {
		return p.SchemaImpl, nil
	}
	return p.Reference, nil
}

// Add any paramHandler's to validate the request to the schema
func (p *Schema) compile(name string, h paramHandler) (paramHandler, error) {
	var err error

	if p.Pattern != "" {
		h, err = p.compileRegex(name, h)
	}

	if err == nil {
		switch p.Type {
		case "boolean":
			h = p.compileBoolean(name, h)

		case "integer":
			h = p.compileInteger(name, h)
		}
	}

	return h, err
}

func (p *Schema) compileBoolean(name string, h paramHandler) paramHandler {
	boolError := fmt.Errorf("Param %s must match \"true\" or \"false\"", name)

	return func(r *rest.Rest) (interface{}, error) {
		v, err := h(r)
		if err != nil {
			return nil, err
		}

		switch v {
		case "true":
			return v, nil

		case "false":
			return v, nil

		default:
			return nil, boolError
		}
	}
}

func (p *Schema) compileInteger(name string, h paramHandler) paramHandler {

	var min int
	if p.Minimum != nil {
		if p.ExclusiveMinimum {
			min = *p.Minimum + 1
		} else {
			min = *p.Minimum
		}
	}

	var max int
	if p.Maximum != nil {
		if p.ExclusiveMaximum {
			max = *p.Maximum - 1
		} else {
			max = *p.Maximum - 1
		}
	}

	var filter func(i int) error
	if p.Minimum != nil && p.Maximum != nil {
		filter = func(i int) error {
			if i >= min && i <= max {
				return nil
			}
			return fmt.Errorf("Value %d out of bounds %d...%d", i, min, max)
		}
	} else if p.Minimum != nil {
		filter = func(i int) error {
			if i >= min {
				return nil
			}
			return fmt.Errorf("Value %d out of bounds %d...", i, min)
		}
	} else if p.Maximum != nil {
		filter = func(i int) error {
			if i <= max {
				return nil
			}
			return fmt.Errorf("Value %d out of bounds ...%d", i, max)
		}
	} else {
		// No min/max set
		filter = func(i int) error {
			return nil
		}
	}

	// Validate the parameter is an integer
	return func(r *rest.Rest) (interface{}, error) {
		v, err := h(r)
		if err != nil {
			return nil, err
		}

		i, err := strconv.Atoi(v.(string))
		if err != nil {
			return nil, err
		}

		err = filter(i)
		if err != nil {
			return nil, err
		}

		return v, nil
	}
}

func (p *Schema) IntMin(name string, min int, h paramHandler) paramHandler {
	minError := fmt.Errorf("Param %s minimum of %d", name, min)

	return func(r *rest.Rest) (interface{}, error) {
		v, err := h(r)
		if err != nil {
			return nil, err
		}

		i := v.(int)

		if i < min {
			return nil, minError
		}

		return i, nil
	}
}

func (p *Schema) IntMax(name string, max int, h paramHandler) paramHandler {
	maxError := fmt.Errorf("Param %s maximum of %d", name, max)

	return func(r *rest.Rest) (interface{}, error) {
		v, err := h(r)
		if err != nil {
			return nil, err
		}

		i := v.(int)

		if i > max {
			return nil, maxError
		}

		return i, nil
	}
}

func (p *Schema) compileRegex(name string, h paramHandler) (paramHandler, error) {
	patternError := fmt.Errorf("Param %s must match \"%s\"", name, p.Pattern)

	exp, err := regexp.Compile(p.Pattern)
	if err != nil {
		return nil, fmt.Errorf("Invalid pattern \"%s\" for %s: %s", p.Pattern, name, err.Error())
	}

	return func(r *rest.Rest) (interface{}, error) {
		v, err := h(r)
		if err != nil {
			return nil, err
		}

		if exp.MatchString(v.(string)) {
			return v, nil
		}

		return nil, patternError
	}, nil
}
