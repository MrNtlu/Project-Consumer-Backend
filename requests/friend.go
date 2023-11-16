package requests

type AnswerFriendRequest struct {
	ID     string `json:"id" binding:"required"`
	Answer int    `json:"answer" binding:"required,min=0,max=2"` //0 deny, 1 accept, 2 ignore
}
