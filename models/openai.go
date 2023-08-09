package models

import (
	"context"
	"os"
	"strings"

	openai "github.com/sashabaranov/go-openai"
	"github.com/sirupsen/logrus"
)

const prompt = `You are a recommendation system. Your job is to recommend list of movies to user based on their previously watched/enjoyed movies. Suggest 10 similar movies.

You have to follow these rules strictly, don't miss anything!
1- Response template should be, Movie name.
2- Suggest/recommend 10 movies.
3- Do not ever suggest movies user already rated/watched.
4- Recommend each movie separately and by their full name. Don't recommend movies in groups or series.
5- Recommend each movie by their full name. Just the name of the movies in list.
6- Don't add anything else to response. Don't add text like sure, yes, based on your previously watched etc. `

const promptAlt = `You are a recommendation system. Recommend 10 movies similar to user's watched/enjoyed. Follow rules strictly:

1- Response template, Movie name
2- Suggest 10 movies.
3- Avoid seen movies.
4- Recommend individually, no groups.
5- Use movie names only.
6- No extra text. `

type OpenAI struct {
	Client *openai.Client
}

func CreateOpenAIClient() *OpenAI {
	client := openai.NewClient(os.Getenv("OPENAI_TOKEN"))

	return &OpenAI{
		Client: client,
	}
}

type OpenAIResponse struct {
	TotalUsedToken int      `bson:"total_used_token" json:"total_used_token"`
	Recommendation []string `bson:"recommendation" json:"recommendation"`
}

func (oa *OpenAI) GetRecommendation(message string) (OpenAIResponse, error) {
	resp, err := oa.Client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: promptAlt,
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: message,
				},
			},
		},
	)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"input": message,
		}).Error("Error: ", err)

		return OpenAIResponse{}, err
	}

	var (
		model  string
		object string

		totalTokens        int
		completionTokens   int
		promptTokens       int
		recommendationList []string
	)

	model = resp.Model
	object = resp.Object
	totalTokens = resp.Usage.TotalTokens
	completionTokens = resp.Usage.CompletionTokens
	promptTokens = resp.Usage.PromptTokens
	recommendationList = strings.Split(resp.Choices[0].Message.Content, "\n")

	logrus.WithFields(logrus.Fields{
		"input":             message,
		"model":             model,
		"object":            object,
		"total_tokens":      totalTokens,
		"completion_tokens": completionTokens,
		"prompt_tokens":     promptTokens,
	}).Info("choices", resp.Choices)

	openAIResponse := OpenAIResponse{
		TotalUsedToken: totalTokens,
		Recommendation: recommendationList,
	}

	return openAIResponse, err
}
