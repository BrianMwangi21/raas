package main

import (
	"context"
	"time"

	chroma "github.com/amikos-tech/chroma-go/pkg/api/v2"
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
	logger.Info("Default handler called.", "ChatID", chatID)

	if _, err := b.SendMessage(ctx, buildMessageParams(chatID, "Welcome to RAAS")); err != nil {
		logger.Error("Send Message failed.", "ChatID", chatID, "Error", err)
	}
}

func addDetailHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	chatID := update.Message.Chat.ID
	userText := update.Message.Text

	logger.Info("Add Detail handler called.", "ChatID", chatID)

	if _, err := b.SendMessage(ctx, buildMessageParams(chatID, "Adding new detail to chromadb")); err != nil {
		logger.Error("Send Message failed.", "ChatID", chatID, "Error", err)
	}

	opCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Get or create the 'details' collection
	col, err := chromaClient.GetOrCreateCollection(opCtx, "details")
	if err != nil {
		logger.Error("GetOrCreateCollection failed", "error", err)
		b.SendMessage(ctx, buildMessageParams(chatID, "Sorry, failed to access the details collection"))
	}

	err = col.Add(
		opCtx,
		chroma.WithIDGenerator(chroma.NewULIDGenerator()),
		chroma.WithTexts(userText),
		chroma.WithMetadatas(
			chroma.NewDocumentMetadata(
				chroma.NewStringAttribute("str", "details"),
			)),
	)
	if err != nil {
		logger.Error("Add to collection failed", "error", err)
		b.SendMessage(ctx, buildMessageParams(chatID, "Sorry, failed to add to the details collection"))
	}
}

func addMomentHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	chatID := update.Message.Chat.ID
	logger.Info("Add Moment handler called.", "ChatID", chatID)

	if _, err := b.SendMessage(ctx, buildMessageParams(chatID, "Add Moment Handler Yet To Be Implemented")); err != nil {
		logger.Error("Send Message failed.", "ChatID", chatID, "Error", err)
	}
}
