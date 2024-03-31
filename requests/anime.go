package requests

type SortFilterAnime struct {
	Status             *string `form:"status" binding:"omitempty,oneof=airing upcoming finished"`
	Year               *int16  `form:"year" binding:"omitempty,number"`
	Season             *string `form:"season" binding:"omitempty,oneof=winter summer fall spring"`
	StreamingPlatforms *string `form:"streaming_platforms"`
	Genres             *string `form:"genres"`
	Demographics       *string `form:"demographics"`
	Themes             *string `form:"themes"`
	Studios            *string `form:"studios"`
	Sort               string  `form:"sort" binding:"required,oneof=top popularity new old"`
	Page               int64   `form:"page" json:"page" binding:"required,number,min=1"`
}

type FilterByStreamingPlatform struct {
	StreamingPlatform string `form:"platform" binding:"required"`
	Sort              string `form:"sort" binding:"required,oneof=popularity new old"`
	Page              int64  `form:"page" json:"page" binding:"required,number,min=1"`
}

type FilterByStudio struct {
	Studio string `form:"studio" binding:"required"`
	Sort   string `form:"sort" binding:"required,oneof=popularity new old"`
	Page   int64  `form:"page" json:"page" binding:"required,number,min=1"`
}
