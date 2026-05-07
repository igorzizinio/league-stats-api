package ai

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/igorzizinio/league-stats-api/internal/ddragon"
	"github.com/igorzizinio/league-stats-api/internal/model"
	"github.com/igorzizinio/league-stats-api/internal/riot"
)

// StreamChunk represents a chunk of streamed response
type StreamChunk struct {
	Content string `json:"content"`
	Done    bool   `json:"done"`
}

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
- PhaseStats
- CombatStats
- ObjectiveStats
- VisionStats
- DataCompleteness
- EventSummary
- ItemSummaries
- ChampionData
- Items

Respond in the user's preferred language: %s

Output sections:
1) Overall Performance (score + summary)
2) Early Game (0-14)
3) Mid Game (14-25)
4) Late Game (25+)
5) Build Analysis
6) Vision Analysis
7) Mechanics Analysis
8) Top 3 Strengths
9) Top 3 Mistakes
10) Improvement Priorities
11) Final Coaching Advice

Evidence requirement:
Every major point must include concrete evidence snippets from the payload.
`, locale)

	userPrompt := fmt.Sprintf(`Below is optimized League of Legends match data.
You are analyzing the player with puuid: %s
Data:
%s`, puuid, string(optimizedBytes))

	return systemPrompt, userPrompt, nil
}

func AnalyzeMatch(shard string, puuid string, matchId string, locale string) (map[string]interface{}, error) {
	url := "https://openrouter.ai/api/v1/chat/completions"

	systemPrompt, userPrompt, err := prepareMatchData(shard, puuid, matchId, locale)
	if err != nil {
		return nil, err
	}

	payload := map[string]any{
		"model": os.Getenv("OPENROUTER_MODEL"),
		"messages": []map[string]any{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": userPrompt},
		},
		"reasoning": map[string]any{"enabled": true},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request payload: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+os.Getenv("OPENROUTER_API_KEY"))

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("API error (status %d): %s", res.StatusCode, string(body))
	}

	var response map[string]any
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, err
	}

	choices, ok := response["choices"].([]any)
	if !ok || len(choices) == 0 {
		return nil, errors.New("invalid API response: missing choices")
	}
	choice0, ok := choices[0].(map[string]any)
	if !ok {
		return nil, errors.New("invalid API response: invalid choice format")
	}
	message, ok := choice0["message"].(map[string]any)
	if !ok {
		return nil, errors.New("invalid API response: missing message")
	}
	content, _ := message["content"].(string)

	return map[string]any{"content": content}, nil
}

// AnalyzeMatchStream streams the AI analysis response chunk by chunk
func AnalyzeMatchStream(shard string, puuid string, matchId string, locale string, onChunk func(chunk StreamChunk) error) error {
	url := "https://openrouter.ai/api/v1/chat/completions"

	systemPrompt, userPrompt, err := prepareMatchData(shard, puuid, matchId, locale)
	if err != nil {
		return err
	}

	payload := map[string]any{
		"model":  os.Getenv("OPENROUTER_MODEL"),
		"stream": true,
		"messages": []map[string]any{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": userPrompt},
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+os.Getenv("OPENROUTER_API_KEY"))
	req.Header.Set("Accept", "text/event-stream")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("API error (status %d): %s", res.StatusCode, string(body))
	}

	reader := bufio.NewReader(res.Body)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("failed to read stream: %w", err)
		}

		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, ":") {
			continue
		}
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			if err := onChunk(StreamChunk{Content: "", Done: true}); err != nil {
				return err
			}
			break
		}

		var chunk map[string]any
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue
		}
		choices, ok := chunk["choices"].([]any)
		if !ok || len(choices) == 0 {
			continue
		}
		choice, ok := choices[0].(map[string]any)
		if !ok {
			continue
		}
		delta, ok := choice["delta"].(map[string]any)
		if !ok {
			continue
		}
		content, ok := delta["content"].(string)
		if !ok || content == "" {
			continue
		}

		if err := onChunk(StreamChunk{Content: content, Done: false}); err != nil {
			return err
		}
	}

	return nil
}

func OptimizeMatchTimeline(
	summonerPuuid string,
	match model.MatchData,
	timeline model.TimelineData,
	items map[string]any,
) (*model.OptimizedMatch, error) {
	var participant model.MatchParticipant
	found := false
	for _, p := range match.Info.Participants {
		if p.Puuid == summonerPuuid {
			participant = p
			found = true
			break
		}
	}
	if !found {
		return nil, errors.New("participant not found in match data")
	}

	participantID := participant.ParticipantId

	phaseStats := map[string]map[string]int{
		"early": {"kills": 0, "deaths": 0, "assists": 0, "itemPurchases": 0, "objectives": 0, "wardsPlaced": 0, "wardsKilled": 0},
		"mid":   {"kills": 0, "deaths": 0, "assists": 0, "itemPurchases": 0, "objectives": 0, "wardsPlaced": 0, "wardsKilled": 0},
		"late":  {"kills": 0, "deaths": 0, "assists": 0, "itemPurchases": 0, "objectives": 0, "wardsPlaced": 0, "wardsKilled": 0},
	}
	combatStats := map[string]int{"kills": 0, "deaths": 0, "assistsFromKillEvents": 0, "specialKills": 0}
	objectiveStats := map[string]int{"eliteMonsterKills": 0, "dragonKills": 0, "baronKills": 0, "riftHeraldKills": 0, "atakhanKills": 0, "towerKills": 0, "inhibitorKills": 0, "plates": 0}
	visionStats := map[string]int{"wardsPlaced": 0, "wardsKilled": 0}
	dataCompleteness := map[string]int{"timelineEventsProcessed": 0, "eventSummaryEntries": 0, "knownItemEvents": 0, "unknownItemEvents": 0}

	itemCounts := make(map[string]map[string]int)
	eventSummary := make([]string, 0, 128)

	for _, frame := range timeline.Info.Frames {
		for _, e := range frame.Events {
			isRelevant := e.ParticipantId == participantID || e.KillerId == participantID || e.VictimId == participantID || containsInt(e.AssistingParticipantIds, participantID)
			if !isRelevant {
				continue
			}

			dataCompleteness["timelineEventsProcessed"]++
			minute := e.Timestamp / 60000
			phase := phaseFromMinute(minute)

			switch e.Type {
			case "CHAMPION_KILL":
				if e.KillerId == participantID {
					combatStats["kills"]++
					phaseStats[phase]["kills"]++
					eventSummary = append(eventSummary, fmt.Sprintf("CHAMPION_KILL kill victim=%s minute=%d", participantNameByID(match.Info.Participants, e.VictimId), minute))
				} else if e.VictimId == participantID {
					combatStats["deaths"]++
					phaseStats[phase]["deaths"]++
					eventSummary = append(eventSummary, fmt.Sprintf("CHAMPION_KILL death killer=%s minute=%d", participantNameByID(match.Info.Participants, e.KillerId), minute))
				} else if containsInt(e.AssistingParticipantIds, participantID) {
					combatStats["assistsFromKillEvents"]++
					phaseStats[phase]["assists"]++
					eventSummary = append(eventSummary, fmt.Sprintf("CHAMPION_KILL assist minute=%d", minute))
				}
			case "CHAMPION_SPECIAL_KILL":
				combatStats["specialKills"]++
				eventSummary = append(eventSummary, fmt.Sprintf("CHAMPION_SPECIAL_KILL type=%s minute=%d", e.KillType, minute))
			case "ITEM_PURCHASED", "ITEM_UNDO", "ITEM_DESTROYED":
				if e.ItemId <= 0 {
					continue
				}
				itemID := strconv.Itoa(e.ItemId)
				if _, ok := itemCounts[itemID]; !ok {
					itemCounts[itemID] = map[string]int{"purchased": 0, "undone": 0, "destroyed": 0}
				}
				action := ""
				if e.Type == "ITEM_PURCHASED" {
					action = "purchased"
					phaseStats[phase]["itemPurchases"]++
				}
				if e.Type == "ITEM_UNDO" {
					action = "undone"
				}
				if e.Type == "ITEM_DESTROYED" {
					action = "destroyed"
				}
				itemCounts[itemID][action]++

				label := fmt.Sprintf("itemId=%s", itemID)
				if itemRaw, ok := items[itemID]; ok {
					if itemMap, ok := itemRaw.(map[string]any); ok {
						if name, ok := itemMap["name"].(string); ok && name != "" {
							label = fmt.Sprintf("item=%s itemId=%s", name, itemID)
						}
						dataCompleteness["knownItemEvents"]++
					} else {
						dataCompleteness["unknownItemEvents"]++
					}
				} else {
					dataCompleteness["unknownItemEvents"]++
				}
				eventSummary = append(eventSummary, fmt.Sprintf("%s %s minute=%d", e.Type, label, minute))
			case "SKILL_LEVEL_UP":
				eventSummary = append(eventSummary, fmt.Sprintf("SKILL_LEVEL_UP skillSlot=%d minute=%d", e.SkillSlot, minute))
			case "ELITE_MONSTER_KILL":
				objectiveStats["eliteMonsterKills"]++
				phaseStats[phase]["objectives"]++
				if e.MonsterType == "DRAGON" {
					objectiveStats["dragonKills"]++
				}
				if e.MonsterType == "BARON_NASHOR" {
					objectiveStats["baronKills"]++
				}
				if e.MonsterType == "RIFTHERALD" {
					objectiveStats["riftHeraldKills"]++
				}
				if e.MonsterType == "ATAKHAN" {
					objectiveStats["atakhanKills"]++
				}
				eventSummary = append(eventSummary, fmt.Sprintf("ELITE_MONSTER_KILL monsterType=%s monsterSubType=%s minute=%d", e.MonsterType, e.MonsterSubType, minute))
			case "BUILDING_KILL":
				phaseStats[phase]["objectives"]++
				if e.BuildingType == "TOWER_BUILDING" {
					objectiveStats["towerKills"]++
				}
				if e.BuildingType == "INHIBITOR_BUILDING" {
					objectiveStats["inhibitorKills"]++
				}
				eventSummary = append(eventSummary, fmt.Sprintf("BUILDING_KILL buildingType=%s laneType=%s minute=%d", e.BuildingType, e.LaneType, minute))
			case "WARD_PLACED":
				visionStats["wardsPlaced"]++
				phaseStats[phase]["wardsPlaced"]++
				eventSummary = append(eventSummary, fmt.Sprintf("WARD_PLACED wardType=%s minute=%d", e.WardType, minute))
			case "WARD_KILL":
				visionStats["wardsKilled"]++
				phaseStats[phase]["wardsKilled"]++
				eventSummary = append(eventSummary, fmt.Sprintf("WARD_KILL wardType=%s minute=%d", e.WardType, minute))
			case "TURRET_PLATE_DESTROYED":
				objectiveStats["plates"]++
				phaseStats[phase]["objectives"]++
				eventSummary = append(eventSummary, fmt.Sprintf("TURRET_PLATE_DESTROYED laneType=%s minute=%d", e.LaneType, minute))
			default:
				eventSummary = append(eventSummary, fmt.Sprintf("%s minute=%d", e.Type, minute))
			}
		}
	}

	if len(eventSummary) > 180 {
		eventSummary = eventSummary[len(eventSummary)-180:]
	}
	dataCompleteness["eventSummaryEntries"] = len(eventSummary)

	itemIDs := []int{participant.Item0, participant.Item1, participant.Item2, participant.Item3, participant.Item4, participant.Item5, participant.Item6}
	namedItems := make([]string, 0, len(itemIDs))
	itemSummaries := make(map[string]map[string]any)
	usedItemsMap := make(map[string]any)

	for slot, id := range itemIDs {
		if id == 0 {
			continue
		}
		key := strconv.Itoa(id)
		counts := map[string]int{"purchased": 0, "undone": 0, "destroyed": 0}
		if c, ok := itemCounts[key]; ok {
			counts = c
		}

		summary := map[string]any{"id": key, "slot": slot, "counts": counts, "exists": false}
		if itemRaw, ok := items[key]; ok {
			if itemMap, ok := itemRaw.(map[string]any); ok {
				summary["exists"] = true
				if n, ok := itemMap["name"].(string); ok {
					summary["name"] = n
					namedItems = append(namedItems, fmt.Sprintf("slot%d:%s (ID:%d)", slot, n, id))
				}
				if gold, ok := itemMap["gold"]; ok {
					summary["gold"] = gold
				}
				if from, ok := itemMap["from"]; ok {
					summary["components"] = from
				}
				if into, ok := itemMap["into"]; ok {
					summary["into"] = into
				}
				if tags, ok := itemMap["tags"]; ok {
					summary["tags"] = tags
				}
				usedItemsMap[key] = itemMap
			}
		}
		if _, ok := summary["name"]; !ok {
			namedItems = append(namedItems, fmt.Sprintf("slot%d:itemId:%d", slot, id))
		}
		itemSummaries[key] = summary
	}

	sort.Strings(namedItems)

	championRaw := ddragon.GetChampionData(participant.ChampionName, "en_US")
	championData := make(map[string]any)
	if championRaw != nil {
		if spells, ok := championRaw["spells"].([]any); ok {
			championData["spells"] = spells
		}
		if passive, ok := championRaw["passive"]; ok {
			championData["passive"] = passive
		}
		if stats, ok := championRaw["stats"].(map[string]any); ok {
			championData["stats"] = stats
		}
	}

	gameContext := map[string]any{
		"matchId":       match.Metadata.MatchId,
		"queueId":       match.Info.QueueId,
		"gameMode":      match.Info.GameMode,
		"gameType":      match.Info.GameType,
		"gameDurationS": match.Info.GameDuration,
		"gameDurationM": match.Info.GameDuration / 60,
	}

	cs := participant.TotalMinionsKilled + participant.NeutralMinionsKilled
	playerSnapshot := map[string]any{
		"participantId":  participant.ParticipantId,
		"teamId":         participant.TeamId,
		"champion":       participant.ChampionName,
		"role":           chooseFirst(participant.TeamPosition, participant.Role),
		"result":         participant.Win,
		"level":          participant.ChampLevel,
		"kills":          participant.Kills,
		"deaths":         participant.Deaths,
		"assists":        participant.Assists,
		"cs":             cs,
		"goldEarned":     participant.GoldEarned,
		"goldSpent":      participant.GoldSpent,
		"damageToChamps": participant.TotalDamageDealtToChampions,
		"visionScore":    participant.VisionScore,
	}

	return &model.OptimizedMatch{
		Name:             fmt.Sprintf("%s#%s", participant.RiotIdGameName, participant.RiotIdTagline),
		Champion:         participant.ChampionName,
		Role:             chooseFirst(participant.TeamPosition, participant.Role),
		KDA:              fmt.Sprintf("%d/%d/%d", participant.Kills, participant.Deaths, participant.Assists),
		TotalDamage:      participant.TotalDamageDealtToChampions,
		VisionScore:      participant.VisionScore,
		Items:            namedItems,
		ItemsFull:        usedItemsMap,
		ChampionData:     championData,
		ItemSummaries:    itemSummaries,
		EventSummary:     eventSummary,
		GameContext:      gameContext,
		PlayerSnapshot:   playerSnapshot,
		PhaseStats:       phaseStats,
		CombatStats:      combatStats,
		ObjectiveStats:   objectiveStats,
		VisionStats:      visionStats,
		DataCompleteness: dataCompleteness,
	}, nil
}

func phaseFromMinute(minute int) string {
	if minute < 14 {
		return "early"
	}
	if minute < 25 {
		return "mid"
	}
	return "late"
}

func participantNameByID(participants []model.MatchParticipant, participantID int) string {
	if participantID <= 0 || participantID > len(participants) {
		return "unknown"
	}
	return participants[participantID-1].ChampionName
}

func containsInt(values []int, target int) bool {
	for _, v := range values {
		if v == target {
			return true
		}
	}
	return false
}

func chooseFirst(a, b string) string {
	if a != "" {
		return a
	}
	return b
}
