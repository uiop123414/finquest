# FinQuest 🏆

Геймифицированный AI-ассистент для личных финансов. Отслеживайте расходы, ставьте цели, получайте советы от нейросети и зарабатывайте опыт за финансовую дисциплину.

## Документация

### Эксплуатация

| # | Документ | Описание |
|---|----------|----------|
| — | [docs/DEPLOYMENT.md](./docs/DEPLOYMENT.md) | Инструкция по развёртыванию: локально, на сервере, nginx + TLS, troubleshooting |
| — | [docs/TEST_REPORT.md](./docs/TEST_REPORT.md) | Отчёт о тестировании: 11 автотестов + 5 ручных сценариев (35 шагов) |

### Оценка качества

| # | Документ | Описание |
|---|----------|----------|
| 15 | [docs/15_kpi_evaluation.md](./docs/15_kpi_evaluation.md) | Оценка проекта по 79 KPI: управление, качество кода, DevOps, академические критерии |

### Управление проектом

| # | Документ | Описание |
|---|----------|----------|
| 03 | [docs/03_project_plan.md](./docs/03_project_plan.md) | План проекта: вехи, спринты, трудоёмкость, ответственные |
| 04 | [docs/04_raci.md](./docs/04_raci.md) | Матрица ответственности RACI |
| 05 | [docs/05_risk_register.md](./docs/05_risk_register.md) | Реестр рисков (10 рисков, тепловая карта) |
| 06 | [docs/06_user_stories.md](./docs/06_user_stories.md) | Use Cases / User Stories (6 историй, таблица UC) |
| 07 | [docs/07_product_backlog.md](./docs/07_product_backlog.md) | Product Backlog (35 задач, MoSCoW приоритеты) |

### UML-диаграммы

| # | Документ | Описание |
|---|----------|----------|
| 08 | [docs/08_use_case_diagram.md](./docs/08_use_case_diagram.md) | Диаграмма вариантов использования |
| 09 | [docs/09_class_diagram.md](./docs/09_class_diagram.md) | Диаграмма классов |
| 10 | [docs/10_sequence_diagram.md](./docs/10_sequence_diagram.md) | Диаграмма последовательности (3 сценария) |
| 11 | [docs/11_activity_diagram.md](./docs/11_activity_diagram.md) | Диаграмма деятельности |
| 12 | [docs/12_component_diagram.md](./docs/12_component_diagram.md) | Диаграмма компонентов |
| 13 | [docs/13_deployment_diagram.md](./docs/13_deployment_diagram.md) | Диаграмма развёртывания (локальное + продакшн) |
| 14 | [docs/14_er_diagram.md](./docs/14_er_diagram.md) | ER-диаграмма / схема базы данных (11 таблиц) |

## Стек

| Слой | Технологии |
|------|-----------|
| Backend | Go 1.22 · Gin · sqlx · pgx/v5 |
| Frontend | React 18 · TypeScript · Vite · Tailwind CSS · Recharts |
| База данных | PostgreSQL 16 |
| Инфраструктура | Docker Compose · nginx · golang-migrate |
| AI | Claude Haiku (Anthropic) / Gemini 2.0 Flash (Google, бесплатно) — советы и автокатегоризация |

## Быстрый старт (1 шаг)

```bash
make start
```

Команда скопирует `.env`, соберёт и запустит все контейнеры.

| Адрес | Сервис |
|-------|--------|
| http://localhost:3000 | Веб-интерфейс |
| http://localhost:8000 | REST API |
| http://localhost:8000/health | Health check |

**Демо-аккаунт** (101 транзакция за 6 месяцев уже загружена):
```
Email:  demo@finquest.ru
Пароль: demo123
```

> Требования: Docker + Docker Compose. Порты 3000, 8000, 5432 должны быть свободны.

## Функциональность

### Транзакции
- Добавление доходов и расходов с категорией, датой и заметкой
- Фильтрация по категории, пагинация (20 записей на страницу)
- Удаление транзакций
- Импорт из CSV с AI-автокатегоризацией (формат: `date,amount,type,note`)

### Аналитика / Дашборд
- Фильтр по месяцу (последние 12 месяцев или всё время)
- KPI-карточки: доходы, расходы, баланс, норма сбережений
- Итоговая оценка финансового поведения
- Круговая диаграмма расходов по категориям
- Линейный график доходы/расходы за 12 месяцев
- **Мнение нейросети** — персональный совет по расходам (кнопка «Получить совет»)

### Цели
- Создание финансовых целей с прогресс-баром
- Пополнение накоплений по цели
- Редактирование, закрытие и возобновление целей
- Индикатор просроченных дедлайнов

### Геймификация
- **+10 XP** за каждую добавленную транзакцию
- **+5 XP** за каждую транзакцию при импорте CSV
- **+20 XP** за создание первой цели
- Автоматический расчёт уровня (каждые 100 XP = новый уровень)
- Ачивки: «Первый шаг», «Десятка», «Сотня», «Мечтатель», «Опытный», «Импортёр»

### Авторизация
- Регистрация / вход по email + пароль (bcrypt, cost 12)
- JWT access token (30 мин) + refresh token (7 дней)
- Автообновление токена при 401

## Структура проекта

```
univ/
├── backend/
│   ├── config/           # загрузка конфигурации из env
│   ├── db/
│   │   ├── db.go         # подключение к PostgreSQL
│   │   └── migrations/   # SQL-миграции (001–005)
│   ├── handlers/         # HTTP-хендлеры (Gin)
│   ├── middleware/        # JWT auth middleware
│   ├── models/           # структуры БД
│   ├── services/         # бизнес-логика (gamification, AI)
│   ├── integration_test.go
│   └── main.go
├── frontend/
│   ├── src/
│   │   ├── api/          # Axios клиент с JWT interceptors
│   │   ├── components/   # Layout, XpBar, AchievementToast
│   │   └── pages/        # Dashboard, Transactions, Goals, Investments, Credits, ...
│   └── nginx.conf
├── docs/
│   ├── 03_project_plan.md
│   ├── 04_raci.md
│   ├── 05_risk_register.md
│   ├── 06_user_stories.md
│   ├── 07_product_backlog.md
│   ├── 08_use_case_diagram.md
│   ├── 09_class_diagram.md
│   ├── 10_sequence_diagram.md
│   ├── 11_activity_diagram.md
│   ├── 12_component_diagram.md
│   ├── 13_deployment_diagram.md
│   ├── 14_er_diagram.md
│   ├── DEPLOYMENT.md
│   └── TEST_REPORT.md
├── docker-compose.yml
├── Makefile
└── .env.example
```

## Переменные окружения

Скопируйте `.env.example` → `.env` и заполните:

```env
DATABASE_URL=postgresql://finquest:finquest@db:5432/finquest?sslmode=disable
JWT_SECRET=your-secret-here          # любая случайная строка
PORT=8000
ANTHROPIC_API_KEY=                   # опционально — для AI советов и автокатегоризации
```

Без `ANTHROPIC_API_KEY` приложение работает полностью, но:
- AI-совет использует встроенную rule-based логику
- Импорт CSV категоризируется по ключевым словам

## Команды Makefile

```bash
make start          # собрать и запустить все контейнеры
make down           # остановить контейнеры
make logs           # логи всех сервисов
make test           # запустить unit-тесты
make lint           # golangci-lint + ESLint
make migrate        # применить миграции (локально)
make migrate-down   # откатить последнюю миграцию
```

### Интеграционные тесты

Требуют запущенной PostgreSQL:

```bash
TEST_DATABASE_URL="postgresql://finquest:finquest@localhost:5432/finquest?sslmode=disable" \
  go test -tags integration -v ./...
```

## API (кратко)

| Метод | Путь | Описание |
|-------|------|----------|
| POST | `/api/v1/auth/register` | Регистрация |
| POST | `/api/v1/auth/login` | Вход |
| POST | `/api/v1/auth/refresh` | Обновление токена |
| GET | `/api/v1/transactions` | Список (фильтры: category_id, date_from, date_to, limit, offset) |
| POST | `/api/v1/transactions` | Создать транзакцию |
| PATCH | `/api/v1/transactions/:id` | Обновить |
| DELETE | `/api/v1/transactions/:id` | Удалить |
| POST | `/api/v1/transactions/import` | Импорт CSV |
| GET | `/api/v1/analytics/summary` | Сводка (опционально: ?period=YYYY-MM) |
| GET | `/api/v1/analytics/over-time` | Динамика по месяцам |
| GET | `/api/v1/goals` | Список целей |
| POST | `/api/v1/goals` | Создать цель |
| PATCH | `/api/v1/goals/:id` | Обновить / закрыть цель |
| DELETE | `/api/v1/goals/:id` | Удалить цель |
| GET | `/api/v1/gamification/profile` | XP, уровень, ачивки |
| GET | `/api/v1/ai/advice` | AI-совет по расходам |
| POST | `/api/v1/ai/chat` | Чат с AI (SSE) |
