package controllers

import (
	"app/db"
)

type MovieController struct {
	Database *db.MongoDB
}

func NewMovieController(mongoDB *db.MongoDB) MovieController {
	return MovieController{
		Database: mongoDB,
	}
}
