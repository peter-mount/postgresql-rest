package openapi

type Schema struct {
	Reference  `yaml:",inline"`
	SchemaImpl `yaml:",inline"`
}

type SchemaImpl struct {
	Type       string             `yaml:"type"`
	Format     interface{}        `yaml:"format,omitempty"`
	Minimum    interface{}        `yaml:"minimum,omitempty"`
	Maximum    interface{}        `yaml:"maximum,omitempty"`
	Enum       interface{}        `yaml:"enum,omitempty"`
	Default    interface{}        `yaml:"default,omitempty"`
	Properties map[string]*Schema `yaml:"properties,omitempty"`
	Items      *Schema            `yaml:"items,omitempty"`
}

func (p *Schema) MarshalYAML() (interface{}, error) {
	if p.Reference.Ref == "" {
		return p.SchemaImpl, nil
	}
	return p.Reference, nil
}
