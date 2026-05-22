package handlers

import (
	"context"
	"encoding/csv"
	"finquest/repository"
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
	userID, _ := uuid.Parse(c.MustGet("userID").(string))

	f := repository.TransactionFilter{}
	f.DateFrom = c.Query("date_from")
	f.DateTo = c.Query("date_to")
	f.CategoryID = c.Query("category_id")
	if l := c.Query("limit"); l != "" {
		f.Limit, _ = strconv.Atoi(l)
	}
	if o := c.Query("offset"); o != "" {
		f.Offset, _ = strconv.Atoi(o)
	}

	txs, err := h.Transactions.List(c.Request.Context(), userID, f)
	if err != nil {
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
	if _, err := time.Parse("2006-01-02", req.Date); err != nil {
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

	tx, err := h.Transactions.Create(c.Request.Context(), userID, req.Amount, req.Type, catID, req.Date, req.Note)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	_ = services.AwardXP(c.Request.Context(), h.Users, h.XPEvents, h.Achievements, h.Transactions, userID, 10, "transaction_added")
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

	var catID *string
	if req.CategoryID != nil && *req.CategoryID != "" {
		if _, err := uuid.Parse(*req.CategoryID); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category_id"})
			return
		}
		catID = req.CategoryID
	}

	tx, err := h.Transactions.Update(c.Request.Context(), txID, userID, req.Amount, req.Type, catID, req.Date, req.Note)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "transaction not found"})
		return
	}
	c.JSON(http.StatusOK, tx)
}

func (h *Handler) DeleteTransaction(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	txID := c.Param("id")

	n, err := h.Transactions.Delete(c.Request.Context(), txID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
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

	catNames, _ := h.Categories.ListNamesForUser(c.Request.Context(), userID)

	txForAI := make([]services.TransactionForAI, len(rows))
	for i, r := range rows {
		txForAI[i] = services.TransactionForAI{Note: r.Note}
	}
	catResults := services.CategorizeBatch(c.Request.Context(), h.Cfg.AnthropicKey, txForAI, catNames)

	imported := 0
	for i, row := range rows {
		var catID *uuid.UUID
		if catResults[i].CategoryName != "" {
			id, err := h.Categories.FindIDByName(c.Request.Context(), catResults[i].CategoryName, userID)
			if err == nil {
				catID = &id
			}
		}
		conf := catResults[i].Confidence
		var confPtr *float64
		if conf > 0 {
			confPtr = &conf
		}
		ok, err := h.Transactions.ImportOne(c.Request.Context(), userID, row.Amount, row.Type, catID, row.Date, row.Note, row.ExternalID, confPtr)
		if err == nil && ok {
			imported++
		}
	}

	if imported > 0 {
		_ = services.AwardXP(context.Background(), h.Users, h.XPEvents, h.Achievements, h.Transactions, userID, imported*5, "csv_import")
		_ = h.Achievements.Unlock(context.Background(), userID, "first_import")
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
	for i, rec := range records[1:] {
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
			Amount: amount, Type: txType, Date: date, Note: note,
			ExternalID: strconv.Itoa(i),
		})
	}
	return rows, nil
}
