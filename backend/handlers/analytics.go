package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type categoryAmount struct {
	Category string  `db:"category" json:"category"`
	Amount   float64 `db:"amount" json:"amount"`
}

func (h *Handler) GetSummary(c *gin.Context) {
	userID := c.MustGet("userID").(string)

	period := c.Query("period") // e.g. "2024-01"
	query := `WHERE t.user_id = $1`
	args := []interface{}{userID}
	if period != "" {
		query += ` AND to_char(t.date, 'YYYY-MM') = $2`
		args = append(args, period)
	}

	var summary struct {
		Income  float64 `db:"income" json:"income"`
		Expense float64 `db:"expense" json:"expense"`
		Balance float64 `db:"balance" json:"balance"`
	}
	err := h.DB.QueryRowxContext(c.Request.Context(), `
		SELECT
			COALESCE(SUM(CASE WHEN type='income'  THEN amount ELSE 0 END), 0) AS income,
			COALESCE(SUM(CASE WHEN type='expense' THEN amount ELSE 0 END), 0) AS expense,
			COALESCE(SUM(CASE WHEN type='income'  THEN amount ELSE -amount END), 0) AS balance
		FROM transactions t `+query, args...,
	).StructScan(&summary)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	byCategory := make([]categoryAmount, 0)
	catArgs := append([]interface{}{userID}, args[1:]...)
	catQuery := `
		SELECT COALESCE(cat.name, 'Прочее') AS category,
		       SUM(t.amount) AS amount
		FROM transactions t
		LEFT JOIN categories cat ON cat.id = t.category_id
		WHERE t.user_id = $1 AND t.type = 'expense'`
	if period != "" {
		catQuery += ` AND to_char(t.date, 'YYYY-MM') = $2`
	}
	catQuery += ` GROUP BY cat.name ORDER BY amount DESC`
	_ = h.DB.SelectContext(c.Request.Context(), &byCategory, catQuery, catArgs...)

	c.JSON(http.StatusOK, gin.H{
		"income":      summary.Income,
		"expense":     summary.Expense,
		"balance":     summary.Balance,
		"by_category": byCategory,
	})
}

type timePoint struct {
	Period string  `db:"period" json:"period"`
	Income float64 `db:"income" json:"income"`
	Expense float64 `db:"expense" json:"expense"`
}

func (h *Handler) GetOverTime(c *gin.Context) {
	userID := c.MustGet("userID").(string)

	points := make([]timePoint, 0)
	err := h.DB.SelectContext(c.Request.Context(), &points, `
		SELECT
			to_char(date, 'YYYY-MM') AS period,
			SUM(CASE WHEN type='income'  THEN amount ELSE 0 END) AS income,
			SUM(CASE WHEN type='expense' THEN amount ELSE 0 END) AS expense
		FROM transactions
		WHERE user_id = $1
		GROUP BY period
		ORDER BY period DESC
		LIMIT 12`,
		userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, points)
}
