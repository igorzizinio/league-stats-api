package service

import (
	"encoding/json"
	"fmt"
	"legue-stats-api/model"
	"net/http"
	"net/url"
)

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
	versions, _ := GetVersions()
	version := versions[0]

	url := fmt.Sprintf(
		"https://ddragon.leagueoflegends.com/cdn/%s/data/%s/item.json",
		url.QueryEscape(version),
		url.QueryEscape(locale),
	)

	req, err := http.NewRequest("GET", url, nil)
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
