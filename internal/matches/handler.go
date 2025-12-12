package matches

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/igorzizinio/league-stats-api/internal/ai"
	"github.com/igorzizinio/league-stats-api/internal/model"
	"github.com/igorzizinio/league-stats-api/internal/riot"
)

func RegisterRoutes(router *gin.Engine) {
	router.GET("/matchlist/:riotRegion/:puuid", GetMatchlistByPuuid)
	router.POST("/match/:riotRegion/:matchId/analyze/:participantPuuid", AnalyzeMatch)
	router.POST("/match/:riotRegion/:matchId/analyze/:participantPuuid/stream", AnalyzeMatchStream)
	router.GET("/match/:riotRegion/:matchId", GetMatchById)
}

func GetMatchById(ctx *gin.Context) {
	riotRegion := ctx.Param("riotRegion")
	matchId := ctx.Param("matchId")

	match, err := riot.GetMatchById(riotRegion, matchId)
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

	matches, err := riot.GetMatchlistByPuuid(riotRegion, puuid, &req)
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

// AnalyzeMatchStream handles streaming AI analysis via Server-Sent Events
func AnalyzeMatchStream(ctx *gin.Context) {
	riotRegion := ctx.Param("riotRegion")
	matchId := ctx.Param("matchId")
	participantPuuid := ctx.Param("participantPuuid")

	locale := ctx.DefaultQuery("locale", "en_US")

	// Set headers for Server-Sent Events
	ctx.Header("Content-Type", "text/event-stream")
	ctx.Header("Cache-Control", "no-cache")
	ctx.Header("Connection", "keep-alive")
	ctx.Header("Transfer-Encoding", "chunked")
	ctx.Header("Access-Control-Allow-Origin", "*")

	// Flush headers immediately
	ctx.Writer.Flush()

	err := ai.AnalyzeMatchStream(riotRegion, participantPuuid, matchId, locale, func(chunk ai.StreamChunk) error {
		if chunk.Done {
			// Send done event
			_, err := ctx.Writer.Write([]byte("data: {\"done\":true}\n\n"))
			ctx.Writer.Flush()
			return err
		}

		// Escape newlines and special characters for JSON
		data := fmt.Sprintf("data: {\"content\":%q,\"done\":false}\n\n", chunk.Content)
		_, err := ctx.Writer.Write([]byte(data))
		if err != nil {
			return err
		}
		ctx.Writer.Flush()
		return nil
	})

	if err != nil {
		fmt.Println("Error streaming match analysis:", err)
		// If we haven't started streaming yet, we can return an error
		// Otherwise, we just close the connection
		ctx.Writer.Write([]byte(fmt.Sprintf("data: {\"error\":%q}\n\n", err.Error())))
		ctx.Writer.Flush()
	}
}
