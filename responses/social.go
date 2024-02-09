package responses

type SocialPreview struct {
	Reviews     []ReviewDetails `bson:"reviews" json:"reviews"`
	Leaderboard []Leaderboard   `bson:"leaderboard" json:"leaderboard"`
	CustomLists []CustomList    `bson:"custom_lists" json:"custom_lists"`
}
