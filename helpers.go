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

func queryCollection(ctx context.Context, collectionName string, userText string) (string, error) {
	collection, err := getCollection(ctx, collectionName)
	if err != nil {
		return "", err
	}

	opCtx, cancel := context.WithTimeoutCause(ctx, 30*time.Second, errors.New("ChromaDB queryCollection timeout"))
	defer cancel()

	query, err := collection.Query(
		opCtx,
		chroma.WithQueryTexts(userText),
		chroma.WithIncludeQuery(chroma.IncludeDocuments, chroma.IncludeMetadatas),
	)
	if err != nil {
		logger.Error("ChromaDB failed to query collection.", "Error", err)
		return "", err
	}

	docsGroup := query.GetDocumentsGroups()
	if len(docsGroup) == 0 || len(docsGroup[0]) == 0 {
		return "", nil
	}

	return docsGroup[0][0].ContentString(), nil
}
