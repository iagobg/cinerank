# 🎬 CineRank

A modern movie review web application built with Go, HTMX, Templ, Tailwind CSS, and PostgreSQL.

## Features

- **Movie Management**: Add and browse movies with detailed information
- **User Reviews**: Write and read movie reviews with 5-star ratings
- **Responsive Design**: Modern, mobile-first UI with Tailwind CSS
- **Interactive UI**: Dynamic interactions powered by HTMX
- **PostgreSQL Database**: Robust data persistence with Neon or any PostgreSQL database
- **Docker Support**: Easy deployment with Docker containerization
- **RESTful API**: JSON API endpoints for programmatic access

## Tech Stack

- **Backend**: Go 1.25+ with standard library HTTP server
- **Frontend**: Templ templates + HTMX + Tailwind CSS
- **Database**: PostgreSQL (Neon compatible)
- **Deployment**: Docker with multi-stage builds
- **CSS Processing**: Tailwind CLI
- **Database Migrations**: golang-migrate

## Quick Start

### Prerequisites

- Go 1.25 or later
- Node.js 20 or later
- PostgreSQL database (or Neon account)
- Docker (optional, for containerized deployment)

### 1. Clone and Setup

```bash
git clone <your-repo-url>
cd cinerank

# Install development dependencies
make install-deps

# Install npm dependencies
npm install
```

### 2. Database Setup

#### Option A: Using Neon (Recommended)

1. Create a free account at [Neon](https://neon.tech)
2. Create a new database
3. Copy your connection string

#### Option B: Local PostgreSQL

```bash
# Install PostgreSQL locally
# Create a database named 'cinerank'
createdb cinerank
```

### 3. Environment Configuration

```bash
# Copy the example environment file
cp .env.example .env

# Edit .env and update DATABASE_URL with your connection string
# For Neon: postgresql://username:password@hostname/database?sslmode=require
# For local: postgresql://username:password@localhost/cinerank?sslmode=disable
```

### 4. Run Migrations

```bash
# Set your DATABASE_URL environment variable
export DATABASE_URL="your_postgresql_connection_string"

# Run database migrations
make migrate-up
```

### 5. Development

Start the development servers (run each in a separate terminal):

```bash
# Terminal 1: Watch and compile Templ templates
templ generate --watch

# Terminal 2: Watch and compile Tailwind CSS
npm run dev:css

# Terminal 3: Run the Go server
make run
```

Visit `http://localhost:8080` to see your movie review app!

## Docker Deployment

### Build and Run with Docker

```bash
# Build the Docker image
make docker-build

# Run with your database URL
export DATABASE_URL="your_postgresql_connection_string"
make docker-run
```

### Using Docker Compose (Development)

```bash
# Start with local PostgreSQL
docker-compose -f docker-compose.dev.yml up --build

# The app will be available at http://localhost:8080
# PostgreSQL will be available at localhost:5432
```

## API Endpoints

### Movies
- `GET /api/movies` - List all movies with stats
- `POST /api/movies` - Create a new movie
- `GET /api/movies/{id}` - Get movie by ID

### Reviews
- `GET /api/reviews` - Get recent reviews
- `GET /api/reviews?movie_id={id}` - Get reviews for a specific movie
- `POST /api/reviews` - Create a new review

### Example API Usage

```bash
# Get all movies
curl http://localhost:8080/api/movies

# Add a new movie
curl -X POST http://localhost:8080/api/movies \
  -H "Content-Type: application/json" \
  -d '{
    "title": "The Matrix",
    "director": "The Wachowskis",
    "year": 1999,
    "genre": "Sci-Fi",
    "plot": "A computer programmer discovers reality is a simulation.",
    "imdb_rating": 8.7
  }'

# Add a review
curl -X POST http://localhost:8080/api/reviews \
  -H "Content-Type: application/json" \
  -d '{
    "movie_id": 1,
    "user_name": "John Doe",
    "rating": 5,
    "title": "Mind-blowing!",
    "content": "This movie changed everything I thought about reality."
  }'
```

## Project Structure

```
cinerank/
├── cmd/
│   └── server/          # Application entry point
├── internal/
│   ├── database/        # Database layer
│   ├── handlers/        # HTTP handlers
│   ├── models/          # Data models
│   └── ui/             # Templ templates
├── migrations/          # Database migrations
├── static/
│   └── css/            # Stylesheets
├── docker-compose.dev.yml
├── Dockerfile
├── Makefile
└── README.md
```

## Available Commands

```bash
# Development
make dev          # Show development startup instructions
make install-deps # Install development dependencies
make templ        # Generate templ templates
make css          # Build Tailwind CSS
make run          # Run the server locally

# Database
make migrate-up   # Run database migrations
make migrate-down # Rollback database migrations
make db-reset     # Reset database (drop all tables and recreate)

# Docker
make docker-build # Build Docker image
make docker-run   # Run Docker container
make clean        # Clean up build artifacts

# Help
make help         # Show all available commands
```

## Environment Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `DATABASE_URL` | PostgreSQL connection string | `postgresql://user:pass@host/db?sslmode=require` |
| `PORT` | Server port (default: 8080) | `8080` |
| `ENVIRONMENT` | Environment mode | `development` or `production` |

## Deployment

### Deploying to Production

1. **Build the Docker image:**
   ```bash
   docker build -t cinerank:latest .
   ```

2. **Deploy to your platform:**
   - **Railway**: Connect your GitHub repo and set `DATABASE_URL`
   - **Render**: Connect repo, set build command to `docker build`
   - **Fly.io**: Use `flyctl deploy`
   - **DigitalOcean App Platform**: Connect repo and configure environment

3. **Required Environment Variables:**
   - `DATABASE_URL`: Your Neon or PostgreSQL connection string
   - `PORT`: Usually set automatically by the platform

### Neon Database Configuration

1. Create a Neon project at [console.neon.tech](https://console.neon.tech)
2. Copy the connection string from your dashboard
3. Set `DATABASE_URL` in your deployment platform
4. The app will automatically run migrations on startup

## Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature-name`
3. Make your changes and test locally
4. Commit with descriptive messages
5. Push and create a Pull Request

## License

This project is licensed under the MIT License. See the LICENSE file for details.

## Troubleshooting

### Common Issues

**Templates not updating:**
- Make sure `templ generate --watch` is running
- Check that `*_templ.go` files are being generated

**CSS not updating:**
- Ensure `npm run dev:css` is running
- Verify `static/css/output.css` is being updated

**Database connection errors:**
- Verify your `DATABASE_URL` is correct
- For Neon, ensure you're using `sslmode=require`
- For local PostgreSQL, you might need `sslmode=disable`

**Docker build fails:**
- Ensure all dependencies are properly listed in `go.mod`
- Check that the Dockerfile can access all necessary files

### Getting Help

- Check existing [Issues](../../issues)
- Create a new issue with detailed information
- Include logs and error messages

---

Built with ❤️ using Go, HTMX, Templ, and Tailwind CSS