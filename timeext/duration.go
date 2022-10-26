package timeext

import "time"

func FromSeconds(v int) time.Duration {
	return time.Duration(int64(v) * int64(time.Second))
}

func FromSecondsInt32(v int32) time.Duration {
	return time.Duration(int64(v) * int64(time.Second))
}

func FromSecondsInt64(v int64) time.Duration {
	return time.Duration(v * int64(time.Second))
}

func FromSecondsFloat32(v float32) time.Duration {
	return time.Duration(int64(v * float32(time.Second)))
}

func FromSecondsFloat64(v float64) time.Duration {
	return time.Duration(int64(v * float64(time.Second)))
}

func FromSecondsFloat(v float64) time.Duration {
	return time.Duration(int64(v * float64(time.Second)))
}

func FromMinutes(v int) time.Duration {
	return time.Duration(int64(v) * int64(time.Minute))
}

func FromMinutesFloat(v float64) time.Duration {
	return time.Duration(int64(v * float64(time.Minute)))
}

func FromMinutesFloat64(v float64) time.Duration {
	return time.Duration(int64(v * float64(time.Minute)))
}

func FromHoursFloat64(v float64) time.Duration {
	return time.Duration(int64(v * float64(time.Hour)))
}

func FromDays(v int) time.Duration {
	return time.Duration(int64(v) * int64(24) * int64(time.Hour))
}

func FromMilliseconds(v int) time.Duration {
	return time.Duration(int64(v) * int64(time.Millisecond))
}

func FromMillisecondsFloat(v float64) time.Duration {
	return time.Duration(int64(v * float64(time.Millisecond)))
}
