package main

import (
	"log"

	"github.com/jimmyjames85/proxysqlapi/pkg/server"
	"github.com/kelseyhightower/envconfig"
)

func main() {

	cfg := server.Config{}
	envconfig.MustProcess("PROXYSQLAPI", &cfg)
	srv, err := server.New(cfg)

	if err != nil {
		log.Fatalf("err loading config: %v", err)
	}

	err = srv.Serve()

	if err != nil {
		log.Fatalf("could not start server: %v", err)
	}

}
