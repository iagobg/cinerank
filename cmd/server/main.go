package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"cinerank/internal/database"
	"cinerank/internal/handlers"
)

func main() {
	// Connect to database
	db, err := database.Connect()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Create handler
	h := handlers.NewHandler(db)

	// Create HTTP router
	mux := http.NewServeMux()

	// HTML routes
	mux.HandleFunc("/", h.HomePage)
	mux.HandleFunc("/search", h.SearchMovies)
	mux.HandleFunc("/movie/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/movie/") {
			h.MoviePage(w, r)
		} else {
			http.NotFound(w, r)
		}
	})
	mux.HandleFunc("/add-movie", h.AddMovieForm)
	mux.HandleFunc("/movies", h.CreateMovie)
	mux.HandleFunc("/review-form", h.AddReviewForm)
	mux.HandleFunc("/reviews", h.CreateReview)
	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			h.Login(w, r)
		} else {
			h.LoginForm(w, r)
		}
	})
	mux.HandleFunc("/logout", h.Logout)
	mux.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			h.Register(w, r)
		} else {
			h.RegisterForm(w, r)
		}
	})
	mux.HandleFunc("/admin", h.AdminPanel)
	mux.HandleFunc("/admin/delete-user/", h.DeleteUser)
	mux.HandleFunc("/admin/delete-movie/", h.DeleteMovie)

	// API routes
	mux.HandleFunc("/api/movies", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			h.APIGetMovies(w, r)
		case http.MethodPost:
			h.APICreateMovie(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	
	mux.HandleFunc("/api/movies/", h.APIGetMovie)
	mux.HandleFunc("/api/reviews", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			h.APIGetReviews(w, r)
		case http.MethodPost:
			h.APICreateReview(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Static assets
	fs := http.FileServer(http.Dir("static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// Get port from environment or default to 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🎬 CineRank server starting on port %s", port)
	log.Printf("📊 Database connected successfully")
	log.Printf("🌐 Visit http://localhost:%s to get started", port)
	
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}