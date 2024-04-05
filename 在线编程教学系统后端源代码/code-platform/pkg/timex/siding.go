package timex

import "time"

func StartOfYearInLocation(t time.Time, location *time.Location) time.Time {
	t = t.In(location)
	return time.Date(t.Year(), 1, 1, 0, 0, 0, 0, location)
}

func StartOfYear(t time.Time) time.Time {
	return StartOfYearInLocation(t, ShanghaiLocation)
}

func EndOfYearInLocation(t time.Time, location *time.Location) time.Time {
	t = t.In(location)
	return time.Date(t.Year()+1, 1, 1, 0, 0, 0, 0, location).Add(-time.Nanosecond)
}

func EndOfYear(t time.Time) time.Time {
	return EndOfYearInLocation(t, ShanghaiLocation)
}

func StartOfDayInLocation(t time.Time, location *time.Location) time.Time {
	t = t.In(location)
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, location)
}

func StartOfDay(t time.Time) time.Time {
	return StartOfDayInLocation(t, ShanghaiLocation)
}

func EndOfDayInLocation(t time.Time, location *time.Location) time.Time {
	t = t.In(location)
	return StartOfDayInLocation(t.AddDate(0, 0, 1), location).Add(-time.Nanosecond)
}

func EndOfDay(t time.Time) time.Time {
	return EndOfDayInLocation(t, ShanghaiLocation)
}
