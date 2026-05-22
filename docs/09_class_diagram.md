# Диаграмма классов (Class Diagram)

```mermaid
classDiagram
    class User {
        +UUID ID
        +string Email
        +string HashedPassword
        +int XPTotal
        +int Level
        +time.Time CreatedAt
    }

    class Category {
        +UUID ID
        +*UUID UserID
        +string Name
        +bool IsSystem
    }

    class Transaction {
        +UUID ID
        +UUID UserID
        +float64 Amount
        +string Type
        +*UUID CategoryID
        +time.Time Date
        +string Note
        +string ExternalID
        +float64 AIConfidence
        +time.Time CreatedAt
    }

    class Goal {
        +UUID ID
        +UUID UserID
        +string Name
        +float64 TargetAmount
        +float64 CurrentAmount
        +time.Time Deadline
        +*time.Time CompletedAt
    }

    class Deposit {
        +UUID ID
        +UUID UserID
        +string BankName
        +float64 Amount
        +float64 InterestRate
        +time.Time StartDate
        +time.Time EndDate
        +string Note
        +time.Time CreatedAt
        +YearlyIncome() float64
    }

    class Credit {
        +UUID ID
        +UUID UserID
        +string Type
        +string BankName
        +float64 TotalAmount
        +float64 RemainingBalance
        +float64 InterestRate
        +float64 MonthlyPayment
        +string Note
        +time.Time CreatedAt
    }

    class Achievement {
        +UUID ID
        +string Code
        +string Name
        +string Description
        +*time.Time EarnedAt
    }

    class XPEvent {
        +UUID ID
        +UUID UserID
        +int Delta
        +string Reason
        +time.Time CreatedAt
    }

    class GamificationService {
        +AwardXP(userID, delta, reason) error
        +GetProfile(userID) GamificationProfile
        -checkAchievements(tx, userID) error
    }

    class Handler {
        +DB *sqlx.DB
        +Cfg *Config
        +Register(c)
        +Login(c)
        +GetTransactions(c)
        +CreateTransaction(c)
        +GetAIAdvice(c)
        +GetDeposits(c)
        +GetCredits(c)
        +GetGoals(c)
    }

    class Config {
        +string DatabaseURL
        +string JWTSecret
        +string Port
        +string AnthropicKey
        +string GeminiKey
    }

    User "1" --> "0..*" Transaction : has
    User "1" --> "0..*" Goal : has
    User "1" --> "0..*" Deposit : has
    User "1" --> "0..*" Credit : has
    User "1" --> "0..*" XPEvent : earns
    User "0..*" --> "0..*" Achievement : earns

    Transaction "0..*" --> "0..1" Category : belongs to

    Handler --> GamificationService : uses
    Handler --> Config : uses
    GamificationService --> XPEvent : creates
    GamificationService --> Achievement : unlocks
```

## Пакеты (Go packages)

| Пакет | Классы / Структуры | Ответственность |
|-------|-------------------|----------------|
| `models` | User, Transaction, Category, Goal, Deposit, Credit, Achievement | Структуры данных, соответствующие таблицам БД |
| `handlers` | Handler | HTTP-хендлеры, разбор запросов, формирование ответов |
| `services` | GamificationService | Бизнес-логика XP, уровней, ачивок |
| `config` | Config | Загрузка переменных окружения |
| `middleware` | AuthMiddleware | Проверка JWT, установка userID в контекст |
| `db` | — | Подключение к PostgreSQL (sqlx) |
