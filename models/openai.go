package models

import (
	"app/responses"
	"context"
	"os"
	"strings"

	openai "github.com/sashabaranov/go-openai"
	"github.com/sirupsen/logrus"
)

const prompt = `You are a recommendation system. Recommend movies, tv series, anime and games similar to user's watched/enjoyed. Follow rules strictly!:

1- Use this format: Content Type: Content Name (e.g. Movie: Movie Name)
2- Suggest 3 movies, 3 TV series, 3 anime, and 3 games.
3- Avoid seen/played.
4- Recommend individually, no groups.
5- Use names only.
6- No extra text.

List of movies, tv series, anime and games user rated (* means, no score given),
`

const message = "Recommend movies, tv series, anime and games."

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
	OpenAIResponse OpenAIResponse       `bson:"openai_response" json:"openai_response"`
	Movies         []responses.Movie    `bson:"movies" json:"movies"`
	TVSeries       []responses.TVSeries `bson:"tv" json:"tv"`
	Anime          []responses.Anime    `bson:"anime" json:"anime"`
	Games          []responses.Game     `bson:"game" json:"game"`
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
