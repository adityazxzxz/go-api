package controllers

import (
	"go-api/models"
	"go-api/requests"
	"go-api/resources"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func (idb *InDB) Profile(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var profile resources.Profile

	err := idb.DB.Table("users u").
		Select("*").
		Where("id", userID).
		Scan(&profile).Error

	if err != nil {
		result := resources.Response{
			Error:   true,
			Message: "Internal Error",
			Data:    nil,
		}
		c.JSON(http.StatusInternalServerError, result)
		return
	}

	if (profile == resources.Profile{}) {
		result := resources.Response{
			Error:   true,
			Message: "Data not found",
			Data:    nil,
		}
		c.JSON(http.StatusBadRequest, result)
		return
	}

	response := resources.Response{
		Error:   false,
		Message: "Profile successful",
		Data:    profile,
	}

	c.JSON(http.StatusOK, response)
}

func (idb *InDB) UpdateProfile(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var profile models.User
	var request requests.Update

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	err := idb.DB.First(&profile, userID).Error

	if err != nil {
		result := resources.Response{
			Error:   true,
			Message: "Internal Error",
			Data:    nil,
		}
		c.JSON(http.StatusInternalServerError, result)
		return
	}

	if (profile == models.User{}) {
		result := resources.Response{
			Error:   true,
			Message: "Data not found",
			Data:    nil,
		}
		c.JSON(http.StatusBadRequest, result)
		return
	}

	now := time.Now()
	updates := models.User{
		FirstName: request.FirstName,
		LastName:  request.LastName,
		UpdatedAt: now.Unix(),
	}

	err = idb.DB.Model(&profile).Updates(updates).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, resources.Response{
			Error:   true,
			Message: "Update profile failed",
			Data:    nil,
		})

		return
	}

	response := resources.Response{
		Error:   false,
		Message: "Update profile successful",
		Data:    profile,
	}

	c.JSON(http.StatusOK, response)
}
