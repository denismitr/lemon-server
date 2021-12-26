package main

import (
	"flag"
	"github.com/denismitr/lemon-server/internal/server"
	"github.com/denismitr/lemon-server/internal/server/serverpb"
	"log"
	"os"
)

var version = "dev"

func main() {
	var cfgFile = flag.String("config", "", "Path to config yaml file")
	var env = flag.String("env", "dev", "Environment to run server in. Supported values (dev, prod)")
	flag.Parse()

	serverEnv, err := server.CreateEnvironment(*env)
	if err != nil {
		log.Fatal(err.Error())
	}

	factory := serverpb.NewFactory()
	factory.WithVersion(version).WithEnvironment(serverEnv)
	if *cfgFile != "" {
		factory.WithYamlConfig(*cfgFile)
	}

	srv, err := factory.BuildGrpcServer()
	if err != nil {
		log.Fatal(err.Error())
	}

	signalCh := make(chan os.Signal, 1)
	if err := srv.RunUntilSigterm(signalCh); err != nil {
		log.Fatal(err.Error())
	}
}
