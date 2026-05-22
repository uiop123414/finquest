.PHONY: install dev backend frontend up down migrate migrate-down lint lint-go lint-js test test-go test-js build start logs help

# ─── Быстрый старт ───────────────────────────────────────────────────────────
start:
	@[ -f .env ] || cp .env.example .env
	docker-compose up --build -d
	@echo ""
	@echo "✓  FinQuest запущен!"
	@echo "   Фронт  →  http://localhost:3000"
	@echo "   API    →  http://localhost:8000"
	@echo ""
	@echo "   Демо-аккаунт (транзакции, депозиты, кредиты, цели уже загружены):"
	@echo "   Email   →  demo@finquest.ru"
	@echo "   Пароль  →  demo123"
	@echo ""
	@echo "   AI советник:"
	@echo "   Бесплатно  →  добавьте GEMINI_API_KEY в .env"
	@echo "   Платно     →  добавьте ANTHROPIC_API_KEY в .env"
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
lint: lint-go lint-js

lint-go:
	cd backend && golangci-lint run --config .golangci.yml ./...

lint-js:
	cd frontend && npm run lint

# ─── Тесты ────────────────────────────────────────────────────────────────────
test: test-go test-js

test-go:
	cd backend && go test ./... -v

test-js:
	cd frontend && npm run test

# Интеграционные тесты (поднимают postgres через testcontainers, Docker обязателен)
test-integration:
	cd backend && go test -tags=integration -v ./...

# ─── Docker ───────────────────────────────────────────────────────────────────
up:
	docker-compose up --build -d

down:
	docker-compose down

down-volumes:
	docker-compose down -v

logs:
	docker-compose logs -f

# ─── Помощь ───────────────────────────────────────────────────────────────────
help:
	@echo "Доступные команды:"
	@echo "  make start             — собрать и запустить всё через Docker"
	@echo "  make dev               — локальный запуск backend + frontend"
	@echo "  make install           — установить npm-зависимости"
	@echo "  make build             — скомпилировать backend"
	@echo "  make migrate           — применить миграции БД"
	@echo "  make migrate-down      — откатить последнюю миграцию"
	@echo "  make lint              — запустить golangci-lint + eslint"
	@echo "  make test              — unit-тесты Go + JS"
	@echo "  make test-integration  — интеграционные тесты (нужен Docker, БД поднимается автоматически)"
	@echo "  make logs              — логи Docker-контейнеров"
	@echo "  make down              — остановить контейнеры"
	@echo "  make down-volumes      — остановить контейнеры + удалить данные БД"