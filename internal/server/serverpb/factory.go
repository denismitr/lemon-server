package serverpb

import (
	"github.com/denismitr/lemon-server/internal/database"
	"github.com/denismitr/lemon-server/internal/server"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

//var ErrNoConfigSource = errors.New("no configuration source specified e.g. yaml config path")

var ErrDisabled = errors.New("grpc server is disabled in configuration")

const DefaultPort = "3099"

type Factory struct {
	yamlConfigPath string
	env            server.Environment
	version        string
}

func NewFactory() *Factory {
	return &Factory{
		env:     server.Dev,
		version: "dev",
	}
}

func (f *Factory) WithYamlConfig(path string) *Factory {
	f.yamlConfigPath = path
	return f
}

func (f *Factory) WithEnvironment(env server.Environment) *Factory {
	f.env = env
	return f
}

func (f *Factory) WithVersion(version string) *Factory {
	f.version = version
	return f
}

func (f *Factory) BuildGrpcServer() (*GrpcServer, error) {
	var slg *zap.SugaredLogger
	if f.env == server.Dev || f.env == server.Test {
		if lg, err := zap.NewDevelopmentConfig().Build(); err != nil {
			return nil, err
		} else {
			slg = lg.Sugar()
		}
	} else {
		if lg, err := zap.NewProductionConfig().Build(); err != nil {
			return nil, err
		} else {
			slg = lg.Sugar()
		}
	}

	var grpcCfg = server.GrpcConfig{
		Enabled:    true,
		Port:       DefaultPort,
		Reflection: f.env == server.Dev,
	}

	if f.yamlConfigPath != "" {
		if cfg, err := server.NewConfigFromYaml(f.yamlConfigPath); err != nil {
			return nil, err
		} else {
			grpcCfg = cfg.Grpc
			grpcCfg.Version = f.version
		}
	}

	if !grpcCfg.Enabled {
		return nil, ErrDisabled
	}

	s := database.NewStore()
	db := database.NewEngine(s)

	grpcHandlers := NewHandlers(slg, db)
	return New(f.env, grpcCfg, slg, grpcHandlers), nil
}
