package progress

import (
	"fmt"
	"io"
	"os"
	"strings"
)

type ProgressBar struct {
	current int64
	total   int64
	writer  io.Writer
}

func New(total int64) *ProgressBar {
	return &ProgressBar{
		current: 0,
		total:   total,
		writer:  os.Stdout,
	}
}

func (p *ProgressBar) SetWriter(w io.Writer) {
	p.writer = w
}

func (p *ProgressBar) SetTotal(total int64) {
	p.total = total
}

func (p *ProgressBar) Increment(n int64) {
	p.current += n
}

func (p *ProgressBar) Current() int64 {
	return p.current
}

func (p *ProgressBar) Finish() {
	p.current = p.total
}

func (p *ProgressBar) IsFinished() bool {
	return p.current >= p.total
}

func (p *ProgressBar) Render() {
	if p.total == 0 {
		return
	}
	percent := float64(p.current) / float64(p.total) * 100
	width := 20
	filled := int(float64(width) * float64(p.current) / float64(p.total))
	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
	fmt.Fprintf(p.writer, "\r%s %6.2f%%", bar, percent)
}
