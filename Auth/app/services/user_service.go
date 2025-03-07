package services

import (
	"LService/models"
	"time"
)

func CreateUser(user *models.User) {
	models.DB.Create(user)
}

func GetUserProfile(login string) (*models.User, error) {
	var user models.User
	if err := models.DB.Where("login = ?", login).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func UpdateUserProfile(login string, updateData map[string]interface{}) (*models.User, error) {
	var user models.User
	if err := models.DB.Where("login = ?", login).First(&user).Error; err != nil {
		return nil, err
	}
	delete(updateData, "Login")
	delete(updateData, "Password")
	updateData["UpdatedAt"] = time.Now()
	models.DB.Model(&user).Updates(updateData)
	return &user, nil
}
