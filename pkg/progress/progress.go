package progress

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

type ProgressBar struct {
	current   int64
	total     int64
	writer    io.Writer
	callback  func(transferred int64)
	startTime time.Time
}

func New(total int64) *ProgressBar {
	return &ProgressBar{
		current:   0,
		total:     total,
		writer:    os.Stdout,
		startTime: time.Now(),
	}
}

func (p *ProgressBar) SetWriter(w io.Writer) {
	p.writer = w
}

func (p *ProgressBar) SetCallback(cb func(transferred int64)) {
	p.callback = cb
}

func (p *ProgressBar) SetTotal(total int64) {
	p.total = total
}

func (p *ProgressBar) SetStartTime(t time.Time) {
	p.startTime = t
}

func (p *ProgressBar) Increment(n int64) {
	p.current += n
	if p.callback != nil {
		p.callback(p.current)
	}
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

func (p *ProgressBar) GetSpeed() float64 {
	elapsed := time.Since(p.startTime)
	if elapsed == 0 {
		return 0
	}
	return float64(p.current) / elapsed.Seconds()
}

func (p *ProgressBar) GetTimeRemaining() time.Duration {
	speed := p.GetSpeed()
	if speed == 0 {
		return 0
	}
	remaining := p.total - p.current
	return time.Duration(float64(remaining) / speed * float64(time.Second))
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
