package main

import (
	"github.com/denismitr/lemon-server/internal/database"
	"github.com/denismitr/lemon-server/internal/server/serverpb"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"syscall"
)

var version = "dev"

func main() {
	lg, err := zap.NewDevelopmentConfig().Build()
	if err != nil {
		panic(err)
	}

	lgs := lg.Sugar()

	if err := run(lgs); err != nil {
		lgs.Error(err)
	}

	lgs.Info("GRPC server quit")
}

func run(lg *zap.SugaredLogger) error {
	s := database.NewStore()
	db := database.NewEngine(s)

	grpcHandlers := serverpb.NewHandlers(lg, db)
	grpcServer := serverpb.New(serverpb.Config{
		Port:       "3009",
		Version:    version,
		Reflection: true,
	}, lg, grpcHandlers)

	errCh := make(chan error)
	go func() {
		if err := grpcServer.Start(); err != nil {
			lg.Error(err)
			errCh <- err
		} else {
			close(errCh)
		}
	}()

	signalCh := make(chan os.Signal, 1)
	go func() {
		<-signalCh
		grpcServer.Shutdown()
	}()

	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)

	return <-errCh
}
