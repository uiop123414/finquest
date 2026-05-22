package repository

import (
	"context"
	"finquest/models"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// psql — построитель запросов с плейсхолдерами PostgreSQL ($1, $2, ...).
var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

// ── Transaction ───────────────────────────────────────────────────────────────

type transactionRepo struct{ db *sqlx.DB }

func NewTransactionRepo(db *sqlx.DB) TransactionRepo { return &transactionRepo{db: db} }

func (r *transactionRepo) List(ctx context.Context, userID uuid.UUID, f TransactionFilter) ([]models.Transaction, error) {
	q := psql.Select("*").From("transactions").Where(sq.Eq{"user_id": userID})
	if f.DateFrom != "" {
		q = q.Where(sq.GtOrEq{"date": f.DateFrom})
	}
	if f.DateTo != "" {
		q = q.Where(sq.LtOrEq{"date": f.DateTo})
	}
	if f.CategoryID != "" {
		q = q.Where(sq.Eq{"category_id": f.CategoryID})
	}
	limit := 50
	if f.Limit > 0 {
		limit = f.Limit
	}
	q = q.OrderBy("date DESC").Limit(uint64(limit)).Offset(uint64(f.Offset))

	query, args, err := q.ToSql()
	if err != nil {
		return nil, err
	}
	txs := make([]models.Transaction, 0)
	return txs, r.db.SelectContext(ctx, &txs, query, args...)
}

func (r *transactionRepo) Create(ctx context.Context, userID uuid.UUID, amount float64, txType string, categoryID *uuid.UUID, date, note string) (models.Transaction, error) {
	query, args, err := psql.Insert("transactions").
		Columns("user_id", "amount", "type", "category_id", "date", "note").
		Values(userID, amount, txType, categoryID, date, note).
		Suffix("RETURNING *").ToSql()
	if err != nil {
		return models.Transaction{}, err
	}
	var tx models.Transaction
	return tx, r.db.QueryRowxContext(ctx, query, args...).StructScan(&tx)
}

func (r *transactionRepo) Update(ctx context.Context, id, userID string, amount *float64, txType, categoryID, date, note *string) (models.Transaction, error) {
	q := psql.Update("transactions").Where(sq.Eq{"id": id, "user_id": userID})
	if amount != nil {
		q = q.Set("amount", *amount)
	}
	if txType != nil {
		q = q.Set("type", *txType)
	}
	if categoryID != nil {
		q = q.Set("category_id", *categoryID)
	}
	if date != nil {
		q = q.Set("date", sq.Expr("?::date", *date))
	}
	if note != nil {
		q = q.Set("note", *note)
	}
	query, args, err := q.Suffix("RETURNING *").ToSql()
	if err != nil {
		return models.Transaction{}, err
	}
	var tx models.Transaction
	return tx, r.db.QueryRowxContext(ctx, query, args...).StructScan(&tx)
}

func (r *transactionRepo) Delete(ctx context.Context, id, userID string) (int64, error) {
	query, args, err := psql.Delete("transactions").Where(sq.Eq{"id": id, "user_id": userID}).ToSql()
	if err != nil {
		return 0, err
	}
	res, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (r *transactionRepo) ImportOne(ctx context.Context, userID uuid.UUID, amount float64, txType string, categoryID *uuid.UUID, date interface{}, note, externalID string, aiConfidence *float64) (bool, error) {
	query, args, err := psql.Insert("transactions").
		Columns("user_id", "amount", "type", "category_id", "date", "note", "external_id", "ai_confidence").
		Values(userID, amount, txType, categoryID, date, note, externalID, aiConfidence).
		Suffix("ON CONFLICT (user_id, external_id) DO NOTHING").ToSql()
	if err != nil {
		return false, err
	}
	res, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return false, err
	}
	n, _ := res.RowsAffected()
	return n > 0, nil
}

func (r *transactionRepo) CountByUser(ctx context.Context, userID uuid.UUID) (int, error) {
	query, args, err := psql.Select("COUNT(*)").From("transactions").Where(sq.Eq{"user_id": userID}).ToSql()
	if err != nil {
		return 0, err
	}
	var count int
	return count, r.db.QueryRowContext(ctx, query, args...).Scan(&count)
}

// ── Goal ──────────────────────────────────────────────────────────────────────

type goalRepo struct{ db *sqlx.DB }

func NewGoalRepo(db *sqlx.DB) GoalRepo { return &goalRepo{db: db} }

func (r *goalRepo) List(ctx context.Context, userID string) ([]models.Goal, error) {
	query, args, err := psql.Select("*").From("goals").
		Where(sq.Eq{"user_id": userID}).
		OrderBy("completed_at NULLS FIRST", "deadline").ToSql()
	if err != nil {
		return nil, err
	}
	goals := make([]models.Goal, 0)
	return goals, r.db.SelectContext(ctx, &goals, query, args...)
}

func (r *goalRepo) Create(ctx context.Context, userID uuid.UUID, name string, target, current float64, deadline string) (models.Goal, error) {
	query, args, err := psql.Insert("goals").
		Columns("user_id", "name", "target_amount", "current_amount", "deadline").
		Values(userID, name, target, current, deadline).
		Suffix("RETURNING *").ToSql()
	if err != nil {
		return models.Goal{}, err
	}
	var goal models.Goal
	return goal, r.db.QueryRowxContext(ctx, query, args...).StructScan(&goal)
}

func (r *goalRepo) Update(ctx context.Context, id, userID string, name *string, target, current *float64, deadline *string, completedAt interface{}) (models.Goal, error) {
	q := psql.Update("goals").Where(sq.Eq{"id": id, "user_id": userID})
	if name != nil {
		q = q.Set("name", *name)
	}
	if target != nil {
		q = q.Set("target_amount", *target)
	}
	if current != nil {
		q = q.Set("current_amount", *current)
	}
	if deadline != nil {
		q = q.Set("deadline", sq.Expr("?::date", *deadline))
	}
	if completedAt != nil {
		q = q.Set("completed_at", completedAt)
	}
	query, args, err := q.Suffix("RETURNING *").ToSql()
	if err != nil {
		return models.Goal{}, err
	}
	var goal models.Goal
	return goal, r.db.QueryRowxContext(ctx, query, args...).StructScan(&goal)
}

func (r *goalRepo) Delete(ctx context.Context, id, userID string) (int64, error) {
	query, args, err := psql.Delete("goals").Where(sq.Eq{"id": id, "user_id": userID}).ToSql()
	if err != nil {
		return 0, err
	}
	res, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// ── Deposit ───────────────────────────────────────────────────────────────────

type depositRepo struct{ db *sqlx.DB }

func NewDepositRepo(db *sqlx.DB) DepositRepo { return &depositRepo{db: db} }

func (r *depositRepo) List(ctx context.Context, userID string) ([]models.Deposit, error) {
	query, args, err := psql.Select("*").From("deposits").
		Where(sq.Eq{"user_id": userID}).OrderBy("end_date").ToSql()
	if err != nil {
		return nil, err
	}
	items := make([]models.Deposit, 0)
	return items, r.db.SelectContext(ctx, &items, query, args...)
}

func (r *depositRepo) Create(ctx context.Context, userID uuid.UUID, bankName string, amount, rate float64, startDate, endDate, note string) (models.Deposit, error) {
	query, args, err := psql.Insert("deposits").
		Columns("user_id", "bank_name", "amount", "interest_rate", "start_date", "end_date", "note").
		Values(userID, bankName, amount, rate, startDate, endDate, note).
		Suffix("RETURNING *").ToSql()
	if err != nil {
		return models.Deposit{}, err
	}
	var dep models.Deposit
	return dep, r.db.QueryRowxContext(ctx, query, args...).StructScan(&dep)
}

func (r *depositRepo) Update(ctx context.Context, id, userID string, bankName *string, amount, rate *float64, startDate, endDate, note *string) (models.Deposit, error) {
	q := psql.Update("deposits").Where(sq.Eq{"id": id, "user_id": userID})
	if bankName != nil {
		q = q.Set("bank_name", *bankName)
	}
	if amount != nil {
		q = q.Set("amount", *amount)
	}
	if rate != nil {
		q = q.Set("interest_rate", *rate)
	}
	if startDate != nil {
		q = q.Set("start_date", sq.Expr("?::date", *startDate))
	}
	if endDate != nil {
		q = q.Set("end_date", sq.Expr("?::date", *endDate))
	}
	if note != nil {
		q = q.Set("note", *note)
	}
	query, args, err := q.Suffix("RETURNING *").ToSql()
	if err != nil {
		return models.Deposit{}, err
	}
	var dep models.Deposit
	return dep, r.db.QueryRowxContext(ctx, query, args...).StructScan(&dep)
}

func (r *depositRepo) Delete(ctx context.Context, id, userID string) (int64, error) {
	query, args, err := psql.Delete("deposits").Where(sq.Eq{"id": id, "user_id": userID}).ToSql()
	if err != nil {
		return 0, err
	}
	res, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// ── Credit ────────────────────────────────────────────────────────────────────

type creditRepo struct{ db *sqlx.DB }

func NewCreditRepo(db *sqlx.DB) CreditRepo { return &creditRepo{db: db} }

func (r *creditRepo) List(ctx context.Context, userID string) ([]models.Credit, error) {
	query, args, err := psql.Select("*").From("credits").
		Where(sq.Eq{"user_id": userID}).OrderBy("created_at DESC").ToSql()
	if err != nil {
		return nil, err
	}
	items := make([]models.Credit, 0)
	return items, r.db.SelectContext(ctx, &items, query, args...)
}

func (r *creditRepo) Create(ctx context.Context, userID uuid.UUID, creditType, bankName string, total, remaining, rate, monthly float64, note string) (models.Credit, error) {
	query, args, err := psql.Insert("credits").
		Columns("user_id", "type", "bank_name", "total_amount", "remaining_balance", "interest_rate", "monthly_payment", "note").
		Values(userID, creditType, bankName, total, remaining, rate, monthly, note).
		Suffix("RETURNING *").ToSql()
	if err != nil {
		return models.Credit{}, err
	}
	var cr models.Credit
	return cr, r.db.QueryRowxContext(ctx, query, args...).StructScan(&cr)
}

func (r *creditRepo) Update(ctx context.Context, id, userID string, bankName *string, total, remaining, rate, monthly *float64, note *string) (models.Credit, error) {
	q := psql.Update("credits").Where(sq.Eq{"id": id, "user_id": userID})
	if bankName != nil {
		q = q.Set("bank_name", *bankName)
	}
	if total != nil {
		q = q.Set("total_amount", *total)
	}
	if remaining != nil {
		q = q.Set("remaining_balance", *remaining)
	}
	if rate != nil {
		q = q.Set("interest_rate", *rate)
	}
	if monthly != nil {
		q = q.Set("monthly_payment", *monthly)
	}
	if note != nil {
		q = q.Set("note", *note)
	}
	query, args, err := q.Suffix("RETURNING *").ToSql()
	if err != nil {
		return models.Credit{}, err
	}
	var cr models.Credit
	return cr, r.db.QueryRowxContext(ctx, query, args...).StructScan(&cr)
}

func (r *creditRepo) Delete(ctx context.Context, id, userID string) (int64, error) {
	query, args, err := psql.Delete("credits").Where(sq.Eq{"id": id, "user_id": userID}).ToSql()
	if err != nil {
		return 0, err
	}
	res, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// ── Category ──────────────────────────────────────────────────────────────────

type categoryRepo struct{ db *sqlx.DB }

func NewCategoryRepo(db *sqlx.DB) CategoryRepo { return &categoryRepo{db: db} }

func (r *categoryRepo) ListForUser(ctx context.Context, userID string) ([]models.Category, error) {
	query, args, err := psql.Select("*").From("categories").
		Where(sq.Or{sq.Eq{"user_id": nil}, sq.Eq{"user_id": userID}}).
		OrderBy("is_system DESC", "name").ToSql()
	if err != nil {
		return nil, err
	}
	cats := make([]models.Category, 0)
	return cats, r.db.SelectContext(ctx, &cats, query, args...)
}

func (r *categoryRepo) ListNamesForUser(ctx context.Context, userID uuid.UUID) ([]string, error) {
	query, args, err := psql.Select("name").From("categories").
		Where(sq.Or{sq.Eq{"user_id": nil}, sq.Eq{"user_id": userID}}).ToSql()
	if err != nil {
		return nil, err
	}
	var names []string
	return names, r.db.SelectContext(ctx, &names, query, args...)
}

func (r *categoryRepo) FindIDByName(ctx context.Context, name string, userID uuid.UUID) (uuid.UUID, error) {
	query, args, err := psql.Select("id").From("categories").
		Where(sq.Eq{"name": name}).
		Where(sq.Or{sq.Eq{"user_id": nil}, sq.Eq{"user_id": userID}}).
		Limit(1).ToSql()
	if err != nil {
		return uuid.UUID{}, err
	}
	var id uuid.UUID
	return id, r.db.QueryRowContext(ctx, query, args...).Scan(&id)
}

func (r *categoryRepo) Create(ctx context.Context, userID uuid.UUID, name string) (models.Category, error) {
	query, args, err := psql.Insert("categories").
		Columns("user_id", "name", "is_system").
		Values(userID, name, false).
		Suffix("RETURNING *").ToSql()
	if err != nil {
		return models.Category{}, err
	}
	var cat models.Category
	return cat, r.db.QueryRowxContext(ctx, query, args...).StructScan(&cat)
}

// ── User ──────────────────────────────────────────────────────────────────────

type userRepo struct{ db *sqlx.DB }

func NewUserRepo(db *sqlx.DB) UserRepo { return &userRepo{db: db} }

func (r *userRepo) FindByEmail(ctx context.Context, email string) (models.User, error) {
	query, args, err := psql.Select("*").From("users").Where(sq.Eq{"email": email}).Limit(1).ToSql()
	if err != nil {
		return models.User{}, err
	}
	var u models.User
	return u, r.db.QueryRowxContext(ctx, query, args...).StructScan(&u)
}

func (r *userRepo) Create(ctx context.Context, email, hashedPassword string) (models.User, error) {
	query, args, err := psql.Insert("users").
		Columns("email", "hashed_password").
		Values(email, hashedPassword).
		Suffix("RETURNING *").ToSql()
	if err != nil {
		return models.User{}, err
	}
	var u models.User
	return u, r.db.QueryRowxContext(ctx, query, args...).StructScan(&u)
}

func (r *userRepo) AddXP(ctx context.Context, userID uuid.UUID, delta int) (models.User, error) {
	query, args, err := psql.Update("users").
		Set("xp_total", sq.Expr("xp_total + ?", delta)).
		Set("level", sq.Expr("(xp_total + ?) / 100 + 1", delta)).
		Where(sq.Eq{"id": userID}).
		Suffix("RETURNING *").ToSql()
	if err != nil {
		return models.User{}, err
	}
	var u models.User
	return u, r.db.QueryRowxContext(ctx, query, args...).StructScan(&u)
}

func (r *userRepo) GetProfile(ctx context.Context, userID uuid.UUID) (models.User, error) {
	query, args, err := psql.Select("*").From("users").Where(sq.Eq{"id": userID}).ToSql()
	if err != nil {
		return models.User{}, err
	}
	var u models.User
	return u, r.db.QueryRowxContext(ctx, query, args...).StructScan(&u)
}

// ── Achievement ───────────────────────────────────────────────────────────────

type achievementRepo struct{ db *sqlx.DB }

func NewAchievementRepo(db *sqlx.DB) AchievementRepo { return &achievementRepo{db: db} }

func (r *achievementRepo) ListForUser(ctx context.Context, userID uuid.UUID) ([]models.Achievement, error) {
	query := `
		SELECT a.id, a.code, a.name, a.description, ua.created_at AS earned_at
		FROM achievements a
		LEFT JOIN user_achievements ua ON ua.achievement_id = a.id AND ua.user_id = $1
		ORDER BY a.id`
	achievements := make([]models.Achievement, 0)
	return achievements, r.db.SelectContext(ctx, &achievements, query, userID)
}

func (r *achievementRepo) Unlock(ctx context.Context, userID uuid.UUID, code string) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO user_achievements (user_id, achievement_id)
		SELECT $1, id FROM achievements WHERE code = $2
		ON CONFLICT DO NOTHING`, userID, code)
	return err
}

// ── XPEvent ───────────────────────────────────────────────────────────────────

type xpEventRepo struct{ db *sqlx.DB }

func NewXPEventRepo(db *sqlx.DB) XPEventRepo { return &xpEventRepo{db: db} }

func (r *xpEventRepo) Insert(ctx context.Context, userID uuid.UUID, delta int, reason string) error {
	query, args, err := psql.Insert("xp_events").
		Columns("user_id", "delta", "reason").
		Values(userID, delta, reason).ToSql()
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx, query, args...)
	return err
}

// Убедимся что time импортирован (используется в моделях)
var _ = time.Now
