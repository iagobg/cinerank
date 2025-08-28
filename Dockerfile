# ---- assets stage (Tailwind) ----
FROM node:20-alpine AS assets
WORKDIR /app
COPY package.json tailwind.config.js ./
COPY static ./static
RUN npm ci || npm install
RUN npx tailwindcss -i ./static/css/input.css -o ./static/css/output.css --minify

# ---- go build stage ----
FROM golang:1.25-alpine AS builder
WORKDIR /app

# Install build deps
RUN apk add --no-cache git

# Install templ and golang-migrate
RUN go install github.com/a-h/templ/cmd/templ@latest
RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Cache modules
COPY go.mod go.sum ./
RUN go mod download

# Copy sources
COPY . .

# Generate templ components
RUN templ generate
RUN go mod tidy

# Copy compiled CSS from assets stage
COPY --from=assets /app/static/css/output.css ./static/css/output.css

# Build server
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server

# ---- runtime stage ----
FROM alpine:latest
WORKDIR /app

# Install ca-certificates for HTTPS requests and postgresql-client for migrations
RUN apk --no-cache add ca-certificates postgresql-client

# Copy migration tool
COPY --from=builder /go/bin/migrate /usr/local/bin/migrate

# Copy application files
COPY --from=builder /app/server ./server
COPY --from=builder /app/static ./static
COPY --from=builder /app/migrations ./migrations

# Create entrypoint script
RUN echo '#!/bin/sh' > /app/entrypoint.sh && \
    echo 'set -e' >> /app/entrypoint.sh && \
    echo '' >> /app/entrypoint.sh && \
    echo '# Wait for database to be ready' >> /app/entrypoint.sh && \
    echo 'echo "Waiting for database..."' >> /app/entrypoint.sh && \
    echo 'until pg_isready -h $(echo $DATABASE_URL | sed "s/.*@\([^:]*\):.*/\1/") -U $(echo $DATABASE_URL | sed "s/.*:\/\/\([^:]*\):.*/\1/"); do' >> /app/entrypoint.sh && \
    echo '  echo "Database is unavailable - sleeping"' >> /app/entrypoint.sh && \
    echo '  sleep 1' >> /app/entrypoint.sh && \
    echo 'done' >> /app/entrypoint.sh && \
    echo '' >> /app/entrypoint.sh && \
    echo '# Run migrations' >> /app/entrypoint.sh && \
    echo 'echo "Running database migrations..."' >> /app/entrypoint.sh && \
    echo 'migrate -path ./migrations -database "$DATABASE_URL" up' >> /app/entrypoint.sh && \
    echo '' >> /app/entrypoint.sh && \
    echo '# Start the application' >> /app/entrypoint.sh && \
    echo 'echo "Starting CineRank server..."' >> /app/entrypoint.sh && \
    echo 'exec ./server' >> /app/entrypoint.sh && \
    chmod +x /app/entrypoint.sh

EXPOSE 8080

# Create non-root user
RUN adduser -D -s /bin/sh cinerank
USER cinerank

ENTRYPOINT ["/app/entrypoint.sh"]