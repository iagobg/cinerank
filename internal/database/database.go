package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

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
func (db *DB) GetAllMoviesWithStats(searchQuery string) ([]models.MovieWithStats, error) {
	query := `
		SELECT 
			m.id, m.title, m.director, m.year, m.plot, 
			m.poster_url, m.imdb_rating, m.created_at, m.updated_at,
			COALESCE(STRING_AGG(t.name, ', '), '') as tags,
			COALESCE(COUNT(r.id), 0) as review_count,
			COALESCE(AVG(r.rating::float), 0) as average_rating
		FROM movies m
		LEFT JOIN movie_tags mt ON m.id = mt.movie_id
		LEFT JOIN tags t ON mt.tag_id = t.id
		LEFT JOIN reviews r ON m.id = r.movie_id
	`

	args := []interface{}{}
	if searchQuery != "" {
		searchQuery = "%" + strings.ToLower(searchQuery) + "%"
		query += ` WHERE LOWER(m.title) LIKE $1 OR EXISTS (
			SELECT 1 FROM movie_tags mt2 JOIN tags t2 ON mt2.tag_id = t2.id 
			WHERE mt2.movie_id = m.id AND LOWER(t2.name) LIKE $1
		)`
		args = append(args, searchQuery)
	}

	query += `
		GROUP BY m.id, m.title, m.director, m.year, m.plot, 
				 m.poster_url, m.imdb_rating, m.created_at, m.updated_at
		ORDER BY m.created_at DESC
	`

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var movies []models.MovieWithStats
	for rows.Next() {
		var m models.MovieWithStats
		var tagsStr string
		err := rows.Scan(
			&m.ID, &m.Title, &m.Director, &m.Year, &m.Plot,
			&m.PosterURL, &m.IMDBRating, &m.CreatedAt, &m.UpdatedAt,
			&tagsStr,
			&m.ReviewCount, &m.AverageRating,
		)
		if err != nil {
			return nil, err
		}
		if tagsStr != "" {
			m.Tags = strings.Split(tagsStr, ", ")
		}
		movies = append(movies, m)
	}

	return movies, nil
}

func (db *DB) GetMovieByID(id int) (*models.Movie, error) {
	query := `
		SELECT m.id, m.title, m.director, m.year, m.plot, m.poster_url, m.imdb_rating, m.created_at, m.updated_at,
			   COALESCE(STRING_AGG(t.name, ', '), '') as tags
		FROM movies m
		LEFT JOIN movie_tags mt ON m.id = mt.movie_id
		LEFT JOIN tags t ON mt.tag_id = t.id
		WHERE m.id = $1
		GROUP BY m.id
	`

	var m models.Movie
	var tagsStr string
	err := db.QueryRow(query, id).Scan(
		&m.ID, &m.Title, &m.Director, &m.Year, &m.Plot,
		&m.PosterURL, &m.IMDBRating, &m.CreatedAt, &m.UpdatedAt,
		&tagsStr,
	)
	if err != nil {
		return nil, err
	}
	if tagsStr != "" {
		m.Tags = strings.Split(tagsStr, ", ")
	}

	return &m, nil
}

func (db *DB) CreateMovie(req models.CreateMovieRequest) (*models.Movie, error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO movies (title, director, year, plot, poster_url, imdb_rating, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
		RETURNING id, title, director, year, plot, poster_url, imdb_rating, created_at, updated_at
	`

	var m models.Movie
	err = tx.QueryRow(query, req.Title, req.Director, req.Year, req.Plot, req.PosterURL, req.IMDBRating).Scan(
		&m.ID, &m.Title, &m.Director, &m.Year, &m.Plot,
		&m.PosterURL, &m.IMDBRating, &m.CreatedAt, &m.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	for _, tagName := range req.Tags {
		if tagName == "" {
			continue
		}
		tagID, err := db.getOrCreateTag(tx, tagName)
		if err != nil {
			return nil, err
		}
		_, err = tx.Exec("INSERT INTO movie_tags (movie_id, tag_id) VALUES ($1, $2) ON CONFLICT DO NOTHING", m.ID, tagID)
		if err != nil {
			return nil, err
		}
		m.Tags = append(m.Tags, tagName)
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &m, nil
}

func (db *DB) getOrCreateTag(tx *sql.Tx, name string) (int, error) {
	var tagID int
	err := tx.QueryRow("SELECT id FROM tags WHERE name = $1", name).Scan(&tagID)
	if err == sql.ErrNoRows {
		err = tx.QueryRow("INSERT INTO tags (name) VALUES ($1) RETURNING id", name).Scan(&tagID)
		if err != nil {
			return 0, err
		}
	} else if err != nil {
		return 0, err
	}
	return tagID, nil
}

func (db *DB) DeleteMovie(id int) error {
	_, err := db.Exec("DELETE FROM movies WHERE id = $1", id)
	return err
}

// Review operations
func (db *DB) GetReviewsByMovieID(movieID int) ([]models.Review, error) {
	query := `
		SELECT r.id, r.movie_id, r.user_id, r.rating, r.title, r.content, r.created_at, r.updated_at,
			   u.username
		FROM reviews r
		JOIN users u ON r.user_id = u.id
		WHERE r.movie_id = $1
		ORDER BY r.created_at DESC
	`

	rows, err := db.Query(query, movieID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reviews []models.Review
	for rows.Next() {
		var r models.Review
		var username string
		err := rows.Scan(
			&r.ID, &r.MovieID, &r.UserID, &r.Rating, &r.Title,
			&r.Content, &r.CreatedAt, &r.UpdatedAt,
			&username,
		)
		if err != nil {
			return nil, err
		}
		r.User = &models.User{Username: username}
		reviews = append(reviews, r)
	}

	return reviews, nil
}

func (db *DB) CreateReview(req models.CreateReviewRequest, userID int) (*models.Review, error) {
	query := `
		INSERT INTO reviews (movie_id, user_id, rating, title, content, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		RETURNING id, movie_id, user_id, rating, title, content, created_at, updated_at
	`

	var r models.Review
	err := db.QueryRow(query, req.MovieID, userID, req.Rating, req.Title, req.Content).Scan(
		&r.ID, &r.MovieID, &r.UserID, &r.Rating, &r.Title,
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
			r.id, r.movie_id, r.user_id, r.rating, r.title, r.content, r.created_at, r.updated_at,
			m.title as movie_title,
			u.username
		FROM reviews r
		JOIN movies m ON r.movie_id = m.id
		JOIN users u ON r.user_id = u.id
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
		var username string
		err := rows.Scan(
			&r.ID, &r.MovieID, &r.UserID, &r.Rating, &r.Title,
			&r.Content, &r.CreatedAt, &r.UpdatedAt, &movieTitle, &username,
		)
		if err != nil {
			return nil, err
		}

		r.Movie = &models.Movie{Title: movieTitle}
		r.User = &models.User{Username: username}
		reviews = append(reviews, r)
	}

	return reviews, nil
}

// User operations
func (db *DB) CreateUser(req models.RegisterRequest) (*models.User, error) {
	query := `
		INSERT INTO users (username, email, password_hash, role, created_at, updated_at)
		VALUES ($1, $2, $3, 'user', NOW(), NOW())
		RETURNING id, username, email, role, created_at, updated_at
	`

	var u models.User
	err := db.QueryRow(query, req.Username, req.Email, req.Password).Scan(
		&u.ID, &u.Username, &u.Email, &u.Role, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &u, nil
}

func (db *DB) GetUserByEmail(email string) (*models.User, error) {
	query := `
		SELECT id, username, email, password_hash, role, created_at, updated_at
		FROM users WHERE email = $1
	`

	var u models.User
	err := db.QueryRow(query, email).Scan(
		&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.Role, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &u, nil
}

func (db *DB) GetUserByID(id int) (*models.User, error) {
	query := `
		SELECT id, username, email, role, created_at, updated_at
		FROM users WHERE id = $1
	`

	var u models.User
	err := db.QueryRow(query, id).Scan(
		&u.ID, &u.Username, &u.Email, &u.Role, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &u, nil
}

func (db *DB) GetAllUsers() ([]models.User, error) {
	query := `
		SELECT id, username, email, role, created_at, updated_at
		FROM users ORDER BY created_at DESC
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		err := rows.Scan(
			&u.ID, &u.Username, &u.Email, &u.Role, &u.CreatedAt, &u.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}

	return users, nil
}

func (db *DB) DeleteUser(id int) error {
	_, err := db.Exec("DELETE FROM users WHERE id = $1", id)
	return err
}