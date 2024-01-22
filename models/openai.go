package models

import (
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

const summaryPrompt = `You are a content summarizer. Your only job is to give brief summary of given content to the user. User should understand the content from the summary. Follow rules strictly!:

1- Only the summary of the content.
2- No longer than 750 words.
3- No extra unnecessary text and information.
4- Try to add information like genre, content like it.`

const generalOpinionPrompt = `You are a content reviewer. Your only job is to give summary of the reviews about the content to the user. Return the summary of what people think about it with positive and negatives. Follow rules strictly!:

1- Try to get all reviews from various sites and give average.
2- No extra unnecessary text and information. Just average reviews  by users and what they say about it.
3- No longer than 500 words.`

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

type AssistantResponse struct {
	TotalUsedToken int    `bson:"total_used_token" json:"total_used_token"`
	Response       string `bson:"response" json:"response"`
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

func (oa *OpenAI) GetSummary(name, contentType string) (AssistantResponse, error) {
	resp, err := oa.Client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: summaryPrompt,
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: name + ", " + contentType,
				},
			},
		},
	)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"name":         name,
			"content_type": contentType,
		}).Error("Error: ", err)

		return AssistantResponse{}, err
	}

	var (
		model  string
		object string

		totalTokens      int
		completionTokens int
		promptTokens     int
		summary          string
	)

	model = resp.Model
	object = resp.Object
	totalTokens = resp.Usage.TotalTokens
	completionTokens = resp.Usage.CompletionTokens
	promptTokens = resp.Usage.PromptTokens
	summary = resp.Choices[0].Message.Content

	logrus.WithFields(logrus.Fields{
		"input":             message,
		"model":             model,
		"object":            object,
		"total_tokens":      totalTokens,
		"completion_tokens": completionTokens,
		"prompt_tokens":     promptTokens,
	}).Info("choices", resp.Choices)

	openAIResponse := AssistantResponse{
		TotalUsedToken: totalTokens,
		Response:       summary,
	}

	return openAIResponse, err
}

func (oa *OpenAI) GetPublicOpinion(name, contentType string) (AssistantResponse, error) {
	resp, err := oa.Client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: generalOpinionPrompt,
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: name + ", " + contentType,
				},
			},
		},
	)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"name":         name,
			"content_type": contentType,
		}).Error("Error: ", err)

		return AssistantResponse{}, err
	}

	var (
		model  string
		object string

		totalTokens      int
		completionTokens int
		promptTokens     int
		summary          string
	)

	model = resp.Model
	object = resp.Object
	totalTokens = resp.Usage.TotalTokens
	completionTokens = resp.Usage.CompletionTokens
	promptTokens = resp.Usage.PromptTokens
	summary = resp.Choices[0].Message.Content

	logrus.WithFields(logrus.Fields{
		"input":             message,
		"model":             model,
		"object":            object,
		"total_tokens":      totalTokens,
		"completion_tokens": completionTokens,
		"prompt_tokens":     promptTokens,
	}).Info("choices", resp.Choices)

	openAIResponse := AssistantResponse{
		TotalUsedToken: totalTokens,
		Response:       summary,
	}

	return openAIResponse, err
}
