package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"notifier_service/internal/config"
	"notifier_service/internal/consumer"
	"notifier_service/internal/cron"
	"notifier_service/internal/db/mongo"
	authgrpc "notifier_service/internal/grpc"
	"notifier_service/internal/handlers"
	"notifier_service/internal/health"
	"notifier_service/internal/repository"
	"notifier_service/internal/routes"
	"notifier_service/internal/service"
	"notifier_service/internal/telegram"
	"notifier_service/pkg/log"
)

func main() {
	log.InitLogger()
	config.LoadConfig()

	mongo.ConnectMongo()

	authClient, err := authgrpc.NewAuthClient(config.AuthGrpcAddr)
	if err != nil {
		log.Logger.Fatal().Err(err).Msg("Failed to connect to auth_service")
	}
	defer authClient.Close()

	bot, err := telegram.NewBot(config.TelegramBotToken, authClient)
	if err != nil {
		log.Logger.Fatal().Err(err).Msg("Failed to create Telegram bot")
	}
	bot.Start()
	defer bot.Stop()

	notifRepo := repository.NewNotificationRepository()
	deadlineRepo := repository.NewDeadlineRepository()
	smtpSvc := service.NewSMTPService()
	telegramChannel := service.NewTelegramChannel(bot)
	notifierSvc := service.NewNotifierService(notifRepo, deadlineRepo, authClient, smtpSvc, telegramChannel)

	deadlineChecker := cron.NewDeadlineChecker(deadlineRepo, notifierSvc, 1*time.Minute)
	deadlineChecker.Start()
	defer deadlineChecker.Stop()

	nc, err := consumer.NewConsumer(notifierSvc)
	if err != nil {
		log.Logger.Fatal().Err(err).Msg("Failed to create NATS consumer")
	}

	if err := nc.Start(); err != nil {
		log.Logger.Fatal().Err(err).Msg("Failed to start NATS consumer")
	}

	h := handlers.NewHandler(notifierSvc)
	healthChecker := health.NewChecker(mongo.MongoDB, nc.Conn())

	router := routes.SetupRoutes(h, healthChecker, authClient)

	srv := &http.Server{
		Addr:    config.ServerPort,
		Handler: router,
	}

	go func() {
		log.Logger.Info().Str("port", config.ServerPort).Msg("Notifier server started")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Logger.Fatal().Err(err).Msg("Server error")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Logger.Info().Msg("Shutting down...")

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
		log.Logger.Error().Err(err).Msg("HTTP server shutdown error")
	}
	log.Logger.Info().Msg("HTTP server stopped")

	deadlineChecker.Stop()
	log.Logger.Info().Msg("Deadline checker stopped")
	nc.Close()
	mongo.CloseMongoDB()

	log.Logger.Info().Msg("Shutdown complete")
}
