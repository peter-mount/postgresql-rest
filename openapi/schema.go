package openapi

import (
	"fmt"
	"github.com/peter-mount/golib/rest"
	"regexp"
	"strconv"
	"strings"
)

type Schema struct {
	Reference  `yaml:",inline"`
	SchemaImpl `yaml:",inline"`
}

type SchemaImpl struct {
	Type                 string             `yaml:"type"`
	Format               interface{}        `yaml:"format,omitempty"`
	Pattern              string             `yaml:"pattern,omitempty"`
	MinLength            *int               `yaml:"minLenth,omitempty"`
	MaxLength            *int               `yaml:"maxLength,omitempty"`
	Minimum              *int               `yaml:"minimum,omitempty"`
	Maximum              *int               `yaml:"maximum,omitempty"`
	ExclusiveMinimum     bool               `yaml:"exclusiveMinimum,omitempty"`
	ExclusiveMaximum     bool               `yaml:"exclusiveMaximum,omitempty"`
	Enum                 []string           `yaml:"enum,omitempty"`
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

	// Although pattern is strictly for strings, we allow it for all types as they are strings at this point
	if p.Pattern != "" {
		h, err = p.compileRegex(name, h)
	}

	if err == nil {
		switch p.Type {
		case "boolean":
			h = p.compileBoolean(name, h)

		case "integer":
			h = p.compileInteger(name, h)

		case "string":
			h = p.compileString(name, h)
		}

	}

	if err == nil && len(p.Enum) > 0 {
		// Support enums once validated
		h = p.compileEnum(name, h, p.Enum)
	}

	return h, err
}

func (p *Schema) compileEnum(name string, h paramHandler, enum []string) paramHandler {
	enumError := Error400("%s not in %s", name, strings.Join(enum, ", "))

	return func(r *rest.Rest) (interface{}, error) {
		v, err := h(r)
		if err != nil {
			return nil, err
		}

		s := v.(string)
		for _, e := range enum {
			if s == e {
				return s, nil
			}
		}

		return nil, enumError
	}
}

func (p *Schema) compileBoolean(name string, h paramHandler) paramHandler {
	boolError := Error400("%s must match \"true\" or \"false\"", name)

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

func (p *Schema) compileString(name string, h paramHandler) paramHandler {

	var min int
	if p.MinLength != nil {
		min = *p.MinLength
	}

	var max int
	if p.MaxLength != nil {
		max = *p.MaxLength - 1
	}

	var filter func(string) error
	if p.MinLength != nil && p.MaxLength != nil {
		filter = func(s string) error {
			l := len(s)
			if l >= min && l <= max {
				return nil
			}
			return Error400("%s out of bounds len %d...%d", name, min, max)
		}
	} else if p.MinLength != nil {
		filter = func(s string) error {
			if len(s) >= min {
				return nil
			}
			return Error400("%s out of bounds len min %d", name, min)
		}
	} else if p.Maximum != nil {
		filter = func(s string) error {
			if len(s) <= max {
				return nil
			}
			return Error400("%s %d out of bounds max %d", name, max)
		}
	} else {
		// No min/max set
		filter = func(s string) error {
			return nil
		}
	}

	// Validate the parameter is an integer
	return func(r *rest.Rest) (interface{}, error) {
		v, err := h(r)
		if err != nil {
			return nil, err
		}

		err = filter(v.(string))
		if err != nil {
			return nil, err
		}

		return v, nil
	}

}

// compileInteger ensures an integer parameter is that and, if configured valid within a specified range
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
			return Error400("Value %d out of bounds %d...%d", i, min, max)
		}
	} else if p.Minimum != nil {
		filter = func(i int) error {
			if i >= min {
				return nil
			}
			return Error400("Value %d out of bounds %d...", i, min)
		}
	} else if p.Maximum != nil {
		filter = func(i int) error {
			if i <= max {
				return nil
			}
			return Error400("Value %d out of bounds ...%d", i, max)
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

// compileRegex handles the pattern matching
func (p *Schema) compileRegex(name string, h paramHandler) (paramHandler, error) {
	exp, err := regexp.Compile(p.Pattern)
	if err != nil {
		return nil, fmt.Errorf("Invalid pattern \"%s\" for %s: %s", p.Pattern, name, err.Error())
	}

	patternError := Error400("Param %s must match \"%s\"", name, p.Pattern)

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
