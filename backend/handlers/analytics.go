package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type categoryAmount struct {
	Category string  `db:"category" json:"category"`
	Amount   float64 `db:"amount"   json:"amount"`
}

func (h *Handler) GetSummary(c *gin.Context) {
	userID := c.MustGet("userID").(string)

	period := c.Query("period")
	rang := c.Query("range")

	whereBase := `WHERE t.user_id = $1`
	args := []interface{}{userID}

	switch {
	case period != "":
		whereBase += ` AND to_char(t.date, 'YYYY-MM') = $2`
		args = append(args, period)
	case rang == "6m":
		whereBase += ` AND t.date >= (NOW() - INTERVAL '6 months')::date`
	case rang == "1y":
		whereBase += ` AND t.date >= (NOW() - INTERVAL '1 year')::date`
	}

	var summary struct {
		Income  float64 `db:"income"  json:"income"`
		Expense float64 `db:"expense" json:"expense"`
		Balance float64 `db:"balance" json:"balance"`
	}
	err := h.DB.QueryRowxContext(c.Request.Context(), `
		SELECT
			COALESCE(SUM(CASE WHEN type='income'  THEN amount ELSE 0 END), 0) AS income,
			COALESCE(SUM(CASE WHEN type='expense' THEN amount ELSE 0 END), 0) AS expense,
			COALESCE(SUM(CASE WHEN type='income'  THEN amount ELSE -amount END), 0) AS balance
		FROM transactions t `+whereBase, args...,
	).StructScan(&summary)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	byCategory := make([]categoryAmount, 0)
	catArgs := append([]interface{}{userID}, args[1:]...)
	catWhere := `WHERE t.user_id = $1 AND t.type = 'expense'`
	switch {
	case period != "":
		catWhere += ` AND to_char(t.date, 'YYYY-MM') = $2`
	case rang == "6m":
		catWhere += ` AND t.date >= (NOW() - INTERVAL '6 months')::date`
	case rang == "1y":
		catWhere += ` AND t.date >= (NOW() - INTERVAL '1 year')::date`
	}
	catQuery := `
		SELECT COALESCE(cat.name, 'Прочее') AS category,
		       SUM(t.amount) AS amount
		FROM transactions t
		LEFT JOIN categories cat ON cat.id = t.category_id ` +
		catWhere + ` GROUP BY cat.name ORDER BY amount DESC`
	_ = h.DB.SelectContext(c.Request.Context(), &byCategory, catQuery, catArgs...)

	c.JSON(http.StatusOK, gin.H{
		"income":      summary.Income,
		"expense":     summary.Expense,
		"balance":     summary.Balance,
		"by_category": byCategory,
	})
}

type timePoint struct {
	Period  string  `db:"period"  json:"period"`
	Income  float64 `db:"income"  json:"income"`
	Expense float64 `db:"expense" json:"expense"`
}

func (h *Handler) GetOverTime(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	period := c.Query("period")
	rang := c.Query("range")

	points := make([]timePoint, 0)

	if period != "" {
		err := h.DB.SelectContext(c.Request.Context(), &points, `
			SELECT
				to_char(date, 'DD') AS period,
				SUM(CASE WHEN type='income'  THEN amount ELSE 0 END) AS income,
				SUM(CASE WHEN type='expense' THEN amount ELSE 0 END) AS expense
			FROM transactions
			WHERE user_id = $1 AND to_char(date, 'YYYY-MM') = $2
			GROUP BY date ORDER BY date`,
			userID, period,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, points)
		return
	}

	var query string
	switch rang {
	case "6m":
		query = `
			SELECT to_char(date, 'YYYY-MM') AS period,
			       SUM(CASE WHEN type='income'  THEN amount ELSE 0 END) AS income,
			       SUM(CASE WHEN type='expense' THEN amount ELSE 0 END) AS expense
			FROM transactions WHERE user_id = $1
			  AND date >= (NOW() - INTERVAL '6 months')::date
			GROUP BY period ORDER BY period`
	case "all":
		query = `
			SELECT to_char(date, 'YYYY-MM') AS period,
			       SUM(CASE WHEN type='income'  THEN amount ELSE 0 END) AS income,
			       SUM(CASE WHEN type='expense' THEN amount ELSE 0 END) AS expense
			FROM transactions WHERE user_id = $1
			GROUP BY period ORDER BY period`
	default:
		query = `
			SELECT to_char(date, 'YYYY-MM') AS period,
			       SUM(CASE WHEN type='income'  THEN amount ELSE 0 END) AS income,
			       SUM(CASE WHEN type='expense' THEN amount ELSE 0 END) AS expense
			FROM transactions WHERE user_id = $1
			  AND date >= (NOW() - INTERVAL '1 year')::date
			GROUP BY period ORDER BY period`
	}

	if err := h.DB.SelectContext(c.Request.Context(), &points, query, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, points)
}
