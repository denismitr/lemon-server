package serverpb

import (
	"github.com/denismitr/lemon-server/internal/database"
	"github.com/denismitr/lemon-server/internal/server"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

//var ErrNoConfigSource = errors.New("no configuration source specified e.g. yaml config path")

var ErrDisabled = errors.New("grpc server is disabled in configuration")

type Factory struct {
	yamlConfigPath string
	dotenvPath     string
	env            server.Environment
	build          string
}

func NewFactory() *Factory {
	return &Factory{
		env:   server.Dev,
		build: "dev",
	}
}

func (f *Factory) WithYamlConfig(path string) *Factory {
	f.yamlConfigPath = path
	return f
}

func (f *Factory) WithDotEnv(path string) *Factory {
	f.dotenvPath = path
	return f
}

func (f *Factory) WithEnvironment(env server.Environment) *Factory {
	f.env = env
	return f
}

func (f *Factory) WithBuildVersion(build string) *Factory {
	f.build = build
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

	cfg, err := server.NewConfig(f.env, f.build, f.yamlConfigPath, f.dotenvPath)
	if err != nil {
		return nil, err
	}

	if !cfg.Grpc.Enabled {
		return nil, ErrDisabled
	}

	s := database.NewStore()
	db := database.NewEngine(s, slg)

	grpcHandlers := NewHandlers(slg, db)
	return New(f.env, cfg, slg, grpcHandlers), nil
}
