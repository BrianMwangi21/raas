package main

import (
	"context"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func buildMessageParams(chatID int64, message string) *bot.SendMessageParams {
	return &bot.SendMessageParams{
		ChatID:    chatID,
		Text:      message,
		ParseMode: models.ParseModeMarkdown,
	}
}

func withChatIDCheck(next bot.HandlerFunc) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		if update.Message == nil {
			return // Ignore non-message updates
		}

		chatID := update.Message.Chat.ID

		if chatID != TELEGRAM_CHAT_ID {
			b.SendMessage(ctx, buildMessageParams(chatID, "Sorry, we can't dance together 💃🕺"))
			logger.Warn("Blocked message from unauthorized chat.", "ChatID", chatID)
			return
		}

		// Authorized, pass to the real handler
		next(ctx, b, update)
	}
}

func defaultHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	chatID := update.Message.Chat.ID

	b.SendMessage(ctx, buildMessageParams(chatID, "Welcome to RAAS"))
	logger.Info("Default handler called and responded.", "ChatID", chatID)
}
