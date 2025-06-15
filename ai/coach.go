package ai

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"legue-stats-api/model"
	"legue-stats-api/service"
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
		You are a League of Legends coach. Your job is to analyze a player's performance based on their match and timeline data. Provide strategic, tactical, and mechanical feedback to help the player improve. Use a friendly but technically accurate tone. Never make up information — only use what is in the data.
      	You can break your feedback into early game, mid game, and late game insights if possible.
      	Also give an 'AI Score' out of 100 based on the player's performance in early, mid and late game if possible.
      	Note: Use user locale: %s
	  `, locale)

	userPrompt := fmt.Sprintf(`
		Here is the optmized data (match, timeline), the player is "participantId": %s

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

	var response map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"content": response["choices"].([]interface{})[0].(map[string]interface{})["message"].(map[string]interface{})["content"],
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
				eventSummary = append(eventSummary, fmt.Sprintf("Dead to %s at %d min", killer, minute))
			}
		case "ITEM_PURCHASED":
			if item, ok := items[e.ItemId]; ok {
				eventSummary = append(eventSummary, fmt.Sprintf("Item buy %s at %d min", item.Name, minute))
			}
		case "SKILL_LEVEL_UP":
			eventSummary = append(eventSummary, fmt.Sprintf("Up skill slot %d at %d min", e.SkillSlot, minute))
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
	filteredItems := make([]int, 0)
	for _, id := range itemIDs {
		if id != 0 {
			filteredItems = append(filteredItems, id)
		}
	}

	return &model.OptimizedMatch{
		Name:         fmt.Sprintf("%s#%s", participant.RiotIdGameName, participant.RiotIdTagline),
		Champion:     participant.ChampionName,
		Role:         chooseFirst(participant.TeamPosition, participant.Role),
		KDA:          fmt.Sprintf("%d/%d/%d", participant.Kills, participant.Deaths, participant.Assists),
		TotalDamage:  participant.TotalDamageDealtToChampions,
		VisionScore:  participant.VisionScore,
		Items:        filteredItems,
		EventSummary: eventSummary,
	}, nil
}

func chooseFirst(a, b string) string {
	if a != "" {
		return a
	}
	return b
}
