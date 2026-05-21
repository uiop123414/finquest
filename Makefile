.PHONY: install dev backend frontend up down migrate lint test start

# ─── Быстрый старт ───────────────────────────────────────────────────────────
start:
	@test -f .env || cp .env.example .env
	docker-compose up --build -d
	@echo ""
	@echo "✓  FinQuest запущен!"
	@echo "   Фронт  →  http://localhost:3000"
	@echo "   API    →  http://localhost:8000"
	@echo ""
	@echo "   Демо-аккаунт (100 транзакций уже загружены):"
	@echo "   Email   →  demo@finquest.ru"
	@echo "   Пароль  →  demo123"
	@echo ""
	@echo "   Логи   →  make logs"

# ─── Разработка (локально, без Docker) ───────────────────────────────────────
install:
	cd frontend && npm install

dev:
	@which concurrently > /dev/null || npm install -g concurrently
	concurrently "make backend" "make frontend"

backend:
	cd backend && go run ./main.go

frontend:
	cd frontend && npm run dev

# ─── Сборка ───────────────────────────────────────────────────────────────────
build:
	cd backend && go build -o bin/finquest ./main.go

# ─── Миграции (локально, без Docker) ─────────────────────────────────────────
migrate:
	docker run --rm \
		-v $(PWD)/backend/db/migrations:/migrations \
		--network host \
		migrate/migrate \
		-path=/migrations \
		-database "$(DATABASE_URL)" \
		up

migrate-down:
	docker run --rm \
		-v $(PWD)/backend/db/migrations:/migrations \
		--network host \
		migrate/migrate \
		-path=/migrations \
		-database "$(DATABASE_URL)" \
		down 1

# ─── Качество ─────────────────────────────────────────────────────────────────
lint:
	cd backend && golangci-lint run ./...
	cd frontend && npm run lint

test:
	cd backend && go test ./... -v
	cd frontend && npm run test

# ─── Docker ───────────────────────────────────────────────────────────────────
up:
	docker-compose up --build -d

down:
	docker-compose down

logs:
	docker-compose logs -f
