package langext

import "strings"

func StrPadLeft(str string, pad string, padlen int) string {
	if pad == "" {
		pad = " "
	}

	if len(str) >= padlen {
		return str
	}

	return strings.Repeat(pad, padlen-len(str))[0:(padlen-len(str))] + str
}

func StrPadRight(str string, pad string, padlen int) string {
	if pad == "" {
		pad = " "
	}

	if len(str) >= padlen {
		return str
	}

	return str + strings.Repeat(pad, padlen-len(str))[0:(padlen-len(str))]
}

func StrRunePadLeft(str string, pad string, padlen int) string {
	if pad == "" {
		pad = " "
	}

	if len([]rune(str)) >= padlen {
		return str
	}

	return strings.Repeat(pad, padlen-len([]rune(str)))[0:(padlen-len([]rune(str)))] + str
}

func StrRunePadRight(str string, pad string, padlen int) string {
	if pad == "" {
		pad = " "
	}

	if len([]rune(str)) >= padlen {
		return str
	}

	return str + strings.Repeat(pad, padlen-len([]rune(str)))[0:(padlen-len([]rune(str)))]
}
