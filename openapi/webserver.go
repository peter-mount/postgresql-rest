package openapi

type Webserver struct {
	// Port defines the port to expose the http server, defaults to 8080
	Port int `yaml:"port"`
	// ExposeOpenAPI contains the path to the static directory containing swagger-ui
	ExposeOpenAPI string `yaml:"exposeOpenAPI"`
}
