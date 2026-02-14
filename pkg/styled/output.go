package styled

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
)

// Color definitions
type Color int

const (
	ColorNone Color = iota
	ColorRed
	ColorGreen
	ColorYellow
	ColorBlue
	ColorCyan
	ColorWhite
	ColorGray
	ColorBold
)

// Style combines foreground color and attributes
type Style struct {
	Fg     Color
	Bg     Color
	Bold   bool
	Italic bool
}

var (
	// Default styles
	Success Style = Style{Fg: ColorGreen, Bold: true}
	Error   Style = Style{Fg: ColorRed, Bold: true}
	Info    Style = Style{Fg: ColorBlue}
	Warning Style = Style{Fg: ColorYellow, Bold: true}
	Header  Style = Style{Fg: ColorCyan, Bold: true}
	Muted   Style = Style{Fg: ColorGray}
)

// Output provides styled terminal output
type Output struct {
	w io.Writer
}

// NewOutput creates a new styled output
func NewOutput() *Output {
	return &Output{w: os.Stdout}
}

// SetOutput sets the output writer
func (o *Output) SetOutput(w io.Writer) {
	o.w = w
}

// color codes
var (
	colorCodes = map[Color]string{
		ColorNone:   "\033[0m",
		ColorRed:    "\033[31m",
		ColorGreen:  "\033[32m",
		ColorYellow: "\033[33m",
		ColorBlue:   "\033[34m",
		ColorCyan:   "\033[36m",
		ColorWhite:  "\033[37m",
		ColorGray:   "\033[90m",
		ColorBold:   "\033[1m",
	}
)

// Print prints styled text
func (o *Output) Print(style Style, format string, args ...interface{}) {
	o.w.Write([]byte(o.styleText(style, format)))
	if !strings.HasSuffix(format, "\n") && !strings.HasSuffix(format, "\r") {
		o.w.Write([]byte("\n"))
	}
}

// Printf prints formatted styled text
func (o *Output) Printf(style Style, format string, args ...interface{}) {
	o.w.Write([]byte(o.styleText(style, fmt.Sprintf(format, args...))))
}

// Println prints styled text with newline
func (o *Output) Println(style Style, args ...interface{}) {
	o.w.Write([]byte(o.styleText(style, fmt.Sprint(args...))))
	o.w.Write([]byte("\n"))
}

// PrintBox prints text in a box
func (o *Output) PrintBox(style Style, title string, content string) {
	o.w.Write([]byte("\n"))
	o.w.Write([]byte(o.styleText(style, "┌"+"─"+strings.Repeat("─", len(title)+2)+"┐")))
	o.w.Write([]byte("\n"))
	o.w.Write([]byte(o.styleText(style, fmt.Sprintf("│ %s │", title))))
	o.w.Write([]byte("\n"))
	o.w.Write([]byte(o.styleText(style, "├"+"─"+strings.Repeat("─", len(title)+2)+"┤")))
	o.w.Write([]byte("\n"))
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		o.w.Write([]byte(o.styleText(style, fmt.Sprintf("│ %-*s │", len(title)+2, line))))
		o.w.Write([]byte("\n"))
	}
	o.w.Write([]byte(o.styleText(style, "└"+"─"+strings.Repeat("─", len(title)+2)+"┘")))
	o.w.Write([]byte("\n"))
}

// Section prints a section header
func (o *Output) Section(title string) {
	o.w.Write([]byte("\n"))
	o.w.Write([]byte(o.styleText(Header, fmt.Sprintf("━━ %s ━━", strings.ToUpper(title)))))
	o.w.Write([]byte("\n"))
}

// Subsection prints a subsection header
func (o *Output) Subsection(title string) {
	o.w.Write([]byte("\n"))
	o.w.Write([]byte(o.styleText(Info, fmt.Sprintf("▶ %s", title))))
	o.w.Write([]byte("\n"))
}

// KeyValue prints a key-value pair
func (o *Output) KeyValue(key string, value interface{}) {
	o.w.Write([]byte(o.styleText(Muted, fmt.Sprintf("  %s:", key))))
	o.w.Write([]byte(" "))
	o.w.Write([]byte(fmt.Sprintf("%v\n", value)))
}

// SuccessMsg prints a success message
func (o *Output) SuccessMsg(msg string) {
	o.w.Write([]byte(o.styleText(Success, "✓")))
	o.w.Write([]byte(" "))
	o.w.Write([]byte(msg))
	o.w.Write([]byte("\n"))
}

// ErrorMsg prints an error message
func (o *Output) ErrorMsg(msg string) {
	o.w.Write([]byte(o.styleText(Error, "✗")))
	o.w.Write([]byte(" "))
	o.w.Write([]byte(msg))
	o.w.Write([]byte("\n"))
}

// WarningMsg prints a warning message
func (o *Output) WarningMsg(msg string) {
	o.w.Write([]byte(o.styleText(Warning, "⚠")))
	o.w.Write([]byte(" "))
	o.w.Write([]byte(msg))
	o.w.Write([]byte("\n"))
}

// ProgressBar prints a progress bar
func (o *Output) ProgressBar(current, total int64, width int) {
	percent := float64(current) / float64(total) * 100
	filled := int(float64(width) * float64(current) / float64(total))
	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
	o.w.Write([]byte(fmt.Sprintf("\r%s %6.2f%%", bar, percent)))
	if current >= total {
		o.w.Write([]byte("\n"))
	}
}

// ResultTable prints a styled result table
func (o *Output) ResultTable(results map[string]string) {
	t := table.NewWriter()
	t.SetOutputMirror(o.w)
	t.AppendHeader(table.Row{"Field", "Value"})
	for k, v := range results {
		t.AppendRow(table.Row{k, v})
	}
	t.SetStyle(table.StyleRounded)
	t.Render()
}

// Summary prints a summary line
func (o *Output) Summary(success, failed int) {
	o.w.Write([]byte("\n"))
	if failed == 0 {
		o.w.Write([]byte(o.styleText(Success, fmt.Sprintf("✓ %d uploaded successfully", success))))
	} else if success == 0 {
		o.w.Write([]byte(o.styleText(Error, fmt.Sprintf("✗ %d failed, %d successful", failed, success))))
	} else {
		o.w.Write([]byte(o.styleText(Warning, fmt.Sprintf("⚠ %d successful, %d failed", success, failed))))
	}
	o.w.Write([]byte("\n"))
}

// Separator prints a separator line
func (o *Output) Separator() {
	o.w.Write([]byte(o.styleText(Muted, strings.Repeat("─", 60))))
	o.w.Write([]byte("\n"))
}

// Divider prints a divider
func (o *Output) Divider() {
	o.w.Write([]byte("\n"))
}

// styleText applies style to text
func (o *Output) styleText(style Style, text string) string {
	var codes []string
	if style.Fg != ColorNone {
		codes = append(codes, colorCodes[style.Fg])
	}
	if style.Bold {
		codes = append(codes, colorCodes[ColorBold])
	}
	if len(codes) > 0 {
		return strings.Join(codes, "") + text + colorCodes[ColorNone]
	}
	return text
}

// CleanText removes ANSI escape codes
func CleanText(text string) string {
	clean := text
	for _, code := range colorCodes {
		clean = strings.ReplaceAll(clean, code, "")
	}
	return clean
}
