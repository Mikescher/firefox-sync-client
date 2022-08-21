package langext

import "strings"

func StrPadRight(str string, pad string, padlen int) string {
	if pad == "" {
		pad = " "
	}

	if len(str) >= padlen {
		return str
	}

	return str + strings.Repeat(pad, padlen-len(str))[0:(padlen-len(str))]
}

func StrPadLeft(str string, pad string, padlen int) string {
	if pad == "" {
		pad = " "
	}

	if len(str) >= padlen {
		return str
	}

	return strings.Repeat(pad, padlen-len(str))[0:(padlen-len(str))] + str
}
