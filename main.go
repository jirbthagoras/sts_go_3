package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"os"

	"gorm.io/gorm"
)

// TokenStore manages active tokens
type TokenStore struct {
	mu     sync.RWMutex
	tokens map[string]time.Time // token -> expiry time
}

// NewTokenStore creates a new token store
func NewTokenStore() *TokenStore {
	return &TokenStore{
		tokens: make(map[string]time.Time),
	}
}

// GenerateToken creates a new random token
func (ts *TokenStore) GenerateToken() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// AddToken adds a token with expiry time
func (ts *TokenStore) AddToken(token string) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	ts.tokens[token] = time.Now().Add(24 * time.Hour) // 24 hour expiry
}

// ValidateToken checks if token is valid and not expired
func (ts *TokenStore) ValidateToken(token string) bool {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	expiry, exists := ts.tokens[token]
	if !exists {
		return false
	}
	if time.Now().After(expiry) {
		// Token expired, remove it
		delete(ts.tokens, token)
		return false
	}
	return true
}

// RemoveToken removes a token (for logout)
func (ts *TokenStore) RemoveToken(token string) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	delete(ts.tokens, token)
}

// Global services
var filmService *FilmService
var userService *UserService
var tokenStore *TokenStore
var db *gorm.DB

// CORS middleware
func enableCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
}

// Authentication middleware
func requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w)

		if r.Method == "OPTIONS" {
			return
		}

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "Authorization header required"})
			return
		}

		// Check for Bearer token format
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid authorization header format"})
			return
		}

		token := parts[1]
		if !tokenStore.ValidateToken(token) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid or expired token"})
			return
		}

		next(w, r)
	}
}

// loginHandler handles user login
func loginHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)

	if r.Method == "OPTIONS" {
		return
	}

	if r.Method != "POST" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Method not allowed"})
		return
	}

	var loginReq LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&loginReq); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid JSON"})
		return
	}

	if loginReq.Username == "" || loginReq.Password == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Username and password are required"})
		return
	}

	if !userService.ValidateUser(loginReq.Username, loginReq.Password) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid credentials"})
		return
	}

	// Generate token
	token := tokenStore.GenerateToken()
	tokenStore.AddToken(token)

	response := LoginResponse{Token: token}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// logoutHandler handles user logout
func logoutHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)

	if r.Method == "OPTIONS" {
		return
	}

	if r.Method != "POST" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Method not allowed"})
		return
	}

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Authorization header required"})
		return
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid authorization header format"})
		return
	}

	token := parts[1]
	tokenStore.RemoveToken(token)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(SuccessResponse{Message: "Logged out successfully"})
}

// getFilmsHandler handles getting all films
func getFilmsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Method not allowed"})
		return
	}

	films, err := filmService.GetAllFilms()
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to retrieve films"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(films)
}

// addFilmHandler handles adding a new film
func addFilmHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Method not allowed"})
		return
	}

	var filmReq FilmRequest
	if err := json.NewDecoder(r.Body).Decode(&filmReq); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid JSON"})
		return
	}

	// Validate required fields
	if filmReq.Title == "" || filmReq.Director == "" || filmReq.Year == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Title, director, and year are required"})
		return
	}

	newFilm, err := filmService.CreateFilm(filmReq)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to create film"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newFilm)
}

// updateFilmHandler handles updating a film
func updateFilmHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Method not allowed"})
		return
	}

	// Extract ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/api/films/")
	id, err := strconv.Atoi(path)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid film ID"})
		return
	}

	var filmReq FilmRequest
	if err := json.NewDecoder(r.Body).Decode(&filmReq); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid JSON"})
		return
	}

	// Validate required fields
	if filmReq.Title == "" || filmReq.Director == "" || filmReq.Year == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Title, director, and year are required"})
		return
	}

	updatedFilm, err := filmService.UpdateFilm(uint(id), filmReq)
	if err != nil {
		if err.Error() == "film not found" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "Film not found"})
		} else {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to update film"})
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedFilm)
}

// deleteFilmHandler handles deleting a film
func deleteFilmHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Method not allowed"})
		return
	}

	// Extract ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/api/films/")
	id, err := strconv.Atoi(path)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid film ID"})
		return
	}

	err = filmService.DeleteFilm(uint(id))
	if err != nil {
		if err.Error() == "film not found" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "Film not found"})
		} else {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to delete film"})
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Route handler to distinguish between different endpoints (protected)
func filmsHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	if path == "/api/films" {
		switch r.Method {
		case "GET":
			getFilmsHandler(w, r)
		case "POST":
			addFilmHandler(w, r)
		default:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "Method not allowed"})
		}
	} else if strings.HasPrefix(path, "/api/films/") {
		switch r.Method {
		case "PUT":
			updateFilmHandler(w, r)
		case "DELETE":
			deleteFilmHandler(w, r)
		default:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "Method not allowed"})
		}
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Not found"})
	}
}

// swaggerHandler serves the swagger YAML file and UI
func swaggerHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/swagger/" || r.URL.Path == "/swagger/index.html" {
		// Serve Swagger UI HTML
		html := `<!DOCTYPE html>
<html>
<head>
    <title>API Documentation</title>
    <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@3.25.0/swagger-ui.css" />
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@3.25.0/swagger-ui-bundle.js"></script>
    <script>
        SwaggerUIBundle({
            url: '/swagger.yaml',
            dom_id: '#swagger-ui',
            presets: [
                SwaggerUIBundle.presets.apis,
                SwaggerUIBundle.presets.standalone
            ]
        });
    </script>
</body>
</html>`
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(html))
	} else if r.URL.Path == "/swagger.yaml" {
		// Serve the YAML file
		yamlContent, err := os.ReadFile("swagger.yaml")
		if err != nil {
			http.Error(w, "Swagger YAML file not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/x-yaml")
		w.Write(yamlContent)
	} else {
		http.Error(w, "Not found", http.StatusNotFound)
	}
}

// Serve static files (HTML)
func staticHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		http.ServeFile(w, r, "index.html")
	} else {
		http.Error(w, "Not found", http.StatusNotFound)
	}
}

func main() {
	// Load environment variables from .env file
	if err := loadEnv(); err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
		log.Println("Continuing with system environment variables...")
	} else {
		log.Println("✅ Successfully loaded .env file")
	}

	// Connect to database
	var err error
	db, err = ConnectDatabase()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Run migrations
	if err := MigrateDatabase(db); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// Initialize services
	filmService = NewFilmService(db)
	userService = NewUserService(db)
	tokenStore = NewTokenStore()

	// Seed database with initial data
	if err := SeedDatabase(db); err != nil {
		log.Printf("Warning: Failed to seed database: %v", err)
	}

	// Seed users
	if err := userService.SeedUsers(); err != nil {
		log.Printf("Warning: Failed to seed users: %v", err)
	}

	// Register handlers
	http.HandleFunc("/api/login", loginHandler)
	http.HandleFunc("/api/logout", logoutHandler)
	http.HandleFunc("/api/films", requireAuth(filmsHandler))
	http.HandleFunc("/api/films/", requireAuth(filmsHandler))
	http.HandleFunc("/swagger/", swaggerHandler)
	http.HandleFunc("/swagger.yaml", swaggerHandler)
	http.HandleFunc("/", staticHandler)

	fmt.Println("🎬 Film REST API Server starting on http://localhost:8080")
	fmt.Println("🔐 Authentication Endpoints:")
	fmt.Println("   POST   /api/login     - User login")
	fmt.Println("   POST   /api/logout    - User logout")
	fmt.Println("📋 Protected API Endpoints:")
	fmt.Println("   GET    /api/films     - Get all films (requires auth)")
	fmt.Println("   POST   /api/films     - Add new film (requires auth)")
	fmt.Println("   PUT    /api/films/{id} - Update film (requires auth)")
	fmt.Println("   DELETE /api/films/{id} - Delete film (requires auth)")
	fmt.Println("📚 API Documentation: http://localhost:8080/swagger/")
	fmt.Println("🌐 Web Interface: http://localhost:8080")
	fmt.Println("👤 Default users: admin/admin123, user1/password123, demo/demo456")
	fmt.Println("🗄️  Database: PostgreSQL")

	log.Fatal(http.ListenAndServe(":8080", nil))
}
