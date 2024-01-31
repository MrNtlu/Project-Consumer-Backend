package models

import (
	"app/db"
	"app/requests"
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type CustomListModel struct {
	Collection *mongo.Collection
}

func NewCustomListModel(mongoDB *db.MongoDB) *CustomListModel {
	return &CustomListModel{
		Collection: mongoDB.Database.Collection("custom-lists"),
	}
}

type CustomList struct {
	ID          primitive.ObjectID  `bson:"_id,omitempty" json:"_id"`
	UserID      string              `bson:"user_id" json:"user_id"`
	Name        string              `bson:"name" json:"name"`
	Description *string             `bson:"description" json:"description"`
	Likes       []string            `bson:"likes" json:"likes"`
	IsPrivate   bool                `bson:"is_private" json:"is_private"`
	Content     []CustomListContent `bson:"content" json:"content"`
	CreatedAt   time.Time           `bson:"created_at" json:"created_at"`
}

type CustomListContent struct {
	Order                int     `bson:"order" json:"order"`
	ContentID            string  `bson:"content_id" json:"content_id"`
	ContentExternalID    *string `bson:"content_external_id" json:"content_external_id"`
	ContentExternalIntID *int64  `bson:"content_external_int_id" json:"content_external_int_id"`
	ContentType          string  `bson:"content_type" json:"content_type"` // anime, movie, tv or game
}

//Order https://github.com/MrNtlu/Asset-Manager/blob/master/models/favouriteInvesting.go
/* TODO Endpoints
* - Get custom list by user id -sort by popularity, name, created at
* - Add to custom list (pass content type, content id etc.)
* - Remove from custom list
* - Bulk delete from list
* - Update custom list
* - Reorder custom list
* - Delete custom list
 */

func createCustomListObject(userID, name string, description *string, isPrivate bool, content []requests.CustomListContent) *CustomList {
	return &CustomList{
		UserID:      userID,
		Name:        name,
		Description: description,
		Likes:       []string{},
		IsPrivate:   isPrivate,
		Content:     convertRequestToModel(content),
		CreatedAt:   time.Now().UTC(),
	}
}

func (customListModel *CustomListModel) CreateCustomList(uid string, data requests.CreateCustomList) (CustomList, error) {
	customList := createCustomListObject(
		uid,
		data.Name,
		data.Description,
		*data.IsPrivate,
		data.Content,
	)

	var (
		insertedID *mongo.InsertOneResult
		err        error
	)

	if insertedID, err = customListModel.Collection.InsertOne(context.TODO(), customList); err != nil {
		logrus.WithFields(logrus.Fields{
			"custom_list": customList,
		}).Error("failed to create new custom list: ", err)

		return CustomList{}, fmt.Errorf("Failed to create custom list.")
	}

	customList.ID = insertedID.InsertedID.(primitive.ObjectID)

	return *customList, nil
}

func convertRequestToModel(content []requests.CustomListContent) []CustomListContent {
	var modelList []CustomListContent

	for i := 0; i < len(content); i++ {
		modelList = append(modelList, CustomListContent{
			Order:                content[i].Order,
			ContentID:            content[i].ContentID,
			ContentExternalID:    content[i].ContentExternalID,
			ContentExternalIntID: content[i].ContentExternalIntID,
		})
	}

	return modelList
}
