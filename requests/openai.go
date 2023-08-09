package requests

type OpenAIRecommendation struct {
	Input string `form:"input" binding:"required"`
}
