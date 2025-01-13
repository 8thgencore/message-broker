package app

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/8thgencore/message-broker/internal/broker"
	"github.com/8thgencore/message-broker/internal/config"
	pb "github.com/8thgencore/message-broker/pkg/broker/v1"
)

type App struct {
	cfg     *config.Config
	logger  *slog.Logger
	broker  *broker.Broker
	grpcSrv *grpc.Server
	httpSrv *http.Server
}

func New(cfg *config.Config) *App {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	return &App{
		cfg:    cfg,
		logger: logger,
		broker: broker.NewBroker(cfg.Queues),
	}
}

func (a *App) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Инициализация gRPC сервера
	a.grpcSrv = grpc.NewServer()
	pb.RegisterBrokerServiceServer(a.grpcSrv, newGRPCServer(a.broker, a.logger))

	// Запуск gRPC сервера
	grpcListener, err := net.Listen("tcp", fmt.Sprintf(":%d", a.cfg.Server.GRPCPort))
	if err != nil {
		return fmt.Errorf("failed to listen grpc: %w", err)
	}

	go func() {
		a.logger.Info("starting grpc server", "port", a.cfg.Server.GRPCPort)
		if err := a.grpcSrv.Serve(grpcListener); err != nil {
			a.logger.Error("failed to serve grpc", "error", err)
			cancel()
		}
	}()

	// Инициализация HTTP сервера с gRPC-Gateway
	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	if err := pb.RegisterBrokerServiceHandlerFromEndpoint(
		ctx,
		mux,
		fmt.Sprintf("localhost:%d", a.cfg.Server.GRPCPort),
		opts,
	); err != nil {
		return fmt.Errorf("failed to register gateway: %w", err)
	}

	a.httpSrv = &http.Server{
		Addr:    fmt.Sprintf(":%d", a.cfg.Server.HTTPPort),
		Handler: mux,
	}

	// Запуск HTTP сервера
	go func() {
		a.logger.Info("starting http server", "port", a.cfg.Server.HTTPPort)
		if err := a.httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.logger.Error("failed to serve http", "error", err)
			cancel()
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	select {
	case v := <-quit:
		a.logger.Info("received signal", "signal", v.String())
	case <-ctx.Done():
		a.logger.Info("received context done")
	}

	return a.Stop()
}

func (a *App) Stop() error {
	// Graceful shutdown gRPC server
	a.grpcSrv.GracefulStop()

	// Graceful shutdown HTTP server
	if err := a.httpSrv.Shutdown(context.Background()); err != nil {
		return fmt.Errorf("failed to shutdown http server: %w", err)
	}

	return nil
} 