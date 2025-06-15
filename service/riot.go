package service

import (
	"encoding/json"
	"fmt"
	"legue-stats-api/model"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/patrickmn/go-cache"
)

var matchCache = cache.New(24*time.Hour, 1*time.Hour)

func GetChampionRotation(region string) (*map[string]any, error) {
	apiKey := os.Getenv("RIOT_API_KEY")

	url := fmt.Sprintf("https://%s.api.riotgames.com/lol/platform/v3/champion-rotations", region)

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("X-Riot-Token", apiKey)

	client := &http.Client{}

	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var data map[string]any

	json.NewDecoder(resp.Body).Decode(&data)
	return &data, nil
}

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

func GetAccountByPuuid(region string, puuid string) (*model.AccountData, error) {
	apiKey := os.Getenv("RIOT_API_KEY")
	url := fmt.Sprintf("https://%s.api.riotgames.com/riot/account/v1/accounts/by-puuid/%s", region, puuid)
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

func GetSummonerChampionMasteries(region string, puuid string) ([]map[string]interface{}, error) {
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
	var data []map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&data)
	return data, nil
}

func GetMatchlistByPuuid(riotRegion string, puuid string, options *model.GetMatchesOptions) ([]string, error) {
	apiKey := os.Getenv("RIOT_API_KEY")
	baseUrl := fmt.Sprintf("https://%s.api.riotgames.com/lol/match/v5/matches/by-puuid/%s/ids", riotRegion, puuid)

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

func GetMatchById(riotRegion string, matchId string) (*model.MatchData, error) {

	if cached, found := matchCache.Get(matchId); found {
		return cached.(*model.MatchData), nil
	}

	apiKey := os.Getenv("RIOT_API_KEY")

	url := fmt.Sprintf("https://%s.api.riotgames.com/lol/match/v5/matches/%s", riotRegion, matchId)
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

	matchCache.Set(matchId, data, cache.DefaultExpiration)

	return &data, nil
}

func GetTimelineByMatchId(riotRegion string, matchId string) (*model.TimelineData, error) {
	apiKey := os.Getenv("RIOT_API_KEY")

	url := fmt.Sprintf("https://%s.api.riotgames.com/lol/match/v5/matches/%s/timeline", riotRegion, matchId)
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
