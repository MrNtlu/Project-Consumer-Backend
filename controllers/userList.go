package controllers

import "app/db"

type UserListController struct {
	Database *db.MongoDB
}

func NewUserListController(mongoDB *db.MongoDB) UserListController {
	return UserListController{
		Database: mongoDB,
	}
}
