package main

import (
	"errors"
	"gorm.io/gorm"
)

// FilmService handles film-related database operations
type FilmService struct {
	db *gorm.DB
}

// NewFilmService creates a new film service
func NewFilmService(db *gorm.DB) *FilmService {
	return &FilmService{db: db}
}

// GetAllFilms retrieves all films from database
func (fs *FilmService) GetAllFilms() ([]Film, error) {
	var films []Film
	err := fs.db.Find(&films).Error
	return films, err
}

// GetFilmByID retrieves a film by ID
func (fs *FilmService) GetFilmByID(id uint) (*Film, error) {
	var film Film
	err := fs.db.First(&film, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("film not found")
		}
		return nil, err
	}
	return &film, nil
}

// CreateFilm creates a new film
func (fs *FilmService) CreateFilm(filmReq FilmRequest) (*Film, error) {
	film := Film{
		Title:    filmReq.Title,
		Director: filmReq.Director,
		Year:     filmReq.Year,
		Genre:    filmReq.Genre,
	}
	
	err := fs.db.Create(&film).Error
	if err != nil {
		return nil, err
	}
	
	return &film, nil
}

// UpdateFilm updates an existing film
func (fs *FilmService) UpdateFilm(id uint, filmReq FilmRequest) (*Film, error) {
	var film Film
	err := fs.db.First(&film, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("film not found")
		}
		return nil, err
	}
	
	// Update fields
	film.Title = filmReq.Title
	film.Director = filmReq.Director
	film.Year = filmReq.Year
	film.Genre = filmReq.Genre
	
	err = fs.db.Save(&film).Error
	if err != nil {
		return nil, err
	}
	
	return &film, nil
}

// DeleteFilm soft deletes a film
func (fs *FilmService) DeleteFilm(id uint) error {
	result := fs.db.Delete(&Film{}, id)
	if result.Error != nil {
		return result.Error
	}
	
	if result.RowsAffected == 0 {
		return errors.New("film not found")
	}
	
	return nil
}

// UserService handles user-related database operations
type UserService struct {
	db *gorm.DB
}

// NewUserService creates a new user service
func NewUserService(db *gorm.DB) *UserService {
	return &UserService{db: db}
}

// GetUserByUsername retrieves a user by username
func (us *UserService) GetUserByUsername(username string) (*User, error) {
	var user User
	err := us.db.Where("username = ?", username).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

// ValidateUser validates user credentials
func (us *UserService) ValidateUser(username, password string) bool {
	user, err := us.GetUserByUsername(username)
	if err != nil {
		return false
	}
	return user.Password == password
}

// CreateUser creates a new user (for future use)
func (us *UserService) CreateUser(username, password string) (*User, error) {
	user := User{
		Username: username,
		Password: password,
	}
	
	err := us.db.Create(&user).Error
	if err != nil {
		return nil, err
	}
	
	return &user, nil
}

// SeedUsers creates initial users if they don't exist
func (us *UserService) SeedUsers() error {
	users := []User{
		{Username: "admin", Password: "admin123"},
		{Username: "user1", Password: "password123"},
		{Username: "demo", Password: "demo456"},
	}
	
	for _, user := range users {
		var existingUser User
		err := us.db.Where("username = ?", user.Username).First(&existingUser).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// User doesn't exist, create it
			if err := us.db.Create(&user).Error; err != nil {
				return err
			}
		}
	}
	
	return nil
}
