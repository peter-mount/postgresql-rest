package openapi

type Response struct {
	Reference    `yaml:",inline"`
	ResponseImpl `yaml:",inline"`
}

type ResponseImpl struct {
	Description string                   `yaml:"description,omitempty"`
	Content     map[string]ResponseEntry `yaml:"content,omitempty"`
}

type ResponseEntry struct {
	Schema *Schema `yaml:"schema,omitempty"`
}
