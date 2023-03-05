package requests

type Login struct {
	EmailAddress string `json:"email_address" binding:"required,email"`
	Password     string `json:"password" binding:"required"`
}

// TODO Fill
type Register struct {
}
