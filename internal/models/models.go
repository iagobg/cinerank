package models

import (
	"time"
)

type Movie struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Director    string    `json:"director"`
	Year        int       `json:"year"`
	Plot        string    `json:"plot"`
	PosterURL   string    `json:"poster_url"`
	IMDBRating  float64   `json:"imdb_rating"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Tags        []string  `json:"tags,omitempty"`
}

type Review struct {
	ID        int       `json:"id"`
	MovieID   int       `json:"movie_id"`
	UserID    int       `json:"user_id"`
	Rating    int       `json:"rating"` // 1-5 stars
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Movie     *Movie    `json:"movie,omitempty"`
	User      *User     `json:"user,omitempty"`
}

type User struct {
	ID           int       `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	PasswordHash string    `json:"-"` // Not exposed in JSON
}

type Tag struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type MovieWithStats struct {
	Movie
	ReviewCount   int     `json:"review_count"`
	AverageRating float64 `json:"average_rating"`
}

type CreateMovieRequest struct {
	Title      string   `json:"title"`
	Director   string   `json:"director"`
	Year       int      `json:"year"`
	Tags       []string `json:"tags"`
	Plot       string   `json:"plot"`
	PosterURL  string   `json:"poster_url"`
	IMDBRating float64  `json:"imdb_rating"`
}

type CreateReviewRequest struct {
	MovieID int    `json:"movie_id"`
	Rating  int    `json:"rating"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}