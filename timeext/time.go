package timeext

import (
	"math"
	"time"
)

var TimezoneBerlin *time.Location

func init() {
	var err error
	TimezoneBerlin, err = time.LoadLocation("Europe/Berlin")
	if err != nil {
		panic(err)
	}
}

// TimeToDatePart returns a timestamp at the start of the day which contains t (= 00:00:00)
func TimeToDatePart(t time.Time) time.Time {
	t = t.In(TimezoneBerlin)
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// TimeToWeekStart returns a timestamp at the start of the week which contains t (= Monday 00:00:00)
func TimeToWeekStart(t time.Time) time.Time {
	t = TimeToDatePart(t)

	delta := time.Duration(((int64(t.Weekday()) + 6) % 7) * 24 * int64(time.Hour))
	t = t.Add(-1 * delta)

	return t
}

// TimeToMonthStart returns a timestamp at the start of the month which contains t (= yyyy-MM-00 00:00:00)
func TimeToMonthStart(t time.Time) time.Time {
	t = t.In(TimezoneBerlin)
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
}

// TimeToMonthEnd returns a timestamp at the end of the month which contains t (= yyyy-MM-31 23:59:59.999999999)
func TimeToMonthEnd(t time.Time) time.Time {
	return TimeToMonthStart(t).AddDate(0, 1, 0).Add(-1)
}

// TimeToYearStart returns a timestamp at the start of the year which contains t (= yyyy-01-01 00:00:00)
func TimeToYearStart(t time.Time) time.Time {
	t = t.In(TimezoneBerlin)
	return time.Date(t.Year(), 1, 1, 0, 0, 0, 0, t.Location())
}

// TimeToYearEnd returns a timestamp at the end of the month which contains t (= yyyy-12-31 23:59:59.999999999)
func TimeToYearEnd(t time.Time) time.Time {
	return TimeToYearStart(t).AddDate(1, 0, 0).Add(-1)
}

// IsSameDayIncludingDateBoundaries returns true if t1 and t2 are part of the same day (TZ/Berlin), the boundaries of the day are
// inclusive, this means 2021-09-15T00:00:00 is still part of the day 2021-09-14
func IsSameDayIncludingDateBoundaries(t1 time.Time, t2 time.Time) bool {
	dp1 := TimeToDatePart(t1)
	dp2 := TimeToDatePart(t2)

	if dp1.Equal(dp2) {
		return true
	}

	if dp1.AddDate(0, 0, 1).Equal(dp2) && dp2.Equal(t2) {
		return true
	}

	return false
}

// IsDatePartEqual returns true if a and b have the same date part (`yyyy`, `MM` and `dd`)
func IsDatePartEqual(a time.Time, b time.Time) bool {
	yy1, mm1, dd1 := a.In(TimezoneBerlin).Date()
	yy2, mm2, dd2 := b.In(TimezoneBerlin).Date()

	return yy1 == yy2 && mm1 == mm2 && dd1 == dd2
}

// WithTimePart returns a timestamp with the date-part (`yyyy`, `MM`, `dd`) from base
// and the time (`HH`, `mm`, `ss`) from the parameter
func WithTimePart(base time.Time, hour, minute, second int) time.Time {
	datepart := TimeToDatePart(base)

	delta := time.Duration(hour*int(time.Hour) + minute*int(time.Minute) + second*int(time.Second))

	return datepart.Add(delta)
}

// CombineDateAndTime returns a timestamp with the date-part (`yyyy`, `MM`, `dd`) from the d parameter
// and the time (`HH`, `mm`, `ss`) from the t parameter
func CombineDateAndTime(d time.Time, t time.Time) time.Time {
	datepart := TimeToDatePart(d)

	delta := time.Duration(t.Hour()*int(time.Hour) + t.Minute()*int(time.Minute) + t.Second()*int(time.Second) + t.Nanosecond()*int(time.Nanosecond))

	return datepart.Add(delta)
}

// IsSunday returns true if t is a sunday (in TZ/Berlin)
func IsSunday(t time.Time) bool {
	if t.In(TimezoneBerlin).Weekday() == time.Sunday {
		return true
	}
	return false
}

func DurationFromTime(hours int, minutes int, seconds int) time.Duration {
	return time.Duration(hours*int(time.Hour) + minutes*int(time.Minute) + seconds*int(time.Second))
}

func Min(a time.Time, b time.Time) time.Time {
	if a.UnixNano() < b.UnixNano() {
		return a
	} else {
		return b
	}
}

func Max(a time.Time, b time.Time) time.Time {
	if a.UnixNano() > b.UnixNano() {
		return a
	} else {
		return b
	}
}

func UnixFloatSeconds(v float64) time.Time {
	sec, dec := math.Modf(v)
	return time.Unix(int64(sec), int64(dec*(1e9)))
}
