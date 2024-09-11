package tmutil

import "time"

// CurrDate returns current date string(YYYY-MM-DD)
func CurrDate() string {
	return time.Now().Format("2006-01-02")
}

// CurrMonth returns current month string(YYYY-MM-01)
func CurrMonth() string {
	return time.Now().Format("2006-01-01")[:8] + "01"
}

// CurrYear returns current year string(YYYY-01-01)
func CurrYear() string {
	return time.Now().Format("2006-01-02")[:4] + "-01-01"
}

// CurrHour returns current hour string(HH)
func CurrHour() string {
	return time.Now().Format("15")
}

// CurrWeek returns the start date of the current week (Sunday)
func CurrWeek() string {
	week := time.Now().AddDate(0, 0, -1)

	for week.Weekday() != time.Sunday {
		week = week.AddDate(0, 0, -1)
	}

	return week.Format("2006-01-02")
}

func CurrWeek2() string {
	week := time.Now().AddDate(0, 0, -1)

	for week.Weekday() != time.Sunday {
		week = week.AddDate(0, 0, -1)
	}
	week = week.AddDate(0, 0, 1)
	return week.Format("2006-01-02")
}
