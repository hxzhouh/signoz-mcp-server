package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	logpkg "github.com/SigNoz/signoz-mcp-server/pkg/log"
	"github.com/stretchr/testify/require"
)

func TestLegacyMetricsFallback(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v2/metrics":
			_, _ = w.Write([]byte("<!doctype html><html></html>"))
		case "/api/v1/metrics":
			require.Equal(t, http.MethodPost, r.Method)
			var request map[string]any
			require.NoError(t, json.NewDecoder(r.Body).Decode(&request))
			require.EqualValues(t, 1000, request["start"])
			require.Contains(t, request["filters"], "items")
			_, _ = w.Write([]byte(`{"status":"success","data":{"metrics":[{"metric_name":"requests","type":"sum"}],"total":1}}`))
		case "/api/v1/metrics/requests/metadata":
			_, _ = w.Write([]byte(`{"status":"success","data":{"name":"requests","metadata":{"metric_type":"sum","temporality":"Cumulative","monotonic":true}}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	c := NewClient(logpkg.New("error"), server.URL, "test", "SIGNOZ-API-KEY", nil)
	rows, err := c.ListMetrics(context.Background(), 1000, 2000, 10, "requests", "")
	require.NoError(t, err)
	require.Contains(t, string(rows), `"metricName":"requests"`)
	meta, err := c.ListMetrics(context.Background(), 0, 0, 10, "requests", "")
	require.NoError(t, err)
	require.Contains(t, string(meta), `"isMonotonic":true`)
	require.Contains(t, string(meta), `"temporality":"cumulative"`)
}

func TestLegacyTopMetricsFallback(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v2/metrics/treemap" {
			_, _ = w.Write([]byte("<html>app</html>"))
			return
		}
		require.Equal(t, "/api/v1/metrics/treemap", r.URL.Path)
		_, _ = w.Write([]byte(`{"status":"success","data":{"samples":[{"metric_name":"requests","total_value":20}],"timeseries":null}}`))
	}))
	defer server.Close()
	c := NewClient(logpkg.New("error"), server.URL, "test", "SIGNOZ-API-KEY", nil)
	data, err := c.GetTopMetrics(context.Background(), 1000, 2000, 10)
	require.NoError(t, err)
	require.Contains(t, string(data), `"metricName":"requests"`)
	require.Contains(t, string(data), `"totalValue":20`)
}
