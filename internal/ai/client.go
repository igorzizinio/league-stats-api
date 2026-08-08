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
	"strings"
)

// StreamChunk represents a chunk of streamed response
type StreamChunk struct {
	Content string `json:"content"`
	Done    bool   `json:"done"`
}

const openRouterURL = "https://openrouter.ai/api/v1/chat/completions"

var httpClient = &http.Client{}

// buildOpenRouterRequest builds the POST request shared by AnalyzeMatch and
// AnalyzeMatchStream. `extra` lets each caller layer on fields the other
// doesn't use (e.g. "reasoning" or "stream") without duplicating the rest of
// the payload/header setup.
func buildOpenRouterRequest(systemPrompt, userPrompt string, extra map[string]any) (*http.Request, error) {
	payload := map[string]any{
		"model": os.Getenv("OPENROUTER_MODEL"),
		"messages": []map[string]any{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": userPrompt},
		},
	}
	for k, v := range extra {
		payload[k] = v
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request payload: %w", err)
	}

	req, err := http.NewRequest("POST", openRouterURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+os.Getenv("OPENROUTER_API_KEY"))
	return req, nil
}

func AnalyzeMatch(shard string, puuid string, matchId string, locale string) (map[string]interface{}, error) {
	systemPrompt, userPrompt, err := prepareMatchData(shard, puuid, matchId, locale)
	if err != nil {
		return nil, err
	}

	req, err := buildOpenRouterRequest(systemPrompt, userPrompt, map[string]any{
		"reasoning": map[string]any{"enabled": true},
	})
	if err != nil {
		return nil, err
	}

	res, err := httpClient.Do(req)
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
	systemPrompt, userPrompt, err := prepareMatchData(shard, puuid, matchId, locale)
	if err != nil {
		return err
	}

	req, err := buildOpenRouterRequest(systemPrompt, userPrompt, map[string]any{
		"stream": true,
	})
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "text/event-stream")

	res, err := httpClient.Do(req)
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
