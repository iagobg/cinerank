package models

import (
	"time"
)

type Movie struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Director    string    `json:"director"`
	Year        int       `json:"year"`
	Genre       string    `json:"genre"`
	Plot        string    `json:"plot"`
	PosterURL   string    `json:"poster_url"`
	IMDBRating  float64   `json:"imdb_rating"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Review struct {
	ID        int       `json:"id"`
	MovieID   int       `json:"movie_id"`
	UserName  string    `json:"user_name"`
	Rating    int       `json:"rating"` // 1-5 stars
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Movie     *Movie    `json:"movie,omitempty"`
}

type MovieWithStats struct {
	Movie
	ReviewCount   int     `json:"review_count"`
	AverageRating float64 `json:"average_rating"`
}

type CreateMovieRequest struct {
	Title     string  `json:"title"`
	Director  string  `json:"director"`
	Year      int     `json:"year"`
	Genre     string  `json:"genre"`
	Plot      string  `json:"plot"`
	PosterURL string  `json:"poster_url"`
	IMDBRating float64 `json:"imdb_rating"`
}

type CreateReviewRequest struct {
	MovieID  int    `json:"movie_id"`
	UserName string `json:"user_name"`
	Rating   int    `json:"rating"`
	Title    string `json:"title"`
	Content  string `json:"content"`
}