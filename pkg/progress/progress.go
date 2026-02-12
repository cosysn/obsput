package progress

import (
	"io"
	"os"
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
