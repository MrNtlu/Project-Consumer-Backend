package requests

type Feedback struct {
	Feedback string `json:"feedback" binding:"required"`
}
