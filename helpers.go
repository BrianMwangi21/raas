package main

import (
	"context"
	"crypto/rand"
	"errors"
	"math/big"
	"strings"
	"time"

	chroma "github.com/amikos-tech/chroma-go/pkg/api/v2"
	"github.com/go-telegram/bot"
	"github.com/openai/openai-go/v2"
)

func getCollection(ctx context.Context, collectionName string) (chroma.Collection, error) {
	opCtx, cancel := context.WithTimeoutCause(ctx, 10*time.Second, errors.New("ChromaDB getCollection timeout"))
	defer cancel()

	collection, err := chromaClient.GetOrCreateCollection(opCtx, collectionName)
	if err != nil {
		logger.Error("ChromaDB failed to use get collection.", "Error", err)
		return nil, err
	}

	return collection, nil
}

func queryCollection(ctx context.Context, collectionName string, userText string) ([]string, error) {
	collection, err := getCollection(ctx, collectionName)
	if err != nil {
		return nil, err
	}

	opCtx, cancel := context.WithTimeoutCause(ctx, 30*time.Second, errors.New("ChromaDB queryCollection timeout"))
	defer cancel()

	query, err := collection.Query(
		opCtx,
		chroma.WithQueryTexts(userText),
		chroma.WithIncludeQuery(chroma.IncludeDocuments, chroma.IncludeMetadatas),
		chroma.WithNResults(3),
	)
	if err != nil {
		logger.Error("ChromaDB failed to query collection.", "Error", err)
		return nil, err
	}

	docsGroup := query.GetDocumentsGroups()
	if len(docsGroup) == 0 || len(docsGroup[0]) == 0 {
		return nil, nil
	}

	var results []string

	for _, doc := range docsGroup[0] {
		if doc.ContentString() != "" {
			results = append(results, doc.ContentString())
		}
	}

	return results, nil
}

func generateChatResponse(ctx context.Context, details []string, moments []string, userText string) (string, error) {
	logger.Info("Generating chat response.")

	userMsg := &strings.Builder{}
	userMsg.WriteString("User text:\n")
	userMsg.WriteString(userText)
	userMsg.WriteString("\n\nKnown Details:\n")
	userMsg.WriteString(bulletize(details, 20))
	userMsg.WriteString("\nShared Moments:\n")
	userMsg.WriteString(bulletize(moments, 20))

	logger.Info("Created user message.", "User Message to AI", userMsg.String())

	opCtx, cancel := context.WithTimeoutCause(ctx, 1*time.Minute, errors.New("OpenAI generation timeout"))
	defer cancel()

	response, err := openaiClient.Chat.Completions.New(
		opCtx,
		openai.ChatCompletionNewParams{
			Messages: []openai.ChatCompletionMessageParamUnion{
				openai.SystemMessage(systemChatResponsePrompt),
				openai.UserMessage(userMsg.String()),
			},
			Model:       openai.ChatModelGPT4_1Nano,
			Temperature: openai.Float(0.85),
		},
	)
	if err != nil {
		logger.Error("OpenAI failed generation.", "Error", err)
		return "", err
	}

	if len(response.Choices) == 0 {
		return "Unfortunately, there's no response from OpenAI", nil
	}

	return strings.TrimSpace(response.Choices[0].Message.Content), nil
}

func getRandomCollectionEntry(ctx context.Context, collectionName string) (string, error) {
	collection, err := getCollection(ctx, collectionName)
	if err != nil {
		return "", err
	}

	countCtx, countCancel := context.WithTimeoutCause(ctx, 10*time.Second, errors.New("ChromaDB getCount timeout"))
	defer countCancel()

	total, err := collection.Count(countCtx)
	if err != nil {
		logger.Error("ChromaDB failed to get count.", "Error", err)
		return "", err
	}

	nBig, _ := rand.Int(rand.Reader, big.NewInt(int64(total)))
	offset := int(nBig.Int64())

	opCtx, cancel := context.WithTimeoutCause(ctx, 1*time.Minute, errors.New("ChromaDB get timeout"))
	defer cancel()

	get, err := collection.Get(
		opCtx,
		chroma.WithOffsetGet(offset),
		chroma.WithIncludeGet(chroma.IncludeDocuments, chroma.IncludeMetadatas),
		chroma.WithLimitGet(1),
	)
	if err != nil {
		logger.Error("ChromaDB failed to get entry at offset.", "Error", err)
		return "", err
	}

	docsGroup := get.GetDocuments()
	return docsGroup[0].ContentString(), nil
}

func generateRandomNuggetResponse(ctx context.Context, detail string, moment string) (string, error) {
	logger.Info("Generating chat response.")

	userMsg := &strings.Builder{}
	userMsg.WriteString("\n\nKnown Detail:\n")
	userMsg.WriteString(detail)
	userMsg.WriteString("\nShared Moment:\n")
	userMsg.WriteString(moment)

	logger.Info("Created user message.", "User Message to AI", userMsg.String())

	opCtx, cancel := context.WithTimeoutCause(ctx, 1*time.Minute, errors.New("OpenAI generation timeout"))
	defer cancel()

	response, err := openaiClient.Chat.Completions.New(
		opCtx,
		openai.ChatCompletionNewParams{
			Messages: []openai.ChatCompletionMessageParamUnion{
				openai.SystemMessage(systemRandomNuggetPrompt),
				openai.UserMessage(userMsg.String()),
			},
			Model:       openai.ChatModelGPT4_1Nano,
			Temperature: openai.Float(0.85),
		},
	)
	if err != nil {
		logger.Error("OpenAI failed generation.", "Error", err)
		return "", err
	}

	if len(response.Choices) == 0 {
		return "Unfortunately, there's no response from OpenAI", nil
	}

	return strings.TrimSpace(response.Choices[0].Message.Content), nil
}

func bulletize(items []string, max int) string {
	if len(items) == 0 {
		return "• (none)\n"
	}
	if max > 0 && len(items) > max {
		items = items[:max]
	}
	var b strings.Builder
	for _, it := range items {
		line := strings.TrimSpace(it)
		if line == "" {
			continue
		}
		if len(line) > 300 {
			line = line[:300] + "…"
		}
		b.WriteString("• ")
		b.WriteString(line)
		b.WriteString("\n")
	}
	out := b.String()
	if out == "" {
		return "• (none)\n"
	}
	return out
}

func startThreeHourScheduler(ctx context.Context, b *bot.Bot) {
	go func() {
		for {
			now := time.Now().UTC()
			hours := now.Hour()
			nextHours := ((hours / 3) + 1) * 3
			var next time.Time
			if nextHours >= 24 {
				next = now.Truncate(24*time.Hour).AddDate(0, 0, 1)
			} else {
				next = now.Truncate(24 * time.Hour).Add(time.Duration(nextHours) * time.Hour)
			}
			wait := time.Until(next)

			logger.Info("Next random nugget (UTC).", "at", next.Format(time.RFC3339), "in", wait.String())

			timer := time.NewTimer(wait)
			select {
			case <-ctx.Done():
				timer.Stop()
				logger.Info("Three-hour scheduler stopped.")
				return
			case <-timer.C:
				runCtx, cancel := context.WithTimeout(ctx, 45*time.Second)
				randomNuggetHandler(runCtx, b)
				cancel()
			}
		}
	}()
}
