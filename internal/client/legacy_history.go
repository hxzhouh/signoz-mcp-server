package client

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/SigNoz/signoz-mcp-server/pkg/types"
)

const legacyHistoryPrefix = "legacy-v1:"

type legacyHistoryCursor struct {
	RuleID string `json:"ruleId"`
	Start  int64  `json:"start"`
	End    int64  `json:"end"`
	State  string `json:"state"`
	Order  string `json:"order"`
	Limit  int    `json:"limit"`
	Offset int    `json:"offset"`
}

func (s *SigNoz) getLegacyAlertHistory(ctx context.Context, ruleID string, req types.AlertHistoryRequest) (json.RawMessage, error) {
	if req.FilterExpression != "" {
		return nil, errors.New("legacy alert history does not support filter expressions")
	}
	cursor := legacyHistoryCursor{RuleID: ruleID, Start: req.Start, End: req.End, State: req.State, Order: req.Order, Limit: req.Limit}
	if cursor.Limit <= 0 {
		cursor.Limit = 20
	}
	if req.Cursor != "" {
		if !strings.HasPrefix(req.Cursor, legacyHistoryPrefix) {
			return nil, errors.New("invalid legacy alert history cursor")
		}
		decoded, err := base64.RawURLEncoding.DecodeString(strings.TrimPrefix(req.Cursor, legacyHistoryPrefix))
		if err != nil || json.Unmarshal(decoded, &cursor) != nil || cursor.RuleID != ruleID || cursor.Start <= 0 || cursor.End <= cursor.Start || cursor.Limit <= 0 || cursor.Limit > 10000 || cursor.Offset < 0 {
			return nil, errors.New("invalid legacy alert history cursor")
		}
	}
	body, err := json.Marshal(map[string]any{"start": cursor.Start, "end": cursor.End, "state": cursor.State, "order": cursor.Order, "limit": cursor.Limit, "offset": cursor.Offset})
	if err != nil {
		return nil, err
	}
	result, err := s.doReplaySafePost(ctx, s.baseURL+"/api/v1/rules/"+url.PathEscape(ruleID)+"/history/timeline", body, DefaultQueryTimeout)
	if err != nil {
		return nil, err
	}
	var response struct {
		Status string                     `json:"status"`
		Data   map[string]json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(result, &response); err != nil {
		return nil, fmt.Errorf("invalid legacy alert history response: %w", err)
	}
	var items []json.RawMessage
	var total int
	if response.Data == nil || json.Unmarshal(response.Data["items"], &items) != nil || json.Unmarshal(response.Data["total"], &total) != nil {
		return nil, errors.New("invalid legacy alert history response: missing items or total")
	}
	if cursor.Offset+len(items) < total && len(items) > 0 {
		cursor.Offset += len(items)
		encoded, err := json.Marshal(cursor)
		if err != nil {
			return nil, err
		}
		next, err := json.Marshal(legacyHistoryPrefix + base64.RawURLEncoding.EncodeToString(encoded))
		if err != nil {
			return nil, err
		}
		response.Data["nextCursor"] = next
	}
	response.Data["compatibility"] = json.RawMessage(`"legacy-v1-offset-pagination"`)
	return json.Marshal(response)
}
