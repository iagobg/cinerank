-- Create users table
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'user' CHECK (role IN ('user', 'admin')),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create tags table
CREATE TABLE IF NOT EXISTS tags (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL
);

-- Create movie_tags junction table
CREATE TABLE IF NOT EXISTS movie_tags (
    movie_id INTEGER NOT NULL REFERENCES movies(id) ON DELETE CASCADE,
    tag_id INTEGER NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (movie_id, tag_id)
);

-- Alter reviews to use user_id
ALTER TABLE reviews ADD COLUMN user_id INTEGER REFERENCES users(id) ON DELETE CASCADE;

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_movie_tags_movie_id ON movie_tags(movie_id);
CREATE INDEX IF NOT EXISTS idx_movie_tags_tag_id ON movie_tags(tag_id);
CREATE INDEX IF NOT EXISTS idx_reviews_user_id ON reviews(user_id);

-- Insert sample users (passwords are bcrypt hashes for 'password123')
INSERT INTO users (username, email, password_hash, role) VALUES
('MovieBuff2024', 'moviebuff@example.com', '$2a$10$K.0Hwpswrmq6JvLMMVaGJeC2gE26m8lo4Q5JgZCGH9z5RqK0fHgyq', 'user'),
('CinemaLover', 'cinemalover@example.com', '$2a$10$K.0Hwpswrmq6JvLMMVaGJeC2gE26m8lo4Q5JgZCGH9z5RqK0fHgyq', 'user'),
('FilmCritic99', 'filmcritic@example.com', '$2a$10$K.0Hwpswrmq6JvLMMVaGJeC2gE26m8lo4Q5JgZCGH9z5RqK0fHgyq', 'user'),
('TarantinoFan', 'tarantinofan@example.com', '$2a$10$K.0Hwpswrmq6JvLMMVaGJeC2gE26m8lo4Q5JgZCGH9z5RqK0fHgyq', 'user'),
('SciFiEnthusiast', 'scifienthusiast@example.com', '$2a$10$K.0Hwpswrmq6JvLMMVaGJeC2gE26m8lo4Q5JgZCGH9z5RqK0fHgyq', 'user'),
('AdminUser', 'admin@example.com', '$2a$10$K.0Hwpswrmq6JvLMMVaGJeC2gE26m8lo4Q5JgZCGH9z5RqK0fHgyq', 'admin')
ON CONFLICT DO NOTHING;

-- Insert sample tags
INSERT INTO tags (name) VALUES
('Drama'), ('Crime'), ('Sci-Fi'), ('Action')
ON CONFLICT DO NOTHING;

-- Insert sample movie_tags (assuming movie IDs 1-5)
INSERT INTO movie_tags (movie_id, tag_id) VALUES
(1, 1),  -- Shawshank: Drama
(2, 2),  -- Godfather: Crime
(3, 2),  -- Pulp Fiction: Crime
(4, 3),  -- Inception: Sci-Fi
(5, 4)   -- Dark Knight: Action
ON CONFLICT DO NOTHING;

-- Update sample reviews with user_ids (assuming review IDs 1-5 correspond to users 1-5)
UPDATE reviews SET user_id = 1 WHERE id = 1;
UPDATE reviews SET user_id = 2 WHERE id = 2;
UPDATE reviews SET user_id = 3 WHERE id = 3;
UPDATE reviews SET user_id = 4 WHERE id = 4;
UPDATE reviews SET user_id = 5 WHERE id = 5;

-- Remove user_name from reviews after migration
ALTER TABLE reviews DROP COLUMN IF EXISTS user_name;

-- Insert sample reviews if not exists (but since initial has them, assume they are there)