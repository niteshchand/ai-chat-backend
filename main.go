package main

import (
	"ai-chat-backend/db"
	"ai-chat-backend/handlers"
	"ai-chat-backend/middleware"
	"log"
	"os"

	"fmt"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	fmt.Println("GROQ KEY:", os.Getenv("GROQ_API_KEY"))
	fmt.Println("GEMINI KEY:", os.Getenv("GEMINI_API_KEY"))

	if os.Getenv("GROQ_API_KEY") == "" {
		log.Fatal("GROQ_API_KEY not set")
	}

	db.Connect()

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{"http://localhost:3000", "https://ai-chat-frontend-iota.vercel.app"},
		AllowMethods: []string{"POST", "GET", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Content-Type", "Authorization"},
	}))
	// Public routes
	r.POST("/register", handlers.Register)
	r.POST("/login", handlers.Login)

	// Protected routes
	protected := r.Group("/")
	protected.Use(middleware.AuthMiddleware())
	{
		// Conversation routes — no rate limit
		protected.POST("/conversations", handlers.CreateConversation)
		protected.GET("/conversations", handlers.GetConversations)
		protected.GET("/conversations/:id", handlers.GetConversationMessages)
		protected.PUT("/conversations/:id", handlers.UpdateConversation)
		protected.DELETE("/conversations/:id", handlers.DeleteConversation)
		protected.GET("/usage", handlers.GetUsageHandler)
		protected.POST("/search", handlers.SearchHandler)

		// Protected routes
		protected.POST("/upload", handlers.UploadHandler)

		// History — no rate limit
		protected.GET("/history", handlers.GetHistoryHandler)

		// AI routes — rate limited 🔒
		aiRoutes := protected.Group("/")
		aiRoutes.Use(middleware.RateLimitMiddleware())
		{
			aiRoutes.POST("/chat", handlers.ChatHandler)
			aiRoutes.POST("/chat/stream", handlers.StreamHandler)
		}
	}

	r.Run(":8080")
}
