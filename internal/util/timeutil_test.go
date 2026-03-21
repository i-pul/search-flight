package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseRFC3339(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantUTC struct{ hour, min int }
		wantErr bool
	}{
		{
			name:    "WIB +07:00 converts to UTC",
			input:   "2025-12-15T06:00:00+07:00",
			wantUTC: struct{ hour, min int }{23, 0},
		},
		{
			name:    "WITA +08:00 converts to UTC",
			input:   "2025-12-15T08:50:00+08:00",
			wantUTC: struct{ hour, min int }{0, 50},
		},
		{
			name:    "early WIB time converts to UTC",
			input:   "2025-12-15T04:45:00+07:00",
			wantUTC: struct{ hour, min int }{21, 45},
		},
		{
			name:    "invalid string returns error",
			input:   "not-a-date",
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseRFC3339(tc.input)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			utc := got.UTC()
			assert.Equal(t, tc.wantUTC.hour, utc.Hour(), "hour")
			assert.Equal(t, tc.wantUTC.min, utc.Minute(), "minute")
		})
	}
}

func TestParseWithIANA(t *testing.T) {
	tests := []struct {
		name     string
		datetime string
		zone     string
		wantErr  bool
	}{
		{
			name:     "Asia/Jakarta parses successfully",
			datetime: "2025-12-15T05:30:00",
			zone:     "Asia/Jakarta",
		},
		{
			name:     "Asia/Makassar parses successfully",
			datetime: "2025-12-15T08:15:00",
			zone:     "Asia/Makassar",
		},
		{
			name:     "unknown timezone returns error",
			datetime: "2025-12-15T05:30:00",
			zone:     "Invalid/Zone",
			wantErr:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ParseWithIANA(tc.datetime, tc.zone)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}

	t.Run("arrival is after departure across timezones", func(t *testing.T) {
		dept, err := ParseWithIANA("2025-12-15T05:30:00", "Asia/Jakarta")
		require.NoError(t, err)
		arr, err := ParseWithIANA("2025-12-15T08:15:00", "Asia/Makassar")
		require.NoError(t, err)
		assert.True(t, arr.After(dept))
	})
}

func TestParseBatikAirTime(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{name: "+0700 no-colon offset parses successfully", input: "2025-12-15T07:15:00+0700"},
		{name: "+0800 no-colon offset parses successfully", input: "2025-12-15T10:00:00+0800"},
		{name: "invalid string returns error", input: "not-a-date", wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ParseBatikAirTime(tc.input)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}

	t.Run("matches equivalent RFC3339 unix timestamp", func(t *testing.T) {
		batik, err := ParseBatikAirTime("2025-12-15T07:15:00+0700")
		require.NoError(t, err)
		rfc, err := ParseRFC3339("2025-12-15T07:15:00+07:00")
		require.NoError(t, err)
		assert.Equal(t, rfc.Unix(), batik.Unix())
	})
}

func TestParseTravelTime(t *testing.T) {
	tests := []struct {
		input   string
		want    int
		wantErr bool
	}{
		{"1h 45m", 105, false},
		{"2h 0m", 120, false},
		{"3h 5m", 185, false},
		{"45m", 45, false},
		{"2h", 120, false},
		{"not valid", 0, true},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			got, err := ParseTravelTime(tc.input)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		input int
		want  string
	}{
		{0, "0m"},
		{45, "45m"},
		{60, "1h 0m"},
		{105, "1h 45m"},
		{110, "1h 50m"},
		{185, "3h 5m"},
		{230, "3h 50m"},
	}

	for _, tc := range tests {
		t.Run(tc.want, func(t *testing.T) {
			assert.Equal(t, tc.want, FormatDuration(tc.input))
		})
	}
}
