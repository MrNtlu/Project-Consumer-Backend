package requests

type SortAnime struct {
	Sort      string `form:"sort" binding:"required,oneof=popularity date"`
	SortOrder int8   `form:"type" json:"type" binding:"required,oneof=1 -1"`
	Page      int64  `form:"page" json:"page" binding:"required,number,min=1"`
}

type SortByYearSeasonAnime struct {
	Year      int16  `form:"year" binding:"required,number"`
	Season    string `form:"season" binding:"required,oneof=winter summer fall spring"`
	Sort      string `form:"sort" binding:"required,oneof=popularity date"`
	SortOrder int8   `form:"type" json:"type" binding:"required,oneof=1 -1"`
	Page      int64  `form:"page" json:"page" binding:"required,number,min=1"`
}
