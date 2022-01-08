package main

import (
	"flag"
	"github.com/denismitr/lemon-server/internal/server"
	"github.com/denismitr/lemon-server/internal/server/serverpb"
	"log"
)

var build = "dev"

func main() {
	var yamlFile = flag.String("yaml-config", "", "Path to config yaml file")
	var env = flag.String("environment", "dev", "Environment to run server in. Supported values (dev, prod, test)")
	var dotenvFile = flag.String("dotenv", "", "Path to .env file")

	flag.Parse()

	serverEnv, err := server.CreateEnvironment(*env)
	if err != nil {
		log.Fatal(err.Error())
	}

	factory := serverpb.NewFactory()
	factory.WithBuildVersion(build).WithEnvironment(serverEnv)

	if *yamlFile != "" {
		factory.WithYamlConfig(*yamlFile)
	} else if *dotenvFile != "" {
		factory.WithDotEnv(*dotenvFile)
	} else {
		log.Fatal("Configuration must be provided via --yaml-config or --dotenv")
	}

	srv, err := factory.BuildGrpcServer()
	if err != nil {
		log.Fatal(err.Error())
	}

	if err := srv.RunUntilTerminated(); err != nil {
		log.Fatal(err.Error())
	}
}
