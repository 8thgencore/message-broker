package app

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/rs/cors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"

	"github.com/8thgencore/message-broker/internal/app/provider"
	"github.com/8thgencore/message-broker/internal/config"
	brokerv1 "github.com/8thgencore/message-broker/pkg/pb/broker/v1"
)

func (a *App) initDeps(ctx context.Context) error {
	inits := []func(context.Context) error{
		a.initConfig,
		a.initLogger,
		a.initServiceProvider,
		a.initGRPCServer,
		a.initHTTPServer,
	}

	for _, f := range inits {
		if err := f(ctx); err != nil {
			return err
		}
	}

	return nil
}

func (a *App) initConfig(_ context.Context) error {
	cfg, err := config.Load("config/config.yaml")
	if err != nil {
		return err
	}
	a.cfg = cfg

	return nil
}

func (a *App) initLogger(_ context.Context) error {
	a.logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	return nil
}

func (a *App) initServiceProvider(_ context.Context) error {
	a.serviceProvider = provider.NewServiceProvider(a.cfg, a.logger)

	return nil
}

func (a *App) initGRPCServer(_ context.Context) error {
	a.logger.Info("[grpc-server] Initializing...")

	a.grpcServer = grpc.NewServer(
		grpc.Creds(insecure.NewCredentials()),
	)

	reflection.Register(a.grpcServer)

	brokerv1.RegisterBrokerServiceServer(a.grpcServer, a.serviceProvider.BrokerDelivery())

	a.logger.Info("[grpc-server] Successfully initialized")

	return nil
}

func (a *App) initHTTPServer(ctx context.Context) error {
	a.logger.Info("[http-server] Initializing...")

	mux := runtime.NewServeMux()

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(16 * 1024 * 1024)), // 16MB
	}

	if err := brokerv1.RegisterBrokerServiceHandlerFromEndpoint(ctx, mux, a.cfg.GRPCAddress(), opts); err != nil {
		return err
	}

	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "DELETE", "PATCH", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Content-Type", "Content-Length", "Authorization"},
		AllowCredentials: true,
	})

	a.httpServer = &http.Server{
		Addr:              a.cfg.HTTPAddress(),
		Handler:           corsMiddleware.Handler(mux),
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
	}

	a.logger.Info("[http-server] Successfully initialized")

	return nil
}
