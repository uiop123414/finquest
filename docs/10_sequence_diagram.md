# Диаграмма последовательности (Sequence Diagram)

## Сценарий 1 — Добавление транзакции с начислением XP

```mermaid
sequenceDiagram
    actor User as Пользователь
    participant FE as Frontend (React)
    participant MW as JWT Middleware
    participant H as Handler
    participant GS as GamificationService
    participant DB as PostgreSQL

    User->>FE: Заполняет форму транзакции
    FE->>MW: POST /api/v1/transactions\n{Authorization: Bearer <token>}
    MW->>DB: SELECT users WHERE id = <userID from JWT>
    DB-->>MW: user found
    MW->>H: Передаёт запрос + userID в контекст

    H->>DB: INSERT INTO transactions (...)
    DB-->>H: transaction.id

    H->>GS: AwardXP(userID, +10, "transaction")
    GS->>DB: BEGIN TRANSACTION
    GS->>DB: INSERT INTO xp_events (...)
    GS->>DB: UPDATE users SET xp_total=xp_total+10, level=...
    GS->>DB: SELECT COUNT(*) FROM transactions WHERE user_id=...
    DB-->>GS: count
    GS->>DB: INSERT INTO user_achievements ON CONFLICT DO NOTHING
    GS->>DB: COMMIT
    DB-->>GS: ok
    GS-->>H: nil (success)

    H-->>FE: 201 Created {transaction}
    FE->>FE: Обновляет список транзакций
    FE->>FE: Обновляет XP-бар
    FE-->>User: Транзакция добавлена, +10 XP
```

---

## Сценарий 2 — Получение AI-совета

```mermaid
sequenceDiagram
    actor User as Пользователь
    participant FE as Frontend
    participant H as Handler (ai.go)
    participant DB as PostgreSQL
    participant AI as LLM API\n(Gemini/Claude)

    User->>FE: Нажимает «Получить совет»
    FE->>H: GET /api/v1/ai/advice
    
    H->>DB: SELECT транзакции за 30 дней по категориям
    DB-->>H: txRows[]
    H->>DB: SELECT * FROM deposits WHERE user_id=...
    DB-->>H: deps[]
    H->>DB: SELECT * FROM credits WHERE user_id=...
    DB-->>H: crds[]
    H->>DB: SELECT * FROM goals WHERE completed_at IS NULL LIMIT 5
    DB-->>H: goals[]

    H->>H: buildPrompt(context)

    alt ANTHROPIC_API_KEY задан
        H->>AI: POST api.anthropic.com/v1/messages\n{model: claude-haiku, prompt}
        AI-->>H: {content: [{text: "..."}]}
    else GEMINI_API_KEY задан
        H->>AI: POST generativelanguage.googleapis.com\n/v1beta/models/gemini-2.0-flash:generateContent
        AI-->>H: {candidates: [{content: {parts: [{text}]}}]}
    else Нет ключей
        H->>H: buildRuleBasedAdvice(income, expense, debt, ...)
    end

    H-->>FE: 200 OK {advice: "...", context: "..."}
    FE-->>User: Отображает совет
```

---

## Сценарий 3 — Автообновление JWT токена

```mermaid
sequenceDiagram
    participant FE as Frontend (Axios)
    participant BE as Backend

    FE->>BE: GET /api/v1/transactions\n{Authorization: Bearer <expired_token>}
    BE-->>FE: 401 Unauthorized

    Note over FE: Axios interceptor перехватывает 401
    FE->>BE: POST /api/v1/auth/refresh\n{refresh_token: "..."}
    
    alt refresh_token валиден
        BE-->>FE: 200 OK {access_token: "...", refresh_token: "..."}
        FE->>FE: Сохранить новые токены в localStorage
        FE->>BE: GET /api/v1/transactions\n{Authorization: Bearer <new_token>}
        BE-->>FE: 200 OK {data: [...]}
    else refresh_token истёк
        BE-->>FE: 401 Unauthorized
        FE->>FE: localStorage.clear()
        FE->>FE: Редирект на /login
    end
```
