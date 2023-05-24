package responses

type GoogleToken struct {
	Email string `bson:"email" json:"email"`
	ID    string `bson:"sub" json:"sub"`
}
