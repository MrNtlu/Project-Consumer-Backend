package responses

type PaginationResponse struct {
	Metadata PaginationMetaData `bson:"metadata" json:"metadata"`
}

type PaginationMetaData struct {
	Count       int64 `bson:"count" json:"count"`
	Total       int64 `bson:"total" json:"total"`
	Page        int64 `bson:"page" json:"page"`
	CanPaginate bool  `bson:"can_paginate" json:"can_paginate"`
}
