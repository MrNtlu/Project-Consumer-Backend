package models

import (
	"context"
	"os"

	openai "github.com/sashabaranov/go-openai"
)

type OpenAI struct {
	Client *openai.Client
}

func CreateOpenAIClient() *OpenAI {
	client := openai.NewClient(os.Getenv("OPENAI_TOKEN"))

	return &OpenAI{
		Client: client,
	}
}

func (oa *OpenAI) GetRecommendation(message string) (openai.ChatCompletionResponse, error) {
	resp, err := oa.Client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role: openai.ChatMessageRoleSystem,
					Content: `You are a recommendation system. Your job is to recommend list of movies to user based on their previously watched/enjoyed movies. Suggest 10 similar movies.

					You have to follow these rules strictly, don't miss anything!
					1- Response template should be, 1. Movie name - Why recommended and very small summary.
					2- Suggest/recommend 10 movies.
					3- Do not ever suggest movies user already rated/watched.
					4- Recommend each movie separately and by their full name. Don't recommend movies in groups or series.
					5- Recommend each movie by their full name. Just the name of the movies in list.
					6- Don't add anything else to response. Don't add text like sure, yes, based on your previously watched etc. `,
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: message,
				},
			},
		},
	)
	if err != nil {
		return openai.ChatCompletionResponse{}, err
	}

	return resp, err
}
