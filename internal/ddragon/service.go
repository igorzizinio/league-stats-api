package ddragon

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/igorzizinio/league-stats-api/internal/model"
	"github.com/patrickmn/go-cache"
)

var cacheItems = cache.New(24*time.Hour, 1*time.Hour)
var cacheFullItems = cache.New(24*time.Hour, 1*time.Hour)
var cacheChampions = cache.New(24*time.Hour, 1*time.Hour)

func LoadStaticItemsData() map[string]model.ItemData {
	items := FetchItems("en_US")
	cacheItems.Set("items", items, cache.DefaultExpiration)

	return items
}

func GetVersions() ([]string, error) {

	url := "https://ddragon.leagueoflegends.com/api/versions.json"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	var versions []string
	if err := json.NewDecoder(resp.Body).Decode(&versions); err != nil {
		return nil, err
	}

	if len(versions) == 0 {
		return nil, fmt.Errorf("no versions found")
	}

	return versions, nil
}

func GetItems(locale string) map[string]model.ItemData {
	cacheKey := "items_" + locale
	if data, found := cacheItems.Get(cacheKey); found {
		items := data.(map[string]model.ItemData)
		return items
	}
	items := FetchItems(locale)
	cacheItems.Set(cacheKey, items, cache.DefaultExpiration)
	return items
}

func FetchItems(locale string) map[string]model.ItemData {
	versions, err := GetVersions()
	if err != nil || len(versions) == 0 {
		fmt.Println("Error fetching ddragon versions:", err)
		return nil
	}
	version := versions[0]

	urlStr := fmt.Sprintf(
		"https://ddragon.leagueoflegends.com/cdn/%s/data/%s/item.json",
		version,
		locale,
	)

	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return nil
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error making request:", err)
		return nil
	}
	defer resp.Body.Close()

	var data model.DDragonItemData
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		fmt.Println("Error decoding response:", err)
		return nil
	}

	return data.Data
}

// GetFullItems returns the raw items data (with full fields) cached per-locale
func GetFullItems(locale string) map[string]any {
	cacheKey := "full_items_" + locale
	if data, found := cacheFullItems.Get(cacheKey); found {
		return data.(map[string]any)
	}
	items := FetchFullItems(locale)
	cacheFullItems.Set(cacheKey, items, cache.DefaultExpiration)
	return items
}

// FetchFullItems fetches the full item json and returns the raw data.data map
func FetchFullItems(locale string) map[string]any {
	versions, err := GetVersions()
	if err != nil || len(versions) == 0 {
		fmt.Println("Error fetching ddragon versions:", err)
		return nil
	}
	version := versions[0]

	urlStr := fmt.Sprintf(
		"https://ddragon.leagueoflegends.com/cdn/%s/data/%s/item.json",
		version,
		locale,
	)

	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return nil
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error making request:", err)
		return nil
	}
	defer resp.Body.Close()

	var data struct {
		Data map[string]any `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		fmt.Println("Error decoding response:", err)
		return nil
	}

	return data.Data
}

// GetChampionData returns cached champion data for a champion name and locale
func GetChampionData(champion string, locale string) map[string]any {
	cacheKey := fmt.Sprintf("champion_%s_%s", strings.ToLower(champion), locale)
	if data, found := cacheChampions.Get(cacheKey); found {
		return data.(map[string]any)
	}
	ch := FetchChampionData(champion, locale)
	if ch != nil {
		cacheChampions.Set(cacheKey, ch, cache.DefaultExpiration)
	}
	return ch
}

// FetchChampionData fetches champion JSON from DDragon and returns the champion object
func FetchChampionData(champion string, locale string) map[string]any {
	versions, err := GetVersions()
	if err != nil || len(versions) == 0 {
		fmt.Println("Error fetching ddragon versions:", err)
		return nil
	}
	version := versions[0]

	urlStr := fmt.Sprintf(
		"https://ddragon.leagueoflegends.com/cdn/%s/data/%s/champion/%s.json",
		version,
		locale,
		champion,
	)

	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return nil
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error making request:", err)
		return nil
	}
	defer resp.Body.Close()

	var data struct {
		Data map[string]any `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		fmt.Println("Error decoding champion response:", err)
		return nil
	}

	// try exact key, then case-insensitive match
	if ch, ok := data.Data[champion]; ok {
		chMap := ch.(map[string]any)
		return chMap
	}
	lower := strings.ToLower(champion)
	for k, v := range data.Data {
		if strings.ToLower(k) == lower {
			if m, ok := v.(map[string]any); ok {
				return m
			}
		}
	}

	// fallback: return the first entry
	for _, v := range data.Data {
		if m, ok := v.(map[string]any); ok {
			return m
		}
	}

	return nil
}
