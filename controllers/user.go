package controllers

import (
	"app/db"
	"app/helpers"
	"app/models"
	"app/requests"
	"app/utils"
	"net/http"
	"time"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sethvargo/go-password/password"
)

type UserController struct {
	Database *db.MongoDB
}

func NewUserController(mongoDB *db.MongoDB) UserController {
	return UserController{
		Database: mongoDB,
	}
}

var (
	errAlreadyRegistered = "User already registered."
	errPasswordNoMatch   = "Passwords do not match."
	errNoUser            = "Sorry, couldn't find user."
	errOAuthUser         = "Sorry, you can't do this action."
	errMailAlreadySent   = "Password reset mail already sent, you have to wait 5 minutes before sending another. Please check spam mails."
	// errPremiumFeature    = "This feature requires premium membership."
)

// Register
// @Summary User Registration
// @Description Allows users to register
// @Tags auth
// @Accept application/json
// @Produce application/json
// @Param register body requests.Register true "User registration info"
// @Success 201 {string} string
// @Failure 400 {string} string
// @Failure 500 {string} string
// @Router /auth/register [post]
func (u *UserController) Register(c *gin.Context) {
	var data requests.Register
	if shouldReturn := bindJSONData(&data, c); shouldReturn {
		return
	}

	userModel := models.NewUserModel(u.Database)
	user, _ := userModel.FindUserByEmail(data.EmailAddress)

	if user.EmailAddress != "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": errAlreadyRegistered,
		})

		return
	}

	createdUser, err := userModel.CreateUser(data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	userListModel := models.NewUserListModel(u.Database)
	if err := userListModel.CreateUserList(createdUser.ID.Hex(), createdUser.Username); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Registered successfully."})
}

// User Info
// @Summary User basic info
// @Description Returns basic user information
// @Tags user
// @Accept application/json
// @Produce application/json
// @Security ApiKeyAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {object} models.User "User"
// @Router /user/basic [get]
func (u *UserController) GetBasicUserInfo(c *gin.Context) {
	uid := jwt.ExtractClaims(c)["id"].(string)

	userModel := models.NewUserModel(u.Database)
	userInfo, err := userModel.FindUserByID(uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully fetched basic user info.", "data": userInfo})
}

// User Info
// @Summary User membership info
// @Description Returns users membership & stats
// @Tags user
// @Accept application/json
// @Produce application/json
// @Security ApiKeyAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {object} responses.UserInfo "User Info"
// @Router /user/info [get]
func (u *UserController) GetUserInfo(c *gin.Context) {
	uid := jwt.ExtractClaims(c)["id"].(string)

	userModel := models.NewUserModel(u.Database)
	userInfo, err := userModel.GetUserInfo(uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	userLevel, _ := userModel.GetUserLevel(uid)
	userInfo.Level = userLevel

	userInteractionModel := models.NewUserInteractionModel(u.Database)

	consumeLaterList, err := userInteractionModel.GetConsumeLater(uid, requests.SortFilterConsumeLater{
		Sort: "new",
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}
	userInfo.ConsumeLater = consumeLaterList

	userListModel := models.NewUserListModel(u.Database)

	userStats, err := userListModel.GetUserListStats(uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}
	userInfo.AnimeCount = userStats.AnimeCount
	userInfo.GameCount = userStats.GameCount
	userInfo.MovieCount = userStats.MovieCount
	userInfo.TVCount = userStats.TVCount
	userInfo.MovieWatchedTime = userStats.MovieWatchedTime
	userInfo.AnimeWatchedEpisodes = userStats.AnimeWatchedEpisodes
	userInfo.TVWatchedEpisodes = userStats.TVWatchedEpisodes
	userInfo.GameTotalHoursPlayed = userStats.GameTotalHoursPlayed

	c.JSON(http.StatusOK, gin.H{"message": "Successfully fetched user info.", "data": userInfo})
}

// Update FCM Token
// @Summary Updates FCM User Token
// @Description Depending on logged in device fcm token will be updated
// @Tags user
// @Accept application/json
// @Produce application/json
// @Param changefcmtoken body requests.ChangeFCMToken true "Set token"
// @Security ApiKeyAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {string} string
// @Failure 500 {string} string
// @Router /user/token [patch]
func (u *UserController) UpdateFCMToken(c *gin.Context) {
	var data requests.ChangeFCMToken
	if shouldReturn := bindJSONData(&data, c); shouldReturn {
		return
	}

	uid := jwt.ExtractClaims(c)["id"].(string)
	userModel := models.NewUserModel(u.Database)

	user, err := userModel.FindUserByID(uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	if user.FCMToken != data.FCMToken {
		user.FCMToken = data.FCMToken
		if err = userModel.UpdateUser(user); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})

			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully updated FCM Token."})
}

// Update User Image
// @Summary Updates user image
// @Description User can update their image
// @Tags user
// @Accept application/json
// @Produce application/json
// @Param changeimage body requests.ChangeImage true "Change Image"
// @Security ApiKeyAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {string} string
// @Failure 500 {string} string
// @Router /user/image [patch]
func (u *UserController) UpdateUser(c *gin.Context) {
	var data requests.ChangeImage
	if shouldReturn := bindJSONData(&data, c); shouldReturn {
		return
	}

	uid := jwt.ExtractClaims(c)["id"].(string)
	userModel := models.NewUserModel(u.Database)

	user, err := userModel.FindUserByID(uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	if user.Image != data.Image {
		user.Image = data.Image
		if err = userModel.UpdateUser(user); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})

			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully updated image."})
}

// Change User Membership
// @Summary Change User Membership
// @Description User membership status will be updated depending on subscription status
// @Tags user
// @Accept application/json
// @Produce application/json
// @Param changemembership body requests.ChangeMembership true "Set Membership"
// @Security ApiKeyAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {string} string
// @Failure 500 {string} string
// @Router /user/membership [patch]
func (u *UserController) ChangeUserMembership(c *gin.Context) {
	var data requests.ChangeMembership
	if shouldReturn := bindJSONData(&data, c); shouldReturn {
		return
	}

	uid := jwt.ExtractClaims(c)["id"].(string)

	userModel := models.NewUserModel(u.Database)
	if err := userModel.UpdateUserMembership(uid, data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully updated membership."})
}

// Change Notification Preference
// @Summary Change User Notification Preference
// @Description Users can change their notification preference
// @Tags user
// @Accept application/json
// @Produce application/json
// @Param changenotification body requests.ChangeNotification true "Set notification"
// @Security ApiKeyAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {string} string
// @Failure 500 {string} string
// @Router /user/notification [patch]
func (u *UserController) ChangeNotificationPreference(c *gin.Context) {
	var data requests.ChangeNotification
	if shouldReturn := bindJSONData(&data, c); shouldReturn {
		return
	}

	uid := jwt.ExtractClaims(c)["id"].(string)
	userModel := models.NewUserModel(u.Database)

	user, err := userModel.FindUserByID(uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	user.AppNotification = *data.AppNotification
	user.MailNotification = *data.MailNotification

	if err = userModel.UpdateUser(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully changed notification preference."})
}

// Change Password
// @Summary Change User Password
// @Description Users can change their password
// @Tags user
// @Accept application/json
// @Produce application/json
// @Param ChangePassword body requests.ChangePassword true "Set new password"
// @Security ApiKeyAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {string} string
// @Failure 400 {string} string
// @Failure 500 {string} string
// @Router /user/password [patch]
func (u *UserController) ChangePassword(c *gin.Context) {
	var data requests.ChangePassword
	if shouldReturn := bindJSONData(&data, c); shouldReturn {
		return
	}

	uid := jwt.ExtractClaims(c)["id"].(string)
	userModel := models.NewUserModel(u.Database)

	user, err := userModel.FindUserByID(uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	if user.IsOAuthUser {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": errOAuthUser,
		})

		return
	}

	if err = utils.CheckPassword([]byte(user.Password), []byte(data.OldPassword)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"error": errPasswordNoMatch},
		})

		return
	}

	user.Password = utils.HashPassword(data.NewPassword)
	if err = userModel.UpdateUser(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully changed password."})
}

// Forgot Password
// @Summary Will be used when user forgot password
// @Description Users can change their password when they forgot
// @Tags user
// @Accept application/json
// @Produce application/json
// @Param ForgotPassword body requests.ForgotPassword true "User's email"
// @Success 200 {string} string
// @Failure 400 {string} string "Couldn't find any user"
// @Failure 500 {string} string
// @Router /user/forgot-password [post]
func (u *UserController) ForgotPassword(c *gin.Context) {
	var data requests.ForgotPassword
	if shouldReturn := bindJSONData(&data, c); shouldReturn {
		return
	}

	userModel := models.NewUserModel(u.Database)

	user, err := userModel.FindUserByEmail(data.EmailAddress)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": errNoUser,
		})

		return
	}

	if user.EmailAddress == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": errNoUser,
		})

		return
	}

	if user.IsOAuthUser {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": errOAuthUser,
		})

		return
	}

	var resetToken string
	if user.PasswordResetToken == "" {
		resetToken = uuid.NewString()
		user.PasswordResetToken = resetToken

		if err = userModel.UpdateUser(user); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})

			return
		}

		const scheduleTime = 5 * time.Minute

		time.AfterFunc(scheduleTime, func() {
			user.PasswordResetToken = ""
			userModel.UpdateUser(user)
		})
	} else {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": errMailAlreadySent,
		})

		return
	}

	if err := helpers.SendForgotPasswordEmail(resetToken, user.EmailAddress); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully send password reset email."})
}

func (u *UserController) ConfirmPasswordReset(c *gin.Context) {
	token := c.Query("token")
	email := c.Query("mail")

	userModel := models.NewUserModel(u.Database)

	user, err := userModel.FindUserByResetTokenAndEmail(token, email)
	if err != nil {
		http.ServeFile(c.Writer, c.Request, "assets/error_password_reset.html")
		return
	}

	if user.EmailAddress == "" {
		http.ServeFile(c.Writer, c.Request, "assets/error_password_reset.html")
		return
	}

	if user.IsOAuthUser {
		http.ServeFile(c.Writer, c.Request, "assets/error_password_reset.html")
		return
	}

	const (
		passwordLength = 10
		numDigits      = 4
	)

	generatedPass, err := password.Generate(passwordLength, numDigits, 0, true, false)
	if err != nil {
		generatedPass = user.EmailAddress + "_Password"
	}

	user.Password = utils.HashPassword(generatedPass)
	user.PasswordResetToken = ""

	if err = userModel.UpdateUser(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	if err := helpers.SendPasswordChangedEmail(generatedPass, user.EmailAddress); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	http.ServeFile(c.Writer, c.Request, "assets/confirm_password.html")
}

// Delete User
// @Summary Deletes user information
// @Description Deletes everything related to user
// @Tags user
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {string} string
// @Error 500 {string} string
// @Router /user [delete]
func (u *UserController) DeleteUser(c *gin.Context) {
	uid := jwt.ExtractClaims(c)["id"].(string)

	userModel := models.NewUserModel(u.Database)
	if err := userModel.DeleteUserByID(uid); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	userListModel := models.NewUserListModel(u.Database)
	userInteractionModel := models.NewUserInteractionModel(u.Database)
	logsModel := models.NewLogsModel(u.Database)

	go userListModel.DeleteUserListByUserID(uid)
	go userInteractionModel.DeleteAllConsumeLaterByUserID(uid)
	go logsModel.DeleteLogsByUserID(uid)

	c.JSON(http.StatusOK, gin.H{"message": "Successfully deleted user."})
}
