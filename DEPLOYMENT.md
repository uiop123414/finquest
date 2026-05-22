# Deployment Guide — FinQuest

## Требования

| Инструмент | Версия | Для чего |
|-----------|--------|----------|
| Docker | 24+ | контейнеры |
| Docker Compose | 2.x (плагин `docker compose`) | оркестрация |
| Make | любая | удобные команды |
| Свободные порты | 3000, 8000, 5432 | веб, API, БД |

---

## Локальный запуск (Docker Compose)

### 1. Клонировать репозиторий

```bash
git clone <repo-url>
cd univ
```

### 2. Создать `.env`

```bash
cp .env.example .env
```

Минимально необходимые переменные:

```env
DATABASE_URL=postgresql://finquest:finquest@db:5432/finquest?sslmode=disable
JWT_SECRET=change-me-in-production
PORT=8000
ANTHROPIC_API_KEY=          # опционально
```

### 3. Запустить

```bash
make start
```

Команда выполняет: `docker compose up --build -d`

Контейнеры запускаются в порядке:
1. `db` — PostgreSQL (healthcheck pg_isready)
2. `migrate` — применяет SQL-миграции (001–003), завершается
3. `backend` — Go API (стартует после migrate)
4. `frontend` — nginx + React (стартует после backend)

Готовность: ~30–60 сек при первом запуске (скачивание образов + компиляция Go).

### 4. Проверить

```bash
curl http://localhost:8000/health
# {"status":"ok"}
```

Открыть http://localhost:3000

### 5. Остановить

```bash
make down          # остановить контейнеры, сохранить данные
docker compose down -v  # + удалить PostgreSQL volume
```

---

## Продакшн-развёртывание на сервере

### Подготовка сервера (Ubuntu 22.04)

```bash
# Docker
curl -fsSL https://get.docker.com | sh
sudo usermod -aG docker $USER
newgrp docker

# Docker Compose plugin
sudo apt-get install docker-compose-plugin

# Make
sudo apt-get install make
```

### Настройка переменных окружения

На сервере создайте `/opt/finquest/.env`:

```env
DATABASE_URL=postgresql://finquest:STRONG_PASSWORD@db:5432/finquest?sslmode=disable
JWT_SECRET=<64-символьная случайная строка>
PORT=8000
ANTHROPIC_API_KEY=sk-ant-...
```

Сгенерировать JWT_SECRET:
```bash
openssl rand -hex 32
```

### Запуск

```bash
cd /opt/finquest
make start
```

### Nginx как reverse proxy (опционально, для домена + TLS)

Установите certbot и настройте `/etc/nginx/sites-available/finquest`:

```nginx
server {
    listen 80;
    server_name your-domain.com;
    return 301 https://$host$request_uri;
}

server {
    listen 443 ssl;
    server_name your-domain.com;

    ssl_certificate     /etc/letsencrypt/live/your-domain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/your-domain.com/privkey.pem;

    # Frontend
    location / {
        proxy_pass http://localhost:3000;
        proxy_set_header Host $host;
    }

    # API (опционально — если хотите API на отдельном поддомене)
    location /api/ {
        proxy_pass http://localhost:8000;
        proxy_set_header Host $host;
        proxy_read_timeout 60s;
    }
}
```

```bash
sudo certbot --nginx -d your-domain.com
sudo systemctl reload nginx
```

---

## Миграции базы данных

Миграции применяются автоматически при `make start` через контейнер `migrate/migrate`.

### Применить вручную

```bash
make migrate
```

### Откатить последнюю миграцию

```bash
make migrate-down
```

### Создать новую миграцию

```bash
# Создайте файлы вручную:
touch backend/db/migrations/004_your_change.up.sql
touch backend/db/migrations/004_your_change.down.sql
```

Формат имени: `NNN_description.up.sql` / `NNN_description.down.sql`

---

## Обновление приложения

```bash
git pull
make start   # пересобирает изменённые образы, применяет новые миграции
```

Данные PostgreSQL сохраняются в Docker volume `pgdata` — `down` без `-v` их не удаляет.

---

## Логи и мониторинг

```bash
make logs                        # все сервисы
docker compose logs backend -f   # только бэкенд
docker compose logs db -f        # только БД
```

Health check API:
```bash
curl http://localhost:8000/health
```

---

## Структура контейнеров

```
docker-compose.yml
├── db        postgres:16-alpine   → порт 5432
├── migrate   migrate/migrate      → однократно, выходит после UP
├── backend   ./backend/Dockerfile → порт 8000
└── frontend  ./frontend/Dockerfile → порт 3000 (nginx)
```

Все данные PostgreSQL хранятся в именованном volume:
```yaml
volumes:
  pgdata:
```

---

## Troubleshooting

| Симптом | Причина | Решение |
|---------|---------|---------|
| `backend` не стартует | `migrate` ещё не завершился | Подождать 10–15 сек, `make logs` |
| 502 Bad Gateway | `backend` упал | `docker compose restart backend` |
| `ERR_CONNECTION_REFUSED :3000` | `frontend` не собрался | `make logs` → ищите ошибки npm/nginx |
| Миграция падает | БД недоступна | Проверить `docker compose ps db`, `make logs db` |
| AI-совет не работает | Нет `ANTHROPIC_API_KEY` | Добавить ключ в `.env`, `make start` |
