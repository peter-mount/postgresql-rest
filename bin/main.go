package main

import (
	"github.com/peter-mount/golib/kernel"
	dbrest "github.com/peter-mount/postgresql-rest"
	"log"
)

func main() {
	err := kernel.Launch(&dbrest.DBRest{})
	if err != nil {
		log.Fatal(err)
	}
}
