package main

import (
	"context"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func defaultHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	chatID := update.Message.Chat.ID

	logger.Info("Default handler called.", "chatID", chatID)

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    chatID,
		Text:      "Welcome to RAAS",
		ParseMode: models.ParseModeMarkdown,
	})
}
