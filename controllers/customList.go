package controllers

import (
	"app/db"
	"app/models"
	"app/requests"
	"app/responses"
	"net/http"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

type CustomListController struct {
	Database *db.MongoDB
}

func NewCustomListController(mongoDB *db.MongoDB) CustomListController {
	return CustomListController{
		Database: mongoDB,
	}
}

const errNoContent = "You need to add at least 1 entry."
const errCustomListPremium = "Free members can create up to 5 lists, you can get premium membership to create more."
const errCustomListPremiumLimit = "You've reached your limit, sorry 🙏"
const errCustomListContentPremium = "Free members can create up to 10 content to their list, you can get premium membership to add more."

// Create Custom List
// @Summary Create Custom List
// @Description Creates Custom List
// @Tags custom_list
// @Accept application/json
// @Produce application/json
// @Param createcustomlist body requests.CreateCustomList true "Create Custom List"
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 201 {object} models.CustomList
// @Failure 404 {string} string
// @Failure 500 {string} string
// @Router /custom-list [post]
func (cl *CustomListController) CreateCustomList(c *gin.Context) {
	var data requests.CreateCustomList
	if shouldReturn := bindJSONData(&data, c); shouldReturn {
		return
	}

	uid := jwt.ExtractClaims(c)["id"].(string)
	userModel := models.NewUserModel(cl.Database)
	customListModel := models.NewCustomListModel(cl.Database)

	isPremium, _ := userModel.IsUserPremium(uid)
	count := customListModel.GetCustomListCount(uid)

	if len(data.Content) <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": errNoContent,
		})

		return
	}

	if (isPremium && count >= models.CustomListPremiumLimit) || (!isPremium && count >= models.CustomListFreeLimit) {
		if isPremium {
			c.JSON(http.StatusForbidden, gin.H{
				"error": errCustomListPremiumLimit,
			})

			return
		}

		c.JSON(http.StatusForbidden, gin.H{
			"error": errCustomListPremium,
		})

		return
	}

	var (
		createdCustomList models.CustomList
		err               error
	)

	if createdCustomList, err = customListModel.CreateCustomList(uid, data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Successfully created.", "data": createdCustomList})
}

// Update Custom List
// @Summary Update Custom List Details
// @Description Updates Custom List Name, Description, Privacy Status
// @Tags custom_list
// @Accept application/json
// @Produce application/json
// @Param updatecustomlist body requests.UpdateCustomList true "Update Custom List"
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {object} models.CustomList
// @Failure 404 {string} string
// @Failure 500 {string} string
// @Router /custom-list [patch]
func (cl *CustomListController) UpdateCustomList(c *gin.Context) {
	var data requests.UpdateCustomList
	if shouldReturn := bindJSONData(&data, c); shouldReturn {
		return
	}

	uid := jwt.ExtractClaims(c)["id"].(string)
	customListModel := models.NewCustomListModel(cl.Database)

	if len(data.Content) <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": errNoContent,
		})

		return
	}

	var (
		updatedCustomList models.CustomList
		err               error
	)

	customList, err := customListModel.GetBaseCustomList(uid, data.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	if customList.UserID == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": ErrNotFound})
		return
	}

	if updatedCustomList, err = customListModel.UpdateCustomList(uid, data, customList); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully updated.", "data": updatedCustomList})
}

// Add Content to Custom List
// @Summary Add Content to Custom List
// @Description Add New Content to Custom List
// @Tags custom_list
// @Accept application/json
// @Produce application/json
// @Param addtocustomlist body requests.AddToCustomList true "Add To Custom List"
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {object} models.CustomList
// @Failure 404 {string} string
// @Failure 500 {string} string
// @Router /custom-list/add [patch]
func (cl *CustomListController) UpdateAddContentToCustomList(c *gin.Context) {
	var data requests.AddToCustomList
	if shouldReturn := bindJSONData(&data, c); shouldReturn {
		return
	}

	uid := jwt.ExtractClaims(c)["id"].(string)
	customListModel := models.NewCustomListModel(cl.Database)

	var (
		updatedCustomList models.CustomList
		err               error
	)

	customList, err := customListModel.GetBaseCustomList(uid, data.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	if customList.UserID == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": ErrNotFound})
		return
	}

	if updatedCustomList, err = customListModel.UpdateAddContentToCustomList(uid, data, customList); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully added.", "data": updatedCustomList})
}

// Like/Dislike Custom List
// @Summary Like/Dislike Custom List
// @Description Like and Dislike Custom List
// @Tags custom_list
// @Accept application/json
// @Produce application/json
// @Param id body requests.ID true "ID"
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {object} responses.CustomList
// @Failure 404 {string} string
// @Failure 500 {string} string
// @Router /custom-list/like [patch]
func (cl *CustomListController) LikeCustomList(c *gin.Context) {
	var data requests.ID
	if shouldReturn := bindJSONData(&data, c); shouldReturn {
		return
	}

	uid := jwt.ExtractClaims(c)["id"].(string)
	customListModel := models.NewCustomListModel(cl.Database)

	var (
		updatedCustomList responses.CustomList
		err               error
	)

	customList, err := customListModel.GetBaseCustomListResponse(&uid, data.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	if customList.UserID == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": ErrNotFound})
		return
	}

	if customList.UserID == uid {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "You cannot like your own list.",
		})

		return
	}

	if updatedCustomList, err = customListModel.LikeCustomList(uid, data, customList); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully liked.", "data": updatedCustomList})
}

// Bookmark Custom List
// @Summary Bookmark Custom List
// @Description Bookmark Custom List
// @Tags custom_list
// @Accept application/json
// @Produce application/json
// @Param id body requests.ID true "ID"
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {object} responses.CustomList
// @Failure 404 {string} string
// @Failure 500 {string} string
// @Router /custom-list/bookmark [patch]
func (cl *CustomListController) BookmarkCustomList(c *gin.Context) {
	var data requests.ID
	if shouldReturn := bindJSONData(&data, c); shouldReturn {
		return
	}

	uid := jwt.ExtractClaims(c)["id"].(string)
	customListModel := models.NewCustomListModel(cl.Database)

	var (
		updatedCustomList responses.CustomList
		err               error
	)

	customList, err := customListModel.GetBaseCustomListResponse(&uid, data.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	if customList.UserID == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": ErrNotFound})
		return
	}

	if customList.UserID == uid {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "You cannot bookmark your own list.",
		})

		return
	}

	if updatedCustomList, err = customListModel.BookmarkCustomList(uid, data, customList); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully bookmarked.", "data": updatedCustomList})
}

// Get Custom List by User ID
// @Summary Get Custom List by User ID
// @Description Get Custom List by User ID with or without authentication
// @Tags custom_list
// @Accept application/json
// @Produce application/json
// @Param sortcustomlistuid body requests.SortCustomListUID true "Sort Custom List UID"
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {array} responses.CustomList
// @Failure 404 {string} string
// @Failure 500 {string} string
// @Router /custom-list [get]
func (cl *CustomListController) GetCustomListsByUserID(c *gin.Context) {
	var data requests.SortCustomListUID
	if err := c.ShouldBindQuery(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": validatorErrorHandler(err),
		})

		return
	}

	uid, OK := c.Get("uuid")
	customListModel := models.NewCustomListModel(cl.Database)

	if OK && uid != nil {
		userId := uid.(string)

		customLists, err := customListModel.GetCustomListsByUserID(&userId, data, false, false)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})

			return
		}

		c.JSON(http.StatusOK, gin.H{"data": customLists})
	} else {
		customLists, err := customListModel.GetCustomListsByUserID(nil, data, false, false)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})

			return
		}

		c.JSON(http.StatusOK, gin.H{"data": customLists})
	}
}

// Get Custom List
// @Summary Get Custom List
// @Description Get Custom List with or without authentication
// @Tags custom_list
// @Accept application/json
// @Produce application/json
// @Param sortcustomlist body requests.SortCustomList true "Sort Custom List"
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {array} responses.CustomList
// @Failure 404 {string} string
// @Failure 500 {string} string
// @Router /custom-list/social [get]
func (cl *CustomListController) GetCustomLists(c *gin.Context) {
	var data requests.SortCustomList
	if err := c.ShouldBindQuery(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": validatorErrorHandler(err),
		})

		return
	}

	uid, OK := c.Get("uuid")
	customListModel := models.NewCustomListModel(cl.Database)

	if OK && uid != nil {
		userId := uid.(string)

		customLists, err := customListModel.GetCustomLists(&userId, data)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})

			return
		}

		c.JSON(http.StatusOK, gin.H{"data": customLists})
	} else {
		customLists, err := customListModel.GetCustomLists(nil, data)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})

			return
		}

		c.JSON(http.StatusOK, gin.H{"data": customLists})
	}
}

// Get Liked Custom List
// @Summary Get Liked Custom List
// @Description Get Liked Custom List
// @Tags custom_list
// @Accept application/json
// @Produce application/json
// @Param id body requests.SortLikeBookmarkCustomList true "ID"
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {array} responses.CustomList
// @Failure 404 {string} string
// @Failure 500 {string} string
// @Router /custom-list/liked [get]
func (cl *CustomListController) GetLikedCustomLists(c *gin.Context) {
	var data requests.SortLikeBookmarkCustomList
	if err := c.ShouldBindQuery(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": validatorErrorHandler(err),
		})

		return
	}

	uid := jwt.ExtractClaims(c)["id"].(string)
	customListModel := models.NewCustomListModel(cl.Database)

	likedCustomLists, err := customListModel.GetLikedCustomLists(uid, data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{"data": likedCustomLists})
}

// Get Bookmarked Custom List
// @Summary Get Bookmarked Custom List
// @Description Get Bookmarked Custom List
// @Tags custom_list
// @Accept application/json
// @Produce application/json
// @Param id body requests.SortLikeBookmarkCustomList true "ID"
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {array} responses.CustomList
// @Failure 404 {string} string
// @Failure 500 {string} string
// @Router /custom-list/bookmarked [get]
func (cl *CustomListController) GetBookmarkedCustomLists(c *gin.Context) {
	var data requests.SortLikeBookmarkCustomList
	if err := c.ShouldBindQuery(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": validatorErrorHandler(err),
		})

		return
	}

	uid := jwt.ExtractClaims(c)["id"].(string)
	customListModel := models.NewCustomListModel(cl.Database)

	bookmarkedCustomLists, err := customListModel.GetBookmarkedCustomLists(uid, data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{"data": bookmarkedCustomLists})
}

// Get Custom List Details
// @Summary Get Custom List Details
// @Description Get Custom List Details with or without authentication
// @Tags custom_list
// @Accept application/json
// @Produce application/json
// @Param id body requests.ID true "ID"
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {array} responses.CustomList
// @Failure 404 {string} string
// @Failure 500 {string} string
// @Router /custom-list/details [get]
func (cl *CustomListController) GetCustomListDetails(c *gin.Context) {
	var data requests.ID
	if err := c.ShouldBindQuery(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": validatorErrorHandler(err),
		})

		return
	}

	uid, OK := c.Get("uuid")
	customListModel := models.NewCustomListModel(cl.Database)

	if OK && uid != nil {
		userId := uid.(string)

		customListDetails, err := customListModel.GetCustomListDetails(&userId, data.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})

			return
		}

		c.JSON(http.StatusOK, gin.H{"data": customListDetails})
	} else {
		customListDetails, err := customListModel.GetCustomListDetails(nil, data.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})

			return
		}

		c.JSON(http.StatusOK, gin.H{"data": customListDetails})
	}
}

// Delete Bulk Content From Custom List
// @Summary Delete Bulk Content From Custom List
// @Description Deletes Bulk Content from Custom List
// @Tags custom_list
// @Accept application/json
// @Produce application/json
// @Param bulkdeletecustomlist body requests.BulkDeleteCustomList true "Bulk Delete Custom List"
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {object} models.CustomList
// @Failure 500 {string} string
// @Router /custom-list/content [delete]
func (cl *CustomListController) DeleteBulkContentFromCustomListByID(c *gin.Context) {
	var data requests.BulkDeleteCustomList
	if shouldReturn := bindJSONData(&data, c); shouldReturn {
		return
	}

	uid := jwt.ExtractClaims(c)["id"].(string)
	customListModel := models.NewCustomListModel(cl.Database)

	isDeleted, err := customListModel.DeleteBulkContentFromCustomListByID(uid, data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	if !isDeleted {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to delete.",
		})

		return
	}

	customList, err := customListModel.GetBaseCustomList(uid, data.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully deleted.", "data": customList})
}

// Delete Custom List
// @Summary Delete Custom List
// @Description Deletes Custom List
// @Tags custom_list
// @Accept application/json
// @Produce application/json
// @Param id body requests.ID true "ID"
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {string} string
// @Failure 404 {string} string
// @Failure 500 {string} string
// @Router /custom-list [delete]
func (cl *CustomListController) DeleteCustomListByID(c *gin.Context) {
	var data requests.ID
	if shouldReturn := bindJSONData(&data, c); shouldReturn {
		return
	}

	uid := jwt.ExtractClaims(c)["id"].(string)
	customListModel := models.NewCustomListModel(cl.Database)

	isDeleted, err := customListModel.DeleteCustomListByID(uid, data.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	if isDeleted {
		c.JSON(http.StatusOK, gin.H{"message": "Custom list deleted successfully."})

		return
	}

	c.JSON(http.StatusNotFound, gin.H{"error": ErrNotFound})
}
