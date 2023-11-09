package controllers

import (
	"app/db"
	"app/models"
	"app/requests"
	"app/responses"
	"encoding/json"
	"net/http"
	"os"
	"strings"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

type OAuth2Controller struct {
	Database *db.MongoDB
}

func NewOAuth2Controller(mongoDB *db.MongoDB) OAuth2Controller {
	return OAuth2Controller{
		Database: mongoDB,
	}
}

var (
	oauthStateString    = "consumer-login"
	errFailedLogin      = "Failed to login"
	errWrongLoginMethod = "Failed to login. This email is already registered with different login method."
)

const tokenExpiration = 259200

// OAuth2 Google Login
// @Summary OAuth2 Google Login
// @Description Gets user info from google and creates/finds user and returns token
// @Tags oauth2
// @Accept application/json
// @Produce application/json
// @Success 200 {string} string "Token"
// @Failure 400 {string} string
// @Failure 401 {string} string
// @Failure 500 {string} string
// @Router /oauth/google [post]
func (o *OAuth2Controller) OAuth2GoogleLogin(jwt *jwt.GinJWTMiddleware) gin.HandlerFunc {
	return func(c *gin.Context) {
		var data requests.GoogleLogin
		if shouldReturn := bindJSONData(&data, c); shouldReturn {
			return
		}

		//Old https://www.googleapis.com/oauth2/v3/tokeninfo?access_token=
		response, err := http.Get("https://oauth2.googleapis.com/tokeninfo?id_token=" + data.Token)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": errFailedLogin,
			})

			return
		}
		defer response.Body.Close()

		var googleToken responses.GoogleToken
		if err := json.NewDecoder(response.Body).Decode(&googleToken); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})

			return
		}

		if googleToken.Email == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"message": errFailedLogin, "code": http.StatusUnauthorized})
			return
		}

		userModel := models.NewUserModel(o.Database)

		var user models.User
		user, _ = userModel.FindUserByEmail(googleToken.Email)

		if user.EmailAddress == "" {
			username := strings.Split(googleToken.Email, "@")[0]

			oAuthUser, err := userModel.CreateOAuthUser(googleToken.Email, username, data.FCMToken, data.Image, nil, 0)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			user = *oAuthUser
		}

		if !user.IsOAuthUser || (user.IsOAuthUser && user.OAuthType != nil && *user.OAuthType != 0) {
			c.JSON(http.StatusInternalServerError, gin.H{"error": errWrongLoginMethod})
			return
		}

		userListModel := models.NewUserListModel(o.Database)
		baseUserList, _ := userListModel.GetBaseUserListByUserID(user.ID.Hex())

		if baseUserList.UserID == "" {
			if err := userListModel.CreateUserList(user.ID.Hex(), user.Username); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})

				return
			}
		}

		token, _, err := jwt.TokenGenerator(user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.SetCookie("jwt", token, tokenExpiration, "/", os.Getenv("BASE_URI"), true, true)
		c.JSON(http.StatusOK, gin.H{"access_token": token})
	}
}
