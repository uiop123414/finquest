package handlers

import (
	"finquest/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (h *Handler) GetCategories(c *gin.Context) {
	userID := c.MustGet("userID").(string)

	cats := make([]models.Category, 0)
	err := h.DB.SelectContext(c.Request.Context(), &cats,
		`SELECT * FROM categories WHERE user_id IS NULL OR user_id = $1 ORDER BY is_system DESC, name`,
		userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, cats)
}

type createCategoryRequest struct {
	Name string `json:"name" binding:"required"`
}

func (h *Handler) CreateCategory(c *gin.Context) {
	userID, _ := uuid.Parse(c.MustGet("userID").(string))

	var req createCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var cat models.Category
	err := h.DB.QueryRowxContext(c.Request.Context(),
		`INSERT INTO categories (user_id, name, is_system) VALUES ($1, $2, false) RETURNING *`,
		userID, req.Name,
	).StructScan(&cat)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, cat)
}
