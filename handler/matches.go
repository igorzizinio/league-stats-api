package handler

import (
	"fmt"
	"legue-stats-api/ai"
	"legue-stats-api/model"
	"legue-stats-api/service"

	"github.com/gin-gonic/gin"
)

func GetMatchById(ctx *gin.Context) {
	riotRegion := ctx.Param("riotRegion")
	matchId := ctx.Param("matchId")

	match, err := service.GetMatchById(riotRegion, matchId)
	if err != nil {
		ctx.JSON(500, gin.H{"error": "Failed to retrieve match"})
		return
	}
	ctx.JSON(200, match)
}

func GetMatchlistByPuuid(ctx *gin.Context) {
	riotRegion := ctx.Param("riotRegion")
	puuid := ctx.Param("puuid")

	var req model.GetMatchesOptions

	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(400, gin.H{"error": "Invalid query parameters", "details": err.Error()})
		return
	}

	matches, err := service.GetMatchlistByPuuid(riotRegion, puuid, &req)
	if err != nil {
		ctx.JSON(500, gin.H{"error": "Failed to retrieve matches"})
		return
	}
	ctx.JSON(200, matches)
}

func AnalyzeMatch(ctx *gin.Context) {
	riotRegion := ctx.Param("riotRegion")
	matchId := ctx.Param("matchId")
	participantPuuid := ctx.Param("participantPuuid")

	locale := ctx.DefaultQuery("locale", "en_US")

	result, err := ai.AnalyzeMatch(riotRegion, participantPuuid, matchId, locale)
	if err != nil {
		fmt.Println("Error analyzing match:", err)
		ctx.JSON(500, gin.H{"error": "Failed to analyze match"})
		return
	}
	ctx.JSON(200, result)
}
