package timex_test

import (
	"testing"
	"time"

	. "code-platform/pkg/timex"

	"github.com/stretchr/testify/require"
)

func TestStartOfYearInLocation(t *testing.T) {
	for _, c := range []struct {
		date         time.Time
		location     *time.Location
		label        string
		expectedYear int
	}{
		{label: "normal", date: time.Date(2022, 5, 1, 1, 1, 1, 1, ShanghaiLocation), location: ShanghaiLocation, expectedYear: 2022},
		{label: "utc", date: time.Date(2022, 1, 1, 1, 1, 1, 1, time.UTC), location: time.UTC, expectedYear: 2022},
		{label: "last year", date: time.Date(2022, 1, 1, 1, 1, 1, 1, ShanghaiLocation).In(time.UTC), location: time.UTC, expectedYear: 2021},
		{label: "next year", date: time.Date(2021, 12, 31, 19, 1, 1, 1, time.UTC).In(ShanghaiLocation), location: ShanghaiLocation, expectedYear: 2022},
	} {
		startOfYear := StartOfYearInLocation(c.date, c.location)
		require.Equal(t, c.expectedYear, startOfYear.Year(), c.label)
	}
}

func TestEndOfYearInLocation(t *testing.T) {
	for _, c := range []struct {
		date         time.Time
		location     *time.Location
		label        string
		expectedYear int
	}{
		{label: "normal", date: time.Date(2022, 5, 1, 1, 1, 1, 1, ShanghaiLocation), location: ShanghaiLocation, expectedYear: 2022},
		{label: "utc", date: time.Date(2022, 1, 1, 1, 1, 1, 1, time.UTC), location: time.UTC, expectedYear: 2022},
		{label: "last year", date: time.Date(2022, 1, 1, 1, 1, 1, 1, ShanghaiLocation).In(time.UTC), location: time.UTC, expectedYear: 2021},
		{label: "next year", date: time.Date(2021, 12, 31, 19, 1, 1, 1, time.UTC).In(ShanghaiLocation), location: ShanghaiLocation, expectedYear: 2022},
	} {
		endOfYear := EndOfYearInLocation(c.date, c.location)
		require.Equal(t, c.expectedYear, endOfYear.Year(), c.label)
	}
}

func TestStartOfDayInLocation(t *testing.T) {
	for _, c := range []struct {
		date        time.Time
		location    *time.Location
		label       string
		expectedDay int
	}{
		{label: "normal", date: time.Date(2022, 5, 1, 1, 1, 1, 1, ShanghaiLocation), location: ShanghaiLocation, expectedDay: 1},
		{label: "utc", date: time.Date(2022, 1, 1, 1, 1, 1, 1, time.UTC), location: time.UTC, expectedDay: 1},
		{label: "yesterday", date: time.Date(2022, 5, 1, 1, 1, 1, 1, ShanghaiLocation).In(time.UTC), location: time.UTC, expectedDay: 30},
		{label: "tomorrow", date: time.Date(2022, 4, 30, 19, 1, 1, 1, time.UTC).In(ShanghaiLocation), location: ShanghaiLocation, expectedDay: 1},
	} {
		startOfDay := StartOfDayInLocation(c.date, c.location)
		require.Equal(t, c.expectedDay, startOfDay.Day(), c.label)
	}
}

func TestEndOfDayInLocation(t *testing.T) {
	for _, c := range []struct {
		date        time.Time
		location    *time.Location
		label       string
		expectedDay int
	}{
		{label: "normal", date: time.Date(2022, 5, 1, 1, 1, 1, 1, ShanghaiLocation), location: ShanghaiLocation, expectedDay: 1},
		{label: "utc", date: time.Date(2022, 1, 1, 1, 1, 1, 1, time.UTC), location: time.UTC, expectedDay: 1},
		{label: "yesterday", date: time.Date(2022, 5, 1, 1, 1, 1, 1, ShanghaiLocation).In(time.UTC), location: time.UTC, expectedDay: 30},
		{label: "tomorrow", date: time.Date(2022, 4, 30, 19, 1, 1, 1, time.UTC).In(ShanghaiLocation), location: ShanghaiLocation, expectedDay: 1},
	} {
		endOfDay := EndOfDayInLocation(c.date, c.location)
		require.Equal(t, c.expectedDay, endOfDay.Day(), c.label)
	}
}
