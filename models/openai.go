package models

import (
	"app/responses"
	"context"
	"os"
	"strings"

	openai "github.com/sashabaranov/go-openai"
	"github.com/sirupsen/logrus"
)

const prompt = `You are a recommendation system. Recommend 10 movies similar to user's watched/enjoyed. Follow rules strictly:

1- Response template, Movie name\n
2- Suggest 10 movies.
3- Avoid seen movies.
4- Recommend individually, no groups.
5- Use movie names only.
6- No extra text.

List of movies user rated,
`

const message = "Recommend movies."

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

type OpenAIMovieResponse struct {
	OpenAIResponse OpenAIResponse    `bson:"openai_response" json:"openai_response"`
	Movies         []responses.Movie `bson:"movies" json:"movies"`
}

func (oa *OpenAI) GetRecommendation(watchList string) (OpenAIResponse, error) {
	resp, err := oa.Client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: prompt + watchList,
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
			"watch_list": watchList,
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
