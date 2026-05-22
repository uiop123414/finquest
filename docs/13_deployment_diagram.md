# Диаграмма развёртывания (Deployment Diagram)

## Локальное развёртывание (Docker Compose)

```mermaid
graph TB
    subgraph Host["💻 Хост-машина (macOS / Linux / Windows)"]
        subgraph DockerEngine["Docker Engine"]
            subgraph Network["Bridge Network: finquest_default"]

                subgraph DB["Контейнер: db\npostgres:16-alpine"]
                    PG["PostgreSQL 16\nПорт: 5432 (внутренний)"]
                    Vol[("Volume: pgdata\n/var/lib/postgresql/data")]
                end

                subgraph Migrate["Контейнер: migrate\nmigrate/migrate (one-shot)"]
                    MigBin["golang-migrate\n→ применяет 001–005.sql\n→ завершается (exit 0)"]
                    MigFiles["Mount:\n./backend/db/migrations\n→ /migrations"]
                end

                subgraph Backend["Контейнер: backend\ngolang:1.22-alpine"]
                    GoApp["Go HTTP Server\nГин :8000"]
                    EnvB[".env → переменные окружения\nDATABASE_URL, JWT_SECRET\nGEMINI_API_KEY"]
                end

                subgraph Frontend["Контейнер: frontend\nnginx:alpine"]
                    NginxSrv["nginx\nПорт: 80 (внутренний)"]
                    ReactDist["React dist/\n(статические файлы)"]
                    NginxConf["nginx.conf:\nresolver 127.0.0.11\nproxy_pass backend:8000"]
                end
            end
        end

        PortMap["Проброс портов:\n3000 → frontend:80\n8000 → backend:8000\n5432 → db:5432"]
    end

    subgraph UserBrowser["🌐 Браузер пользователя"]
        Browser["http://localhost:3000"]
    end

    subgraph ExternalAPIs["☁️ Внешние API (опционально)"]
        GeminiAPI["Google Gemini API\ngenerativelanguage.googleapis.com"]
        AnthropicAPI["Anthropic API\napi.anthropic.com"]
    end

    Browser -->|":3000"| PortMap
    PortMap --> NginxSrv
    NginxSrv --> ReactDist
    NginxSrv -->|"proxy /api/"| GoApp
    GoApp -->|"sqlx :5432"| PG
    GoApp -->|"HTTPS"| GeminiAPI
    GoApp -->|"HTTPS"| AnthropicAPI
    MigBin -->|"SQL migrations"| PG
    PG --- Vol

    Migrate -->|"depends_on: db healthy"| DB
    Backend -->|"depends_on: migrate"| Migrate
    Frontend -->|"depends_on: backend"| Backend
```

---

## Продакшн-развёртывание (VPS + внешний nginx)

```mermaid
graph TB
    subgraph Internet["🌍 Интернет"]
        Client["Клиент (браузер)"]
    end

    subgraph VPS["🖥️ VPS Ubuntu 22.04"]
        subgraph NginxHost["Системный nginx (reverse proxy)"]
            NginxLB["nginx + Let's Encrypt TLS\nyour-domain.com:443"]
        end

        subgraph DockerCompose["Docker Compose (порты закрыты снаружи)"]
            FE["frontend\nnginx:alpine :3000"]
            BE["backend\n:8000"]
            DB["PostgreSQL\n:5432 (только внутри)"]
        end

        EnvFile[".env\n/opt/finquest/.env"]
    end

    Client -->|"HTTPS :443"| NginxLB
    NginxLB -->|"proxy_pass :3000"| FE
    NginxLB -->|"proxy_pass /api/ :8000"| BE
    FE -->|"proxy /api/"| BE
    BE --> DB
    EnvFile -.-> BE
```

---

## Порядок запуска контейнеров

```mermaid
sequenceDiagram
    participant DC as Docker Compose
    participant DB as db (postgres)
    participant MG as migrate
    participant BE as backend
    participant FE as frontend

    DC->>DB: docker run postgres:16-alpine
    Note over DB: healthcheck: pg_isready\nждёт готовности БД

    DB-->>DC: healthy ✅

    DC->>MG: docker run migrate/migrate up
    MG->>DB: Применяет миграции 001–005
    DB-->>MG: ok
    MG-->>DC: exit 0 ✅

    DC->>BE: docker run backend
    Note over BE: go build + запуск :8000
    BE-->>DC: HTTP :8000 доступен ✅

    DC->>FE: docker run frontend (nginx)
    Note over FE: nginx раздаёт React dist/
    FE-->>DC: HTTP :3000 доступен ✅
```

## Требования к окружению

| Компонент | Минимальные требования | Рекомендуемые |
|-----------|----------------------|---------------|
| Docker Engine | 24+ | 26+ |
| Docker Compose | 2.x (plugin) | 2.x |
| RAM | 512 MB | 1 GB |
| CPU | 1 vCPU | 2 vCPU |
| Диск | 2 GB | 5 GB |
| ОС | Linux / macOS / Windows (WSL2) | Ubuntu 22.04 |
| Открытые порты | 3000, 8000 | 3000, 8000 |
