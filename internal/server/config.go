package server

import (
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
)

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
	Grpc GrpcConfig `yaml:"grpc"`
}

type GrpcConfig struct {
	Enabled    bool   `yaml:"enabled"`
	Port       string `yaml:"port"`
	Reflection bool   `yaml:"reflection"`
	Version    string `yaml:"-"`
}

// todo: add validation for GrpcConfig

func NewConfigFromYaml(path string) (*Config, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, errors.Wrapf(err, "could not get absolute path from %s", path)
	}

	data, err := os.ReadFile(absPath)
	if err != nil {
		return nil, errors.Wrapf(err, "could not read config file %s", absPath)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, errors.Wrapf(err, "could not parse yaml in %s", absPath)
	}

	return &cfg, nil
}
