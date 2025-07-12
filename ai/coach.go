package ai

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"league-stats-api/model"
	"league-stats-api/service"
	"net/http"
	"os"
)

func AnalyzeMatch(shard string, puuid string, matchId string, locale string) (map[string]interface{}, error) {

	url := "https://openrouter.ai/api/v1/chat/completions"

	match, _ := service.GetMatchById(shard, matchId)
	timeline, _ := service.GetTimelineByMatchId(shard, matchId)

	optmized, err := OptimizeMatchTimeline(puuid, *match, *timeline, service.GetItems("en_US"))

	if err != nil {
		return nil, fmt.Errorf("failed to optimize match timeline: %w", err)
	}

	otmizedString, err := json.Marshal(optmized)

	if err != nil {
		return nil, fmt.Errorf("failed to get match data: %w", err)
	}

	systemPrompt := fmt.Sprintf(`
		You are a professional League of Legends coach AI.
		Your role is to analyze a player's match performance based strictly on the provided match data and timeline.
		
		Give detailed and constructive feedback to help the player improve.
		Your analysis must be:
		- Technically accurate and relevant to the role/champion played
		- Split into Early Game (0–14min), Mid Game (14–25min), and Late Game (25min+), if data allows
		- Focused on key aspects: laning phase, roaming, vision control, teamfights, objective control, positioning, mechanics, and decision making

		Also include:
		- A friendly but direct coaching tone
		- An AI Performance Score (0–100) for each phase, and an overall score

		Important rules:
		- NEVER guess or make up information
		- ONLY use what is provided in the match and timeline data
		- Respect the user's locale and respond accordingly (language, terminology): %s
	`, locale)

	userPrompt := fmt.Sprintf(`
		Below is optimized League of Legends data from one match, including match summary and timeline events.
		You are analyzing the player with: "participantId": %s
		Data:
		%s
	`, puuid, string(otmizedString))

	payload := map[string]any{
		"model": os.Getenv("OPENROUTER_MODEL"),
		"messages": []map[string]any{
			{
				"role":    "system",
				"content": systemPrompt,
			},
			{
				"role":    "user",
				"content": userPrompt,
			},
		},
	}

	jsonData, err := json.Marshal(payload)

	if err != nil {
		panic(err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))

	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+os.Getenv("OPENROUTER_API_KEY"))

	client := &http.Client{}
	res, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	var response map[string]any
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, err
	}

	return map[string]any{
		"content": response["choices"].([]any)[0].(map[string]any)["message"].(map[string]any)["content"],
	}, err

}

func OptimizeMatchTimeline(
	summonerPuuid string,
	match model.MatchData,
	timeline model.TimelineData,
	items map[string]model.ItemData,
) (*model.OptimizedMatch, error) {
	var participant model.MatchParticipant
	found := false

	for _, p := range match.Info.Participants {
		if p.Puuid == summonerPuuid {
			participant = p
			found = true
		}
	}
	if !found {
		return nil, errors.New("participant not found in match data")
	}

	participantID := participant.ParticipantId
	var relevantEvents []model.MatchEvent

	for _, frame := range timeline.Info.Frames {
		for _, event := range frame.Events {
			if event.ParticipantId == participantID ||
				event.KillerId == participantID ||
				event.VictimId == participantID {
				relevantEvents = append(relevantEvents, event)
			}
		}
	}

	var eventSummary []string
	for _, e := range relevantEvents {
		minute := e.Timestamp / 60000
		switch e.Type {
		case "CHAMPION_KILL":
			if e.KillerId == participantID {
				victim := match.Info.Participants[e.VictimId-1].ChampionName
				eventSummary = append(eventSummary, fmt.Sprintf("Kill %s at %d min", victim, minute))
			} else {
				killer := match.Info.Participants[e.KillerId-1].ChampionName
				eventSummary = append(eventSummary, fmt.Sprintf("Death to %s at %d min", killer, minute))
			}
		case "CHAMPION_SPECIAL_KILL":
			eventSummary = append(eventSummary, fmt.Sprintf("Got a %s at %d min", e.KillType, minute))
		case "ITEM_PURCHASED":
			if item, ok := items[e.ItemId]; ok {
				eventSummary = append(eventSummary, fmt.Sprintf("Bought %s at %d min", item.Name, minute))
			}
		case "ITEM_UNDO":
			if item, ok := items[e.ItemId]; ok {
				eventSummary = append(eventSummary, fmt.Sprintf("Undid %s at %d min", item.Name, minute))
			}
		case "ITEM_DESTROYED":
			if item, ok := items[e.ItemId]; ok {
				eventSummary = append(eventSummary, fmt.Sprintf("Destroyed %s at %d min", item.Name, minute))
			}
		case "SKILL_LEVEL_UP":
			eventSummary = append(eventSummary, fmt.Sprintf("Upgraded skill slot %d at %d min", e.SkillSlot, minute))
		case "ELITE_MONSTER_KILL":
			eventSummary = append(eventSummary, fmt.Sprintf("Took %s at %d min", e.MonsterType, minute)) // DRAGON, BARON_NASHOR, etc.
		case "BUILDING_KILL":
			eventSummary = append(eventSummary, fmt.Sprintf("Destroyed %s at %d min", e.BuildingType, minute)) // TOWER_BUILDING, INHIBITOR_BUILDING
		case "WARD_PLACED":
			eventSummary = append(eventSummary, fmt.Sprintf("Placed ward at %d min", minute))
		case "WARD_KILL":
			eventSummary = append(eventSummary, fmt.Sprintf("Destroyed ward at %d min", minute))
		case "TURRET_PLATE_DESTROYED":
			eventSummary = append(eventSummary, fmt.Sprintf("Destroyed turret plate at %d min", minute))
		case "CHAMPION_TRANSFORM":
			eventSummary = append(eventSummary, fmt.Sprintf("Transformed champion at %d min", minute)) // Ex: Kayn forma azul/vermelha
		default:
			eventSummary = append(eventSummary, fmt.Sprintf("%s at %d min", e.Type, minute))
		}
	}

	itemIDs := []int{
		participant.Item0,
		participant.Item1,
		participant.Item2,
		participant.Item3,
		participant.Item4,
		participant.Item5,
		participant.Item6,
	}

	namedItems := make([]string, 0)
	for _, id := range itemIDs {
		if id != 0 {
			if item, ok := items[string(rune(id))]; ok {
				namedItems = append(namedItems, fmt.Sprintf("%s (ID: %s)", item.Name, fmt.Sprint(id)))
			}
		}
	}

	return &model.OptimizedMatch{
		Name:         fmt.Sprintf("%s#%s", participant.RiotIdGameName, participant.RiotIdTagline),
		Champion:     participant.ChampionName,
		Role:         chooseFirst(participant.TeamPosition, participant.Role),
		KDA:          fmt.Sprintf("%d/%d/%d", participant.Kills, participant.Deaths, participant.Assists),
		TotalDamage:  participant.TotalDamageDealtToChampions,
		VisionScore:  participant.VisionScore,
		Items:        namedItems,
		EventSummary: eventSummary,
	}, nil
}

func chooseFirst(a, b string) string {
	if a != "" {
		return a
	}
	return b
}
