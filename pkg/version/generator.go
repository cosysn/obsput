package version

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type Generator struct {
	counter int64
}

func NewGenerator() *Generator {
	return &Generator{}
}

func (g *Generator) Generate() string {
	commit := g.getShortCommit()
	now := time.Now()
	date := now.Format("20060102")
	timestamp := now.Format("150405")
	g.counter++
	return fmt.Sprintf("v1.0.0-%s-%s-%s-%d", commit, date, timestamp, g.counter)
}

func (g *Generator) getShortCommit() string {
	cmd := exec.Command("git", "rev-parse", "--short", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(output))
}
