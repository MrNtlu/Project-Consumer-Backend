package helpers

import (
	"app/db"
	"app/models"
	"app/requests"
	"app/utils"
	"errors"
	"log"
	"net/http"
	"os"
	"time"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

//lint:file-ignore ST1005 Ignore all

var identityKey = "id"
var (
	errMissingAuth   = errors.New("Missing email or password")
	errIncorrectAuth = errors.New("Incorrect email or password")
	errEmptyPassword = errors.New("Password is empty")
)

func SetupJWTHandler(mongoDB *db.MongoDB) *jwt.GinJWTMiddleware {
	// port := os.Getenv("PORT")
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	authMiddleware, err := jwt.New(&jwt.GinJWTMiddleware{
		Realm:       "project-consumer",
		Key:         []byte(os.Getenv("JWT_SECRET_KEY")),
		Timeout:     time.Hour * 730,  // 1 month
		MaxRefresh:  time.Hour * 2190, // 3 months
		IdentityKey: identityKey,
		Authenticator: func(c *gin.Context) (interface{}, error) {
			var data requests.Login
			if err := c.Bind(&data); err != nil {
				return "", errMissingAuth
			}

			userModel := models.NewUserModel(mongoDB)
			user, err := userModel.FindUserByEmail(data.EmailAddress)
			if err != nil {
				return "", errIncorrectAuth
			}

			if user.Password == "" {
				return "", errEmptyPassword
			}

			if err := utils.CheckPassword([]byte(user.Password), []byte(data.Password)); err != nil {
				logrus.WithFields(logrus.Fields{
					"email": data.EmailAddress,
					"uid":   user.ID,
				}).Error("failed to check password: ", err)

				return "", errIncorrectAuth
			}

			if data.FCMToken != nil && user.FCMToken != *data.FCMToken {
				user.FCMToken = *data.FCMToken
				go userModel.UpdateUser(user)
			}

			return user, nil
		},
		PayloadFunc: func(data interface{}) jwt.MapClaims {
			if user, ok := data.(models.User); ok {
				return jwt.MapClaims{
					identityKey: user.ID,
				}
			}
			return jwt.MapClaims{}
		},
		Unauthorized: func(c *gin.Context, code int, message string) {
			c.JSON(code, gin.H{
				"code":    code,
				"message": message,
			})
		},
		LoginResponse: func(c *gin.Context, code int, token string, expire time.Time) {
			c.SetCookie("access_token", token, 2592000, "/", os.Getenv("BASE_URI"), true, true)
			c.JSON(http.StatusOK, gin.H{"access_token": token})
		},
		RefreshResponse: func(c *gin.Context, code int, token string, expire time.Time) {
			c.SetCookie("access_token", token, 2592000, "/", os.Getenv("BASE_URI"), true, true)
			c.JSON(http.StatusOK, gin.H{"access_token": token})
		},
		TokenLookup:    "header: Authorization, cookie: access_token",
		TimeFunc:       time.Now,
		SendCookie:     true,
		SecureCookie:   false,          // non HTTPS dev environments
		CookieHTTPOnly: true,           // JS can't modify
		CookieName:     "access_token", // default jwt
	})

	if err != nil {
		log.Fatal("JWT Error:" + err.Error())
	}

	return authMiddleware
}
