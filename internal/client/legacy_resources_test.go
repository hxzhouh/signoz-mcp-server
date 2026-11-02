package client

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNormalizeLegacyRules(t *testing.T) {
	result, err := normalizeLegacyRules(json.RawMessage(`{"status":"success","data":{"rules":[{"id":"a"}]}}`))
	require.NoError(t, err)
	require.JSONEq(t, `{"status":"success","data":[{"id":"a"}]}`, string(result))
	_, err = normalizeLegacyRules(json.RawMessage(`{"data":{}}`))
	require.Error(t, err)
}

func TestNormalizeLegacyDashboardsPagination(t *testing.T) {
	data := json.RawMessage(`{"status":"success","data":[{"id":"a","data":{"title":"A"}},{"id":"b","locked":true,"data":{"title":"B"}}]}`)
	result, err := normalizeLegacyDashboards(data, 1, 1)
	require.NoError(t, err)
	require.JSONEq(t, `{"status":"success","data":{"dashboards":[{"id":"b","name":"B","description":"","locked":true}],"total":2,"format":"legacy-v1"}}`, string(result))
	result, err = normalizeLegacyDashboards(data, 1, 5)
	require.NoError(t, err)
	require.Contains(t, string(result), `"dashboards":[]`)
}
