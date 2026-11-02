package client

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"testing"

	"github.com/SigNoz/signoz-mcp-server/pkg/types"
	"github.com/stretchr/testify/require"
)

func TestPrepareTraceFieldsSelectsOnlyDiscoveredAttributes(t *testing.T) {
	c := &MockClient{GetFieldKeysFn: func(context.Context, string, string, string, string, string, string) (json.RawMessage, error) {
		return json.RawMessage(`{"data":{"complete":true,"keys":{"service.name":[{"fieldContext":"resource"}],"db.statement":[{"fieldContext":"attribute"}],"host.name":[{"fieldContext":"resource"}]}}}`), nil
	}}
	payload := types.BuildTracesQueryPayload(1, 2, "", 10, 0)
	require.NoError(t, PrepareTraceFields(context.Background(), c, payload, slog.New(slog.NewTextHandler(io.Discard, nil))))
	spec := payload.CompositeQuery.Queries[0].Spec.(types.QuerySpec)
	fields := map[string]bool{}
	for _, field := range spec.SelectFields {
		fields[field.Name] = true
	}
	require.True(t, fields["trace_id"])
	require.True(t, fields["service.name"])
	require.True(t, fields["db.statement"])
	require.True(t, fields["host.name"])
	require.False(t, fields["cloud.account.id"])
	require.False(t, fields["http.request.method"])
}

func TestPrepareTraceFieldsFallbackAndAuthorization(t *testing.T) {
	for _, tc := range []struct {
		name    string
		err     error
		wantErr bool
	}{
		{"unavailable metadata", errors.New("unavailable"), false},
		{"forbidden", &HTTPStatusError{StatusCode: http.StatusForbidden}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := &MockClient{GetFieldKeysFn: func(context.Context, string, string, string, string, string, string) (json.RawMessage, error) {
				return nil, tc.err
			}}
			payload := types.BuildTracesQueryPayload(1, 2, "", 10, 0)
			err := PrepareTraceFields(context.Background(), c, payload, slog.New(slog.NewTextHandler(io.Discard, nil)))
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			for _, field := range payload.CompositeQuery.Queries[0].Spec.(types.QuerySpec).SelectFields {
				require.Equal(t, "span", field.FieldContext)
			}
		})
	}
}
