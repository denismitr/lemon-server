package serverpb

import (
	"context"
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
	cfg      server.GrpcConfig
	receiver *GrpcHandlers
	lg       *zap.SugaredLogger
	shutdown chan os.Signal
}

func New(
	env server.Environment,
	cfg server.GrpcConfig,
	lg *zap.SugaredLogger,
	receiver *GrpcHandlers,
) *GrpcServer {
	return &GrpcServer{
		env:      env,
		cfg:      cfg,
		lg:       lg,
		receiver: receiver,
		shutdown: make(chan os.Signal),
	}
}

func (srv *GrpcServer) RunUntilSigterm(signalCh chan os.Signal) error {
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

	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)

	return <-errCh
}

func (srv *GrpcServer) Start() error {
	i := createRequestLoggerInterceptor(srv.lg)

	grpcSrv := grpc.NewServer(grpc.UnaryInterceptor(i))
	command.RegisterReceiverServer(grpcSrv, srv.receiver)

	if srv.cfg.Reflection {
		reflection.Register(grpcSrv)
	}

	listener, err := net.Listen("tcp", ":"+srv.cfg.Port)
	if err != nil {
		return err
	}

	fatalErrCh := make(chan error)

	go func() {
		srv.lg.Debugf(
			"Starting LemonDB GRPC server version '%s' on port :%s in '%s' environment",
			srv.cfg.Version, srv.cfg.Port, srv.env,
		)

		if err := grpcSrv.Serve(listener); err != nil {
			fatalErrCh <- err
		}

		srv.lg.Debugf("Stopping LemonDB GRPC server on port :%s", srv.cfg.Port)
	}()

	for {
		select {
		case <-srv.shutdown:
			grpcSrv.Stop()
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
	srv.shutdown <- syscall.SIGTERM
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
