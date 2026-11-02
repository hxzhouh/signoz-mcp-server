package client

import (
	"encoding/json"
	"errors"
	"fmt"
)

func normalizeLegacyRules(data json.RawMessage) (json.RawMessage, error) {
	var response struct {
		Status string `json:"status"`
		Data   struct {
			Rules json.RawMessage `json:"rules"`
		} `json:"data"`
	}
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, err
	}
	if len(response.Data.Rules) == 0 {
		return nil, errors.New("invalid legacy rules response: missing rules")
	}
	return json.Marshal(map[string]any{"status": response.Status, "data": response.Data.Rules})
}

func normalizeLegacyDashboards(data json.RawMessage, limit, offset int) (json.RawMessage, error) {
	var response struct {
		Data []struct {
			ID     string          `json:"id"`
			Locked bool            `json:"locked"`
			Data   json.RawMessage `json:"data"`
		} `json:"data"`
	}
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("invalid legacy dashboards response: %w", err)
	}
	if response.Data == nil {
		return nil, errors.New("invalid legacy dashboards response: missing data")
	}
	total := len(response.Data)
	if offset < 0 {
		offset = 0
	}
	if offset > total {
		offset = total
	}
	end := total
	if limit > 0 && offset+limit < end {
		end = offset + limit
	}
	items := make([]map[string]any, 0, end-offset)
	for _, dashboard := range response.Data[offset:end] {
		var info struct {
			Title       string `json:"title"`
			Description string `json:"description"`
		}
		if err := json.Unmarshal(dashboard.Data, &info); err != nil {
			return nil, fmt.Errorf("invalid legacy dashboard %q: %w", dashboard.ID, err)
		}
		items = append(items, map[string]any{"id": dashboard.ID, "name": info.Title, "description": info.Description, "locked": dashboard.Locked})
	}
	return json.Marshal(map[string]any{"status": "success", "data": map[string]any{"dashboards": items, "total": total, "format": "legacy-v1"}})
}
