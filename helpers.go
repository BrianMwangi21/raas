package main

import (
	"context"
	"errors"
	"time"

	chroma "github.com/amikos-tech/chroma-go/pkg/api/v2"
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
