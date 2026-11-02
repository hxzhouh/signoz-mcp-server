package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func isMissingAPI(err error) bool {
	var status *HTTPStatusError
	return errors.Is(err, ErrUnsupportedAPI) || (errors.As(err, &status) && status.StatusCode == http.StatusNotFound)
}

func (s *SigNoz) listMetricsV1(ctx context.Context, start, end int64, limit int, searchText, source string) (json.RawMessage, error) {
	if source != "" {
		return nil, fmt.Errorf("legacy SigNoz metrics API does not support source %q", source)
	}
	if start == 0 && end == 0 && searchText != "" {
		result, err := s.doRequest(ctx, http.MethodGet, s.baseURL+"/api/v1/metrics/"+url.PathEscape(searchText)+"/metadata", nil, DefaultQueryTimeout)
		if err == nil {
			var response struct {
				Data struct {
					Name     string `json:"name"`
					Type     string `json:"type"`
					Metadata struct {
						MetricType  string  `json:"metric_type"`
						Temporality *string `json:"temporality"`
						Monotonic   *bool   `json:"monotonic"`
					} `json:"metadata"`
				} `json:"data"`
			}
			if json.Unmarshal(result, &response) == nil && response.Data.Name == searchText {
				metricType := response.Data.Metadata.MetricType
				if metricType == "" {
					metricType = response.Data.Type
				}
				row := map[string]any{"metricName": response.Data.Name, "type": metricType}
				if response.Data.Metadata.Temporality != nil {
					row["temporality"] = strings.ToLower(*response.Data.Metadata.Temporality)
				}
				if response.Data.Metadata.Monotonic != nil {
					row["isMonotonic"] = *response.Data.Metadata.Monotonic
				}
				return json.Marshal(map[string]any{"status": "success", "data": map[string]any{"metrics": []any{row}}})
			}
		} else if !isMissingAPI(err) {
			return nil, err
		}
	}
	if end <= 0 {
		end = time.Now().UnixMilli()
	}
	if start <= 0 {
		start = end - int64(time.Hour/time.Millisecond)
	}
	if limit <= 0 {
		limit = 50
	}
	items := []any{}
	if searchText != "" {
		items = append(items, map[string]any{"key": map[string]string{"key": "metric_name", "type": "tag"}, "op": "contains", "value": searchText})
	}
	body, err := json.Marshal(map[string]any{"start": start, "end": end, "limit": limit, "offset": 0, "filters": map[string]any{"op": "AND", "items": items}})
	if err != nil {
		return nil, err
	}
	result, err := s.doReplaySafePost(ctx, s.baseURL+"/api/v1/metrics", body, DefaultQueryTimeout)
	if err != nil {
		return nil, err
	}
	return normalizeLegacyMetrics(result)
}

func normalizeLegacyMetrics(data json.RawMessage) (json.RawMessage, error) {
	var response struct {
		Status string                     `json:"status"`
		Data   map[string]json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("invalid legacy metrics response: %w", err)
	}
	if response.Data == nil {
		return nil, errors.New("invalid legacy metrics response: missing data")
	}
	for _, key := range []string{"metrics", "samples", "timeseries"} {
		if len(response.Data[key]) == 0 || string(response.Data[key]) == "null" {
			continue
		}
		var rows []map[string]json.RawMessage
		if err := json.Unmarshal(response.Data[key], &rows); err != nil {
			return nil, err
		}
		for _, row := range rows {
			if v, ok := row["metric_name"]; ok {
				row["metricName"] = v
			}
			if v, ok := row["total_value"]; ok {
				row["totalValue"] = v
			}
		}
		encoded, err := json.Marshal(rows)
		if err != nil {
			return nil, err
		}
		response.Data[key] = encoded
	}
	return json.Marshal(response)
}
