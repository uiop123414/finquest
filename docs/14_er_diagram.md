# ER-диаграмма / Схема базы данных

## Entity-Relationship Diagram

```mermaid
erDiagram
    users {
        uuid id PK
        text email UK
        text hashed_password
        int xp_total
        int level
        timestamptz created_at
    }

    categories {
        uuid id PK
        uuid user_id FK
        text name
        bool is_system
    }

    transactions {
        uuid id PK
        uuid user_id FK
        numeric amount
        text type
        uuid category_id FK
        date date
        text note
        text external_id
        numeric ai_confidence
        timestamptz created_at
    }

    goals {
        uuid id PK
        uuid user_id FK
        text name
        numeric target_amount
        numeric current_amount
        date deadline
        timestamptz completed_at
    }

    deposits {
        uuid id PK
        uuid user_id FK
        text bank_name
        numeric amount
        numeric interest_rate
        date start_date
        date end_date
        text note
        timestamptz created_at
    }

    credits {
        uuid id PK
        uuid user_id FK
        text type
        text bank_name
        numeric total_amount
        numeric remaining_balance
        numeric interest_rate
        numeric monthly_payment
        text note
        timestamptz created_at
    }

    achievements {
        uuid id PK
        text code UK
        text name
        text description
    }

    user_achievements {
        uuid id PK
        uuid user_id FK
        uuid achievement_id FK
        timestamptz earned_at
    }

    xp_events {
        uuid id PK
        uuid user_id FK
        int delta
        text reason
        timestamptz created_at
    }

    challenges {
        uuid id PK
        text name
        text description
        int xp_reward
        jsonb condition_json
    }

    user_challenges {
        uuid id PK
        uuid user_id FK
        uuid challenge_id FK
        timestamptz accepted_at
        timestamptz completed_at
    }

    users ||--o{ transactions : "has"
    users ||--o{ goals : "has"
    users ||--o{ deposits : "has"
    users ||--o{ credits : "has"
    users ||--o{ xp_events : "earns"
    users ||--o{ user_achievements : "earns"
    users ||--o{ user_challenges : "accepts"
    users ||--o{ categories : "creates"

    transactions }o--o| categories : "belongs to"
    user_achievements }o--|| achievements : "references"
    user_challenges }o--|| challenges : "references"
```

---

## Описание таблиц

| Таблица | Строк (демо) | Описание |
|---------|:-----------:|----------|
| `users` | 1+ | Учётные записи. `xp_total` и `level` обновляются атомарно через `GamificationService` |
| `categories` | 8 системных + пользовательские | Системные (is_system=true) создаются миграцией: Еда, Транспорт, ЖКХ, Здоровье, Развлечения, Одежда, Зарплата, Прочее |
| `transactions` | 100+ (демо) | Доходы/расходы. `external_id` обеспечивает идемпотентность при CSV-импорте |
| `goals` | 3 (демо) | Финансовые цели. `completed_at IS NULL` = активная цель |
| `deposits` | 3 (демо) | Банковские вклады. Годовой доход = `amount × interest_rate / 100` |
| `credits` | 2 (демо) | Кредиты и карты. `type` ∈ {consumer, card} |
| `achievements` | 6 | Статические записи, созданные миграцией: коды first_transaction, ten_transactions, hundred_transactions, level_5, first_goal, first_import |
| `user_achievements` | — | Связь M:M с уникальностью (user_id, achievement_id) — ачивка выдаётся один раз |
| `xp_events` | — | Лог всех начислений XP: дельта и причина |
| `challenges` | — | Задания (зарезервировано для будущих версий) |
| `user_challenges` | — | Прогресс по заданиям (зарезервировано) |

---

## Ключевые ограничения

| Таблица | Ограничение | Назначение |
|---------|------------|------------|
| `users.email` | UNIQUE | Запрет дублей |
| `transactions.(user_id, external_id)` | UNIQUE | Идемпотентный импорт CSV |
| `user_achievements.(user_id, achievement_id)` | UNIQUE | Каждая ачивка — один раз |
| `transactions.type` | CHECK IN ('income','expense') | Валидация на уровне БД |
| `credits.type` | CHECK IN ('consumer','card') | Валидация типа кредита |
| `deposits.amount` | CHECK > 0 | Сумма должна быть положительной |
| `credits.total_amount` | CHECK > 0 | Лимит/сумма кредита > 0 |
| `categories.user_id → users.id` | ON DELETE CASCADE | Удаление пользователя удаляет его категории |
| `transactions.category_id → categories.id` | ON DELETE SET NULL | Удаление категории не удаляет транзакции |
