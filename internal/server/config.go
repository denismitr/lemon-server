package server

import (
	"github.com/ardanlabs/conf/v2"
	"github.com/ardanlabs/conf/v2/yaml"
	"github.com/joho/godotenv"
	"github.com/pkg/errors"
	"os"
	"path/filepath"
)

var ErrHelpRequested = errors.New("help requested")

type Environment string

const (
	Dev  Environment = "dev"
	Prod Environment = "prod"
	Test Environment = "test"
)

func CreateEnvironment(env string) (Environment, error) {
	switch env {
	case "dev":
		return Dev, nil
	case "prod":
		return Prod, nil
	case "test":
		return Test, nil
	default:
		return "", errors.Errorf("invalid environment %s", env)
	}
}

type Config struct {
	conf.Version
	Environment Environment `conf:"-" yaml:"-"`
	Grpc        GrpcConfig  `yaml:"grpc"`
}

type GrpcConfig struct {
	Enabled    bool   `conf:"default:true,env:GRPC_ENABLED" yaml:"enabled"`
	Port       int    `conf:"default:3099,env:GRPC_PORT" yaml:"port"`
	Reflection bool   `conf:"default:true,env:GRPC_REFLECTION_ENABLED" yaml:"reflection_enabled"`
	Version    string `conf:"default:1,env:GRPC_VERSION" yaml:"version"`
}

// todo: add validation for GrpcConfig

func NewConfig(env Environment, buildVersion string, yamlPath, dotenvPath string) (*Config, error) {
	if dotenvPath != "" {
		if err := godotenv.Load(); err != nil {
			return nil, errors.Wrapf(err, "could not load .env file %s", dotenvPath)
		}
	}

	cfg := Config{
		Version: conf.Version{
			Build: buildVersion,
		},
		Environment: env,
	}

	var parsers []conf.Parsers
	if yamlPath != "" {
		yamlData, err := readYamlFile(yamlPath)
		if err != nil {
			return nil, err
		}

		parsers = append(parsers, yaml.WithData(yamlData))
	}

	_, err := conf.Parse("", &cfg, parsers...)
	if err != nil {
		return nil, errors.Wrap(err, "could not process config")
	}

	return &cfg, nil
}

func readYamlFile(path string) ([]byte, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, errors.Wrapf(err, "could not get absolute path from %s", path)
	}

	data, err := os.ReadFile(absPath)
	if err != nil {
		return nil, errors.Wrapf(err, "could not read config file %s", absPath)
	}

	return data, nil
}
