package utils

import "time"

func GetStartOfDay(d time.Time) time.Time {
	year, month, day := d.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, d.Location())
}

func GetEndOfDay(d time.Time) time.Time {
	year, month, day := d.Date()
	return time.Date(year, month, day, 23, 59, 59, int(time.Second-time.Nanosecond), d.Location())
}
