package dbrest

import (
	"errors"
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
	return a.config.Unmarshal(filename)
}

func (a *DBRest) Start() error {

	b, err := yaml.Marshal(&a.config)
	if err != nil {
		return err
	}
	log.Println("\n" + string(b[:]))

	api := a.config.Publish()
	b, err = yaml.Marshal(api)
	if err != nil {
		return err
	}
	log.Println("\n" + string(b[:]))

	//return nil //a.config.start(a)
	return errors.New("FIN")
}
