package handlers

import (
	"context"
	"encoding/csv"
	"finquest/models"
	"finquest/services"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (h *Handler) GetTransactions(c *gin.Context) {
	userID := c.MustGet("userID").(string)

	query := `SELECT * FROM transactions WHERE user_id = $1`
	args := []interface{}{userID}
	i := 2

	if from := c.Query("date_from"); from != "" {
		query += " AND date >= $" + strconv.Itoa(i)
		args = append(args, from)
		i++
	}
	if to := c.Query("date_to"); to != "" {
		query += " AND date <= $" + strconv.Itoa(i)
		args = append(args, to)
		i++
	}
	if catID := c.Query("category_id"); catID != "" {
		query += " AND category_id = $" + strconv.Itoa(i)
		args = append(args, catID)
		i++
	}

	limit := 50
	offset := 0
	if l := c.Query("limit"); l != "" {
		limit, _ = strconv.Atoi(l)
	}
	if o := c.Query("offset"); o != "" {
		offset, _ = strconv.Atoi(o)
	}
	query += " ORDER BY date DESC LIMIT $" + strconv.Itoa(i) + " OFFSET $" + strconv.Itoa(i+1)
	args = append(args, limit, offset)

	txs := make([]models.Transaction, 0)
	if err := h.DB.SelectContext(c.Request.Context(), &txs, query, args...); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, txs)
}

type createTransactionRequest struct {
	Amount     float64 `json:"amount" binding:"required,gt=0"`
	Type       string  `json:"type" binding:"required,oneof=income expense"`
	CategoryID *string `json:"category_id"`
	Date       string  `json:"date" binding:"required"`
	Note       string  `json:"note"`
}

func (h *Handler) CreateTransaction(c *gin.Context) {
	userID, _ := uuid.Parse(c.MustGet("userID").(string))

	var req createTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date format, use YYYY-MM-DD"})
		return
	}

	var catID *uuid.UUID
	if req.CategoryID != nil && *req.CategoryID != "" {
		id, err := uuid.Parse(*req.CategoryID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category_id"})
			return
		}
		catID = &id
	}

	var tx models.Transaction
	err = h.DB.QueryRowxContext(c.Request.Context(),
		`INSERT INTO transactions (user_id, amount, type, category_id, date, note)
         VALUES ($1, $2, $3, $4, $5, $6) RETURNING *`,
		userID, req.Amount, req.Type, catID, date, req.Note,
	).StructScan(&tx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Award XP
	_ = services.AwardXP(c.Request.Context(), h.DB, userID, 10, "transaction_added")

	c.JSON(http.StatusCreated, tx)
}

type updateTransactionRequest struct {
	Amount     *float64 `json:"amount"`
	Type       *string  `json:"type"`
	CategoryID *string  `json:"category_id"`
	Date       *string  `json:"date"`
	Note       *string  `json:"note"`
}

func (h *Handler) UpdateTransaction(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	txID := c.Param("id")

	var req updateTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var catID *uuid.UUID
	if req.CategoryID != nil && *req.CategoryID != "" {
		id, err := uuid.Parse(*req.CategoryID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category_id"})
			return
		}
		catID = &id
	}

	var tx models.Transaction
	err := h.DB.QueryRowxContext(c.Request.Context(), `
		UPDATE transactions SET
			amount      = COALESCE($1, amount),
			type        = COALESCE($2, type),
			category_id = COALESCE($3, category_id),
			date        = COALESCE($4::date, date),
			note        = COALESCE($5, note)
		WHERE id = $6 AND user_id = $7
		RETURNING *`,
		req.Amount, req.Type, catID, req.Date, req.Note, txID, userID,
	).StructScan(&tx)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "transaction not found"})
		return
	}

	c.JSON(http.StatusOK, tx)
}

func (h *Handler) DeleteTransaction(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	txID := c.Param("id")

	res, err := h.DB.ExecContext(c.Request.Context(),
		`DELETE FROM transactions WHERE id = $1 AND user_id = $2`, txID, userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "transaction not found"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func (h *Handler) ImportTransactions(c *gin.Context) {
	userID, _ := uuid.Parse(c.MustGet("userID").(string))

	file, _, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file required"})
		return
	}
	defer file.Close()

	rows, err := parseCSV(file)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Fetch categories for AI categorization
	var catNames []string
	_ = h.DB.SelectContext(c.Request.Context(), &catNames,
		`SELECT name FROM categories WHERE user_id IS NULL OR user_id = $1`, userID,
	)

	// Categorize
	txForAI := make([]services.TransactionForAI, len(rows))
	for i, r := range rows {
		txForAI[i] = services.TransactionForAI{Note: r.Note}
	}
	catResults := services.CategorizeBatch(c.Request.Context(), h.Cfg.AnthropicKey, txForAI, catNames)

	imported := 0
	for i, row := range rows {
		// Resolve category ID by name
		var catID *uuid.UUID
		if catResults[i].CategoryName != "" {
			var id uuid.UUID
			err := h.DB.QueryRowContext(c.Request.Context(),
				`SELECT id FROM categories WHERE name = $1 AND (user_id IS NULL OR user_id = $2) LIMIT 1`,
				catResults[i].CategoryName, userID,
			).Scan(&id)
			if err == nil {
				catID = &id
			}
		}

		conf := catResults[i].Confidence
		_, err := h.DB.ExecContext(c.Request.Context(), `
			INSERT INTO transactions (user_id, amount, type, category_id, date, note, external_id, ai_confidence)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			ON CONFLICT (user_id, external_id) DO NOTHING`,
			userID, row.Amount, row.Type, catID, row.Date, row.Note, row.ExternalID, conf,
		)
		if err == nil {
			imported++
		}
	}

	if imported > 0 {
		_ = services.AwardXP(context.Background(), h.DB, userID, imported*5, "csv_import")
		// Check first_import achievement
		_, _ = h.DB.ExecContext(context.Background(), `
			INSERT INTO user_achievements (user_id, achievement_id)
			SELECT $1, id FROM achievements WHERE code = 'first_import'
			ON CONFLICT DO NOTHING`, userID,
		)
	}

	c.JSON(http.StatusOK, gin.H{"imported": imported, "total": len(rows)})
}

type csvRow struct {
	Amount     float64
	Type       string
	Date       time.Time
	Note       string
	ExternalID string
}

// parseCSV expects columns: date,amount,type,note (header row required)
// type values: income | expense
func parseCSV(r io.Reader) ([]csvRow, error) {
	reader := csv.NewReader(r)
	reader.TrimLeadingSpace = true

	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(records) < 2 {
		return nil, nil
	}

	var rows []csvRow
	for i, rec := range records[1:] { // skip header
		if len(rec) < 3 {
			continue
		}
		date, err := time.Parse("2006-01-02", strings.TrimSpace(rec[0]))
		if err != nil {
			continue
		}
		amount, err := strconv.ParseFloat(strings.TrimSpace(rec[1]), 64)
		if err != nil {
			continue
		}
		txType := strings.TrimSpace(rec[2])
		if txType != "income" && txType != "expense" {
			continue
		}
		note := ""
		if len(rec) > 3 {
			note = strings.TrimSpace(rec[3])
		}
		rows = append(rows, csvRow{
			Amount:     amount,
			Type:       txType,
			Date:       date,
			Note:       note,
			ExternalID: strconv.Itoa(i),
		})
	}
	return rows, nil
}
