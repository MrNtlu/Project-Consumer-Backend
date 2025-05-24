package db

import (
	"context"
	"log"
	"os"

	"github.com/pinecone-io/go-pinecone/v3/pinecone"
)

func InitPinecone() (pinecone.IndexConnection, pinecone.Client, error) {
	apiKey := os.Getenv("PINECONE_API_KEY")
	indexName := os.Getenv("PINECONE_INDEX")

	clientParams := pinecone.NewClientParams{
		ApiKey: apiKey,
	}
	pc, err := pinecone.NewClient(clientParams)
	if err != nil {
		return pinecone.IndexConnection{}, pinecone.Client{}, err
	}

	idx, err := pc.DescribeIndex(context.Background(), indexName)
	if err != nil {
		return pinecone.IndexConnection{}, pinecone.Client{}, err
	}

	idxConn, err := pc.Index(pinecone.NewIndexConnParams{
		Host: idx.Host,
	})
	if err != nil {
		return pinecone.IndexConnection{}, pinecone.Client{}, err
	}
	log.Println("✅ Pinecone client ready")
	return *idxConn, *pc, nil
}
