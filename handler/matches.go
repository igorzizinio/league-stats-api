package handler

import (
	"fmt"
	"legue-stats-api/ai"
	"legue-stats-api/model"
	"legue-stats-api/service"

	"github.com/gin-gonic/gin"
)

func GetMatchById(ctx *gin.Context) {
	shard := ctx.Param("shard")
	matchId := ctx.Param("matchId")

	match, err := service.GetMatchById(shard, matchId)
	if err != nil {
		ctx.JSON(500, gin.H{"error": "Failed to retrieve match"})
		return
	}
	ctx.JSON(200, match)
}

func GetMatchlistByPuuid(ctx *gin.Context) {
	region := ctx.Param("region")
	puuid := ctx.Param("puuid")

	var req model.GetMatchesOptions
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(400, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	matches, err := service.GetMatchlistByPuuid(region, puuid, &req)
	if err != nil {
		ctx.JSON(500, gin.H{"error": "Failed to retrieve matches"})
		return
	}
	ctx.JSON(200, matches)
}

func AnalyzeMatch(ctx *gin.Context) {
	shard := ctx.Param("shard")
	matchId := ctx.Param("matchId")
	participantPuuid := ctx.Param("participantPuuid")

	result, err := ai.AnalyzeMatch(shard, participantPuuid, matchId)
	if err != nil {
		fmt.Println("Error analyzing match:", err)
		ctx.JSON(500, gin.H{"error": "Failed to analyze match"})
		return
	}
	ctx.JSON(200, result)
}
