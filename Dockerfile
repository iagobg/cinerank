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

# Install templ
RUN go install github.com/a-h/templ/cmd/templ@latest


# Cache modules
COPY go.mod ./
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
FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=builder /app/server ./server
COPY --from=builder /app/static ./static

EXPOSE 8080
USER 65532:65532
ENTRYPOINT ["/app/server"]
