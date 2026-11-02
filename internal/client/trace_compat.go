package client

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/SigNoz/signoz-mcp-server/pkg/types"
)

// PrepareTraceFields removes optional attributes absent from the workspace.
// Some SigNoz versions return 500 instead of null for unknown select fields.
func PrepareTraceFields(ctx context.Context, c Client, payload *types.QueryPayload, logger *slog.Logger) error {
	data, err := c.GetFieldKeys(ctx, "traces", "", "", "", "", "")
	if err != nil {
		var statusErr *HTTPStatusError
		if errors.Is(err, ErrUnauthorized) || (errors.As(err, &statusErr) && (statusErr.StatusCode == http.StatusUnauthorized || statusErr.StatusCode == http.StatusForbidden)) {
			return err
		}
	}
	var response struct {
		Data struct {
			Keys map[string][]struct {
				FieldContext string `json:"fieldContext"`
			} `json:"keys"`
			Complete bool `json:"complete"`
		} `json:"data"`
	}
	if err == nil {
		err = json.Unmarshal(data, &response)
		if err == nil && response.Data.Keys == nil {
			err = errors.New("trace field discovery returned no keys")
		}
	}
	if err != nil || !response.Data.Complete {
		logger.WarnContext(ctx, "trace field discovery unavailable or incomplete; optional fields may be omitted")
	}
	for i := range payload.CompositeQuery.Queries {
		spec, ok := payload.CompositeQuery.Queries[i].Spec.(types.QuerySpec)
		if !ok {
			continue
		}
		selected := make([]types.SelectField, 0, len(spec.SelectFields))
		for _, field := range spec.SelectFields {
			if field.FieldContext == "span" {
				selected = append(selected, field)
				continue
			}
			for _, item := range response.Data.Keys[field.Name] {
				if item.FieldContext == field.FieldContext || (field.FieldContext == "tag" && item.FieldContext == "attribute") {
					selected = append(selected, field)
					break
				}
			}
		}
		spec.SelectFields = selected
		payload.CompositeQuery.Queries[i].Spec = spec
	}
	return nil
}
