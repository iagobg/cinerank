package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"cinerank/internal/models"

	_ "github.com/lib/pq"
)

type DB struct {
	*sql.DB
}

// Connect connects to the PostgreSQL database
func Connect() (*DB, error) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return nil, fmt.Errorf("DATABASE_URL environment variable is not set")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Successfully connected to database")
	return &DB{db}, nil
}

// Movie operations
func (db *DB) GetAllMoviesWithStats() ([]models.MovieWithStats, error) {
	query := `
		SELECT 
			m.id, m.title, m.director, m.year, m.genre, m.plot, 
			m.poster_url, m.imdb_rating, m.created_at, m.updated_at,
			COALESCE(COUNT(r.id), 0) as review_count,
			COALESCE(AVG(r.rating::float), 0) as average_rating
		FROM movies m
		LEFT JOIN reviews r ON m.id = r.movie_id
		GROUP BY m.id, m.title, m.director, m.year, m.genre, m.plot, 
				 m.poster_url, m.imdb_rating, m.created_at, m.updated_at
		ORDER BY m.created_at DESC
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var movies []models.MovieWithStats
	for rows.Next() {
		var m models.MovieWithStats
		err := rows.Scan(
			&m.ID, &m.Title, &m.Director, &m.Year, &m.Genre, &m.Plot,
			&m.PosterURL, &m.IMDBRating, &m.CreatedAt, &m.UpdatedAt,
			&m.ReviewCount, &m.AverageRating,
		)
		if err != nil {
			return nil, err
		}
		movies = append(movies, m)
	}

	return movies, nil
}

func (db *DB) GetMovieByID(id int) (*models.Movie, error) {
	query := `
		SELECT id, title, director, year, genre, plot, poster_url, imdb_rating, created_at, updated_at
		FROM movies WHERE id = $1
	`

	var m models.Movie
	err := db.QueryRow(query, id).Scan(
		&m.ID, &m.Title, &m.Director, &m.Year, &m.Genre, &m.Plot,
		&m.PosterURL, &m.IMDBRating, &m.CreatedAt, &m.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &m, nil
}

func (db *DB) CreateMovie(req models.CreateMovieRequest) (*models.Movie, error) {
	query := `
		INSERT INTO movies (title, director, year, genre, plot, poster_url, imdb_rating, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
		RETURNING id, title, director, year, genre, plot, poster_url, imdb_rating, created_at, updated_at
	`

	var m models.Movie
	err := db.QueryRow(query, req.Title, req.Director, req.Year, req.Genre, req.Plot, req.PosterURL, req.IMDBRating).Scan(
		&m.ID, &m.Title, &m.Director, &m.Year, &m.Genre, &m.Plot,
		&m.PosterURL, &m.IMDBRating, &m.CreatedAt, &m.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &m, nil
}

// Review operations
func (db *DB) GetReviewsByMovieID(movieID int) ([]models.Review, error) {
	query := `
		SELECT id, movie_id, user_name, rating, title, content, created_at, updated_at
		FROM reviews WHERE movie_id = $1
		ORDER BY created_at DESC
	`

	rows, err := db.Query(query, movieID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reviews []models.Review
	for rows.Next() {
		var r models.Review
		err := rows.Scan(
			&r.ID, &r.MovieID, &r.UserName, &r.Rating, &r.Title,
			&r.Content, &r.CreatedAt, &r.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		reviews = append(reviews, r)
	}

	return reviews, nil
}

func (db *DB) CreateReview(req models.CreateReviewRequest) (*models.Review, error) {
	query := `
		INSERT INTO reviews (movie_id, user_name, rating, title, content, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		RETURNING id, movie_id, user_name, rating, title, content, created_at, updated_at
	`

	var r models.Review
	err := db.QueryRow(query, req.MovieID, req.UserName, req.Rating, req.Title, req.Content).Scan(
		&r.ID, &r.MovieID, &r.UserName, &r.Rating, &r.Title,
		&r.Content, &r.CreatedAt, &r.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &r, nil
}

func (db *DB) GetRecentReviews(limit int) ([]models.Review, error) {
	query := `
		SELECT 
			r.id, r.movie_id, r.user_name, r.rating, r.title, r.content, r.created_at, r.updated_at,
			m.title as movie_title
		FROM reviews r
		JOIN movies m ON r.movie_id = m.id
		ORDER BY r.created_at DESC
		LIMIT $1
	`

	rows, err := db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reviews []models.Review
	for rows.Next() {
		var r models.Review
		var movieTitle string
		err := rows.Scan(
			&r.ID, &r.MovieID, &r.UserName, &r.Rating, &r.Title,
			&r.Content, &r.CreatedAt, &r.UpdatedAt, &movieTitle,
		)
		if err != nil {
			return nil, err
		}

		r.Movie = &models.Movie{Title: movieTitle}
		reviews = append(reviews, r)
	}

	return reviews, nil
}