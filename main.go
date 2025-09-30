package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"gopkg.in/yaml.v2"
)

// Film represents a movie with its details
type Film struct {
	ID       int    `json:"id"`
	Title    string `json:"title"`
	Director string `json:"director"`
	Year     int    `json:"year"`
	Genre    string `json:"genre"`
}

// User represents a user from config
type User struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

// Config represents the configuration structure
type Config struct {
	Users []User `yaml:"users"`
}

// LoginRequest represents login request payload
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse represents login response
type LoginResponse struct {
	Token string `json:"token"`
}

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

// FilmStore manages the film data in memory
type FilmStore struct {
	mu     sync.RWMutex
	films  map[int]Film
	nextID int
}

// NewFilmStore creates a new film store with sample data
func NewFilmStore() *FilmStore {
	store := &FilmStore{
		films:  make(map[int]Film),
		nextID: 1,
	}

	// Add some sample films
	sampleFilms := []Film{
		{Title: "The Shawshank Redemption", Director: "Frank Darabont", Year: 1994, Genre: "Drama"},
		{Title: "The Godfather", Director: "Francis Ford Coppola", Year: 1972, Genre: "Crime"},
		{Title: "The Dark Knight", Director: "Christopher Nolan", Year: 2008, Genre: "Action"},
		{Title: "Pulp Fiction", Director: "Quentin Tarantino", Year: 1994, Genre: "Crime"},
		{Title: "Forrest Gump", Director: "Robert Zemeckis", Year: 1994, Genre: "Drama"},
	}

	for _, film := range sampleFilms {
		store.addFilm(film)
	}

	return store
}

func (fs *FilmStore) addFilm(film Film) Film {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	film.ID = fs.nextID
	fs.films[fs.nextID] = film
	fs.nextID++
	return film
}

func (fs *FilmStore) getFilms() []Film {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	films := make([]Film, 0, len(fs.films))
	for _, film := range fs.films {
		films = append(films, film)
	}
	return films
}

func (fs *FilmStore) getFilm(id int) (Film, bool) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	film, exists := fs.films[id]
	return film, exists
}

func (fs *FilmStore) updateFilm(id int, updatedFilm Film) (Film, bool) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	if _, exists := fs.films[id]; !exists {
		return Film{}, false
	}

	updatedFilm.ID = id
	fs.films[id] = updatedFilm
	return updatedFilm, true
}

func (fs *FilmStore) deleteFilm(id int) bool {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	if _, exists := fs.films[id]; !exists {
		return false
	}

	delete(fs.films, id)
	return true
}

// Global stores
var filmStore *FilmStore
var tokenStore *TokenStore
var config Config

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
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}
		
		// Check for Bearer token format
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
			return
		}
		
		token := parts[1]
		if !tokenStore.ValidateToken(token) {
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}
		
		next(w, r)
	}
}

// Load configuration from YAML file
func loadConfig() error {
	file, err := os.Open("config.yaml")
	if err != nil {
		return err
	}
	defer file.Close()
	
	data, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}
	
	return yaml.Unmarshal(data, &config)
}

// Validate user credentials
func validateUser(username, password string) bool {
	for _, user := range config.Users {
		if user.Username == username && user.Password == password {
			return true
		}
	}
	return false
}

// POST /api/login - User login
func loginHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	
	if r.Method == "OPTIONS" {
		return
	}
	
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	var loginReq LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&loginReq); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	
	if loginReq.Username == "" || loginReq.Password == "" {
		http.Error(w, "Username and password are required", http.StatusBadRequest)
		return
	}
	
	if !validateUser(loginReq.Username, loginReq.Password) {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}
	
	// Generate token
	token := tokenStore.GenerateToken()
	tokenStore.AddToken(token)
	
	response := LoginResponse{Token: token}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// POST /api/logout - User logout
func logoutHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	
	if r.Method == "OPTIONS" {
		return
	}
	
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header required", http.StatusUnauthorized)
		return
	}
	
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
		return
	}
	
	token := parts[1]
	tokenStore.RemoveToken(token)
	
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Logged out successfully"))
}

// GET /api/films - Get all films (protected)
func getFilmsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	films := filmStore.getFilms()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(films)
}

// POST /api/films - Add a new film (protected)
func addFilmHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var film Film
	if err := json.NewDecoder(r.Body).Decode(&film); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if film.Title == "" || film.Director == "" || film.Year == 0 {
		http.Error(w, "Title, director, and year are required", http.StatusBadRequest)
		return
	}

	newFilm := filmStore.addFilm(film)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newFilm)
}

// PUT /api/films/{id} - Update a film (protected)
func updateFilmHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/api/films/")
	id, err := strconv.Atoi(path)
	if err != nil {
		http.Error(w, "Invalid film ID", http.StatusBadRequest)
		return
	}

	var film Film
	if err := json.NewDecoder(r.Body).Decode(&film); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if film.Title == "" || film.Director == "" || film.Year == 0 {
		http.Error(w, "Title, director, and year are required", http.StatusBadRequest)
		return
	}

	updatedFilm, exists := filmStore.updateFilm(id, film)
	if !exists {
		http.Error(w, "Film not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedFilm)
}

// DELETE /api/films/{id} - Delete a film (protected)
func deleteFilmHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/api/films/")
	id, err := strconv.Atoi(path)
	if err != nil {
		http.Error(w, "Invalid film ID", http.StatusBadRequest)
		return
	}

	if !filmStore.deleteFilm(id) {
		http.Error(w, "Film not found", http.StatusNotFound)
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
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	} else if strings.HasPrefix(path, "/api/films/") {
		switch r.Method {
		case "PUT":
			updateFilmHandler(w, r)
		case "DELETE":
			deleteFilmHandler(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
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
	// Load configuration
	if err := loadConfig(); err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Initialize stores
	filmStore = NewFilmStore()
	tokenStore = NewTokenStore()

	// Register handlers
	http.HandleFunc("/api/login", loginHandler)
	http.HandleFunc("/api/logout", logoutHandler)
	http.HandleFunc("/api/films", requireAuth(filmsHandler))
	http.HandleFunc("/api/films/", requireAuth(filmsHandler))
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
	fmt.Println("🌐 Web Interface: http://localhost:8080")
	fmt.Println("👤 Default users: admin/admin123, user1/password123, demo/demo456")

	log.Fatal(http.ListenAndServe(":8080", nil))
}
