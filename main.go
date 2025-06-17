package main

import (
	"league-stats-api/handler"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {

	godotenv.Load()

	router := gin.Default()

	router.Use(cors.Default())

	router.GET("/hello", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{
			"message": "Hello!",
		})
	})

	router.GET("/champion-rotation/:region", handler.GetChampionRotation)

	router.GET("/summoner/by-puuid/:region/:puuid", handler.GetSummonerByPuuid)
	router.GET("/summoner/by-riot-id/:region/:gameName/:tagLine", handler.GetSummonerByRiotId)

	router.GET("/summoner/league/:region/:puuid", handler.GetSummonerLeagueByPuuid)
	router.GET(("/summoner/masteries/:region/:puuid"), handler.GetSummonerMasteriesByPuuid)

	router.GET("/matchlist/:riotRegion/:puuid", handler.GetMatchlistByPuuid)

	router.POST("/match/:riotRegion/:matchId/analyze/:participantPuuid", handler.AnalyzeMatch)
	router.GET("/match/:riotRegion/:matchId", handler.GetMatchById)

	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	router.Run("0.0.0.0:" + port)
}
