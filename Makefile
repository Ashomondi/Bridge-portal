.PHONY: help dev up down test backend-run frontend-run ai-run ci migrate-up migrate-down clean

help:
	@echo "Ikonex Academy - Development Commands:"
	@echo "  make dev          - Run all services via Docker Compose"
	@echo "  make up           - Start Postgres and Redis containers in the background"
	@echo "  make down         - Stop all Docker containers"
	@echo "  make backend-run  - Run Go backend locally"
	@echo "  make frontend-run - Run Next.js frontend locally"
	@echo "  make ai-run       - Run Python AI service locally"
	@echo "  make test         - Run tests across backend and AI service"
	@echo "  make ci           - Run the repository CI gate"
	@echo "  make migrate-up   - Apply SQL migrations"
	@echo "  make clean        - Clean temporary build and cache files"

up:
	docker compose up -d postgres redis

down:
	docker compose down

dev:
	docker compose up --build

backend-run:
	cd backend && go run cmd/server/main.go

frontend-run:
	cd frontend && npm run dev

ai-run:
	cd ai_service && uvicorn app.main:app --host 0.0.0.0 --port 8000 --reload

ci:
	@set -eu; \
	cd backend && gofmt -l . > /tmp/ikonex-gofmt.out; \
	if [ -s /tmp/ikonex-gofmt.out ]; then cat /tmp/ikonex-gofmt.out; exit 1; fi; \
	go vet ./...; \
	go test -v ./...; \
	cd ../frontend && npm ci --no-audit --no-fund; \
	npm run lint; \
	npm run build; \
	cd ../ai_service && python3 -m venv .venv && . .venv/bin/activate && \
	pip install --upgrade pip >/dev/null && pip install -r requirements.txt >/dev/null && \
	pytest -q

test:
	@echo "Running Go Backend tests..."
	cd backend && go test -v ./...
	@echo "Running Python AI Service tests..."
	cd ai_service && python3 -m venv .venv && . .venv/bin/activate && pip install -r requirements.txt >/dev/null && pytest -q || true

clean:
	rm -rf backend/bin frontend/.next frontend/node_modules ai_service/__pycache__ ai_service/.venv
