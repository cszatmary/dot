package color

const (
	red    = "\x1b[31m"
	green  = "\x1b[32m"
	yellow = "\x1b[33m"
	cyan   = "\x1b[36m"
	reset  = "\x1b[0m"
)

func Red(str string) string {
	return red + str + reset
}

func Green(str string) string {
	return green + str + reset
}

func Yellow(str string) string {
	return yellow + str + reset
}

func Cyan(str string) string {
	return cyan + str + reset
}
