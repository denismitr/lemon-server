package serverpb

import (
	"context"
	"fmt"
	"github.com/denismitr/lemon-server/internal/server"
	"github.com/denismitr/lemon-server/pkg/command"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
	"os"
	"os/signal"
	"syscall"
)

type GrpcServer struct {
	env      server.Environment
	cfg      *server.Config
	receiver *GrpcHandlers
	lg       *zap.SugaredLogger
	stopCh   chan struct{}
}

func New(
	env server.Environment,
	cfg *server.Config,
	lg *zap.SugaredLogger,
	receiver *GrpcHandlers,
) *GrpcServer {
	return &GrpcServer{
		env:      env,
		cfg:      cfg,
		lg:       lg,
		receiver: receiver,
		stopCh:   make(chan struct{}),
	}
}

func (srv *GrpcServer) RunUntilTerminated() error {
	signalCh := make(chan os.Signal)

	errCh := make(chan error)
	go func() {
		if err := srv.Start(); err != nil {
			srv.lg.Error(err)
			errCh <- err
		} else {
			close(errCh)
		}
	}()

	go func() {
		<-signalCh
		srv.Shutdown()
	}()

	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	return <-errCh
}

func (srv *GrpcServer) Start() error {
	i := createRequestLoggerInterceptor(srv.lg)

	grpcSrv := grpc.NewServer(grpc.UnaryInterceptor(i))
	command.RegisterReceiverServer(grpcSrv, srv.receiver)

	if srv.cfg.Grpc.Reflection {
		reflection.Register(grpcSrv)
	}

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", srv.cfg.Grpc.Port))
	if err != nil {
		return err
	}

	fatalErrCh := make(chan error)

	go func() {
		srv.lg.Debugf(
			"Starting LemonDB GRPC server: build %s, API version '%s', port :%d in '%s' environment",
			srv.cfg.Version.Build, srv.cfg.Grpc.Version, srv.cfg.Grpc.Port, srv.env,
		)

		if err := grpcSrv.Serve(listener); err != nil {
			fatalErrCh <- err
		}

		srv.lg.Debugf("Stopping LemonDB GRPC server on port :%d", srv.cfg.Grpc.Port)
	}()

	for {
		select {
		case <-srv.stopCh:
			grpcSrv.GracefulStop()
			return nil
		case err := <-fatalErrCh:
			if err == nil {
				return nil
			}

			return errors.Wrap(err, "grpc server error")
		}
	}
}

func (srv *GrpcServer) Shutdown() {
	close(srv.stopCh)
}

func createRequestLoggerInterceptor(lg *zap.SugaredLogger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		lg.Debugf("grpc server got request for method %s", info.FullMethod)
		return handler(ctx, req)
	}
}
