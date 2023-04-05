package requests

type SortFilterMovie struct {
	Status              *string `form:"status" binding:"omitempty,oneof=production released planned"`
	Genres              *string `form:"genres"`
	ProductionCompanies *string `form:"production_companies"`
	ReleaseDateFrom     *int    `form:"from" binding:"omitempty,number,min=1900,max=2050"`
	ReleaseDateTo       *int    `form:"to" binding:"omitempty,number,min=1900,max=2050"`
	Sort                string  `form:"sort" binding:"required,oneof=popularity new old"`
	Page                int64   `form:"page" json:"page" binding:"required,number,min=1"`
}
