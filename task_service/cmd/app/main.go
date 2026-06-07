package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"task_service/internal/config"
	"task_service/internal/db/mongo"
	"task_service/internal/db/postgres"
	"task_service/internal/repository"
	"task_service/internal/service"
	"task_service/internal/handlers"
	"task_service/internal/db/redisConn"
	"task_service/internal/events"
	"task_service/internal/routes"
	"task_service/internal/health"
	"task_service/pkg/log"
	authgrpc "task_service/internal/grpc"
)

func main() {
	log.InitLogger()
    config.LoadConfig()

    redisConn.ConnectRedis()
    postgres.ConnectPostgres()
    mongo.ConnectMongo()
    events.ConnectNATS()

    taskRepo := repository.NewTaskRepository()
    noteRepo := repository.NewNoteRepository()

    taskSvc := service.NewTaskService(taskRepo)
    noteSvc := service.NewNoteService(noteRepo)

    h := handlers.NewHandler(taskSvc, noteSvc)
    
	healthChecker := health.NewChecker(
    postgres.DB,
    redisConn.RedisClient,
    mongo.MongoDB,
    events.NatsConn,
	)
	
	authClient, err := authgrpc.NewAuthClient(config.AuthGrpcAddr)
	if err != nil {
	log.Logger.Fatal().Err(err).Msg("Failed to connect to auth_service")
	}
	defer authClient.Close()

	router := routes.SetupRoutes(h, healthChecker, authClient)
	
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

	postgres.ClosePostgres()
	redisConn.CloseRedis()
	mongo.CloseMongoDB()
	events.CloseNATS()
	log.Logger.Info().Msg("Shutdown")
}
