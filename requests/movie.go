package requests

type SortFilterMovie struct {
	Status                      *string `form:"status" binding:"omitempty,oneof=production released planned"`
	Genres                      *string `form:"genres"`
	ProductionCompanies         *string `form:"production_companies"`
	ProductionCountry           *string `form:"production_country"`
	StreamingPlatforms          *string `form:"streaming_platforms"`
	IsStreamingPlatformFiltered *bool   `form:"is_region_filtered"`
	Region                      *string `form:"region"`
	ReleaseDateFrom             *int    `form:"from" binding:"omitempty,number,min=1900,max=2050"`
	ReleaseDateTo               *int    `form:"to" binding:"omitempty,number,min=1900,max=2050"`
	Sort                        string  `form:"sort" binding:"required,oneof=top popularity new old"`
	Page                        int64   `form:"page" json:"page" binding:"required,number,min=1"`
}

type RegionFilters struct {
	Region string `form:"region" binding:"required"`
}

type FilterByStreamingPlatformAndRegion struct {
	Region            string `form:"region" binding:"required"`
	StreamingPlatform string `form:"platform" binding:"required"`
	Sort              string `form:"sort" binding:"required,oneof=popularity new old"`
	Page              int64  `form:"page" json:"page" binding:"required,number,min=1"`
}
