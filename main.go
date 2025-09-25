package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

// Film represents a movie with its details
type Film struct {
	ID       int    `json:"id"`
	Title    string `json:"title"`
	Director string `json:"director"`
	Year     int    `json:"year"`
	Genre    string `json:"genre"`
}

// FilmStore manages the film data in memory
type FilmStore struct {
	mu    sync.RWMutex
	films map[int]Film
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

// Global film store
var filmStore *FilmStore

// CORS middleware
func enableCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

// GET /api/films - Get all films
func getFilmsHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	
	if r.Method == "OPTIONS" {
		return
	}
	
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	films := filmStore.getFilms()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(films)
}

// POST /api/films - Add a new film
func addFilmHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	
	if r.Method == "OPTIONS" {
		return
	}
	
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

// PUT /api/films/{id} - Update a film
func updateFilmHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	
	if r.Method == "OPTIONS" {
		return
	}
	
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

// DELETE /api/films/{id} - Delete a film
func deleteFilmHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	
	if r.Method == "OPTIONS" {
		return
	}
	
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

// Route handler to distinguish between different endpoints
func filmsHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	
	if path == "/api/films" {
		switch r.Method {
		case "GET", "OPTIONS":
			getFilmsHandler(w, r)
		case "POST":
			addFilmHandler(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	} else if strings.HasPrefix(path, "/api/films/") {
		switch r.Method {
		case "PUT", "OPTIONS":
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
	// Initialize the film store
	filmStore = NewFilmStore()
	
	// Register handlers
	http.HandleFunc("/api/films", filmsHandler)
	http.HandleFunc("/api/films/", filmsHandler)
	http.HandleFunc("/", staticHandler)
	
	fmt.Println("🎬 Film REST API Server starting on http://localhost:8080")
	fmt.Println("📋 API Endpoints:")
	fmt.Println("   GET    /api/films     - Get all films")
	fmt.Println("   POST   /api/films     - Add new film")
	fmt.Println("   PUT    /api/films/{id} - Update film")
	fmt.Println("   DELETE /api/films/{id} - Delete film")
	fmt.Println("🌐 Web Interface: http://localhost:8080")
	
	log.Fatal(http.ListenAndServe(":8080", nil))
}
