package dbrest

import (
	"flag"
	"github.com/peter-mount/golib/kernel"
	"github.com/peter-mount/golib/kernel/cron"
	"github.com/peter-mount/golib/rest"
	"github.com/peter-mount/postgresql-rest/openapi"
	"gopkg.in/yaml.v3"
	"log"
	"path/filepath"
)

type DBRest struct {
	configFile *string
	config     *openapi.OpenAPI
	cron       *cron.CronService
	rest       *rest.Server
}

func (a *DBRest) Name() string {
	return "DBRest"
}

func (a *DBRest) Init(k *kernel.Kernel) error {
	a.configFile = flag.String("c", "", "The config file to use")

	service, err := k.AddService(&cron.CronService{})
	if err != nil {
		return err
	}
	a.cron = (service).(*cron.CronService)

	service, err = k.AddService(&rest.Server{})
	if err != nil {
		return err
	}
	a.rest = (service).(*rest.Server)

	return nil
}

func (a *DBRest) PostInit() error {
	if *a.configFile == "" {
		*a.configFile = "config.yaml"
	}

	filename, err := filepath.Abs(*a.configFile)
	if err != nil {
		return err
	}

	a.config = openapi.NewOpenAPI()
	err = a.config.Unmarshal(filename)
	if err != nil {
		return err
	}

	log.Println(a.config.Webserver)
	if a.config.Webserver != nil {
		if a.config.Webserver.Port > 0 {
			a.rest.Port = a.config.Webserver.Port
		}
	}

	return nil
}

func (a *DBRest) Start() error {

	a.config.Start(a.rest)

	if a.config.Webserver != nil && a.config.Webserver.ExposeOpenAPI != "" {
		api := a.config.Publish()
		b, err := yaml.Marshal(api)
		if err != nil {
			return err
		}
		a.rest.Handle("/openapi.yaml", func(r *rest.Rest) error {
			r.ContentType("application/yaml").
				Value(b)
			return nil
		})

		a.rest.Static("/", a.config.Webserver.ExposeOpenAPI)
	}

	return nil
}
