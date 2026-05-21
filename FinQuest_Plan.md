# FinQuest — План реализации

**Проект:** FinQuest — геймифицированный AI-ассистент личных финансов  
**Стек:** React + TypeScript · Go + Gin · PostgreSQL · JWT · Claude API · Docker Compose  
**Срок:** 10 рабочих дней + буфер  
**Источники:** ТЗ (FinQuest TZ.pages) + Notion task board (41 задача, FQ-1…FQ-41)

---

## 1. Структура монорепо

```
finquest/
├── Makefile
├── docker-compose.yml
├── .env.example
├── .github/
│   └── workflows/
│       └── ci.yml              # lint + smoke на каждый PR (FQ-4)
│
├── docs/
│   ├── PROJECT_CHARTER.md      # FQ-5
│   ├── TEST_REPORT.md          # FQ-34
│   ├── DEPLOYMENT.md           # FQ-35
│   └── diagrams/               # 6 диаграмм (FQ-6,10,22,29,36)
│       ├── use_case.png
│       ├── component.png
│       ├── class.png
│       ├── sequence_ai_import.png
│       ├── activity_xp.png
│       └── deployment.png
│
├── backend/
│   ├── Dockerfile
│   ├── go.mod
│   ├── go.sum
│   ├── main.go                  # Gin app, роутеры, CORS
│   ├── schema.sql               # FQ-3: init-скрипт (9 таблиц + seed)
│   ├── config/
│   │   └── config.go            # env-переменные через os.Getenv
│   ├── db/
│   │   ├── db.go                # pgx/sqlx connection pool
│   │   └── migrations/          # golang-migrate файлы
│   │       ├── 001_init.up.sql
│   │       └── 001_init.down.sql
│   ├── models/
│   │   └── models.go            # Go-структуры для всех 9 таблиц
│   ├── handlers/                # HTTP-хендлеры (тонкий слой)
│   │   ├── auth.go              # /api/v1/auth/*
│   │   ├── transactions.go      # /api/v1/transactions/*
│   │   ├── categories.go        # /api/v1/categories
│   │   ├── analytics.go         # /api/v1/analytics/*
│   │   ├── gamification.go      # /api/v1/gamification/profile
│   │   ├── ai.go                # /api/v1/ai/chat (SSE)
│   │   └── goals.go             # /api/v1/goals
│   ├── middleware/
│   │   └── auth.go              # JWT middleware для Gin
│   └── services/                # бизнес-логика
│       ├── ai_categorize.go     # FQ-17: Claude API, few-shot
│       ├── rule_based.go        # FQ-19: fallback по словарю
│       └── gamification.go      # FQ-23,24: calculate_level, award_xp
│
└── frontend/
    ├── Dockerfile
    ├── package.json
    ├── vite.config.ts
    ├── tsconfig.json
    └── src/
        ├── main.tsx
        ├── App.tsx
        ├── api/
        │   └── client.ts            # axios + JWT interceptors
        ├── pages/
        │   ├── LoginPage.tsx         # FQ-9
        │   ├── DashboardPage.tsx     # FQ-21
        │   ├── TransactionsPage.tsx  # FQ-13
        │   ├── ImportPage.tsx        # FQ-16
        │   ├── AchievementsPage.tsx  # FQ-27
        │   └── GoalsPage.tsx         # FQ-30
        └── components/
            ├── XpBar.tsx             # FQ-26
            ├── AchievementToast.tsx  # FQ-28
            └── AiChat.tsx            # FQ-39 (буфер)
```

---

## 2. База данных — 9 таблиц

```sql
-- schema.sql (FQ-3)
users             -- id, email, hashed_password, xp_total, level, created_at
transactions      -- id, user_id, amount, type, category_id, date, note,
                  --   external_id (дедупликация), ai_confidence
categories        -- id, user_id(nullable), name, is_system
xp_events         -- id, user_id, delta, reason, created_at
achievements      -- id, code, name, description
user_achievements -- id, user_id, achievement_id, earned_at
goals             -- id, user_id, name, target_amount, current_amount, deadline
challenges        -- id, name, description, xp_reward, condition_json
user_challenges   -- id, user_id, challenge_id, accepted_at, completed_at
```

Seed: системные категории (Еда, Транспорт, ЖКХ…) + стартовые ачивки.

---

## 3. Go-зависимости (`go.mod`)

```
github.com/gin-gonic/gin              # HTTP-роутер
github.com/gin-contrib/cors           # CORS middleware
github.com/golang-jwt/jwt/v5          # JWT access + refresh
golang.org/x/crypto                   # bcrypt
github.com/jmoiron/sqlx               # SQL + named queries
github.com/lib/pq                     # PostgreSQL driver
github.com/golang-migrate/migrate/v4  # DB-миграции
github.com/google/uuid                # UUID для ID
github.com/joho/godotenv              # .env файл локально
github.com/anthropics/anthropic-sdk-go # Claude API (FQ-17)
encoding/csv                          # stdlib, парсер CSV (FQ-14)
```

---

## 4. API-эндпоинты (`/api/v1/...`)

| Метод | Путь | FQ | Что делает |
|---|---|---|---|
| POST | /auth/register | FQ-8 | Регистрация, bcrypt cost=12 |
| POST | /auth/login | FQ-8 | Access + refresh JWT |
| POST | /auth/refresh | FQ-8 | Обновление access-токена |
| GET | /categories | FQ-12 | Системные + пользовательские |
| POST | /categories | FQ-12 | Создать пользовательскую |
| GET | /transactions | FQ-11 | Фильтры date_from/to/category_id, пагинация |
| POST | /transactions | FQ-11 | Добавить транзакцию |
| PATCH | /transactions/:id | FQ-11 | Изменить |
| DELETE | /transactions/:id | FQ-11 | Удалить |
| POST | /transactions/import | FQ-15 | Multipart CSV, дедупликация по external_id |
| GET | /analytics/summary | FQ-20 | {income, expense, balance, by_category} |
| GET | /analytics/over-time | FQ-40 | Динамика по неделям/месяцам |
| GET | /gamification/profile | FQ-25 | {xp_total, level, level_progress_pct, achievements} |
| POST | /ai/chat | FQ-39 | SSE-стрим ответа Claude (буфер) |
| GET | /goals | FQ-30 | Список целей |
| POST | /goals | FQ-30 | Создать цель |
| GET | /health | — | {status: ok} |

---

## 5. Ключевые модули бэкенда (Go)

### Авторизация (`middleware/auth.go`)
```go
// bcrypt из golang.org/x/crypto, cost=12
// JWT: github.com/golang-jwt/jwt/v5
// access-токен: 30 мин, refresh: 7 дней
// Gin middleware: AuthRequired() — парсит Bearer, кладёт userID в ctx
```

### AI-категоризация (`services/ai_categorize.go`, FQ-17)
```go
// Claude API через anthropic-sdk-go
// Few-shot промпт: список категорий + 3-5 примеров
// func CategorizeBatch(txs []Transaction) ([]CategoryResult, error)
// Unit-тест с моком HTTP-клиента + 1 smoke на 5 транзакциях
```

### Rule-based fallback (`services/rule_based.go`, FQ-19)
```go
// map[string]string{"магнит": "food", "пятёрочка": "food", ...}
// func RuleBasedCategorize(description string) (categoryID string, ok bool)
// Используется если Claude API недоступен или confidence < порога
```

### CSV-парсер (`handlers/transactions.go`, FQ-14)
```go
// encoding/csv из stdlib — нет внешних зависимостей
// func parseCSV(r io.Reader) ([]TransactionRow, error)
// Тесты: валидный файл, кривые колонки, пустой файл
```

### Геймификация (`services/gamification.go`, FQ-23, FQ-24)
```go
// Чистые функции (FQ-23):
func CalculateLevel(xp int) int { return xp / 100 }
func CheckUnlockableAchievements(state UserState) []string { ... }

// Сервис с эффектами (FQ-24):
func AwardXP(ctx context.Context, db *sqlx.DB, userID uuid.UUID, delta int, reason string) error
// → INSERT xp_event → UPDATE user.xp → проверка ачивок
// Integration-тест: импорт CSV → AwardXP → ачивка "first_import"
```

---

## 6. Фронтенд (без изменений)

**Стек:** Vite + React 18 + TypeScript + Tailwind CSS + Recharts + React Router v6

```json
{
  "dependencies": {
    "react": "^18",
    "react-router-dom": "^6",
    "axios": "^1",
    "recharts": "^2",
    "tailwindcss": "^3"
  },
  "devDependencies": {
    "@testing-library/react": "^14",
    "vitest": "^1"
  }
}
```

**Страницы:** Login · Dashboard (Pie chart) · Transactions · Import CSV · Achievements · Goals  
**Компоненты:** XpBar (RTL-тест) · AchievementToast · AiChat (SSE, буфер)

---

## 7. Makefile

```makefile
.PHONY: install dev backend frontend up down migrate lint test

install:
	cd frontend && npm install

dev:
	npm install -g concurrently
	concurrently "make backend" "make frontend"

backend:
	cd backend && go run ./main.go

frontend:
	cd frontend && npm run dev

build:
	cd backend && go build -o bin/finquest ./main.go

migrate:
	cd backend && migrate -path db/migrations -database "$$DATABASE_URL" up

lint:
	cd backend && golangci-lint run ./...
	cd frontend && npm run lint

test:
	cd backend && go test ./...
	cd frontend && npm run test

up:
	docker-compose up --build -d

down:
	docker-compose down
```

```bash
# Запуск с нуля:
cp .env.example .env    # вписать ANTHROPIC_API_KEY, JWT_SECRET, DATABASE_URL
make install
make up                 # postgres + backend + frontend

# Разработка (hot reload):
make migrate && make dev
# Фронт  → http://localhost:5173
# Swagger → http://localhost:8000/docs (gin-swagger)
```

---

## 8. Docker Compose

```yaml
version: "3.9"
services:
  db:
    image: postgres:16-alpine
    environment:
      POSTGRES_DB: finquest
      POSTGRES_USER: finquest
      POSTGRES_PASSWORD: finquest
    volumes:
      - pgdata:/var/lib/postgresql/data
      - ./backend/schema.sql:/docker-entrypoint-initdb.d/schema.sql

  backend:
    build: ./backend
    ports: ["8000:8000"]
    env_file: .env
    depends_on: [db]

  frontend:
    build: ./frontend
    ports: ["3000:80"]
    depends_on: [backend]

volumes:
  pgdata:
```

```bash
# .env.example
DATABASE_URL=postgresql://finquest:finquest@db/finquest?sslmode=disable
ANTHROPIC_API_KEY=sk-ant-...
JWT_SECRET=supersecretkey
PORT=8000
```

### Dockerfile для Go (multi-stage, маленький образ)
```dockerfile
# backend/Dockerfile
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o finquest ./main.go

FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/finquest .
EXPOSE 8000
CMD ["./finquest"]
```

---

## 9. Этапы по дням

### День 1 — Инфраструктура (FQ-1, FQ-2, FQ-3, FQ-4, FQ-5, FQ-6)
- Репо, структура монорепо, защита `main`
- `docker compose up` → 3 контейнера, `/health → {status: ok}`, `:5173` открывается
- `schema.sql` → 9 таблиц + seed
- GitHub Actions CI (golangci-lint + go test + npm lint/test)
- `PROJECT_CHARTER.md` подписан
- Use Case + Component Diagram

### День 2 — Модели и авторизация (FQ-7, FQ-8, FQ-9, FQ-10)
- Go-структуры для всех 9 таблиц + первая миграция (`golang-migrate`)
- `/auth/register`, `/auth/login`, `/auth/refresh` + JWT middleware
- Integration-тесты: регистрация, дубликат email, логин, неверный пароль
- UI: форма логина/регистрации + защищённый layout
- Class Diagram

### День 3 — Транзакции и категории (FQ-11, FQ-12, FQ-13)
- CRUD `/transactions` с фильтрами + пагинацией, тесты
- `/categories` (GET + POST)
- UI: таблица транзакций + форма + фильтр по категории

### День 4 — Импорт CSV (FQ-14, FQ-15, FQ-16)
- `parseCSV()` на stdlib `encoding/csv` + unit-тесты
- POST `/transactions/import` — multipart, дедупликация по `external_id`
- UI: страница импорта — upload → счётчик → таблица

### День 5 — AI-категоризация (FQ-17, FQ-18, FQ-19)
- `CategorizeBatch()` — Claude API, few-shot промпт, unit с моком
- Batch-запрос после импорта → заполнение `category_id` + `ai_confidence`
- `RuleBasedCategorize()` — fallback по словарю, unit-тесты

### День 6 — Аналитика и дашборд (FQ-20, FQ-21, FQ-22)
- GET `/analytics/summary`
- UI: дашборд — 3 карточки + Pie chart
- Sequence Diagram для AI-импорта

### День 7 — Геймификация: логика (FQ-23, FQ-24, FQ-25)
- `CalculateLevel`, `CheckUnlockableAchievements` + unit-тесты
- `AwardXP` + integration-тест: импорт → XP → ачивка
- GET `/gamification/profile`

### День 8 — Геймификация: UI (FQ-26, FQ-27, FQ-28, FQ-29)
- XpBar в шапке (RTL-тест)
- Страница «Ачивки» — сетка карточек (RTL-тест)
- Toast «Получено достижение»
- Activity Diagram для XP-системы

### День 9 — Цели (FQ-30)
- POST/GET `/goals` + страница с прогресс-барами
- (Альтернатива: Challenges FQ-31, P3 — только если успеваем)

### День 10 — Финал (FQ-32…FQ-38)
- Деплой Railway/Yandex Cloud, HTTPS, бэкап БД
- Test Report, Deployment Guide
- 6 UML-диаграмм, 25 пунктов чек-листа
- Тег `v0.1.0` + Release Notes + README
- Презентация: 5-7 слайдов

### Буфер (FQ-39, FQ-40, FQ-41)
- AI-чат SSE (P3, ~8ч) — FQ-39
- Line chart динамики (P2, ~3ч) — FQ-40
- Стабилизация + `v0.2.0` — FQ-41
