package ansi

import "fmt"

const resetCol = "\033[0m"

func Compose(cols ...func(string) string) func(string) string {
	return func(s string) string {
		for _, mod := range cols {
			s = mod(s)
		}
		return s
	}
}

func col(s string, code int) string {
	c := fmt.Sprintf("\033[%dm", code)
	return c + s + resetCol
}

func ColorRed(s string) string {
	return col(s, 31)
}

func ColorWhite(s string) string {
	return col(s, 37)
}

func ColorGreen(s string) string {
	return col(s, 32)
}

func ColorYellow(s string) string {
	return col(s, 33)
}

func ColorCyan(s string) string {
	return col(s, 36)
}

func ColorLight(s string) string {
	return col(s, 97) // Bright white → light
}

// BG
func BgDark(s string) string {
	return col(s, 100)
}

func BgRed(s string) string {
	return col(s, 41)
}

func BgGreen(s string) string {
	return col(s, 42)
}

// modifiers

func Bold(s string) string {
	return col(s, 1)
}

func Underline(s string) string {
	return col(s, 4)
}
