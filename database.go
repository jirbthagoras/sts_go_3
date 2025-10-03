package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// GetDatabaseConfig returns database configuration from environment variables or defaults
func GetDatabaseConfig() DatabaseConfig {
	return DatabaseConfig{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnv("DB_PORT", "5432"),
		User:     getEnv("DB_USER", "postgres"),
		Password: getEnv("DB_PASSWORD", "passsword"),
		DBName:   getEnv("DB_NAME", "postgres"),
		SSLMode:  getEnv("DB_SSLMODE", "disable"),
	}
}

// getEnv gets environment variable or returns default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// ConnectDatabase establishes connection to PostgreSQL database
func ConnectDatabase() (*gorm.DB, error) {
	config := GetDatabaseConfig()

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=Asia/Jakarta",
		config.Host, config.User, config.Password, config.DBName, config.Port, config.SSLMode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	log.Println("✅ Successfully connected to PostgreSQL database")
	return db, nil
}

// MigrateDatabase runs database migrations
func MigrateDatabase(db *gorm.DB) error {
	log.Println("🔄 Running database migrations...")

	err := db.AutoMigrate(&Film{}, &User{})
	if err != nil {
		return fmt.Errorf("failed to migrate database: %v", err)
	}

	log.Println("✅ Database migrations completed successfully")
	return nil
}

// SeedDatabase adds initial data to the database
func SeedDatabase(db *gorm.DB) error {
	log.Println("🌱 Seeding database with initial data...")

	// Check if films already exist
	var count int64
	db.Model(&Film{}).Count(&count)
	if count > 0 {
		log.Println("📋 Database already contains films, skipping seed")
		return nil
	}

	// Add sample films
	sampleFilms := []Film{
		{Title: "The Shawshank Redemption", Director: "Frank Darabont", Year: 1994, Genre: "Drama"},
		{Title: "The Godfather", Director: "Francis Ford Coppola", Year: 1972, Genre: "Crime"},
		{Title: "The Dark Knight", Director: "Christopher Nolan", Year: 2008, Genre: "Action"},
		{Title: "Pulp Fiction", Director: "Quentin Tarantino", Year: 1994, Genre: "Crime"},
		{Title: "Forrest Gump", Director: "Robert Zemeckis", Year: 1994, Genre: "Drama"},
	}

	for _, film := range sampleFilms {
		if err := db.Create(&film).Error; err != nil {
			return fmt.Errorf("failed to seed film: %v", err)
		}
	}

	log.Printf("✅ Successfully seeded %d films to database", len(sampleFilms))
	return nil
}
