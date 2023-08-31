package requests

type SortFilterTVSeries struct {
	Status              *string `form:"status" binding:"omitempty,oneof=production ended airing"`
	Genres              *string `form:"genres"`
	NumSeason           *int    `form:"season" binding:"omitempty,number,min=1"`
	ProductionCompanies *string `form:"production_companies"`
	FirstAirDateFrom    *int    `form:"from" binding:"omitempty,number,min=1900,max=2050"`
	FirstAirDateTo      *int    `form:"to" binding:"omitempty,number,min=1900,max=2050"`
	Sort                string  `form:"sort" binding:"required,oneof=top popularity new old"`
	Page                int64   `form:"page" json:"page" binding:"required,number,min=1"`
}
