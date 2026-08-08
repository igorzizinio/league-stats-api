package ai

import (
	"encoding/json"
	"fmt"

	"github.com/igorzizinio/league-stats-api/internal/ddragon"
	"github.com/igorzizinio/league-stats-api/internal/riot"
)

// prepareMatchData prepares the match data and prompts for AI analysis
func prepareMatchData(shard string, puuid string, matchId string, locale string) (string, string, error) {
	match, err := riot.GetMatchById(shard, matchId)
	if err != nil {
		return "", "", fmt.Errorf("failed to fetch match: %w", err)
	}

	timeline, err := riot.GetTimelineByMatchId(shard, matchId)
	if err != nil {
		return "", "", fmt.Errorf("failed to fetch timeline: %w", err)
	}

	// Keep static data in English to avoid duplicating locale variants.
	fullItems := ddragon.GetFullItems("en_US")
	optimized, err := OptimizeMatchTimeline(puuid, *match, *timeline, fullItems)
	if err != nil {
		return "", "", fmt.Errorf("failed to optimize match timeline: %w", err)
	}

	optimizedBytes, err := json.Marshal(optimized)
	if err != nil {
		return "", "", fmt.Errorf("failed to encode match data: %w", err)
	}

	systemPrompt := fmt.Sprintf(`You are a high-level League of Legends coach AI specialized in evidence-based gameplay analysis.

Your task is to analyze the player's performance strictly using only the provided structured data.

Strict evidence rules:
- Never guess or fabricate information.
- Every conclusion must map to explicit evidence in the payload.
- If evidence is insufficient, explicitly say: "Not enough data available."

Allowed data sources:
- GameContext
- PlayerSnapshot
- OpponentSnapshot
- PhaseStats
- CombatStats
- ObjectiveStats
- VisionStats
- DataCompleteness
- EventSummary
- GoldXpTimeline
- ItemSummaries
- ChampionData
- Items
- Runes
- SkillOrder
- Challenges

Respond in the user's preferred language: %s

Output sections:
1) Overall Performance (score + summary)
2) Early Game (0-14) — use GoldXpTimeline and OpponentSnapshot to judge the laning phase against the direct lane opponent
3) Mid Game (14-25)
4) Late Game (25+)
5) Build Analysis — cross-reference Items/ItemSummaries with Runes and ChampionData
6) Vision Analysis
7) Mechanics Analysis — SkillOrder and EventSummary support this
8) Top 3 Strengths
9) Top 3 Mistakes
10) Improvement Priorities
11) Final Coaching Advice

Evidence requirement:
Every major point must include concrete evidence snippets from the payload. Challenges contains Riot-computed advanced stats (e.g. killParticipation, teamDamagePercentage) that can support claims in any section.
`, locale)

	userPrompt := fmt.Sprintf(`Below is optimized League of Legends match data.
You are analyzing the player with puuid: %s
Data:
%s`, puuid, string(optimizedBytes))

	return systemPrompt, userPrompt, nil
}
