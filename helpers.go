package main

import (
	"context"
	"errors"
	"strings"
	"time"

	chroma "github.com/amikos-tech/chroma-go/pkg/api/v2"
	"github.com/openai/openai-go"
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
		chroma.WithNResults(2),
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

func generateChatResponse(ctx context.Context, detailsResult []string, momentsResult []string, userText string) (string, error) {
	logger.Info("Generating chat response.")

	userMsg := &strings.Builder{}
	userMsg.WriteString("User text:\n")
	userMsg.WriteString(userText)
	userMsg.WriteString("\n\nKnown Details:\n")
	userMsg.WriteString(bulletize(detailsResult, 20))
	userMsg.WriteString("\nShared Moments:\n")
	userMsg.WriteString(bulletize(momentsResult, 20))

	logger.Info("Created user message.", "User Message to AI", userMsg.String())

	opCtx, cancel := context.WithTimeoutCause(ctx, 30*time.Second, errors.New("OpenAI generation timeout"))
	defer cancel()

	response, err := openaiClient.Chat.Completions.New(
		opCtx,
		openai.ChatCompletionNewParams{
			Messages: []openai.ChatCompletionMessageParamUnion{
				openai.SystemMessage(systemPrompt),
				openai.UserMessage(userMsg.String()),
			},
			Temperature: openai.Float(0.7),
			Model:       openai.ChatModelGPT4_1Nano,
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
