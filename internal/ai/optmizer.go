package ai

import (
	"errors"
	"fmt"
	"sort"
	"strconv"

	"github.com/igorzizinio/league-stats-api/internal/ddragon"
	"github.com/igorzizinio/league-stats-api/internal/model"
)

var timelineCheckpointMinutes = []int{5, 10, 15, 20, 25}

var coachRelevantChallenges = []string{
	"killParticipation",
	"teamDamagePercentage",
	"damagePerMinute",
	"goldPerMinute",
	"visionScorePerMinute",
	"laningPhaseGoldExpAdvantage",
	"earlyLaningPhaseGoldExpAdvantage",
	"maxCsAdvantageOnLaneOpponent",
	"maxLevelLeadLaneOpponent",
	"turretPlatesTaken",
	"soloKills",
	"outnumberedKills",
	"epicMonsterSteals",
	"dragonTakedowns",
	"baronTakedowns",
	"riftHeraldTakedowns",
	"skillshotsHit",
	"skillshotsDodged",
	"saveAllyFromDeath",
	"jungleCsBefore10Minutes",
	"laneMinionsFirst10Minutes",
	"visionScoreAdvantageLaneOpponent",
	"killsNearEnemyTurret",
	"effectiveHealAndShielding",
}

type eventEntry struct {
	minute int
	text   string
}

func formatTimestamp(ms int) string {
	totalSeconds := ms / 1000
	minutes := totalSeconds / 60
	seconds := totalSeconds % 60
	return fmt.Sprintf("%d:%02d", minutes, seconds)
}

func findLaneOpponent(participants []model.MatchParticipant, me model.MatchParticipant) (*model.MatchParticipant, bool) {
	myPos := chooseFirst(me.TeamPosition, me.Role)
	if myPos == "" {
		return nil, false
	}
	for i := range participants {
		p := participants[i]
		if p.TeamId == me.TeamId {
			continue
		}
		if chooseFirst(p.TeamPosition, p.Role) == myPos {
			return &p, true
		}
	}
	return nil, false
}

func thinEventSummary(entries []eventEntry, max int, alwaysKeepUntilMinute int) []string {
	if len(entries) <= max {
		out := make([]string, len(entries))
		for i, e := range entries {
			out[i] = e.text
		}
		return out
	}

	var early, rest []eventEntry
	for _, e := range entries {
		if e.minute <= alwaysKeepUntilMinute {
			early = append(early, e)
		} else {
			rest = append(rest, e)
		}
	}

	budget := max - len(early)
	if budget < 0 {
		budget = 0
	}
	if len(rest) > budget {
		rest = rest[len(rest)-budget:]
	}

	combined := append(early, rest...)
	out := make([]string, len(combined))
	for i, e := range combined {
		out[i] = e.text
	}
	return out
}

func buildGoldXpTimeline(frames []model.MatchFrame, participantID int, opponent *model.MatchParticipant, checkpoints []int) []map[string]any {
	result := make([]map[string]any, 0, len(checkpoints))
	pidKey := strconv.Itoa(participantID)

	var oppKey string
	if opponent != nil {
		oppKey = strconv.Itoa(opponent.ParticipantId)
	}

	for _, cpMinute := range checkpoints {
		targetMs := cpMinute * 60000

		var closest *model.MatchFrame
		for i := range frames {
			f := &frames[i]
			if f.Timestamp <= targetMs {
				closest = f
			} else {
				break
			}
		}
		if closest == nil {
			continue
		}

		mine, ok := closest.ParticipantFrames[pidKey]
		if !ok {
			continue
		}

		myCS := mine.MinionsKilled + mine.JungleMinionsKilled
		entry := map[string]any{
			"minute": cpMinute,
			"gold":   mine.TotalGold,
			"xp":     mine.XP,
			"cs":     myCS,
			"level":  mine.Level,
		}

		if opponent != nil {
			if opp, ok := closest.ParticipantFrames[oppKey]; ok {
				oppCS := opp.MinionsKilled + opp.JungleMinionsKilled
				entry["opponentGold"] = opp.TotalGold
				entry["opponentXp"] = opp.XP
				entry["opponentCs"] = oppCS
				entry["opponentLevel"] = opp.Level
				entry["goldDiff"] = mine.TotalGold - opp.TotalGold
				entry["xpDiff"] = mine.XP - opp.XP
				entry["csDiff"] = myCS - oppCS
			}
		}

		result = append(result, entry)
	}

	return result
}

func curatedChallenges(raw map[string]any) map[string]any {
	if raw == nil {
		return nil
	}
	out := make(map[string]any, len(coachRelevantChallenges))
	for _, key := range coachRelevantChallenges {
		if v, ok := raw[key]; ok {
			out[key] = v
		}
	}
	return out
}

func extractRunes(p model.MatchParticipant) map[string]any {
	runes := map[string]any{
		"summoner1Id": p.Summoner1Id,
		"summoner2Id": p.Summoner2Id,
	}

	var primaryPerks, secondaryPerks []int
	for _, style := range p.Perks.Styles {
		switch style.Description {
		case "primaryStyle":
			for _, sel := range style.Selections {
				primaryPerks = append(primaryPerks, sel.Perk)
			}
		case "subStyle":
			for _, sel := range style.Selections {
				secondaryPerks = append(secondaryPerks, sel.Perk)
			}
		}
	}

	keystone := 0
	if len(primaryPerks) > 0 {
		keystone = primaryPerks[0]
	}

	runes["primaryKeystone"] = keystone
	runes["primaryPerks"] = primaryPerks
	runes["secondaryPerks"] = secondaryPerks
	runes["statPerks"] = []int{p.Perks.StatPerks.Offense, p.Perks.StatPerks.Flex, p.Perks.StatPerks.Defense}
	return runes
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
	opponent, hasOpponent := findLaneOpponent(match.Info.Participants, participant)

	phaseStats := map[string]map[string]int{
		"early": {"kills": 0, "deaths": 0, "assists": 0, "itemPurchases": 0, "objectives": 0, "wardsPlaced": 0, "wardsKilled": 0},
		"mid":   {"kills": 0, "deaths": 0, "assists": 0, "itemPurchases": 0, "objectives": 0, "wardsPlaced": 0, "wardsKilled": 0},
		"late":  {"kills": 0, "deaths": 0, "assists": 0, "itemPurchases": 0, "objectives": 0, "wardsPlaced": 0, "wardsKilled": 0},
	}
	combatStats := map[string]int{"kills": 0, "deaths": 0, "assistsFromKillEvents": 0, "specialKills": 0, "soloKills": 0}
	objectiveStats := map[string]int{"eliteMonsterKills": 0, "dragonKills": 0, "baronKills": 0, "riftHeraldKills": 0, "atakhanKills": 0, "towerKills": 0, "inhibitorKills": 0, "plates": 0}
	visionStats := map[string]int{"wardsPlaced": 0, "wardsKilled": 0}
	dataCompleteness := map[string]int{"timelineEventsProcessed": 0, "eventSummaryEntries": 0, "knownItemEvents": 0, "unknownItemEvents": 0, "eventsBeforeThinning": 0}

	itemCounts := make(map[string]map[string]int)
	rawEvents := make([]eventEntry, 0, 256)
	skillOrder := make([]int, 0, 18)

	for _, frame := range timeline.Info.Frames {
		for _, e := range frame.Events {
			isRelevant := e.ParticipantId == participantID || e.KillerId == participantID || e.VictimId == participantID || containsInt(e.AssistingParticipantIds, participantID)
			if !isRelevant {
				continue
			}

			dataCompleteness["timelineEventsProcessed"]++
			minute := e.Timestamp / 60000
			ts := formatTimestamp(e.Timestamp)
			phase := phaseFromMinute(minute)

			switch e.Type {
			case "CHAMPION_KILL":
				if e.KillerId == participantID {
					combatStats["kills"]++
					phaseStats[phase]["kills"]++
					if len(e.AssistingParticipantIds) == 0 {
						combatStats["soloKills"]++
					}
					rawEvents = append(rawEvents, eventEntry{minute, fmt.Sprintf(
						"CHAMPION_KILL kill victim=%s assists=%d bounty=%d streak=%d time=%s",
						participantNameByID(match.Info.Participants, e.VictimId),
						len(e.AssistingParticipantIds),
						e.Bounty,
						e.KillStreakLength,
						ts,
					)})
				} else if e.VictimId == participantID {
					combatStats["deaths"]++
					phaseStats[phase]["deaths"]++
					rawEvents = append(rawEvents, eventEntry{minute, fmt.Sprintf(
						"CHAMPION_KILL death killer=%s bounty=%d time=%s",
						participantNameByID(match.Info.Participants, e.KillerId),
						e.Bounty,
						ts,
					)})
				} else if containsInt(e.AssistingParticipantIds, participantID) {
					combatStats["assistsFromKillEvents"]++
					phaseStats[phase]["assists"]++
					rawEvents = append(rawEvents, eventEntry{minute, fmt.Sprintf("CHAMPION_KILL assist time=%s", ts)})
				}
			case "CHAMPION_SPECIAL_KILL":
				combatStats["specialKills"]++
				rawEvents = append(rawEvents, eventEntry{minute, fmt.Sprintf("CHAMPION_SPECIAL_KILL type=%s time=%s", e.KillType, ts)})
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
				rawEvents = append(rawEvents, eventEntry{minute, fmt.Sprintf("%s %s time=%s", e.Type, label, ts)})
			case "SKILL_LEVEL_UP":
				skillOrder = append(skillOrder, e.SkillSlot)
				rawEvents = append(rawEvents, eventEntry{minute, fmt.Sprintf("SKILL_LEVEL_UP skillSlot=%d time=%s", e.SkillSlot, ts)})
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
				rawEvents = append(rawEvents, eventEntry{minute, fmt.Sprintf("ELITE_MONSTER_KILL monsterType=%s monsterSubType=%s time=%s", e.MonsterType, e.MonsterSubType, ts)})
			case "BUILDING_KILL":
				phaseStats[phase]["objectives"]++
				if e.BuildingType == "TOWER_BUILDING" {
					objectiveStats["towerKills"]++
				}
				if e.BuildingType == "INHIBITOR_BUILDING" {
					objectiveStats["inhibitorKills"]++
				}
				rawEvents = append(rawEvents, eventEntry{minute, fmt.Sprintf("BUILDING_KILL buildingType=%s laneType=%s time=%s", e.BuildingType, e.LaneType, ts)})
			case "WARD_PLACED":
				visionStats["wardsPlaced"]++
				phaseStats[phase]["wardsPlaced"]++
				rawEvents = append(rawEvents, eventEntry{minute, fmt.Sprintf("WARD_PLACED wardType=%s time=%s", e.WardType, ts)})
			case "WARD_KILL":
				visionStats["wardsKilled"]++
				phaseStats[phase]["wardsKilled"]++
				rawEvents = append(rawEvents, eventEntry{minute, fmt.Sprintf("WARD_KILL wardType=%s time=%s", e.WardType, ts)})
			case "TURRET_PLATE_DESTROYED":
				objectiveStats["plates"]++
				phaseStats[phase]["objectives"]++
				rawEvents = append(rawEvents, eventEntry{minute, fmt.Sprintf("TURRET_PLATE_DESTROYED laneType=%s time=%s", e.LaneType, ts)})
			default:
				rawEvents = append(rawEvents, eventEntry{minute, fmt.Sprintf("%s time=%s", e.Type, ts)})
			}
		}
	}

	dataCompleteness["eventsBeforeThinning"] = len(rawEvents)
	eventSummary := thinEventSummary(rawEvents, 400, 15)
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
		"participantId":       participant.ParticipantId,
		"teamId":              participant.TeamId,
		"champion":            participant.ChampionName,
		"role":                chooseFirst(participant.TeamPosition, participant.Role),
		"result":              participant.Win,
		"level":               participant.ChampLevel,
		"kills":               participant.Kills,
		"deaths":              participant.Deaths,
		"assists":             participant.Assists,
		"cs":                  cs,
		"goldEarned":          participant.GoldEarned,
		"goldSpent":           participant.GoldSpent,
		"damageToChamps":      participant.TotalDamageDealtToChampions,
		"visionScore":         participant.VisionScore,
		"damageTaken":         participant.TotalDamageTaken,
		"damageSelfMitigated": participant.DamageSelfMitigated,
		"totalHeal":           participant.TotalHeal,
		"healOnTeammates":     participant.TotalHealsOnTeammates,
		"damageToObjectives":  participant.DamageDealtToObjectives,
		"damageToTurrets":     participant.DamageDealtToTurrets,
		"timeCCingOthers":     participant.TimeCCingOthers,
		"firstBloodKill":      participant.FirstBloodKill,
		"firstTowerKill":      participant.FirstTowerKill,
	}

	var opponentSnapshot map[string]any
	if hasOpponent {
		opponentSnapshot = map[string]any{
			"champion": opponent.ChampionName,
			"kills":    opponent.Kills,
			"deaths":   opponent.Deaths,
			"assists":  opponent.Assists,
			"cs":       opponent.TotalMinionsKilled + opponent.NeutralMinionsKilled,
			"gold":     opponent.GoldEarned,
		}
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
		GoldXpTimeline:   buildGoldXpTimeline(timeline.Info.Frames, participantID, opponent, timelineCheckpointMinutes),
		SkillOrder:       skillOrder,
		Runes:            extractRunes(participant),
		Challenges:       curatedChallenges(participant.Challenges),
		OpponentSnapshot: opponentSnapshot,
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
