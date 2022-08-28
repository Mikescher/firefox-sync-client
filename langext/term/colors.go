package term

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
	colorCyan   = "\033[36m"
	colorGray   = "\033[37m"
	colorWhite  = "\033[97m"
)

func Red(v string) string {
	return colorRed + v + colorReset
}

func Green(v string) string {
	return colorGreen + v + colorReset
}

func Yellow(v string) string {
	return colorYellow + v + colorReset
}

func Blue(v string) string {
	return colorBlue + v + colorReset
}

func Purple(v string) string {
	return colorPurple + v + colorReset
}

func Cyan(v string) string {
	return colorCyan + v + colorReset
}

func Gray(v string) string {
	return colorGray + v + colorReset
}

func White(v string) string {
	return colorWhite + v + colorReset
}
