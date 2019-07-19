package openapi

// Method represents the handler of a method
type Method struct {
	Summary     string       `yaml:"summary,omitempty"`
	Description string       `yaml:"description,omitempty"`
	Parameters  []Parameters `yaml:"parameters,omitempty"`
	Handler     *Handler     `yaml:"handler,omitempty"`
}

type Parameters struct {
	Name     string `yaml:"name"`
	In       string `yaml:"in"`
	Required bool   `yaml:"required"`
	Schema   struct {
		Type    string      `yaml:"type"`
		Format  interface{} `yaml:"format,omitempty"`
		Minimum interface{} `yaml:"minimum,omitempty"`
		Maximum interface{} `yaml:"maximum,omitempty"`
		Enum    interface{} `yaml:"enum,omitempty"`
		Default interface{} `yaml:"default,omitempty"`
	} `yaml:"schema"`
	Description     string      `yaml:"description,omitempty"`
	AllowReserved   bool        `yaml:"allowReserved,omitempty"`
	AllowEmptyValue bool        `yaml:"allowEmptyValue,omitempty"`
	Deprecated      bool        `yaml:"deprecated,omitempty"`
	Example         interface{} `yaml:"example,omitempty"`
	Examples        interface{} `yaml:"examples,omitempty"`
}

// Handler contains non-OpenAPI hander config used by dbrest to implement the API
type Handler struct {
	Function string `yaml:"function"`
	MaxAge   int    `yaml:"maxAge"`
	Json     bool   `yaml:"json"`
}
