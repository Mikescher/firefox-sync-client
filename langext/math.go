package langext

import "golang.org/x/exp/constraints"

func Max[T constraints.Ordered](v1 T, v2 T) T {
	if v1 > v2 {
		return v1
	} else {
		return v2
	}
}

func Min[T constraints.Ordered](v1 T, v2 T) T {
	if v1 < v2 {
		return v1
	} else {
		return v2
	}
}
