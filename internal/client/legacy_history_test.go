package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	logpkg "github.com/SigNoz/signoz-mcp-server/pkg/log"
	"github.com/SigNoz/signoz-mcp-server/pkg/types"
	"github.com/stretchr/testify/require"
)

func TestLegacyAlertHistoryPagination(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			_, _ = w.Write([]byte("<html>app</html>"))
			return
		}
		var body map[string]any
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		require.Equal(t, "/api/v1/rules/rule/history/timeline", r.URL.Path)
		if body["offset"] == float64(0) {
			_, _ = w.Write([]byte(`{"status":"success","data":{"items":[{"state":"firing"}],"total":2,"labels":[]}}`))
			return
		}
		require.EqualValues(t, 1, body["offset"])
		_, _ = w.Write([]byte(`{"status":"success","data":{"items":[{"state":"inactive"}],"total":2,"labels":[]}}`))
	}))
	defer server.Close()
	c := NewClient(logpkg.New("error"), server.URL, "test", "SIGNOZ-API-KEY", nil)
	req := types.AlertHistoryRequest{Start: 1000, End: 2000, Limit: 1, Order: "desc"}
	first, err := c.GetAlertHistory(context.Background(), "rule", req)
	require.NoError(t, err)
	var decoded struct {
		Data struct {
			Cursor string `json:"nextCursor"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(first, &decoded))
	require.NotEmpty(t, decoded.Data.Cursor)
	req.Cursor = decoded.Data.Cursor
	second, err := c.GetAlertHistory(context.Background(), "rule", req)
	require.NoError(t, err)
	require.Contains(t, string(second), "inactive")
	require.NotContains(t, string(second), "nextCursor")
	_, err = c.GetAlertHistory(context.Background(), "other", req)
	require.Error(t, err)
}
