# Диаграмма компонентов (Component Diagram)

```mermaid
graph TB
    subgraph Browser["🌐 Браузер (пользователь)"]
        subgraph React["Frontend — React 18 + TypeScript"]
            Router["React Router v6\n(маршрутизация)"]
            Pages["Pages\nDashboard · Transactions\nGoals · Investments\nCredits · Achievements\nImport · Login"]
            Components["Components\nLayout · XpBar\nAchievementToast"]
            AxiosClient["Axios Client\n(JWT interceptors)"]
            Recharts["Recharts\n(LineChart, PieChart)"]
        end
    end

    subgraph Docker["🐳 Docker Compose"]
        subgraph FrontendContainer["frontend (nginx:alpine)"]
            Nginx["nginx\n:3000"]
            StaticFiles["dist/ (React build)"]
        end

        subgraph BackendContainer["backend (golang:1.22)"]
            GinRouter["Gin Router\n:8000"]

            subgraph Handlers["handlers/"]
                AuthH["auth.go\nregister · login · refresh"]
                TxH["transactions.go\nCRUD · import · paginate"]
                AnalyticsH["analytics.go\nsummary · over-time"]
                GoalsH["goals.go\nCRUD · deposit · close"]
                DepositsH["deposits.go\nCRUD"]
                CreditsH["credits.go\nCRUD"]
                GamH["gamification.go\nprofile"]
                AIH["ai.go\nadvice · chat"]
            end

            subgraph Repos["repository/"]
                TxRepo["TransactionRepo"]
                GoalRepo["GoalRepo"]
                DepRepo["DepositRepo"]
                CrRepo["CreditRepo"]
                UserRepo["UserRepo / AchievementRepo / XPEventRepo"]
            end

            subgraph Services["services/"]
                GamSvc["gamification.go\nAwardXP · checkAchievements"]
                RuleSvc["rule_based.go\ncategorize · advice"]
            end

            JWTMw["middleware/\nJWT Auth + CORS"]
            Config["config/\n.env loader"]
        end

        subgraph MigrateContainer["migrate (one-shot)"]
            Migrate["golang-migrate\n001–005 *.sql"]
        end

        subgraph DBContainer["db (postgres:16-alpine)"]
            Postgres["PostgreSQL 16\n:5432"]
            PGData[("pgdata volume")]
        end
    end

    subgraph External["☁️ Внешние сервисы (опционально)"]
        Gemini["Google Gemini API\ngenerative language"]
        Claude["Anthropic API\nClaude Haiku"]
    end

    %% Browser → nginx
    Browser -->|"HTTP :3000"| Nginx
    Nginx --> StaticFiles
    AxiosClient -->|"HTTP /api/v1/*\n:8000"| GinRouter

    %% Gin routing
    GinRouter --> JWTMw
    JWTMw --> Handlers

    %% Handlers → Services
    TxH --> GamSvc
    GoalsH --> GamSvc
    TxH --> RuleSvc
    AIH --> RuleSvc

    %% Handlers → Repos → DB
    Handlers --> Repos
    Repos -->|"sqlx"| Postgres
    Services -->|"sqlx"| Postgres

    %% AI
    AIH -->|"HTTPS"| Gemini
    AIH -->|"HTTPS"| Claude

    %% DB → volume
    Postgres --- PGData

    %% Migrate → DB
    Migrate -->|"sql"| Postgres

    %% React internals
    Pages --> AxiosClient
    Pages --> Recharts
    Pages --> Components
    Router --> Pages
```

## Описание компонентов

| Компонент | Технология | Ответственность |
|-----------|-----------|----------------|
| **nginx** | nginx:alpine | Раздача статических файлов React, reverse proxy к API |
| **Gin Router** | gin-gonic/gin | HTTP маршрутизация, группировка по префиксу `/api/v1` |
| **JWT Middleware** | golang-jwt/jwt/v5 | Проверка Bearer токена, установка `userID` в контекст |
| **CORS Middleware** | gin-contrib/cors | Разрешает запросы с localhost:5173 / localhost:3000 |
| **repository/** | squirrel + sqlx | Типизированный доступ к БД через интерфейсы репозиториев |
| **handlers/** | Go | Разбор запросов, вызов репозиториев, формирование JSON-ответов |
| **GamificationService** | Go | XP, уровни, проверка и выдача ачивок (атомарно в транзакции БД) |
| **RuleBasedService** | Go | Категоризация CSV по ключевым словам, fallback AI-советы |
| **Axios Client** | axios + interceptors | JWT в заголовках, автообновление токена при 401 |
| **React Router** | react-router-dom v6 | SPA-маршрутизация, защита маршрутов через `RequireAuth` |
| **Recharts** | recharts | LineChart (динамика) и PieChart (категории расходов) |
| **golang-migrate** | migrate/migrate | Применение SQL-миграций при старте |
| **PostgreSQL** | postgres:16-alpine | Хранение всех данных, именованный volume `pgdata` |
