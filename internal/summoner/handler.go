package summoner

import (
	"github.com/gin-gonic/gin"
	"github.com/igorzizinio/league-stats-api/internal/riot"
	"github.com/igorzizinio/league-stats-api/internal/util"
)

func RegisterRoutes(router *gin.Engine) {
	router.GET("/summoner/by-puuid/:region/:puuid", GetSummonerByPuuid)
	router.GET("/summoner/by-riot-id/:region/:gameName/:tagLine", GetSummonerByRiotId)
	router.GET("/summoner/league/:region/:puuid", GetSummonerLeagueByPuuid)
	router.GET("/summoner/masteries/:region/:puuid", GetSummonerMasteriesByPuuid)
}

func GetSummonerByPuuid(ctx *gin.Context) {
	region := ctx.Param("region")
	puuid := ctx.Param("puuid")

	account, err := riot.GetAccountByPuuid(string(util.RiotRegionFromLeague(region)), puuid)
	if err != nil {
		ctx.JSON(500, gin.H{"error": "Failed to retrieve account by PUUID"})
		return
	}

	summoner, err := riot.GetSummonerByPuuid(region, &puuid)
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

	account, err := riot.GetAccountByRiotId(string(riotRegion), gameName, tagLine)
	if err != nil {
		ctx.JSON(500, gin.H{"error": "Failed to retrieve riot account"})
		return
	}

	summoner, err := riot.GetSummonerByPuuid(region, &account.Puuid)
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

	leagues, err := riot.GetSummonerLeagueByPuuid(region, puuid)
	if err != nil {
		ctx.JSON(500, gin.H{"error": "Failed to retrieve summoner leagues"})
		return
	}
	ctx.JSON(200, leagues)
}

func GetSummonerMasteriesByPuuid(ctx *gin.Context) {
	region := ctx.Param("region")
	puuid := ctx.Param("puuid")

	masteries, err := riot.GetSummonerChampionMasteries(region, puuid)
	if err != nil {
		ctx.JSON(500, gin.H{"error": "Failed to retrieve summoner masteries"})
		return
	}
	ctx.JSON(200, masteries)
}
