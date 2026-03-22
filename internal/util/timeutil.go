package util

import (
	"fmt"
	"strings"
	"time"
)

// ParseRFC3339 parses a standard RFC3339 time string with timezone offset (e.g. "+07:00").
// Used by Garuda and AirAsia providers.
func ParseRFC3339(s string) (time.Time, error) {
	return time.Parse(time.RFC3339, s)
}

// ParseWithIANA parses a datetime string without timezone combined with a separate
// IANA timezone name. Used by Lion Air (e.g. "Asia/Jakarta", "Asia/Makassar").
func ParseWithIANA(datetime, ianaZone string) (time.Time, error) {
	loc, err := time.LoadLocation(ianaZone)
	if err != nil {
		return time.Time{}, fmt.Errorf("unknown timezone %q: %w", ianaZone, err)
	}
	return time.ParseInLocation("2006-01-02T15:04:05", datetime, loc)
}

// ParseBatikAirTime parses Batik Air's datetime format which uses "+0700" (no colon).
// Go's reference layout uses "-0700" for zone without colon.
func ParseBatikAirTime(s string) (time.Time, error) {
	return time.Parse("2006-01-02T15:04:05-0700", s)
}

// ParseTravelTime converts a string like "1h 45m" or "3h 5m" to total minutes.
// Used by Batik Air's travelTime field.
func ParseTravelTime(s string) (int, error) {
	s = strings.TrimSpace(s)
	var h, m int

	// "Xh Ym" format
	if n, _ := fmt.Sscanf(s, "%dh %dm", &h, &m); n == 2 {
		return h*60 + m, nil
	}

	// "Xh" format (hours only, string must end with "h")
	if strings.HasSuffix(s, "h") {
		var hours int
		if n, _ := fmt.Sscanf(s, "%dh", &hours); n == 1 {
			return hours * 60, nil
		}
	}

	// "Xm" format (minutes only, no "h" in string)
	if strings.HasSuffix(s, "m") && !strings.Contains(s, "h") {
		var mins int
		if n, _ := fmt.Sscanf(s, "%dm", &mins); n == 1 {
			return mins, nil
		}
	}

	return 0, fmt.Errorf("cannot parse travel time %q", s)
}

// FormatDuration converts total minutes to human-readable string like "1h 45m".
func FormatDuration(totalMinutes int) string {
	if totalMinutes <= 0 {
		return "0m"
	}
	if totalMinutes < 60 {
		return fmt.Sprintf("%dm", totalMinutes)
	}
	return fmt.Sprintf("%dh %dm", totalMinutes/60, totalMinutes%60)
}
