package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"todo/internal/config"
	"todo/internal/db/mongo"
	"todo/internal/db/postgres"
	"todo/internal/repository"
	"todo/internal/service"
	"todo/internal/handlers"
	"todo/internal/db/redisConn"
	"todo/internal/events"
	"todo/internal/routes"
	"todo/internal/health"
	"todo/pkg/log"
)

func main() {
	log.InitLogger()
    config.LoadConfig()

    redisConn.ConnectRedis()
    postgres.ConnectPostgres()
    mongo.ConnectMongo()
    events.ConnectNATS()

    userRepo := repository.NewUserRepository()
    taskRepo := repository.NewTaskRepository()
    noteRepo := repository.NewNoteRepository()
	sessionRepo := repository.NewSessionRepository()

    authSvc := service.NewAuthService(userRepo, sessionRepo)
    taskSvc := service.NewTaskService(taskRepo)
    noteSvc := service.NewNoteService(noteRepo)

    h := handlers.NewHandler(authSvc, taskSvc, noteSvc)
    
	healthChecker := health.NewChecker(
    postgres.DB,
    redisConn.RedisClient,
    mongo.MongoDB,
    events.NatsConn,
	)
	
	router := routes.SetupRoutes(h, healthChecker)
	
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
