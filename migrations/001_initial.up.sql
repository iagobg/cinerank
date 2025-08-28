-- Create movies table
CREATE TABLE IF NOT EXISTS movies (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    director VARCHAR(255) NOT NULL,
    year INTEGER NOT NULL,
    genre VARCHAR(100) NOT NULL,
    plot TEXT,
    poster_url TEXT,
    imdb_rating DECIMAL(3,1),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create reviews table
CREATE TABLE IF NOT EXISTS reviews (
    id SERIAL PRIMARY KEY,
    movie_id INTEGER NOT NULL REFERENCES movies(id) ON DELETE CASCADE,
    user_name VARCHAR(255) NOT NULL,
    rating INTEGER NOT NULL CHECK (rating >= 1 AND rating <= 5),
    title VARCHAR(255) NOT NULL,
    content TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_reviews_movie_id ON reviews(movie_id);
CREATE INDEX IF NOT EXISTS idx_movies_year ON movies(year);
CREATE INDEX IF NOT EXISTS idx_movies_genre ON movies(genre);
CREATE INDEX IF NOT EXISTS idx_reviews_created_at ON reviews(created_at);

-- Insert some sample data
INSERT INTO movies (title, director, year, genre, plot, poster_url, imdb_rating) VALUES
('The Shawshank Redemption', 'Frank Darabont', 1994, 'Drama', 'Two imprisoned men bond over a number of years, finding solace and eventual redemption through acts of common decency.', 'https://m.media-amazon.com/images/M/MV5BNDE3ODcxYzMtY2YzZC00NmNlLWJiNDMtZDViZWM2MzIxZDYwXkEyXkFqcGdeQXVyNjAwNDUxODI@._V1_.jpg', 9.3),
('The Godfather', 'Francis Ford Coppola', 1972, 'Crime', 'The aging patriarch of an organized crime dynasty transfers control of his clandestine empire to his reluctant son.', 'https://m.media-amazon.com/images/M/MV5BM2MyNjYxNmUtYTAwNi00MTYxLWJmNWYtYzZlODY3ZTk3OTFlXkEyXkFqcGdeQXVyNzkwMjQ5NzM@._V1_.jpg', 9.2),
('Pulp Fiction', 'Quentin Tarantino', 1994, 'Crime', 'The lives of two mob hitmen, a boxer, a gangster and his wife intertwine in four tales of violence and redemption.', 'https://m.media-amazon.com/images/M/MV5BNGNhMDIzZTUtNTBlZi00MTRlLWFjM2ItYzViMjE3YzI5MjljXkEyXkFqcGdeQXVyNzkwMjQ5NzM@._V1_.jpg', 8.9),
('Inception', 'Christopher Nolan', 2010, 'Sci-Fi', 'A thief who steals corporate secrets through dream-sharing technology is given the inverse task of planting an idea into the mind of a C.E.O.', 'https://m.media-amazon.com/images/M/MV5BMjAxMzY3NjcxNF5BMl5BanBnXkFtZTcwNTI5OTM0Mw@@._V1_.jpg', 8.8),
('The Dark Knight', 'Christopher Nolan', 2008, 'Action', 'When the menace known as the Joker wreaks havoc and chaos on the people of Gotham, Batman must accept one of the greatest psychological and physical tests of his ability to fight injustice.', 'https://m.media-amazon.com/images/M/MV5BMTMxNTMwODM0NF5BMl5BanBnXkFtZTcwODAyMTk2Mw@@._V1_.jpg', 9.0)
ON CONFLICT DO NOTHING;

-- Insert some sample reviews
INSERT INTO reviews (movie_id, user_name, rating, title, content) VALUES
(1, 'MovieBuff2024', 5, 'Absolute Masterpiece', 'This movie changed my perspective on life. The performances are outstanding and the story is deeply moving.'),
(1, 'CinemaLover', 5, 'Timeless Classic', 'Every time I watch this, I discover something new. Pure cinematic perfection.'),
(2, 'FilmCritic99', 5, 'The Gold Standard', 'Coppola created something truly special here. The character development and storytelling are unmatched.'),
(3, 'TarantinoFan', 4, 'Stylish and Bold', 'Tarantino at his best. The dialogue is razor-sharp and the non-linear narrative keeps you engaged.'),
(4, 'SciFiEnthusiast', 5, 'Mind-Bending Brilliance', 'Nolan creates a complex and layered story that rewards multiple viewings. Visually stunning.')
ON CONFLICT DO NOTHING;