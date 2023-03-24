package models

import (
	"app/db"
	"app/requests"
	"app/responses"
	"app/utils"
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
	EmailAddress       string             `bson:"email" json:"email"`
	Image              *string            `bson:"image" json:"image"`
	Password           string             `bson:"password" json:"-"`
	PasswordResetToken string             `bson:"reset_token" json:"-"`
	CreatedAt          time.Time          `bson:"created_at" json:"-"`
	UpdatedAt          time.Time          `bson:"updated_at" json:"-"`
	IsPremium          bool               `bson:"is_premium" json:"is_premium"`
	IsOAuthUser        bool               `bson:"is_oauth" json:"is_oauth"`
	OAuthType          *int               `bson:"oauth_type" json:"oauth_type"`
	RefreshToken       *string            `bson:"refresh_token" json:"-"`
	FCMToken           string             `bson:"fcm_token" json:"fcm_token"`
	AppNotification    bool               `bson:"app_notification" json:"app_notification"`
	MailNotification   bool               `bson:"mail_notification" json:"mail_notification"`
}

// Create
func createUserObject(emailAddress, username, password, fcmToken string, image *string) *User {
	return &User{
		Username:         username,
		EmailAddress:     emailAddress,
		Image:            image,
		Password:         utils.HashPassword(password),
		CreatedAt:        time.Now().UTC(),
		UpdatedAt:        time.Now().UTC(),
		IsPremium:        false,
		IsOAuthUser:      false,
		AppNotification:  true,
		MailNotification: true,
		FCMToken:         fcmToken,
	}
}

func createOAuthUserObject(emailAddress, username, fcmToken string, refreshToken *string, oAuthType int) *User {
	return &User{
		EmailAddress:     emailAddress,
		Username:         username,
		CreatedAt:        time.Now().UTC(),
		UpdatedAt:        time.Now().UTC(),
		IsPremium:        false,
		IsOAuthUser:      true,
		AppNotification:  true,
		MailNotification: false,
		OAuthType:        &oAuthType,
		RefreshToken:     refreshToken,
		FCMToken:         fcmToken,
	}
}

func (userModel *UserModel) CreateUser(data requests.Register) (*User, error) {
	user := createUserObject(data.EmailAddress, data.Username, data.Password, data.FCMToken, data.Image)

	result, err := userModel.Collection.InsertOne(context.TODO(), user)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"email": data.EmailAddress,
		}).Error("failed to create new user: ", err)

		return nil, fmt.Errorf("Failed to create new user.")
	}

	user.ID = result.InsertedID.(primitive.ObjectID)

	return user, nil
}

func (userModel *UserModel) CreateOAuthUser(emailAddress, username, fcmToken string, refreshToken *string, oAuthType int) (*User, error) {
	user := createOAuthUserObject(emailAddress, username, fcmToken, refreshToken, oAuthType)

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

// Update
func (userModel *UserModel) UpdateUser(user User) error {
	user.UpdatedAt = time.Now().UTC()

	if _, err := userModel.Collection.UpdateOne(context.TODO(), bson.M{"_id": user.ID}, bson.M{"$set": user}); err != nil {
		logrus.WithFields(logrus.Fields{
			"uid": user.ID,
		}).Error("failed to update user: ", err)

		return fmt.Errorf("Failed to update user.")
	}

	return nil
}

func (userModel *UserModel) UpdateUserMembership(uid string, data requests.ChangeMembership) error {
	objectUID, _ := primitive.ObjectIDFromHex(uid)

	if _, err := userModel.Collection.UpdateOne(context.TODO(), bson.M{"_id": objectUID}, bson.M{"$set": bson.M{
		"is_premium": data.IsPremium,
		"updated_at": time.Now().UTC(),
	}}); err != nil {
		logrus.WithFields(logrus.Fields{
			"uid":        uid,
			"is_premium": data.IsPremium,
		}).Error("failed to set membership for user: ", err)

		return fmt.Errorf("Failed to set membership for user.")
	}

	return nil
}

// Checks
func (userModel *UserModel) IsUserPremium(uid string) bool {
	objectUID, _ := primitive.ObjectIDFromHex(uid)

	result := userModel.Collection.FindOne(context.TODO(), bson.M{
		"_id": objectUID,
	})

	var isUserPremium responses.IsUserPremium
	if err := result.Decode(&isUserPremium); err != nil {
		logrus.WithFields(logrus.Fields{
			"uid": uid,
		}).Error("failed to find user by uid: ", err)

		return false
	}

	return isUserPremium.IsPremium || isUserPremium.IsLifetimePremium
}

// Delete
func (userModel *UserModel) DeleteUserByID(uid string) error {
	objectUID, _ := primitive.ObjectIDFromHex(uid)

	if _, err := userModel.Collection.DeleteOne(context.TODO(), bson.M{"_id": objectUID}); err != nil {
		logrus.WithFields(logrus.Fields{
			"uid": uid,
		}).Error("failed to delete user: ", err)

		return fmt.Errorf("Failed to delete user.")
	}

	return nil
}

// Find
func (userModel *UserModel) FindUserByEmail(emailAddress string) (User, error) {
	result := userModel.Collection.FindOne(context.TODO(), bson.M{
		"email": emailAddress,
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

func (userModel *UserModel) FindUserByID(uid string) (User, error) {
	objectUID, _ := primitive.ObjectIDFromHex(uid)

	result := userModel.Collection.FindOne(context.TODO(), bson.M{
		"_id": objectUID,
	})

	var user User
	if err := result.Decode(&user); err != nil {
		logrus.WithFields(logrus.Fields{
			"uid": user.ID,
		}).Error("failed to find user by uid: ", err)

		return User{}, fmt.Errorf("Failed to find user by id.")
	}

	return user, nil
}

func (userModel *UserModel) FindUserByRefreshToken(refreshToken string) (User, error) {
	result := userModel.Collection.FindOne(context.TODO(), bson.M{
		"refresh_token": refreshToken,
	})

	var user User
	if err := result.Decode(&user); err != nil {
		logrus.WithFields(logrus.Fields{
			"refresh_token": refreshToken,
		}).Error("failed to find user by refreshToken: ", err)

		return User{}, fmt.Errorf("Failed to find user by token.")
	}

	return user, nil
}

func (userModel *UserModel) FindUserByResetTokenAndEmail(token, email string) (User, error) {
	result := userModel.Collection.FindOne(context.TODO(), bson.M{
		"reset_token": token,
		"email":       email,
	})

	var user User
	if err := result.Decode(&user); err != nil {
		logrus.WithFields(logrus.Fields{
			"uid":   user.ID,
			"token": token,
		}).Error("failed to find user by reset token: ", err)

		return User{}, fmt.Errorf("Failed to find user by reset token.")
	}

	return user, nil
}
