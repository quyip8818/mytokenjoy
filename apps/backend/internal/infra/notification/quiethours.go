package notification

import (
	"time"

	domainnotification "github.com/tokenjoy/backend/internal/domain/notification"
)

// IsInQuietHours checks if the current time (in the user's timezone) falls within quiet hours.
// Returns false if no quiet hours are configured or if the timezone is invalid.
func IsInQuietHours(qh *domainnotification.QuietHours) bool {
	if qh == nil || qh.Start == "" || qh.End == "" {
		return false
	}

	loc, err := time.LoadLocation(qh.Timezone)
	if err != nil {
		return false
	}

	now := time.Now().In(loc)
	currentMinutes := now.Hour()*60 + now.Minute()

	startMinutes := parseTimeMinutes(qh.Start)
	endMinutes := parseTimeMinutes(qh.End)

	if startMinutes < 0 || endMinutes < 0 {
		return false
	}

	// Handle overnight quiet hours (e.g. 22:00 → 08:00)
	if startMinutes <= endMinutes {
		return currentMinutes >= startMinutes && currentMinutes < endMinutes
	}
	// Overnight: quiet if after start OR before end
	return currentMinutes >= startMinutes || currentMinutes < endMinutes
}

// parseTimeMinutes parses "HH:mm" into total minutes since midnight.
// Returns -1 on failure.
func parseTimeMinutes(s string) int {
	if len(s) != 5 || s[2] != ':' {
		return -1
	}
	h := int(s[0]-'0')*10 + int(s[1]-'0')
	m := int(s[3]-'0')*10 + int(s[4]-'0')
	if h < 0 || h > 23 || m < 0 || m > 59 {
		return -1
	}
	return h*60 + m
}
