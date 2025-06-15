package service

import (
	"encoding/json"
	"fmt"
	"legue-stats-api/model"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

func GetSummonerByPuuid(region string, puuid *string) (*model.SummonerData, error) {
	apiKey := os.Getenv("RIOT_API_KEY")

	url := fmt.Sprintf("https://%s.api.riotgames.com/lol/summoner/v4/summoners/by-puuid/%s", region, *puuid)

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("X-Riot-Token", apiKey)

	client := &http.Client{}

	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var data model.SummonerData
	json.NewDecoder(resp.Body).Decode(&data)

	return &data, nil
}

func GetAccountByRiotId(region string, gameName string, tagLine string) (*model.AccountData, error) {
	apiKey := os.Getenv("RIOT_API_KEY")

	url := fmt.Sprintf("https://%s.api.riotgames.com/riot/account/v1/accounts/by-riot-id/%s/%s", region, gameName, tagLine)

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("X-Riot-Token", apiKey)

	client := &http.Client{}

	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var data model.AccountData
	json.NewDecoder(resp.Body).Decode(&data)

	return &data, nil
}

func GetSummonerLeagueByPuuid(region string, puuid string) ([]map[string]interface{}, error) {
	apiKey := os.Getenv("RIOT_API_KEY")
	url := fmt.Sprintf("https://%s.api.riotgames.com/lol/league/v4/entries/by-puuid/%s", region, puuid)

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("X-Riot-Token", apiKey)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var data []map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&data)
	return data, nil
}

func GetSummonerChampionMastery(region string, puuid string) (map[string]interface{}, error) {
	apiKey := os.Getenv("RIOT_API_KEY")
	url := fmt.Sprintf("https://%s.api.riotgames.com/lol/champion-mastery/v4/champion-masteries/by-puuid/%s", region, puuid)

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("X-Riot-Token", apiKey)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var data map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&data)
	return data, nil
}

func GetMatchlistByPuuid(shard string, puuid string, options *model.GetMatchesOptions) ([]string, error) {
	apiKey := os.Getenv("RIOT_API_KEY")
	baseUrl := fmt.Sprintf("https://%s.api.riotgames.com/lol/match/v5/matches/by-puuid/%s/ids", shard, puuid)

	params := url.Values{}

	if options.StartIndex != nil {
		params.Add("start", strconv.Itoa(*options.StartIndex))
	}
	if options.Count != nil {
		params.Add("count", strconv.Itoa(*options.Count))
	}
	if options.Queue != nil {
		params.Add("queue", strconv.Itoa(*options.Queue))
	}
	if options.StartTime != nil {
		params.Add("startTime", strconv.FormatInt(*options.StartTime, 10))
	}
	if options.EndTime != nil {
		params.Add("endTime", strconv.FormatInt(*options.EndTime, 10))
	}
	if options.Type != nil {
		params.Add("type", *options.Type)
	}

	finalURL := baseUrl + "?" + params.Encode()

	req, _ := http.NewRequest("GET", finalURL, nil)
	req.Header.Set("X-Riot-Token", apiKey)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var data []string
	json.NewDecoder(resp.Body).Decode(&data)
	return data, nil
}

func GetMatchById(shard string, matchId string) (*model.MatchData, error) {
	apiKey := os.Getenv("RIOT_API_KEY")

	url := fmt.Sprintf("https://%s.api.riotgames.com/lol/match/v5/matches/%s", shard, matchId)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("X-Riot-Token", apiKey)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var data model.MatchData
	json.NewDecoder(resp.Body).Decode(&data)
	return &data, nil
}

func GetTimelineByMatchId(shard string, matchId string) (*model.TimelineData, error) {
	apiKey := os.Getenv("RIOT_API_KEY")

	url := fmt.Sprintf("https://%s.api.riotgames.com/lol/match/v5/matches/%s/timeline", shard, matchId)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("X-Riot-Token", apiKey)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var data model.TimelineData
	json.NewDecoder(resp.Body).Decode(&data)
	return &data, nil
}
