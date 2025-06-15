package handler

import (
	"fmt"
	"legue-stats-api/service"

	"github.com/gin-gonic/gin"
)

func GetChampionRotation(ctx *gin.Context) {
	region := ctx.Param("region")

	data, err := service.GetChampionRotation(region)

	if err != nil {
		fmt.Println("Error getting free champion rotation", err)
		ctx.JSON(500, gin.H{"error": "Failed to fetch champion rotation"})
		return
	}

	ctx.JSON(200, data)
}
