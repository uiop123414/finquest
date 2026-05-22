# Product Backlog — FinQuest

**Приоритеты (MoSCoW):**
- 🔴 **Must Have** — критически необходимо для MVP
- 🟡 **Should Have** — важно, но не блокирует запуск
- 🟢 **Could Have** — желательно при наличии времени
- ⚪ **Won't Have** — вне текущего скоупа

---

## Backlog

| # | Задача | Приоритет | Оценка (SP) | Статус | Sprint |
|---|--------|:---------:|:-----------:|--------|--------|
| PB-01 | Регистрация пользователя (email + bcrypt пароль) | 🔴 Must | 3 | ✅ Done | S1 |
| PB-02 | Вход и выдача JWT access + refresh токенов | 🔴 Must | 3 | ✅ Done | S1 |
| PB-03 | Middleware авторизации для защищённых маршрутов | 🔴 Must | 2 | ✅ Done | S1 |
| PB-04 | Docker Compose: db + migrate + backend + frontend | 🔴 Must | 3 | ✅ Done | S1 |
| PB-05 | CRUD транзакций с пагинацией (20/стр) и фильтрами | 🔴 Must | 5 | ✅ Done | S2 |
| PB-06 | Аналитика: summary (доходы/расходы/баланс/категории) | 🔴 Must | 3 | ✅ Done | S2 |
| PB-07 | Дашборд: KPI-карточки + оценка финансового поведения | 🔴 Must | 3 | ✅ Done | S2 |
| PB-08 | Дашборд: линейный график доходы/расходы (period-aware) | 🔴 Must | 3 | ✅ Done | S2 |
| PB-09 | Дашборд: круговая диаграмма расходов по категориям | 🟡 Should | 2 | ✅ Done | S2 |
| PB-10 | Импорт CSV с автокатегоризацией по ключевым словам | 🟡 Should | 4 | ✅ Done | S2 |
| PB-11 | Системные категории (8 шт. по умолчанию) | 🔴 Must | 1 | ✅ Done | S1 |
| PB-12 | Финансовые цели: создание, прогресс-бар, дедлайн | 🔴 Must | 3 | ✅ Done | S3 |
| PB-13 | Цели: пополнение накоплений, редактирование, закрытие | 🟡 Should | 3 | ✅ Done | S3 |
| PB-14 | Цели: пагинация (5 на страницу, активные / выполненные) | 🟢 Could | 2 | ✅ Done | S3 |
| PB-15 | Геймификация: начисление XP за транзакции и цели | 🟡 Should | 3 | ✅ Done | S3 |
| PB-16 | Геймификация: уровни (каждые 100 XP) | 🟡 Should | 2 | ✅ Done | S3 |
| PB-17 | Ачивки: 6 достижений с автоматическим выполнением | 🟡 Should | 3 | ✅ Done | S3 |
| PB-18 | XP-бар в шапке + тост при получении ачивки | 🟢 Could | 2 | ✅ Done | S3 |
| PB-19 | AI-советник: Claude Haiku / Gemini / rule-based fallback | 🟡 Should | 5 | ✅ Done | S4 |
| PB-20 | AI-советник: полный финансовый контекст (tx + dep + cr + goals) | 🟡 Should | 3 | ✅ Done | S4 |
| PB-21 | Страница инвестиций: банковские депозиты (CRUD) | 🟡 Should | 4 | ✅ Done | S4 |
| PB-22 | Страница кредитов: потребительские + карты (CRUD) | 🟡 Should | 4 | ✅ Done | S4 |
| PB-23 | Демо-аккаунт: 101 транзакция + 3 цели | 🟡 Should | 2 | ✅ Done | S4 |
| PB-24 | Демо-аккаунт: 3 депозита + 2 кредита | 🟡 Should | 1 | ✅ Done | S4 |
| PB-25 | Период-aware дашборд: фильтры all/1y/6m/YYYY-MM | 🟡 Should | 3 | ✅ Done | S2 |
| PB-26 | Статический анализ: golangci-lint + ESLint | 🟢 Could | 2 | ✅ Done | S5 |
| PB-27 | Интеграционные тесты (6 функций) | 🟡 Should | 4 | ✅ Done | S5 |
| PB-28 | Unit-тесты Go: геймификация, категоризация | 🟡 Should | 2 | ✅ Done | S5 |
| PB-29 | Unit-тесты Frontend: XpBar, AchievementsPage | 🟢 Could | 2 | ✅ Done | S5 |
| PB-30 | Документация: README + DEPLOYMENT + TEST_REPORT | 🔴 Must | 3 | ✅ Done | S5 |
| PB-31 | Документация: UML-диаграммы, RACI, реестр рисков, backlog | 🔴 Must | 4 | ✅ Done | S5 |
| PB-32 | SSE-чат с AI (/ai/chat streaming endpoint) | 🟢 Could | 3 | ✅ Done | S4 |
| PB-33 | Автообновление JWT (axios interceptor) | 🔴 Must | 2 | ✅ Done | S1 |
| PB-34 | nginx reverse proxy с Docker DNS resolver | 🔴 Must | 1 | ✅ Done | S1 |
| PB-35 | Поддержка Gemini API как бесплатной альтернативы | 🟢 Could | 2 | ✅ Done | S5 |

---

## Статистика

| Приоритет | Кол-во задач | Story Points |
|-----------|:-----------:|:------------:|
| 🔴 Must Have | 12 | 27 |
| 🟡 Should Have | 16 | 43 |
| 🟢 Could Have | 7 | 18 |
| **Итого** | **35** | **88** |

**Все задачи выполнены.** ✅
