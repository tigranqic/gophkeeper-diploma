package main

import (
	"context"
	"fmt"
	"github.com/tigranqic/gophkeeper-diploma/internal/config"
	"github.com/tigranqic/gophkeeper-diploma/internal/proto/gophkeeperv1"
	"github.com/tigranqic/gophkeeper-diploma/internal/server/gservice"
	"github.com/tigranqic/gophkeeper-diploma/internal/server/repository/postgres"
	"github.com/tigranqic/gophkeeper-diploma/pkg/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"net"
	"os"
	"os/signal"
	"syscall"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
)

func main() {
	cfg, err := config.Load(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	log := logger.Init(cfg.LogLevel, cfg.LogFormat)
	log.Info("starting server", zap.String("version", buildVersion), zap.String("build_date", buildDate))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	repo, err := postgres.NewPostgresRepository(ctx, cfg.DatabaseDSN, log)
	if err != nil {
		log.Fatal("failed to connect to repo", zap.Error(err))
	}
	defer repo.Close()

	server := gservice.NewServer(repo, cfg.JWTSecret, log)

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(server.AuthInterceptor),
		grpc.StreamInterceptor(server.StreamAuthInterceptor),
	)
	gophkeeperv1.RegisterAuthServiceServer(grpcServer, server)
	gophkeeperv1.RegisterStorageServiceServer(grpcServer, server)

	listen, err := net.Listen("tcp", cfg.GRPCAddr)
	if err != nil {
		log.Fatal("failed to listen", zap.Error(err))
	}

	go func() {
		log.Info("grpc server listening", zap.String("addr", cfg.GRPCAddr))
		if err := grpcServer.Serve(listen); err != nil {
			log.Error("failed to serve", zap.Error(err))
		}
	}()

	<-ctx.Done()
	log.Info("shutting down")
	grpcServer.GracefulStop()
}
