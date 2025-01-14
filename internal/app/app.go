package app

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"sync"

	"google.golang.org/grpc"

	"github.com/8thgencore/message-broker/internal/app/provider"
	"github.com/8thgencore/message-broker/internal/config"
)

type App struct {
	cfg             *config.Config
	logger          *slog.Logger
	serviceProvider *provider.ServiceProvider
	grpcServer      *grpc.Server
	httpServer      *http.Server
}

func NewApp(ctx context.Context) (*App, error) {
	a := &App{}
	if err := a.initDeps(ctx); err != nil {
		return nil, err
	}
	return a, nil
}

func (a *App) Run() error {
	wg := sync.WaitGroup{}
	wg.Add(2) // gRPC и HTTP серверы

	go func() {
		defer wg.Done()
		if err := a.runGRPCServer(); err != nil {
			a.logger.Error("failed to run gRPC server", "error", err)
		}
	}()

	go func() {
		defer wg.Done()
		if err := a.runHTTPServer(); err != nil {
			a.logger.Error("failed to run HTTP server", "error", err)
		}
	}()

	wg.Wait()

	return nil
}

func (a *App) runGRPCServer() error {
	a.logger.Info("starting grpc server", "port", a.cfg.Server.GRPCPort)

	lis, err := net.Listen("tcp", a.serviceProvider.Config.GRPCAddress())
	if err != nil {
		return err
	}

	return a.grpcServer.Serve(lis)
}

func (a *App) runHTTPServer() error {
	a.logger.Info("starting http server", "port", a.cfg.Server.HTTPPort)

	return a.httpServer.ListenAndServe()
}
