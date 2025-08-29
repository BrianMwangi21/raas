package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"strconv"
	"time"

	chroma "github.com/amikos-tech/chroma-go/pkg/api/v2"
	"github.com/charmbracelet/log"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/joho/godotenv"
	"github.com/openai/openai-go/v2"
	"github.com/openai/openai-go/v2/option"
)

func init() {
	logger = log.NewWithOptions(os.Stderr, log.Options{
		ReportTimestamp: true,
		TimeFormat:      time.RFC3339,
		Prefix:          "[RAAS]",
	})
	logger.Info("Initializing application.")

	err := godotenv.Load()
	if err != nil {
		logger.Error("Error loading .env file.")
		os.Exit(1)
	}

	TELEGRAM_BOT_TOKEN = os.Getenv("TELEGRAM_BOT_TOKEN")
	if TELEGRAM_BOT_TOKEN == "" {
		logger.Error("Telegram Bot Token is not set.")
		os.Exit(1)
	}

	chatIDStr := os.Getenv("TELEGRAM_CHAT_ID")
	if chatIDStr == "" {
		logger.Error("Telegram ChatID is not set.")
		os.Exit(1)
	}

	TELEGRAM_CHAT_ID, err = strconv.ParseInt(chatIDStr, 10, 64)
	if err != nil {
		logger.Error("Invalid Telegram ChatID.", "Error", err)
		os.Exit(1)
	}

	OPENAI_API_KEY = os.Getenv("OPENAI_API_KEY")
	if OPENAI_API_KEY == "" {
		logger.Error("OPENAI_API_KEY is not set.")
		os.Exit(1)
	}

	TENANT_NAME = os.Getenv("TENANT_NAME")
	if TENANT_NAME == "" {
		logger.Error("Tenant Name is not set.")
		os.Exit(1)
	}

	DATABASE_NAME = os.Getenv("DATABASE_NAME")
	if DATABASE_NAME == "" {
		logger.Error("Database Name is not set.")
		os.Exit(1)
	}
}

func initializeChromaClient(ctx context.Context) error {
	client, err := chroma.NewHTTPClient()
	if err != nil {
		return err
	}

	opCtx, cancel := context.WithTimeoutCause(ctx, 30*time.Second, errors.New("ChromaDB setup timeout"))
	defer cancel()

	if err := client.PreFlight(opCtx); err != nil {
		_ = client.Close()
		return err
	}
	if err := client.Heartbeat(opCtx); err != nil {
		_ = client.Close()
		return err
	}

	// Set fixed tenant and database
	tenant := chroma.NewTenant(TENANT_NAME)
	if err := client.UseTenant(opCtx, tenant); err != nil {
		_ = client.Close()
		return err
	}

	db := chroma.NewDatabase(DATABASE_NAME, tenant)
	if err := client.UseDatabase(opCtx, db); err != nil {
		_ = client.Close()
		return err
	}

	chromaClient = client
	logger.Info("ChromaDB connected, tenant/db selected, collections ready.")
	return nil
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := initializeChromaClient(ctx); err != nil {
		logger.Error("Failed to connect to ChromaDB.", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := chromaClient.Close(); err != nil {
			logger.Error("Error closing ChromaDB client.", "error", err)
		}
	}()

	openaiClient = openai.NewClient(
		option.WithAPIKey(OPENAI_API_KEY),
	)

	opts := []bot.Option{
		bot.WithDefaultHandler(withChatIDCheck(defaultHandler)),
	}

	b, err := bot.New(TELEGRAM_BOT_TOKEN, opts...)
	if err != nil {
		logger.Error("Error occured when starting bot.", "Error", err)
		os.Exit(1)
	}
	_, err = b.SetMyCommands(ctx, &bot.SetMyCommandsParams{
		Commands: []models.BotCommand{
			{Command: "add_detail", Description: "Add detail about person"},
			{Command: "add_moment", Description: "Add moment with person"},
		},
	})
	if err != nil {
		logger.Error("Error occured when setting commands.", "Error", err)
		os.Exit(1)
	}

	b.RegisterHandler(bot.HandlerTypeMessageText, "add_detail", bot.MatchTypeCommandStartOnly, withChatIDCheck(addDetailHandler))
	b.RegisterHandler(bot.HandlerTypeMessageText, "add_moment", bot.MatchTypeCommandStartOnly, withChatIDCheck(addMomentHandler))

	startThreeHourScheduler(ctx, b)

	b.Start(ctx)
}
