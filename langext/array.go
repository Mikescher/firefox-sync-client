package langext

func BoolCount(arr ...bool) int {
	c := 0
	for _, v := range arr {
		if v {
			c++
		}
	}
	return c
}
