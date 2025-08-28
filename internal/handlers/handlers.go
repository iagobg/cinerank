package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"cinerank/internal/database"
	"cinerank/internal/models"
	"cinerank/internal/ui"
)

type Handler struct {
	DB *database.DB
}

func NewHandler(db *database.DB) *Handler {
	return &Handler{DB: db}
}

// Home page
func (h *Handler) HomePage(w http.ResponseWriter, r *http.Request) {
	movies, err := h.DB.GetAllMoviesWithStats()
	if err != nil {
		log.Printf("Error fetching movies: %v", err)
		movies = []models.MovieWithStats{}
	}

	recentReviews, err := h.DB.GetRecentReviews(5)
	if err != nil {
		log.Printf("Error fetching recent reviews: %v", err)
		recentReviews = []models.Review{}
	}

	if err := ui.HomePage(movies, recentReviews).Render(r.Context(), w); err != nil {
		http.Error(w, "Error rendering page", http.StatusInternalServerError)
		return
	}
}

// Movie detail page
func (h *Handler) MoviePage(w http.ResponseWriter, r *http.Request) {
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

	if err := ui.MoviePage(movie, reviews).Render(r.Context(), w); err != nil {
		http.Error(w, "Error rendering page", http.StatusInternalServerError)
		return
	}
}

// Add movie form
func (h *Handler) AddMovieForm(w http.ResponseWriter, r *http.Request) {
	if err := ui.AddMovieForm().Render(r.Context(), w); err != nil {
		http.Error(w, "Error rendering form", http.StatusInternalServerError)
		return
	}
}

// Create movie
func (h *Handler) CreateMovie(w http.ResponseWriter, r *http.Request) {
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

	req := models.CreateMovieRequest{
		Title:      r.Form.Get("title"),
		Director:   r.Form.Get("director"),
		Year:       year,
		Genre:      r.Form.Get("genre"),
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
}

// Add review form (HTMX partial)
func (h *Handler) AddReviewForm(w http.ResponseWriter, r *http.Request) {
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
}

// Create review
func (h *Handler) CreateReview(w http.ResponseWriter, r *http.Request) {
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
		MovieID:  movieID,
		UserName: r.Form.Get("user_name"),
		Rating:   rating,
		Title:    r.Form.Get("title"),
		Content:  r.Form.Get("content"),
	}

	review, err := h.DB.CreateReview(req)
	if err != nil {
		log.Printf("Error creating review: %v", err)
		http.Error(w, "Error creating review", http.StatusInternalServerError)
		return
	}

	// Return the new review as HTMX response
	if err := ui.ReviewItem(*review).Render(r.Context(), w); err != nil {
		http.Error(w, "Error rendering review", http.StatusInternalServerError)
		return
	}
}

// API endpoints for JSON responses
func (h *Handler) APIGetMovies(w http.ResponseWriter, r *http.Request) {
	movies, err := h.DB.GetAllMoviesWithStats()
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

	review, err := h.DB.CreateReview(req)
	if err != nil {
		http.Error(w, "Error creating review", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(review)
}