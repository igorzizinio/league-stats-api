package champion

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/igorzizinio/league-stats-api/internal/riot"
)

func RegisterRoutes(router *gin.Engine) {
	router.GET("/champion-rotation/:region", GetChampionRotation)
}

func GetChampionRotation(ctx *gin.Context) {
	region := ctx.Param("region")

	data, err := riot.GetChampionRotation(region)

	if err != nil {
		fmt.Println("Error getting free champion rotation", err)
		ctx.JSON(500, gin.H{"error": "Failed to fetch champion rotation"})
		return
	}

	ctx.JSON(200, data)
}
