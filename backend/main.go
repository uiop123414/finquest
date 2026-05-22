package main

import (
	"finquest/config"
	"finquest/db"
	"finquest/handlers"
	"finquest/middleware"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()
	database := db.Connect(cfg.DatabaseURL)
	defer database.Close()

	r := gin.Default()
	r.Use(middleware.CORS())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	h := handlers.New(database, cfg)

	api := r.Group("/api/v1")

	// Public
	auth := api.Group("/auth")
	auth.POST("/register", h.Register)
	auth.POST("/login", h.Login)
	auth.POST("/refresh", h.Refresh)

	// Protected
	p := api.Group("/", middleware.AuthRequired(cfg.JWTSecret))

	p.GET("/categories", h.GetCategories)
	p.POST("/categories", h.CreateCategory)

	p.GET("/transactions", h.GetTransactions)
	p.POST("/transactions", h.CreateTransaction)
	p.PATCH("/transactions/:id", h.UpdateTransaction)
	p.DELETE("/transactions/:id", h.DeleteTransaction)
	p.POST("/transactions/import", h.ImportTransactions)

	p.GET("/analytics/summary", h.GetSummary)
	p.GET("/analytics/over-time", h.GetOverTime)

	p.GET("/gamification/profile", h.GetGamificationProfile)

	p.GET("/goals", h.GetGoals)
	p.POST("/goals", h.CreateGoal)
	p.PATCH("/goals/:id", h.UpdateGoal)
	p.DELETE("/goals/:id", h.DeleteGoal)

	p.GET("/investments/deposits", h.GetDeposits)
	p.POST("/investments/deposits", h.CreateDeposit)
	p.PATCH("/investments/deposits/:id", h.UpdateDeposit)
	p.DELETE("/investments/deposits/:id", h.DeleteDeposit)

	p.GET("/credits", h.GetCredits)
	p.POST("/credits", h.CreateCredit)
	p.PATCH("/credits/:id", h.UpdateCredit)
	p.DELETE("/credits/:id", h.DeleteCredit)

	p.POST("/ai/chat", h.AIChat)
	p.GET("/ai/advice", h.GetAIAdvice)

	r.Run(":" + cfg.Port)
}
