package main

import (
	"time"
	"gorm.io/gorm"
)

// Film represents a movie with its details and standard database columns
// @Description Film information
type Film struct {
	ID        uint           `json:"id" gorm:"primarykey" example:"1"`
	Title     string         `json:"title" gorm:"not null" example:"The Shawshank Redemption"`
	Director  string         `json:"director" gorm:"not null" example:"Frank Darabont"`
	Year      int            `json:"year" gorm:"not null" example:"1994"`
	Genre     string         `json:"genre" example:"Drama"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// User represents a user from database with standard columns
// @Description User information
type User struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	Username  string         `json:"username" gorm:"uniqueIndex;not null"`
	Password  string         `json:"-" gorm:"not null"` // Hide password in JSON responses
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// LoginRequest represents login request payload
// @Description Login request payload
type LoginRequest struct {
	Username string `json:"username" example:"admin"`
	Password string `json:"password" example:"admin123"`
}

// LoginResponse represents login response
// @Description Login response with token
type LoginResponse struct {
	Token string `json:"token" example:"abc123def456"`
}

// FilmRequest represents film creation/update request
// @Description Film request payload
type FilmRequest struct {
	Title    string `json:"title" example:"The Shawshank Redemption"`
	Director string `json:"director" example:"Frank Darabont"`
	Year     int    `json:"year" example:"1994"`
	Genre    string `json:"genre" example:"Drama"`
}

// ErrorResponse represents error response
// @Description Error response
type ErrorResponse struct {
	Error string `json:"error" example:"Invalid request"`
}

// SuccessResponse represents success response
// @Description Success response
type SuccessResponse struct {
	Message string `json:"message" example:"Operation completed successfully"`
}
