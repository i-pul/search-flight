package flight_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func writeJSON(t *testing.T, v any) string {
	t.Helper()
	b, err := json.Marshal(v)
	require.NoError(t, err)
	return string(b)
}

func decodeError(t *testing.T, body []byte) map[string]any {
	t.Helper()
	var m map[string]any
	require.NoError(t, json.Unmarshal(body, &m))
	return m
}

func validBody(t *testing.T) string {
	t.Helper()
	return writeJSON(t, map[string]any{
		"origin":        "CGK",
		"destination":   "DPS",
		"departureDate": "2025-12-15",
		"passengers":    1,
		"cabinClass":    "economy",
	})
}
