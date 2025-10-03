# 🎬 Film REST API

A simple REST API built with pure Go HTTP for managing film data. This project demonstrates CRUD operations with a clean, modern web interface.

## 📋 Features

- **Pure Go HTTP**: No external frameworks, just standard library
- **Film Management**: Store and manage film data (title, director, year, genre)
- **RESTful API**: Standard HTTP methods for all operations
- **Web Interface**: Beautiful, responsive HTML interface
- **In-Memory Storage**: Thread-safe data storage with mutex
- **CORS Support**: Cross-origin requests enabled
- **Sample Data**: Pre-loaded with popular films

## 🚀 Quick Start

## API Docs Link
https://documenter.getpostman.com/view/34415611/2sB3QGtBXu

1. **Run the server:**
   ```bash
   go run main.go
   ```

2. **Open your browser:**
   Navigate to `http://localhost:8080` to access the web interface

3. **Or use the API directly:**
   The API is available at `http://localhost:8080/api/films`

## 📡 API Endpoints

### GET /api/films
Get all films in the database.

**Response:**
```json
[
  {
    "id": 1,
    "title": "The Shawshank Redemption",
    "director": "Frank Darabont",
    "year": 1994,
    "genre": "Drama"
  }
]
```

### POST /api/films
Add a new film to the database.

**Request Body:**
```json
{
  "title": "Inception",
  "director": "Christopher Nolan",
  "year": 2010,
  "genre": "Sci-Fi"
}
```

**Response:**
```json
{
  "id": 6,
  "title": "Inception",
  "director": "Christopher Nolan",
  "year": 2010,
  "genre": "Sci-Fi"
}
```

### PUT /api/films/{id}
Update an existing film by ID.

**Request Body:**
```json
{
  "title": "Inception (Updated)",
  "director": "Christopher Nolan",
  "year": 2010,
  "genre": "Sci-Fi/Thriller"
}
```

**Response:**
```json
{
  "id": 6,
  "title": "Inception (Updated)",
  "director": "Christopher Nolan",
  "year": 2010,
  "genre": "Sci-Fi/Thriller"
}
```

### DELETE /api/films/{id}
Delete a film by ID.

**Response:** `204 No Content`

## 🧪 Testing the API

### Using curl:

```bash
# Get all films
curl http://localhost:8080/api/films

# Add a new film
curl -X POST http://localhost:8080/api/films \
  -H "Content-Type: application/json" \
  -d '{"title":"Interstellar","director":"Christopher Nolan","year":2014,"genre":"Sci-Fi"}'

# Update a film (replace {id} with actual ID)
curl -X PUT http://localhost:8080/api/films/1 \
  -H "Content-Type: application/json" \
  -d '{"title":"The Shawshank Redemption","director":"Frank Darabont","year":1994,"genre":"Drama/Crime"}'

# Delete a film (replace {id} with actual ID)
curl -X DELETE http://localhost:8080/api/films/1
```

### Using the Web Interface:
1. Open `http://localhost:8080` in your browser
2. Use the intuitive interface to:
   - View all films in a beautiful card layout
   - Add new films with the form
   - Update existing films (click "Edit" on any film card)
   - Delete films (click "Delete" on any film card)

## 🏗️ Project Structure

```
sts_go_3/
├── main.go      # Main server code with all API endpoints
├── index.html   # Web interface for interacting with the API
├── go.mod       # Go module file
└── README.md    # This file
```

## 🎯 Data Model

Each film has the following structure:

```go
type Film struct {
    ID       int    `json:"id"`       // Auto-generated unique identifier
    Title    string `json:"title"`    // Film title (required)
    Director string `json:"director"` // Director name (required)
    Year     int    `json:"year"`     // Release year (required)
    Genre    string `json:"genre"`    // Film genre (optional)
}
```

## 🔧 Technical Details

- **Concurrency Safe**: Uses `sync.RWMutex` for thread-safe operations
- **Error Handling**: Proper HTTP status codes and error messages
- **Validation**: Required field validation for API requests
- **CORS Enabled**: Supports cross-origin requests
- **Clean Architecture**: Separation of concerns with dedicated store methods

## 📦 Sample Data

The API comes pre-loaded with these classic films:
- The Shawshank Redemption (1994)
- The Godfather (1972)
- The Dark Knight (2008)
- Pulp Fiction (1994)
- Forrest Gump (1994)

## 🌟 Features of the Web Interface

- **Responsive Design**: Works on desktop and mobile devices
- **Modern UI**: Beautiful gradient design with smooth animations
- **Real-time Updates**: Automatic refresh after operations
- **Form Validation**: Client-side validation for better UX
- **Loading States**: Visual feedback during API calls
- **Error Handling**: Clear error messages for failed operations

## 🚀 Next Steps

To extend this project, you could add:
- Database persistence (PostgreSQL, MySQL, etc.)
- Authentication and authorization
- Pagination for large datasets
- Search and filtering capabilities
- Unit tests
- Docker containerization
- API rate limiting

---

**Happy coding! 🎬✨**
