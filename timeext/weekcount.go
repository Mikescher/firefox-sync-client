package timeext

import (
	"time"
)

var isoWeekCountAggregate map[int]int

func init() {
	isoWeekCountAggregate = make(map[int]int)
	for y := 1900; y <= time.Now().Year(); y++ {
		GetAggregateIsoWeekCount(y)
	}
}

func GetAggregateIsoWeekCount(year int) int {
	if v, ok := isoWeekCountAggregate[year]; ok {
		return v
	}

	if year == 1900 {
		isoWeekCountAggregate[year] = 0
		return 0
	}

	if year < 1900 {
		s := 0
		for yy := year; yy < 1900; yy++ {
			s += GetIsoWeekCount(yy)
		}
		w := -s
		isoWeekCountAggregate[year] = w
		return w
	}

	w := GetIsoWeekCount(year)

	w += GetAggregateIsoWeekCount(year - 1)

	isoWeekCountAggregate[year] = w

	return w
}

func GetIsoWeekCount(year int) int {
	_, w1 := time.Date(year+0, 12, 27, 0, 0, 0, 0, TimezoneBerlin).ISOWeek()
	_, w2 := time.Date(year+0, 12, 31, 0, 0, 0, 0, TimezoneBerlin).ISOWeek()
	_, w3 := time.Date(year+1, 1, 4, 0, 0, 0, 0, TimezoneBerlin).ISOWeek()

	w1 -= 1
	w2 -= 1
	w3 -= 1

	w := w1
	if w2 > w {
		w = w2
	}
	if w3 > w {
		w = w3
	}

	return w
}

func GetGlobalWeeknumber(t time.Time) int {
	y, w := t.ISOWeek()
	w -= 1
	if y <= 1900 {
		w -= 1
	}
	return GetAggregateIsoWeekCount(y-1) + w
}
