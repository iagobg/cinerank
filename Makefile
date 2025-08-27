APP_NAME=cinerank
IMAGE?=$(APP_NAME):latest

.PHONY: dev templ css run docker-build docker-run docker-push clean

dev:
	@echo "Start dev: run these in separate terminals:"
	@echo "  1) templ generate --watch"
	@echo "  2) npm install && npm run dev:css"
	@echo "  3) go run ./cmd/server"

templ:
	templ generate

css:
	npm run build:css

run:
	go run ./cmd/server

docker-build:
	docker build -t $(IMAGE) .

docker-run:
	docker run --rm -p 8080:8080 $(IMAGE)

docker-push:
	docker push $(IMAGE)

clean:
	rm -rf ./tmp ./dist
