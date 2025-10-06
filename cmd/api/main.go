package main

import (
	"fmt"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/igorzizinio/league-stats-api/internal/champion"
	"github.com/igorzizinio/league-stats-api/internal/matches"
	"github.com/igorzizinio/league-stats-api/internal/summoner"
	"github.com/joho/godotenv"
)

func main() {

	if err := godotenv.Load(); err != nil {
		fmt.Println("No .env file found, using system default environment variables.")
	}

	router := gin.Default()

	router.Use(cors.Default())

	router.GET("/hello", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{
			"message": "Hello!",
		})
	})

	champion.RegisterRoutes(router)
	matches.RegisterRoutes(router)
	summoner.RegisterRoutes(router)

	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	router.Run("0.0.0.0:" + port)
}
