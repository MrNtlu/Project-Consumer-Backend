package controllers

import (
	"app/db"
	"app/helpers"
	"app/models"
	"app/requests"
	"app/responses"
	"app/utils"
	"math"
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
	errPremiumFeature    = "This feature requires premium membership."
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

// Extra Statistics
// @Summary Extra statistics
// @Description Returns extra statistics
// @Tags user
// @Accept application/json
// @Produce application/json
// @Param logstatinterval body requests.LogStatInterval true "Log Stat Interval"
// @Security ApiKeyAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {object} responses.ExtraStatistics "Extra Statistics"
// @Router /user/stats [get]
func (u *UserController) GetExtraStatistics(c *gin.Context) {
	var data requests.LogStatInterval
	if err := c.ShouldBindQuery(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": validatorErrorHandler(err),
		})

		return
	}

	uid := jwt.ExtractClaims(c)["id"].(string)

	userModel := models.NewUserModel(u.Database)
	logsModel := models.NewLogsModel(u.Database)

	isPremium, _ := userModel.IsUserPremium(uid)

	if !isPremium && data.Interval != "weekly" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": errPremiumFeature,
		})

		return
	}

	mostLikedGenres, err := logsModel.MostLikedGenresByLogs(uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	finishedLogStats, err := logsModel.FinishedLogStats(uid, data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	statistics := responses.ExtraStatistics{
		FinishedLogStats: finishedLogStats,
		MostLikedGenres:  mostLikedGenres,
	}

	c.JSON(http.StatusOK, gin.H{"data": statistics})
}

// User Info
// @Summary User basic info
// @Description Returns basic user information
// @Tags user
// @Accept application/json
// @Produce application/json
// @Security ApiKeyAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {object} responses.User "User"
// @Router /user/basic [get]
func (u *UserController) GetBasicUserInfo(c *gin.Context) {
	uid := jwt.ExtractClaims(c)["id"].(string)

	userModel := models.NewUserModel(u.Database)
	userInfo, err := userModel.GetUserByID(uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	logsModel := models.NewLogsModel(u.Database)

	_, currentStreak := logsModel.GetLogStreak(uid)
	userInfo.Streak = currentStreak

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
	logsModel := models.NewLogsModel(u.Database)

	userInfo, err := userModel.GetUserInfo("", uid, false)
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

	reviewsModel := models.NewReviewModel(u.Database)
	reviews, _, err := reviewsModel.GetReviewsByUserID(&uid, requests.SortReviewByUserID{
		UserID: uid,
		Sort:   "popularity",
		Page:   1,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}
	userInfo.Reviews = reviews

	userListModel := models.NewUserListModel(u.Database)
	userStats, err := userListModel.GetUserListStats(uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	friendModel := models.NewFriendModel(u.Database)
	friendRequestCount, err := friendModel.FriendRequestCount(uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}
	userInfo.FriendRequestCount = friendRequestCount

	maxStreak, currentStreak := logsModel.GetLogStreak(uid)
	userInfo.MaxStreak = maxStreak
	userInfo.Streak = currentStreak

	userInfo.IsFriendRequestSent = false
	userInfo.IsFriendRequestReceived = false
	userInfo.IsFriendsWith = false
	userInfo.AnimeCount = userStats.AnimeCount
	userInfo.GameCount = userStats.GameCount
	userInfo.MovieCount = userStats.MovieCount
	userInfo.TVCount = userStats.TVCount
	userInfo.MovieWatchedTime = userStats.MovieWatchedTime
	userInfo.AnimeWatchedEpisodes = userStats.AnimeWatchedEpisodes
	userInfo.TVWatchedEpisodes = userStats.TVWatchedEpisodes
	userInfo.GameTotalHoursPlayed = userStats.GameTotalHoursPlayed

	if userStats.MovieCount != 0 && userStats.MovieTotalScore != 0 {
		userInfo.MovieTotalScore = math.Round(float64(userStats.MovieTotalScore)/float64(userStats.MovieCount)*100) / 100
	} else {
		userInfo.MovieTotalScore = 0
	}

	if userStats.TVCount != 0 && userStats.TVTotalScore != 0 {
		userInfo.TVTotalScore = math.Round((float64(userStats.TVTotalScore)/float64(userStats.TVCount))*100) / 100
	} else {
		userInfo.TVTotalScore = 0
	}

	if userStats.AnimeCount != 0 && userStats.AnimeTotalScore != 0 {
		userInfo.AnimeTotalScore = math.Round((float64(userStats.AnimeTotalScore)/float64(userStats.AnimeCount))*100) / 100
	} else {
		userInfo.AnimeTotalScore = 0
	}

	if userStats.GameCount != 0 && userStats.GameTotalScore != 0 {
		userInfo.GameTotalScore = math.Round((float64(userStats.GameTotalScore)/float64(userStats.GameCount))*100) / 100
	} else {
		userInfo.GameTotalScore = 0
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully fetched user info.", "data": userInfo})
}

// Get Friends
// @Summary Get friends
// @Description Returns friends by user id
// @Tags user
// @Accept application/json
// @Produce application/json
// @Security ApiKeyAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {array} models.User "User"
// @Router /user/friends [get]
func (u *UserController) GetFriends(c *gin.Context) {
	uid := jwt.ExtractClaims(c)["id"].(string)

	userModel := models.NewUserModel(u.Database)

	friends, _ := userModel.GetUserFriends(uid)

	var result []models.User

	if friends != nil {
		result = friends
	} else {
		result = []models.User{}
	}

	c.JSON(http.StatusOK, gin.H{"data": result})
}

// Get Friend Requests
// @Summary Get friend requests
// @Description Returns friend requests by user id
// @Tags user
// @Accept application/json
// @Produce application/json
// @Security ApiKeyAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {array} responses.FriendRequest "Friend Request"
// @Router /user/requests [get]
func (u *UserController) GetFriendRequests(c *gin.Context) {
	uid := jwt.ExtractClaims(c)["id"].(string)

	friendModel := models.NewFriendModel(u.Database)

	friendRequests, err := friendModel.GetFriendRequests(uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{"data": friendRequests})
}

// User Info From Username
// @Summary User info from username
// @Description Returns users stats from username
// @Tags user
// @Accept application/json
// @Produce application/json
// @Security ApiKeyAuth
// @Param getprofile body requests.GetProfile true "Get Profile"
// @Param Authorization header string true "Authentication header"
// @Success 200 {object} responses.UserInfo "User Info"
// @Router /user/profile [get]
func (u *UserController) GetUserInfoFromUsername(c *gin.Context) {
	var data requests.GetProfile
	if err := c.ShouldBindQuery(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": validatorErrorHandler(err),
		})

		return
	}

	userModel := models.NewUserModel(u.Database)
	userListModel := models.NewUserListModel(u.Database)
	reviewsModel := models.NewReviewModel(u.Database)
	customListModel := models.NewCustomListModel(u.Database)
	friendModel := models.NewFriendModel(u.Database)
	logsModel := models.NewLogsModel(u.Database)

	var (
		userInfo responses.UserInfo
		err      error
	)

	userId := jwt.ExtractClaims(c)["id"].(string)

	userInfo, err = userModel.GetUserInfo(data.Username, userId, true)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	if userInfo.Username == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": errNoUser,
		})

		return
	}

	reviews, _, err := reviewsModel.GetReviewsByUserID(&userId, requests.SortReviewByUserID{
		UserID: userInfo.ID.Hex(),
		Sort:   "popularity",
		Page:   1,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}
	userInfo.Reviews = reviews

	customLists, err := customListModel.GetCustomListsByUserID(&userId, requests.SortCustomListUID{
		UserID: userInfo.ID.Hex(),
		Sort:   "popularity",
	}, true, false)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}
	userInfo.CustomLists = customLists

	isFriendsWith, err := userModel.IsFriendsWith(userInfo.ID.Hex(), userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}
	userInfo.IsFriendsWith = isFriendsWith

	isFriendRequestSent, err := friendModel.IsFriendRequestSent(userId, userInfo.ID.Hex())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}
	userInfo.IsFriendRequestSent = isFriendRequestSent

	isFriendRequestReceived, err := friendModel.IsFriendRequestReceived(userInfo.ID.Hex(), userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}
	userInfo.IsFriendRequestReceived = isFriendRequestReceived

	userLevel, _ := userModel.GetUserLevel(userInfo.ID.Hex())
	userInfo.Level = userLevel

	userStats, err := userListModel.GetUserListStats(userInfo.ID.Hex())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	maxStreak, currentStreak := logsModel.GetLogStreak(userInfo.ID.Hex())
	userInfo.MaxStreak = maxStreak
	userInfo.Streak = currentStreak

	userInfo.FriendRequestCount = 0
	userInfo.AnimeCount = userStats.AnimeCount
	userInfo.GameCount = userStats.GameCount
	userInfo.MovieCount = userStats.MovieCount
	userInfo.TVCount = userStats.TVCount
	userInfo.MovieWatchedTime = userStats.MovieWatchedTime
	userInfo.AnimeWatchedEpisodes = userStats.AnimeWatchedEpisodes
	userInfo.TVWatchedEpisodes = userStats.TVWatchedEpisodes
	userInfo.GameTotalHoursPlayed = userStats.GameTotalHoursPlayed

	if userStats.MovieCount != 0 && userStats.MovieTotalScore != 0 {
		userInfo.MovieTotalScore = math.Round(float64(userStats.MovieTotalScore)/float64(userStats.MovieCount)*100) / 100
	} else {
		userInfo.MovieTotalScore = 0
	}

	if userStats.TVCount != 0 && userStats.TVTotalScore != 0 {
		userInfo.TVTotalScore = math.Round((float64(userStats.TVTotalScore)/float64(userStats.TVCount))*100) / 100
	} else {
		userInfo.TVTotalScore = 0
	}

	if userStats.AnimeCount != 0 && userStats.AnimeTotalScore != 0 {
		userInfo.AnimeTotalScore = math.Round((float64(userStats.AnimeTotalScore)/float64(userStats.AnimeCount))*100) / 100
	} else {
		userInfo.AnimeTotalScore = 0
	}

	if userStats.GameCount != 0 && userStats.GameTotalScore != 0 {
		userInfo.GameTotalScore = math.Round((float64(userStats.GameTotalScore)/float64(userStats.GameCount))*100) / 100
	} else {
		userInfo.GameTotalScore = 0
	}

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

// Answer Friend Request
// @Summary Answer Friend Request
// @Description Response friend request object
// @Tags user
// @Accept application/json
// @Produce application/json
// @Param answerfriendrequest body requests.AnswerFriendRequest true "Answer Friend Request"
// @Security ApiKeyAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {string} string
// @Failure 500 {string} string
// @Router /user/request-answer [post]
func (u *UserController) AnswerFriendRequest(c *gin.Context) {
	var data requests.AnswerFriendRequest
	if shouldReturn := bindJSONData(&data, c); shouldReturn {
		return
	}

	uid := jwt.ExtractClaims(c)["id"].(string)
	userModel := models.NewUserModel(u.Database)
	friendModel := models.NewFriendModel(u.Database)

	friendRequest, err := friendModel.GetFriendRequest(data.ID, uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	if friendRequest.ReceiverID == "" {
		c.JSON(http.StatusNotFound, gin.H{
			"error": ErrNotFound,
		})

		return
	}

	if friendRequest.ReceiverID != uid {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": ErrUnauthorized,
		})

		return
	}

	if data.Answer == 0 || data.Answer == 1 {
		if err := friendModel.DeleteFriendRequest(data.ID, uid); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})

			return
		}

		if data.Answer == 1 { //Accept
			userModel.InsertFriend(friendRequest.SenderID, friendRequest.ReceiverID)
			userModel.InsertFriend(friendRequest.ReceiverID, friendRequest.SenderID)

			c.JSON(http.StatusOK, gin.H{"message": "Successfully accepted friend request."})

			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Successfully denied friend request."})
	} else {
		if err := friendModel.IgnoreFriendRequest(data.ID, uid); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})

			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Successfully ignored friend request."})
	}
}

// Send Friend Request
// @Summary Send Friend Request
// @Description Creates user request object and send notification
// @Tags user
// @Accept application/json
// @Produce application/json
// @Param sendfriendrequest body requests.SendFriendRequest true "Send Friend Request"
// @Security ApiKeyAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {string} string
// @Failure 500 {string} string
// @Router /user/request [post]
func (u *UserController) SendFriendRequest(c *gin.Context) {
	var data requests.SendFriendRequest
	if shouldReturn := bindJSONData(&data, c); shouldReturn {
		return
	}

	uid := jwt.ExtractClaims(c)["id"].(string)
	userModel := models.NewUserModel(u.Database)

	sender, err := userModel.FindUserByID(uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	receiver, err := userModel.FindUserByUsername(data.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	if sender.EmailAddress == "" || receiver.EmailAddress == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": errNoUser,
		})

		return
	}

	friendModel := models.NewFriendModel(u.Database)

	if err := friendModel.CreateFriendRequest(sender.ID.Hex(), sender.Username, receiver.ID.Hex(), sender.Username); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	if receiver.AppNotification.FriendRequest {
		go helpers.SendNotification(
			receiver.FCMToken,
			"New Friend Request",
			sender.Username+" has sent you a friend request. Do you want to connect?",
			"https://watchlistfy.com/friend-requests",
			nil, nil,
		)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully sent friend request."})
}

// Change Username
// @Summary Change Username
// @Description Users can change their username
// @Tags user
// @Accept application/json
// @Produce application/json
// @Param changeusername body requests.ChangeUsername true "Change Username"
// @Security ApiKeyAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {string} string
// @Failure 500 {string} string
// @Router /user/username [patch]
func (u *UserController) ChangeUsername(c *gin.Context) {
	var data requests.ChangeUsername
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

	if user.CanChangeUsername {
		user.Username = data.Username
		user.CanChangeUsername = false

		if err = userModel.UpdateUser(user); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})

			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Successfully updated username."})

		return
	} else {
		c.JSON(http.StatusMethodNotAllowed, gin.H{
			"error": "You are not allowed to change your username.",
		})

		return
	}
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

	user.AppNotification.FriendRequest = *data.FriendRequest
	user.AppNotification.ReviewLikes = *data.ReviewLikes

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
			"error": errPasswordNoMatch,
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
	friendModel := models.NewFriendModel(u.Database)
	suggestionModel := models.NewAISuggestionsModel(u.Database)
	customListsModel := models.NewCustomListModel(u.Database)
	reviewsModel := models.NewReviewModel(u.Database)
	logsModel := models.NewLogsModel(u.Database)

	go userListModel.DeleteUserListByUserID(uid)
	go userInteractionModel.DeleteAllConsumeLaterByUserID(uid)
	go friendModel.DeleteAllFriendRequestsByUserID(uid)
	go reviewsModel.DeleteAllReviewsByUserID(uid)
	go customListsModel.DeleteAllCustomListsByUserID(uid)
	go suggestionModel.DeleteAllAISuggestionsByUserID(uid)
	go logsModel.DeleteLogsByUserID(uid)

	c.JSON(http.StatusOK, gin.H{"message": "Successfully deleted user."})
}
