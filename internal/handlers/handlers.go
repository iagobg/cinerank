package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"cinerank/internal/database"
	"cinerank/internal/models"
	"cinerank/internal/ui"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Handler struct {
	DB       *database.DB
	Sessions map[string]SessionData
}

type SessionData struct {
	UserID    int
	ExpiresAt time.Time
}

func NewHandler(db *database.DB) *Handler {
	return &Handler{
		DB:       db,
		Sessions: make(map[string]SessionData),
	}
}

// Middleware to check authentication
func (h *Handler) requireAuth(next func(http.ResponseWriter, *http.Request, *models.User)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionID, err := r.Cookie("session_id")
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		session, exists := h.Sessions[sessionID.Value]
		if !exists || session.ExpiresAt.Before(time.Now()) {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		user, err := h.DB.GetUserByID(session.UserID)
		if err != nil {
			http.Error(w, "User not found", http.StatusInternalServerError)
			return
		}

		next(w, r, user)
	}
}

// Middleware to check admin
func (h *Handler) requireAdmin(next func(http.ResponseWriter, *http.Request, *models.User)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionID, err := r.Cookie("session_id")
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		session, exists := h.Sessions[sessionID.Value]
		if !exists || session.ExpiresAt.Before(time.Now()) {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		user, err := h.DB.GetUserByID(session.UserID)
		if err != nil || user.Role != "admin" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		next(w, r, user)
	}
}

// Home page
func (h *Handler) HomePage(w http.ResponseWriter, r *http.Request) {
	user := h.getUserFromSession(r)

	searchQuery := r.URL.Query().Get("query")
	movies, err := h.DB.GetAllMoviesWithStats(searchQuery)
	if err != nil {
		log.Printf("Error fetching movies: %v", err)
		movies = []models.MovieWithStats{}
	}

	recentReviews, err := h.DB.GetRecentReviews(5)
	if err != nil {
		log.Printf("Error fetching recent reviews: %v", err)
		recentReviews = []models.Review{}
	}

	if err := ui.HomePage(movies, recentReviews, user, searchQuery).Render(r.Context(), w); err != nil {
		http.Error(w, "Error rendering page", http.StatusInternalServerError)
		return
	}
}

// Search movies (HTMX partial)
func (h *Handler) SearchMovies(w http.ResponseWriter, r *http.Request) {
	searchQuery := r.URL.Query().Get("query")
	movies, err := h.DB.GetAllMoviesWithStats(searchQuery)
	if err != nil {
		log.Printf("Error fetching movies: %v", err)
		movies = []models.MovieWithStats{}
	}

	if err := ui.MovieList(movies).Render(r.Context(), w); err != nil {
		http.Error(w, "Error rendering movie list", http.StatusInternalServerError)
	}
}

// Movie detail page
func (h *Handler) MoviePage(w http.ResponseWriter, r *http.Request) {
	user := h.getUserFromSession(r)

	movieIDStr := strings.TrimPrefix(r.URL.Path, "/movie/")
	movieID, err := strconv.Atoi(movieIDStr)
	if err != nil {
		http.Error(w, "Invalid movie ID", http.StatusBadRequest)
		return
	}

	movie, err := h.DB.GetMovieByID(movieID)
	if err != nil {
		http.Error(w, "Movie not found", http.StatusNotFound)
		return
	}

	reviews, err := h.DB.GetReviewsByMovieID(movieID)
	if err != nil {
		log.Printf("Error fetching reviews: %v", err)
		reviews = []models.Review{}
	}

	if err := ui.MoviePage(movie, reviews, user).Render(r.Context(), w); err != nil {
		http.Error(w, "Error rendering page", http.StatusInternalServerError)
		return
	}
}

// Add movie form
func (h *Handler) AddMovieForm(w http.ResponseWriter, r *http.Request) {
	h.requireAuth(func(w http.ResponseWriter, r *http.Request, user *models.User) {
		if err := ui.AddMovieForm(user).Render(r.Context(), w); err != nil {
			http.Error(w, "Error rendering form", http.StatusInternalServerError)
			return
		}
	})(w, r)
}

// Create movie
func (h *Handler) CreateMovie(w http.ResponseWriter, r *http.Request) {
	h.requireAuth(func(w http.ResponseWriter, r *http.Request, user *models.User) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Error parsing form", http.StatusBadRequest)
			return
		}

		year, err := strconv.Atoi(r.Form.Get("year"))
		if err != nil {
			http.Error(w, "Invalid year", http.StatusBadRequest)
			return
		}

		imdbRating, err := strconv.ParseFloat(r.Form.Get("imdb_rating"), 64)
		if err != nil {
			imdbRating = 0.0 // Optional field
		}

		tagsStr := r.Form.Get("tags")
		tags := strings.Split(tagsStr, ",")
		for i := range tags {
			tags[i] = strings.TrimSpace(tags[i])
		}

		req := models.CreateMovieRequest{
			Title:      r.Form.Get("title"),
			Director:   r.Form.Get("director"),
			Year:       year,
			Tags:       tags,
			Plot:       r.Form.Get("plot"),
			PosterURL:  r.Form.Get("poster_url"),
			IMDBRating: imdbRating,
		}

		movie, err := h.DB.CreateMovie(req)
		if err != nil {
			log.Printf("Error creating movie: %v", err)
			http.Error(w, "Error creating movie", http.StatusInternalServerError)
			return
		}

		// Return success message or redirect
		w.Header().Set("HX-Redirect", fmt.Sprintf("/movie/%d", movie.ID))
		w.WriteHeader(http.StatusCreated)
	})(w, r)
}

// Add review form (HTMX partial)
func (h *Handler) AddReviewForm(w http.ResponseWriter, r *http.Request) {
	h.requireAuth(func(w http.ResponseWriter, r *http.Request, user *models.User) {
		movieIDStr := r.URL.Query().Get("movie_id")
		movieID, err := strconv.Atoi(movieIDStr)
		if err != nil {
			http.Error(w, "Invalid movie ID", http.StatusBadRequest)
			return
		}

		if err := ui.ReviewForm(movieID).Render(r.Context(), w); err != nil {
			http.Error(w, "Error rendering form", http.StatusInternalServerError)
			return
		}
	})(w, r)
}

// Create review
func (h *Handler) CreateReview(w http.ResponseWriter, r *http.Request) {
	h.requireAuth(func(w http.ResponseWriter, r *http.Request, user *models.User) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Error parsing form", http.StatusBadRequest)
			return
		}

		movieID, err := strconv.Atoi(r.Form.Get("movie_id"))
		if err != nil {
			http.Error(w, "Invalid movie ID", http.StatusBadRequest)
			return
		}

		rating, err := strconv.Atoi(r.Form.Get("rating"))
		if err != nil || rating < 1 || rating > 5 {
			http.Error(w, "Invalid rating (must be 1-5)", http.StatusBadRequest)
			return
		}

		req := models.CreateReviewRequest{
			MovieID: movieID,
			Rating:  rating,
			Title:   r.Form.Get("title"),
			Content: r.Form.Get("content"),
		}

		review, err := h.DB.CreateReview(req, user.ID)
		if err != nil {
			log.Printf("Error creating review: %v", err)
			http.Error(w, "Error creating review", http.StatusInternalServerError)
			return
		}

		review.User = user

		// Return the new review as HTMX response
		if err := ui.ReviewItem(*review).Render(r.Context(), w); err != nil {
			http.Error(w, "Error rendering review", http.StatusInternalServerError)
			return
		}
	})(w, r)
}

// Login form
func (h *Handler) LoginForm(w http.ResponseWriter, r *http.Request) {
	if err := ui.LoginForm().Render(r.Context(), w); err != nil {
		http.Error(w, "Error rendering form", http.StatusInternalServerError)
	}
}

// Login
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}

	email := r.Form.Get("email")
	password := r.Form.Get("password")

	user, err := h.DB.GetUserByEmail(email)
	if err != nil || bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Create session
	sessionID := uuid.New().String()
	h.Sessions[sessionID] = SessionData{
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Logout
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	sessionID, err := r.Cookie("session_id")
	if err == nil {
		delete(h.Sessions, sessionID.Value)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		HttpOnly: true,
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Register form
func (h *Handler) RegisterForm(w http.ResponseWriter, r *http.Request) {
	if err := ui.RegisterForm().Render(r.Context(), w); err != nil {
		http.Error(w, "Error rendering form", http.StatusInternalServerError)
	}
}

// Register
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}

	username := r.Form.Get("username")
	email := r.Form.Get("email")
	password := r.Form.Get("password")

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Error hashing password", http.StatusInternalServerError)
		return
	}

	user, err := h.DB.CreateUser(models.RegisterRequest{
		Username: username,
		Email:    email,
		Password: string(hash),
	})
	if err != nil {
		log.Printf("Error creating user: %v", err)
		http.Error(w, "Error creating user", http.StatusInternalServerError)
		return
	}

	// Create session
	sessionID := uuid.New().String()
	h.Sessions[sessionID] = SessionData{
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Admin panel
func (h *Handler) AdminPanel(w http.ResponseWriter, r *http.Request) {
	h.requireAdmin(func(w http.ResponseWriter, r *http.Request, user *models.User) {
		users, err := h.DB.GetAllUsers()
		if err != nil {
			log.Printf("Error fetching users: %v", err)
		}

		movies, err := h.DB.GetAllMoviesWithStats("")
		if err != nil {
			log.Printf("Error fetching movies: %v", err)
		}

		if err := ui.AdminPanel(users, movies, user).Render(r.Context(), w); err != nil {
			http.Error(w, "Error rendering admin panel", http.StatusInternalServerError)
		}
	})(w, r)
}

// Delete user (admin)
func (h *Handler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	h.requireAdmin(func(w http.ResponseWriter, r *http.Request, user *models.User) {
		userIDStr := strings.TrimPrefix(r.URL.Path, "/admin/delete-user/")
		userID, err := strconv.Atoi(userIDStr)
		if err != nil {
			http.Error(w, "Invalid user ID", http.StatusBadRequest)
			return
		}

		if userID == user.ID {
			http.Error(w, "Cannot delete yourself", http.StatusBadRequest)
			return
		}

		err = h.DB.DeleteUser(userID)
		if err != nil {
			http.Error(w, "Error deleting user", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/admin", http.StatusSeeOther)
	})(w, r)
}

// Delete movie (admin)
func (h *Handler) DeleteMovie(w http.ResponseWriter, r *http.Request) {
	h.requireAdmin(func(w http.ResponseWriter, r *http.Request, user *models.User) {
		movieIDStr := strings.TrimPrefix(r.URL.Path, "/admin/delete-movie/")
		movieID, err := strconv.Atoi(movieIDStr)
		if err != nil {
			http.Error(w, "Invalid movie ID", http.StatusBadRequest)
			return
		}

		err = h.DB.DeleteMovie(movieID)
		if err != nil {
			http.Error(w, "Error deleting movie", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/admin", http.StatusSeeOther)
	})(w, r)
}

// Helper to get user from session
func (h *Handler) getUserFromSession(r *http.Request) *models.User {
	sessionID, err := r.Cookie("session_id")
	if err != nil {
		return nil
	}

	session, exists := h.Sessions[sessionID.Value]
	if !exists || session.ExpiresAt.Before(time.Now()) {
		return nil
	}

	user, err := h.DB.GetUserByID(session.UserID)
	if err != nil {
		return nil
	}

	return user
}

// API endpoints for JSON responses
func (h *Handler) APIGetMovies(w http.ResponseWriter, r *http.Request) {
	searchQuery := r.URL.Query().Get("query")
	movies, err := h.DB.GetAllMoviesWithStats(searchQuery)
	if err != nil {
		http.Error(w, "Error fetching movies", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(movies)
}

func (h *Handler) APIGetMovie(w http.ResponseWriter, r *http.Request) {
	movieIDStr := strings.TrimPrefix(r.URL.Path, "/api/movies/")
	movieID, err := strconv.Atoi(movieIDStr)
	if err != nil {
		http.Error(w, "Invalid movie ID", http.StatusBadRequest)
		return
	}

	movie, err := h.DB.GetMovieByID(movieID)
	if err != nil {
		http.Error(w, "Movie not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(movie)
}

func (h *Handler) APICreateMovie(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.CreateMovieRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	movie, err := h.DB.CreateMovie(req)
	if err != nil {
		http.Error(w, "Error creating movie", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(movie)
}

func (h *Handler) APIGetReviews(w http.ResponseWriter, r *http.Request) {
	movieIDStr := r.URL.Query().Get("movie_id")
	if movieIDStr == "" {
		// Get recent reviews
		reviews, err := h.DB.GetRecentReviews(10)
		if err != nil {
			http.Error(w, "Error fetching reviews", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(reviews)
		return
	}

	movieID, err := strconv.Atoi(movieIDStr)
	if err != nil {
		http.Error(w, "Invalid movie ID", http.StatusBadRequest)
		return
	}

	reviews, err := h.DB.GetReviewsByMovieID(movieID)
	if err != nil {
		http.Error(w, "Error fetching reviews", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(reviews)
}

func (h *Handler) APICreateReview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.CreateReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Rating < 1 || req.Rating > 5 {
		http.Error(w, "Rating must be between 1 and 5", http.StatusBadRequest)
		return
	}

	// Note: API would need user authentication (e.g., JWT). For simplicity, assuming user_id is provided
	userID := 1 // Placeholder; should be extracted from auth token
	review, err := h.DB.CreateReview(req, userID)
	if err != nil {
		http.Error(w, "Error creating review", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(review)
}