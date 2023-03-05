package models

import (
	"app/db"
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserModel struct {
	Collection *mongo.Collection
}

func NewUserModel(mongoDB *db.MongoDB) *UserModel {
	return &UserModel{
		Collection: mongoDB.Database.Collection("users"),
	}
}

type User struct {
	ID                 primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	Username           string             `bson:"username" json:"username"`
	EmailAddress       string             `bson:"email_address" json:"email_address"`
	Image              *string            `bson:"image" json:"image"`
	Password           string             `bson:"password" json:"-"`
	PasswordResetToken *string            `bson:"reset_token" json:"-"`
	CreatedAt          time.Time          `bson:"created_at" json:"-"`
	UpdatedAt          time.Time          `bson:"updated_at" json:"-"`
	IsPremium          bool               `bson:"is_premium" json:"is_premium"`
	IsOAuthUser        bool               `bson:"is_oauth" json:"is_oauth"`
	OAuthType          int                `bson:"oauth_type" json:"oauth_type"`
	RefreshToken       *string            `bson:"refresh_token" json:"-"`
	FCMToken           string             `bson:"fcm_token" json:"fcm_token"`
	AppNotification    bool               `bson:"app_notification" json:"app_notification"`
	MailNotification   bool               `bson:"mail_notification" json:"mail_notification"`
}

// func createUserObject(emailAddress, username, password, fcmToken string, image *string) *User {
// 	return &User{
// 		Username:         username,
// 		EmailAddress:     emailAddress,
// 		Image:            image,
// 		Password:         utils.HashPassword(password),
// 		CreatedAt:        time.Now().UTC(),
// 		UpdatedAt:        time.Now().UTC(),
// 		IsPremium:        false,
// 		IsOAuthUser:      false,
// 		AppNotification:  true,
// 		MailNotification: true,
// 		OAuthType:        -1,
// 		FCMToken:         fcmToken,
// 	}
// }

func createOAuthUserObject(username, emailAddress, fcmToken string, refreshToken *string, oAuthType int) *User {
	return &User{
		EmailAddress:     emailAddress,
		Username:         username,
		CreatedAt:        time.Now().UTC(),
		UpdatedAt:        time.Now().UTC(),
		IsPremium:        false,
		IsOAuthUser:      true,
		AppNotification:  true,
		MailNotification: false,
		OAuthType:        oAuthType,
		RefreshToken:     refreshToken,
		FCMToken:         fcmToken,
	}
}

// func (userModel *UserModel) CreateUser(data requests.Register) error {
// 	user := createUserObject(data.EmailAddress, data.Currency, data.Password)

// 	if _, err := userModel.Collection.InsertOne(context.TODO(), user); err != nil {
// 		logrus.WithFields(logrus.Fields{
// 			"email": data.EmailAddress,
// 		}).Error("failed to create new user: ", err)

// 		return fmt.Errorf("Failed to create new user.")
// 	}

// 	return nil
// }

func (userModel *UserModel) CreateOAuthUser(emailAddress, username, fcmToken string, refreshToken *string, oAuthType int) (*User, error) {
	user := createOAuthUserObject(username, emailAddress, fcmToken, refreshToken, oAuthType)

	result, err := userModel.Collection.InsertOne(context.TODO(), user)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"email":    emailAddress,
			"username": username,
		}).Error("failed to create new oauth user: ", err)

		return nil, fmt.Errorf("Failed to create new oauth user.")
	}

	user.ID = result.InsertedID.(primitive.ObjectID)

	return user, nil
}

func (userModel *UserModel) FindUserByEmail(emailAddress string) (User, error) {
	result := userModel.Collection.FindOne(context.TODO(), bson.M{
		"email_address": emailAddress,
	})

	var user User
	if err := result.Decode(&user); err != nil {
		logrus.WithFields(logrus.Fields{
			"email": emailAddress,
		}).Error("failed to find user by email: ", err)

		return User{}, fmt.Errorf("Failed to find user by email.")
	}

	return user, nil
}
