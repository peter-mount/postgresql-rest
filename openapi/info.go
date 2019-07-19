package openapi

type Info struct {
	Title          string   `yaml:"title"`
	Description    string   `yaml:"description,omitempty"`
	TermsOfService string   `yaml:"termsOfService,omitempty"`
	Contact        *Contact `yaml:"contact,omitempty"`
	License        *License `yaml:"license,omitempty"`
	Version        string   `yaml:"version"`
}
