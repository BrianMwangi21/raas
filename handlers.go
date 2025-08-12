package main

import (
	"context"
	"errors"
	"strings"
	"time"

	chroma "github.com/amikos-tech/chroma-go/pkg/api/v2"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func buildMessageParams(chatID int64, message string) *bot.SendMessageParams {
	return &bot.SendMessageParams{
		ChatID: chatID,
		Text:   message,
	}
}

func sendMessageToUser(ctx context.Context, b *bot.Bot, message string) {
	sendCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if _, err := b.SendMessage(sendCtx, buildMessageParams(TELEGRAM_CHAT_ID, message)); err != nil {
		logger.Error("Send Message failed.", "Error", err, "Message", message)
	}
}

func withChatIDCheck(next bot.HandlerFunc) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		if update.Message == nil {
			return // Ignore non-message updates
		}

		chatID := update.Message.Chat.ID

		if chatID != TELEGRAM_CHAT_ID {
			// We use chatID here cause it could be different from the one set in the env
			// And we want to shun them off. Don't see how this could ever happen though
			// But it's a good gate!
			b.SendMessage(ctx, buildMessageParams(chatID, "Sorry, we can't dance together 💃🕺"))
			logger.Error("Blocked message from unauthorized chat.", "ChatID", chatID)
			return
		}

		// Authorized, pass to the real handler
		next(ctx, b, update)
	}
}

func defaultHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	logger.Info("Default handler called.")
	userText := strings.TrimSpace(update.Message.Text)

	if userText == "" {
		sendMessageToUser(ctx, b, "Message is empty. Please try again.")
		return
	}

	sendMessageToUser(ctx, b, "Querying chromadb.")

	detailsResult, err := queryCollection(ctx, "details", userText)
	if err != nil {
		sendMessageToUser(ctx, b, "Sorry, failed to query the details collection.")
	}
	logger.Info("Queried details collection.", "Result", detailsResult, "ResultCount", len(detailsResult))

	momentsResult, err := queryCollection(ctx, "moments", userText)
	if err != nil {
		sendMessageToUser(ctx, b, "Sorry, failed to query the moments collection.")
	}
	logger.Info("Queried moments collection.", "Result", momentsResult, "ResultCount", len(momentsResult))

	response, err := generateChatResponse(ctx, detailsResult, momentsResult, userText)
	if err != nil {
		sendMessageToUser(ctx, b, "Sorry, failed to generate chat response.")
		return
	}
	logger.Info("Generated Chat Response.", "Response", response)

	sendMessageToUser(ctx, b, response)
}

func addDetailHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	logger.Info("Add Detail handler called.")
	userText := strings.TrimSpace(strings.TrimPrefix(update.Message.Text, "/add_detail"))

	if userText == "" {
		sendMessageToUser(ctx, b, "Message is empty. Please try again.")
		return
	}

	sendMessageToUser(ctx, b, "Adding new detail to chromadb.")

	collection, err := getCollection(ctx, "details")
	if err != nil {
		sendMessageToUser(ctx, b, "Sorry, failed to access the details collection.")
		return
	}

	opCtx, cancel := context.WithTimeoutCause(ctx, 20*time.Second, errors.New("ChromaDB addToCollection timeout"))
	defer cancel()

	if err = collection.Add(
		opCtx,
		chroma.WithIDGenerator(chroma.NewULIDGenerator()),
		chroma.WithTexts(userText),
		chroma.WithMetadatas(
			chroma.NewDocumentMetadata(chroma.NewStringAttribute("tag", "detail")),
		),
	); err != nil {
		logger.Error("ChromaDB failed to add to collection.", "Error", err)
		sendMessageToUser(ctx, b, "Sorry, failed to add detail to collection.")
		return
	}

	sendMessageToUser(ctx, b, "Amazing news! Detail has been memorized forever!")
	logger.Info("Add Detail handler finished successfully.", "Detail added", userText)
}

func addMomentHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	logger.Info("Add Moment handler called.")
	userText := strings.TrimSpace(strings.TrimPrefix(update.Message.Text, "/add_moment"))

	if userText == "" {
		sendMessageToUser(ctx, b, "Message is empty. Please try again.")
		return
	}

	sendMessageToUser(ctx, b, "Adding new moment to chromadb.")

	collection, err := getCollection(ctx, "moments")
	if err != nil {
		sendMessageToUser(ctx, b, "Sorry, failed to access the moments collection.")
		return
	}

	opCtx, cancel := context.WithTimeoutCause(ctx, 20*time.Second, errors.New("ChromaDB addToCollection timeout"))
	defer cancel()

	if err = collection.Add(
		opCtx,
		chroma.WithIDGenerator(chroma.NewULIDGenerator()),
		chroma.WithTexts(userText),
		chroma.WithMetadatas(
			chroma.NewDocumentMetadata(chroma.NewStringAttribute("tag", "moment")),
		),
	); err != nil {
		logger.Error("ChromaDB failed to add to collection.", "Error", err)
		sendMessageToUser(ctx, b, "Sorry, failed to add moment to collection.")
		return
	}

	sendMessageToUser(ctx, b, "Amazing news! Moment has been memorized forever!")
	logger.Info("Add Moment handler finished successfully.", "Moment added", userText)
}

func randomNuggetHandler(ctx context.Context, b *bot.Bot) {
	logger.Info("Random Nugget Handler called.")

	detailsResult, err := getRandomCollectionEntry(ctx, "details")
	if err != nil {
		sendMessageToUser(ctx, b, "Sorry, failed to query the details collection.")
	}
	logger.Info("Queried random details collection.", "Result", detailsResult)

	momentsResult, err := getRandomCollectionEntry(ctx, "moments")
	if err != nil {
		sendMessageToUser(ctx, b, "Sorry, failed to query the moments collection.")
	}
	logger.Info("Queried random moments collection.", "Result", momentsResult)

	response, err := generateRandomNuggetResponse(ctx, detailsResult, momentsResult)
	if err != nil {
		sendMessageToUser(ctx, b, "Sorry, failed to generate chat response.")
		return
	}
	logger.Info("Generated Chat Response.", "Response", response)

	sendMessageToUser(ctx, b, response)
}
