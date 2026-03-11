package controllers

import (
	"fmt"
	"go-api/models"
	"go-api/requests"
	"go-api/resources"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func (idb *InDB) Register(c *gin.Context) {
	var req requests.RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	user := models.User{
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Phone:     req.Phone,
		Username:  req.Username,
		Password:  string(hashedPassword),
		Status:    1,
		CreatedAt: time.Now().Unix(),
		UpdatedAt: time.Now().Unix(),
	}

	if err := idb.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Registration successful"})
}

func (idb *InDB) Profile(c *gin.Context) {
	userID, _ := c.Get("user_id")
	fmt.Println("user", userID)
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
