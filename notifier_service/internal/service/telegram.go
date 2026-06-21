package service

import (
	"fmt"

	"notifier_service/pkg/log"
)

type TelegramSender interface {
	SendMessage(chatID int64, text string) error
}

type TelegramChannel struct {
	bot TelegramSender
}

func NewTelegramChannel(bot TelegramSender) *TelegramChannel {
	return &TelegramChannel{bot: bot}
}

func (t *TelegramChannel) Send(chatID int64, subject, body string) error {
	text := fmt.Sprintf("*%s*\n\n%s", subject, body)
	if err := t.bot.SendMessage(chatID, text); err != nil {
		log.Logger.Error().Err(err).Int64("chat_id", chatID).Msg("Failed to send telegram message")
		return err
	}
	log.Logger.Info().Int64("chat_id", chatID).Str("subject", subject).Msg("Telegram message sent")
	return nil
}
