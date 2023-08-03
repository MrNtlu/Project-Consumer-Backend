package models

import (
	"app/db"
	"app/requests"
	"app/responses"
	"app/utils"
	"context"
	"fmt"
	"math"
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

const legendContentThreshold = 3

type User struct {
	ID                 primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	Username           string             `bson:"username" json:"username"`
	EmailAddress       string             `bson:"email" json:"email"`
	Image              string             `bson:"image" json:"image"`
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
func createUserObject(emailAddress, username, password, fcmToken, image string) *User {
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

func createOAuthUserObject(emailAddress, username, fcmToken, image string, refreshToken *string, oAuthType int) *User {
	return &User{
		EmailAddress:     emailAddress,
		Username:         username,
		Image:            image,
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

func (userModel *UserModel) CreateOAuthUser(emailAddress, username, fcmToken, image string, refreshToken *string, oAuthType int) (*User, error) {
	user := createOAuthUserObject(emailAddress, username, fcmToken, image, refreshToken, oAuthType)

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

func (userModel *UserModel) GetUserLevel(uid string) (int, error) {
	objectID, _ := primitive.ObjectIDFromHex(uid)

	match := bson.M{"$match": bson.M{
		"_id": objectID,
	}}

	addFields := bson.M{"$addFields": bson.M{
		"user_id": bson.M{
			"$toString": "$_id",
		},
	}}

	facet := bson.M{"$facet": bson.M{
		"lookups": bson.A{
			bson.M{
				"$lookup": bson.M{
					"from": "anime-lists",
					"let": bson.M{
						"user_id": "$user_id",
					},
					"pipeline": bson.A{
						bson.M{
							"$match": bson.M{
								"$expr": bson.M{
									"$eq": bson.A{
										"$user_id",
										"$$user_id",
									},
								},
							},
						},
						bson.M{
							"$group": bson.M{
								"_id": "$status",
								"total": bson.M{
									"$sum": bson.M{
										"$add": bson.A{
											bson.M{
												"$cond": bson.M{
													"if": bson.M{
														"$eq": bson.A{"finished", "$status"},
													},
													"then": 100,
													"else": 50,
												},
											},
											bson.M{
												"$cond": bson.M{
													"if": bson.M{
														"$gt": bson.A{"$score", 0},
													},
													"then": 25,
													"else": 0,
												},
											},
										},
									},
								},
							},
						},
					},
					"as": "anime_list",
				},
			},
			bson.M{
				"$lookup": bson.M{
					"from": "game-lists",
					"let": bson.M{
						"user_id": "$user_id",
					},
					"pipeline": bson.A{
						bson.M{
							"$match": bson.M{
								"$expr": bson.M{
									"$eq": bson.A{
										"$user_id",
										"$$user_id",
									},
								},
							},
						},
						bson.M{
							"$group": bson.M{
								"_id": "$status",
								"total": bson.M{
									"$sum": bson.M{
										"$add": bson.A{
											bson.M{
												"$cond": bson.M{
													"if": bson.M{
														"$eq": bson.A{"finished", "$status"},
													},
													"then": 100,
													"else": 50,
												},
											},
											bson.M{
												"$cond": bson.M{
													"if": bson.M{
														"$gt": bson.A{"$score", 0},
													},
													"then": 25,
													"else": 0,
												},
											},
										},
									},
								},
							},
						},
					},
					"as": "game_list",
				},
			},
			bson.M{
				"$lookup": bson.M{
					"from": "movie-watch-lists",
					"let": bson.M{
						"user_id": "$user_id",
					},
					"pipeline": bson.A{
						bson.M{
							"$match": bson.M{
								"$expr": bson.M{
									"$eq": bson.A{
										"$user_id",
										"$$user_id",
									},
								},
							},
						},
						bson.M{
							"$group": bson.M{
								"_id": "$status",
								"total": bson.M{
									"$sum": bson.M{
										"$add": bson.A{
											bson.M{
												"$cond": bson.M{
													"if": bson.M{
														"$eq": bson.A{"finished", "$status"},
													},
													"then": 100,
													"else": 50,
												},
											},
											bson.M{
												"$cond": bson.M{
													"if": bson.M{
														"$gt": bson.A{"$score", 0},
													},
													"then": 25,
													"else": 0,
												},
											},
										},
									},
								},
							},
						},
					},
					"as": "movie_list",
				},
			},
			bson.M{
				"$lookup": bson.M{
					"from": "tvseries-watch-lists",
					"let": bson.M{
						"user_id": "$user_id",
					},
					"pipeline": bson.A{
						bson.M{
							"$match": bson.M{
								"$expr": bson.M{
									"$eq": bson.A{
										"$user_id",
										"$$user_id",
									},
								},
							},
						},
						bson.M{
							"$group": bson.M{
								"_id": "$status",
								"total": bson.M{
									"$sum": bson.M{
										"$add": bson.A{
											bson.M{
												"$cond": bson.M{
													"if": bson.M{
														"$eq": bson.A{"finished", "$status"},
													},
													"then": 100,
													"else": 50,
												},
											},
											bson.M{
												"$cond": bson.M{
													"if": bson.M{
														"$gt": bson.A{"$score", 0},
													},
													"then": 25,
													"else": 0,
												},
											},
										},
									},
								},
							},
						},
					},
					"as": "tv_list",
				},
			},
			bson.M{
				"$lookup": bson.M{
					"from": "consume-laters",
					"let": bson.M{
						"user_id": "$user_id",
					},
					"pipeline": bson.A{
						bson.M{
							"$match": bson.M{
								"$expr": bson.M{
									"$eq": bson.A{
										"$user_id",
										"$$user_id",
									},
								},
							},
						},
						bson.M{
							"$group": bson.M{
								"_id": "$status",
								"total": bson.M{
									"$sum": 25,
								},
							},
						},
					},
					"as": "later_list",
				},
			},
		},
	}}

	unwind := bson.M{"$unwind": bson.M{
		"path":                       "$lookups",
		"includeArrayIndex":          "index",
		"preserveNullAndEmptyArrays": true,
	}}

	replaceRoot := bson.M{"$replaceRoot": bson.M{
		"newRoot": "$lookups",
	}}

	group := bson.M{"$group": bson.M{
		"_id": nil,
		"total_score": bson.M{
			"$sum": bson.M{
				"$add": bson.A{
					bson.M{
						"$sum": "$anime_list.total",
					},
					bson.M{
						"$sum": "$game_list.total",
					},
					bson.M{
						"$sum": "$movie_list.total",
					},
					bson.M{
						"$sum": "$tv_list.total",
					},
					bson.M{
						"$sum": "$later_list.total",
					},
				},
			},
		},
	}}

	cursor, err := userModel.Collection.Aggregate(context.TODO(), bson.A{
		match, addFields, facet, unwind, replaceRoot, group,
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"uid": uid,
		}).Error("failed to aggregate user level: ", err)

		return 1, fmt.Errorf("Failed to aggregate user level.")
	}

	var userLevel []responses.UserLevel
	if err = cursor.All(context.TODO(), &userLevel); err != nil {
		logrus.WithFields(logrus.Fields{
			"uid": uid,
		}).Error("failed to decode user level: ", err)

		return 1, fmt.Errorf("Failed to decode user level.")
	}

	if len(userLevel) > 0 {
		return int(math.Sqrt(float64(userLevel[0].TotalScore)) * 0.2), nil
	}

	return 1, nil
}

func (userModel *UserModel) GetUserInfo(uid string) (responses.UserInfo, error) {
	objectID, _ := primitive.ObjectIDFromHex(uid)

	match := bson.M{"$match": bson.M{
		"_id": objectID,
	}}

	addFields := bson.M{"$addFields": bson.M{
		"user_id": bson.M{
			"$toString": "$_id",
		},
	}}

	facet := bson.M{"$facet": bson.M{
		"lookups": bson.A{
			bson.M{
				"$lookup": bson.M{
					"from":         "anime-lists",
					"localField":   "user_id",
					"foreignField": "user_id",
					"as":           "anime_list",
				},
			},
			bson.M{
				"$lookup": bson.M{
					"from":         "game-lists",
					"localField":   "user_id",
					"foreignField": "user_id",
					"as":           "game_list",
				},
			},
			bson.M{
				"$lookup": bson.M{
					"from":         "movie-watch-lists",
					"localField":   "user_id",
					"foreignField": "user_id",
					"as":           "movie_list",
				},
			},
			bson.M{
				"$lookup": bson.M{
					"from":         "tvseries-watch-lists",
					"localField":   "user_id",
					"foreignField": "user_id",
					"as":           "tv_list",
				},
			},
		},
	}}

	unwind := bson.M{"$unwind": bson.M{
		"path":                       "$lookups",
		"includeArrayIndex":          "index",
		"preserveNullAndEmptyArrays": true,
	}}

	replaceRoot := bson.M{"$replaceRoot": bson.M{
		"newRoot": "$lookups",
	}}

	set := bson.M{"$set": bson.M{
		"anime_count": bson.M{
			"$size": "$anime_list",
		},
		"movie_count": bson.M{
			"$size": "$movie_list",
		},
		"tv_count": bson.M{
			"$size": "$tv_list",
		},
		"game_count": bson.M{
			"$size": "$game_list",
		},
	}}

	project := bson.M{"$project": bson.M{
		"username":    "$username",
		"email":       "$email",
		"is_premium":  "$is_premium",
		"image":       "$image",
		"anime_count": "$anime_count",
		"movie_count": "$movie_count",
		"tv_count":    "$tv_count",
		"game_count":  "$game_count",
		"fcm_token":   "$fcm_token",
		"legend_anime_list": bson.M{
			"$map": bson.M{
				"input": bson.M{
					"$filter": bson.M{
						"input": "$anime_list",
						"as":    "animes",
						"cond": bson.M{
							"$gte": bson.A{"$$animes.times_finished", legendContentThreshold},
						},
					},
				},
				"as": "anime",
				"in": bson.M{
					"times_finished": "$$anime.times_finished",
					"anime_obj_id": bson.M{
						"$toObjectId": "$$anime.anime_id",
					},
				},
			},
		},
		"legend_movie_list": bson.M{
			"$map": bson.M{
				"input": bson.M{
					"$filter": bson.M{
						"input": "$movie_list",
						"as":    "movies",
						"cond": bson.M{
							"$gte": bson.A{"$$movies.times_finished", legendContentThreshold},
						},
					},
				},
				"as": "movie",
				"in": bson.M{
					"times_finished": "$$movie.times_finished",
					"movie_obj_id": bson.M{
						"$toObjectId": "$$movie.movie_id",
					},
				},
			},
		},
		"legend_tv_list": bson.M{
			"$map": bson.M{
				"input": bson.M{
					"$filter": bson.M{
						"input": "$tv_list",
						"as":    "tvs",
						"cond": bson.M{
							"$gte": bson.A{"$$tvs.times_finished", legendContentThreshold},
						},
					},
				},
				"as": "tv",
				"in": bson.M{
					"times_finished": "$$tv.times_finished",
					"tv_obj_id": bson.M{
						"$toObjectId": "$$tv.tv_id",
					},
				},
			},
		},
		"legend_game_list": bson.M{
			"$map": bson.M{
				"input": bson.M{
					"$filter": bson.M{
						"input": "$game_list",
						"as":    "games",
						"cond": bson.M{
							"$gte": bson.A{"$$games.times_finished", legendContentThreshold},
						},
					},
				},
				"as": "game",
				"in": bson.M{
					"times_finished": "$$game.times_finished",
					"game_obj_id": bson.M{
						"$toObjectId": "$$game.game_id",
					},
				},
			},
		},
	}}

	contentFacet := bson.M{"$facet": bson.M{
		"lookups": bson.A{
			bson.M{
				"$lookup": bson.M{
					"from": "movies",
					"let": bson.M{
						"obj_id":         "$legend_movie_list.movie_obj_id",
						"times_finished": "$legend_movie_list.times_finished",
					},
					"pipeline": bson.A{
						bson.M{
							"$match": bson.M{
								"$expr": bson.M{
									"$in": bson.A{
										"$_id",
										"$$obj_id",
									},
								},
							},
						},
						bson.M{
							"$project": bson.M{
								"image_url":      1,
								"title_original": 1,
								"title_en":       1,
								"times_finished": bson.M{
									"$arrayElemAt": bson.A{
										"$$times_finished",
										bson.M{
											"$indexOfArray": bson.A{
												"$$obj_id",
												"$_id",
											},
										},
									},
								},
							},
						},
					},
					"as": "legend_movie_list",
				},
			},
			bson.M{
				"$lookup": bson.M{
					"from": "tv-series",
					"let": bson.M{
						"obj_id":         "$legend_tv_list.tv_obj_id",
						"times_finished": "$legend_tv_list.times_finished",
					},
					"pipeline": bson.A{
						bson.M{
							"$match": bson.M{
								"$expr": bson.M{
									"$in": bson.A{
										"$_id",
										"$$obj_id",
									},
								},
							},
						},
						bson.M{
							"$project": bson.M{
								"image_url":      1,
								"title_original": 1,
								"title_en":       1,
								"times_finished": bson.M{
									"$arrayElemAt": bson.A{
										"$$times_finished",
										bson.M{
											"$indexOfArray": bson.A{
												"$$obj_id",
												"$_id",
											},
										},
									},
								},
							},
						},
					},
					"as": "legend_tv_list",
				},
			},
			bson.M{
				"$lookup": bson.M{
					"from": "animes",
					"let": bson.M{
						"obj_id":         "$legend_anime_list.anime_obj_id",
						"times_finished": "$legend_anime_list.times_finished",
					},
					"pipeline": bson.A{
						bson.M{
							"$match": bson.M{
								"$expr": bson.M{
									"$in": bson.A{
										"$_id",
										"$$obj_id",
									},
								},
							},
						},
						bson.M{
							"$project": bson.M{
								"image_url":      1,
								"title_original": 1,
								"title_en":       1,
								"title_jp":       1,
								"times_finished": bson.M{
									"$arrayElemAt": bson.A{
										"$$times_finished",
										bson.M{
											"$indexOfArray": bson.A{
												"$$obj_id",
												"$_id",
											},
										},
									},
								},
							},
						},
					},
					"as": "legend_anime_list",
				},
			},
			bson.M{
				"$lookup": bson.M{
					"from": "games",
					"let": bson.M{
						"obj_id":         "$legend_game_list.game_obj_id",
						"times_finished": "$legend_game_list.times_finished",
					},
					"pipeline": bson.A{
						bson.M{
							"$match": bson.M{
								"$expr": bson.M{
									"$in": bson.A{
										"$_id",
										"$$obj_id",
									},
								},
							},
						},
						bson.M{
							"$project": bson.M{
								"image_url":      1,
								"title_original": 1,
								"title":          1,
								"times_finished": bson.M{
									"$arrayElemAt": bson.A{
										"$$times_finished",
										bson.M{
											"$indexOfArray": bson.A{
												"$$obj_id",
												"$_id",
											},
										},
									},
								},
							},
						},
					},
					"as": "legend_game_list",
				},
			},
		},
	}}

	unwindContentFacet := bson.M{"$unwind": bson.M{
		"path":                       "$lookups",
		"includeArrayIndex":          "index",
		"preserveNullAndEmptyArrays": true,
	}}

	finalReplaceRoot := bson.M{"$replaceRoot": bson.M{
		"newRoot": "$lookups",
	}}

	cursor, err := userModel.Collection.Aggregate(context.TODO(), bson.A{
		match, addFields, facet, unwind, replaceRoot,
		set, project, contentFacet, unwindContentFacet, finalReplaceRoot,
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"uid": uid,
		}).Error("failed to aggregate user info: ", err)

		return responses.UserInfo{}, fmt.Errorf("Failed to aggregate user info.")
	}

	var userInfo []responses.UserInfo
	if err = cursor.All(context.TODO(), &userInfo); err != nil {
		logrus.WithFields(logrus.Fields{
			"uid": uid,
		}).Error("failed to decode user info: ", err)

		return responses.UserInfo{}, fmt.Errorf("Failed to decode user info.")
	}

	if len(userInfo) > 0 {
		return userInfo[0], nil
	}

	return responses.UserInfo{}, nil
}
