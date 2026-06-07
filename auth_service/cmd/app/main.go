package main

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"auth_service/internal/config"
	"auth_service/internal/db/postgres"
	"auth_service/internal/db/redisConn"
	authgrpc "auth_service/internal/grpc"
	"auth_service/internal/handlers"
	"auth_service/internal/health"
	"auth_service/internal/repository"
	"auth_service/internal/routes"
	"auth_service/internal/service"
	"auth_service/pkg/log"
	"auth_service/pkg/pb"

	"google.golang.org/grpc"
)

func main() {
	log.InitLogger()
    config.LoadConfig()

    redisConn.ConnectRedis()
    postgres.ConnectPostgres()

    userRepo := repository.NewUserRepository()
	sessionRepo := repository.NewSessionRepository()

    authSvc := service.NewAuthService(userRepo, sessionRepo)

    h := handlers.NewHandler(authSvc)
    
	healthChecker := health.NewChecker(
    postgres.DB,
    redisConn.RedisClient,

	)
	
	router := routes.SetupRoutes(h, healthChecker)
	
	grpcServer := grpc.NewServer()
	pb.RegisterAuthServiceServer(grpcServer, authgrpc.NewAuthServer())

	go func() {
		lis, err := net.Listen("tcp", config.GrpcPort)
		if err != nil {
			log.Logger.Fatal().Err(err).Msg("gRPC listen error")
		}
		log.Logger.Info().Str("port", config.GrpcPort).Msg("gRPC server starting")
		if err := grpcServer.Serve(lis); err != nil {
			log.Logger.Fatal().Err(err).Msg("gRPC server error")
		}
	} ()

	log.Logger.Info().Msg("Starting server")
	srv := &http.Server{
		Addr:    config.ServerPort,
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Logger.Fatal().Err(err).Msg("Sever starting error:")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Logger.Info().Msg("Shutting down")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	go func() {
		force := make(chan os.Signal, 1)
		signal.Notify(force, syscall.SIGINT, syscall.SIGTERM)
		<-force
		log.Logger.Warn().Msg("Forced shutdown")
		os.Exit(1)
	}()

	if err := srv.Shutdown(ctx); err != nil {
		log.Logger.Error().Err(err).Msg("HTTP server shutdown error:")
	}
	log.Logger.Info().Msg("HTTP server stoped")

	grpcServer.GracefulStop()
	log.Logger.Info().Msg("gRPC server stopped")

	postgres.ClosePostgres()
	redisConn.CloseRedis()
	log.Logger.Info().Msg("Shutdown")
}
