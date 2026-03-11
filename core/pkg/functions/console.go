package functions

import (
	"github.com/fatih/color"
)

func PrintStyle(c string, data ...interface{}) {
	var o = color.New(color.FgBlack)
	switch c {
	case "cyan":
		o = color.New(color.FgCyan)

	case "cyan-underline":
		o = color.New(color.FgCyan).Add(color.Underline)

	case "red":
		o = color.New(color.FgRed)

	case "red-underline":
		o = color.New(color.FgRed).Add(color.Underline)

	case "yellow-underline":
		o = color.New(color.FgYellow).Add(color.Underline)

	case "yellow":
		o = color.New(color.FgYellow)

	case "blue":
		o = color.New(color.FgBlue)

	case "green":
		o = color.New(color.FgGreen)
	}

	_, _ = o.Println(data...)
}

func PrintStyleInline(c string, data ...interface{}) {
	var o = color.New(color.FgBlack)
	switch c {
	case "cyan":
		o = color.New(color.FgCyan)

	case "cyan-underline":
		o = color.New(color.FgCyan).Add(color.Underline)

	case "red":
		o = color.New(color.FgRed)

	case "red-underline":
		o = color.New(color.FgRed).Add(color.Underline)

	case "yellow-underline":
		o = color.New(color.FgYellow).Add(color.Underline)

	case "yellow":
		o = color.New(color.FgYellow)

	case "blue":
		o = color.New(color.FgBlue)

	case "green":
		o = color.New(color.FgGreen)
	}

	_, _ = o.Print(data...)
}

const (
	colorReset   = "\033[0m"
	colorRed     = "\033[31m"
	colorGreen   = "\033[32m"
	colorYellow  = "\033[33m"
	colorBlue    = "\033[34m"
	colorMagenta = "\033[35m"
	colorCyan    = "\033[36m"
	colorWhite   = "\033[37m"
)

func SRed(content string) string {
	return colorRed + content + colorReset
}

func SGreen(content string) string {
	return colorGreen + content + colorReset
}

func SYellow(content string) string {
	return colorYellow + content + colorReset
}

func SBlue(content string) string {
	return colorBlue + content + colorReset
}

func SMagenta(content string) string {
	return colorMagenta + content + colorReset
}

func SCyan(content string) string {
	return colorCyan + content + colorReset
}

func SWhite(content string) string {
	return colorWhite + content + colorReset
}
