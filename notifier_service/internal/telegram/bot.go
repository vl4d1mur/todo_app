package telegram

import (
	"context"
	"strings"
	"time"

	authgrpc "notifier_service/internal/grpc"
	"notifier_service/pkg/log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	api        *tgbotapi.BotAPI
	authClient *authgrpc.AuthClient
	stop       chan struct{}
}

func NewBot(token string, authClient *authgrpc.AuthClient) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}

	log.Logger.Info().Str("username", api.Self.UserName).Msg("Telegram bot authorized")

	return &Bot{
		api:        api,
		authClient: authClient,
		stop:       make(chan struct{}),
	}, nil
}

func (b *Bot) Start() {
	go func() {
		u := tgbotapi.NewUpdate(0)
		u.Timeout = 30
		updates := b.api.GetUpdatesChan(u)

		log.Logger.Info().Msg("Telegram bot started polling")

		for {
			select {
			case <-b.stop:
				b.api.StopReceivingUpdates()
				log.Logger.Info().Msg("Telegram bot stopped")
				return
			case update := <-updates:
				if update.Message == nil {
					continue
				}
				b.handleMessage(update.Message)
			}
		}
	}()
}

func (b *Bot) Stop() {
	close(b.stop)
}

func (b *Bot) SendMessage(chatID int64, text string) error {
	msg := tgbotapi.NewMessage(chatID, text)
	_, err := b.api.Send(msg)
	return err
}

func (b *Bot) handleMessage(msg *tgbotapi.Message) {
	if !msg.IsCommand() {
		b.reply(msg.Chat.ID, "Use /start <code> to link your account.")
		return
	}

	switch msg.Command() {
	case "start":
		b.handleStart(msg)
	default:
		b.reply(msg.Chat.ID, "Unknown command. Use /start <code> to link your account.")
	}
}

func (b *Bot) handleStart(msg *tgbotapi.Message) {
	code := strings.TrimSpace(msg.CommandArguments())

	if code == "" {
		b.reply(msg.Chat.ID, "Welcome! To link your account, get a code in the app and send: /start <code>")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	success, err := b.authClient.ActivateTelegram(ctx, code, msg.Chat.ID)
	if err != nil {
		log.Logger.Error().Err(err).Msg("Failed to activate telegram")
		b.reply(msg.Chat.ID, "Failed to activate. Please try again.")
		return
	}

	if !success {
		b.reply(msg.Chat.ID, "Invalid or expired code. Please get a new one in the app.")
		return
	}

	b.reply(msg.Chat.ID, "Your account is linked! You will now receive notifications here.")
}

func (b *Bot) reply(chatID int64, text string) {
	if err := b.SendMessage(chatID, text); err != nil {
		log.Logger.Error().Err(err).Int64("chat_id", chatID).Msg("Failed to send message")
	}
}
