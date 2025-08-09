package main

import (
	"context"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/charmbracelet/log"
	"github.com/go-telegram/bot"
	"github.com/joho/godotenv"
)

var (
	TELEGRAM_BOT_TOKEN string
	TELEGRAM_CHAT_ID   int64
	logger             *log.Logger
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
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	opts := []bot.Option{
		bot.WithDefaultHandler(withChatIDCheck(defaultHandler)),
	}

	b, err := bot.New(TELEGRAM_BOT_TOKEN, opts...)
	if err != nil {
		logger.Error("Error occured when starting bot.", "Error", err)
		os.Exit(1)
	}

	b.Start(ctx)
}
