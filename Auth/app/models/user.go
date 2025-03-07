package models

import (
	"time"

	"gorm.io/gorm"
)

var DB *gorm.DB

type User struct {
	ID        uint   `gorm:"primaryKey"`
	Login     string `gorm:"unique;not null"`
	Password  string `gorm:"not null"`
	Email     string `gorm:"unique;not null"`
	FirstName string
	LastName  string
	BirthDate string
	Phone     string
	CreatedAt time.Time
	UpdatedAt time.Time
}
