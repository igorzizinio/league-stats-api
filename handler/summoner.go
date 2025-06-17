package handler

import (
	"league-stats-api/service"
	"league-stats-api/util"

	"github.com/gin-gonic/gin"
)

func GetSummonerByPuuid(ctx *gin.Context) {
	region := ctx.Param("region")
	puuid := ctx.Param("puuid")

	account, err := service.GetAccountByPuuid(string(util.RiotRegionFromLeague(region)), puuid)
	if err != nil {
		ctx.JSON(500, gin.H{"error": "Failed to retrieve account by PUUID"})
		return
	}

	summoner, err := service.GetSummonerByPuuid(region, &puuid)
	if err != nil {
		ctx.JSON(500, gin.H{"error": "Failed to retrieve summoner"})
		return
	}
	ctx.JSON(200, gin.H{
		"account":    account,
		"summoner":   summoner,
		"region":     region,
		"riotRegion": util.RiotRegionFromLeague(region),
	})

}

func GetSummonerByRiotId(ctx *gin.Context) {
	region := ctx.Param("region")
	riotRegion := util.RiotRegionFromLeague(region)

	gameName := ctx.Param("gameName")
	tagLine := ctx.Param("tagLine")

	account, err := service.GetAccountByRiotId(string(riotRegion), gameName, tagLine)
	if err != nil {
		ctx.JSON(500, gin.H{"error": "Failed to retrieve riot account"})
		return
	}

	summoner, err := service.GetSummonerByPuuid(region, &account.Puuid)
	if err != nil {
		ctx.JSON(500, gin.H{"error": "Failed to retrieve summoner by PUUID"})
		return
	}

	ctx.JSON(200, gin.H{
		"account":    account,
		"summoner":   summoner,
		"region":     region,
		"riotRegion": string(riotRegion),
	})

}

func GetSummonerLeagueByPuuid(ctx *gin.Context) {
	region := ctx.Param("region")
	puuid := ctx.Param("puuid")

	leagues, err := service.GetSummonerLeagueByPuuid(region, puuid)
	if err != nil {
		ctx.JSON(500, gin.H{"error": "Failed to retrieve summoner leagues"})
		return
	}
	ctx.JSON(200, leagues)
}

func GetSummonerMasteriesByPuuid(ctx *gin.Context) {
	region := ctx.Param("region")
	puuid := ctx.Param("puuid")

	masteries, err := service.GetSummonerChampionMasteries(region, puuid)
	if err != nil {
		ctx.JSON(500, gin.H{"error": "Failed to retrieve summoner masteries"})
		return
	}
	ctx.JSON(200, masteries)
}
