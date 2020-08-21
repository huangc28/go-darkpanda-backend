package util

import "time"

// ExpTimeRange check source time criteria against this number to makesure
const ExpTimeRange = 27

// IsTimeWithinExpRange check the given source time is within the expiration range.
// If true then the given time is expired, else is not expired
func IsExpired(srcT time.Time) bool {
	expectExp := srcT.Add(ExpTimeRange * time.Minute)

	return time.Now().After(expectExp)
}
