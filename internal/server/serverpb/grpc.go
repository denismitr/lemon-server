package serverpb

import (
	"context"
	"github.com/denismitr/lemon-server/pkg/command"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
	"os"
	"syscall"
)

type Config struct {
	Port       string
	Reflection bool
	Version    string
}

type GrpcServer struct {
	cfg      Config
	receiver *GrpcHandlers
	lg       *zap.SugaredLogger
	shutdown chan os.Signal
}

func New(cfg Config, lg *zap.SugaredLogger, receiver *GrpcHandlers) *GrpcServer {
	return &GrpcServer{
		cfg:      cfg,
		lg:       lg,
		receiver: receiver,
		shutdown: make(chan os.Signal),
	}
}

func (srv *GrpcServer) Start() error {
	i := func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		srv.lg.Debugf("got request for %s", info.FullMethod)
		return handler(ctx, req)
	}

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
		srv.lg.Debugf("Starting LemonDB GRPC server version %s on port :%s", srv.cfg.Version, srv.cfg.Port)
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
