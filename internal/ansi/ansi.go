package ansi

import "fmt"

const resetCol = "\033[0m"

func col(s string, code int) string {
	c := fmt.Sprintf("\033[%dm", code)
	return c + s + resetCol
}

func ColorRed(s string) string {
	return col(s, 31)
}

func ColorYellow(s string) string {
	return col(s, 33)
}

func ColorCyan(s string) string {
	return col(s, 36)
}
