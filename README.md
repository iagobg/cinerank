# Go + HTMX + Templ + Tailwind Starter

A minimal starter with Docker multi-stage build.

## Features
- Go `net/http`
- Templ components
- HTMX for progressive interactivity
- TailwindCSS (built in Docker assets stage)
- Distroless runtime image
- Makefile for common tasks

## Quickstart (local)
```bash
go install github.com/a-h/templ/cmd/templ@latest
npm install
templ generate
npm run build:css
go run ./cmd/server
# visit http://localhost:8080
```

For live dev, use two terminals:
```bash
templ generate --watch
npm run dev:css
go run ./cmd/server
```

## Docker
Build and run:
```bash
make docker-build
make docker-run
# visit http://localhost:8080
```

Push to a registry:
```bash
IMAGE=ghcr.io/youruser/myapp:latest make docker-build
IMAGE=ghcr.io/youruser/myapp:latest make docker-push
```

Then run on your platform of choice (Render, Fly.io, Railway, k8s, etc.).

## Project Layout
```
myapp/
├── cmd/server/main.go
├── internal/ui/*.templ
├── static/css/input.css -> Tailwind entry
├── static/css/output.css -> built CSS (generated)
├── Dockerfile (multi-stage: node assets + go build + distroless runtime)
├── Makefile
├── go.mod
├── package.json
└── tailwind.config.js
```

## Notes
- Generated Templ `.go` files are excluded from git and rebuilt in Docker.
- Runtime image is non-root on Distroless for security.
- HTMX demo endpoint: `GET /clicked` updates `#result` on the home page.
