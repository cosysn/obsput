package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/jedib0t/go-pretty/v6/table"
)

type VersionItem struct {
	Version string
	Size    string
	Date    string
	Commit  string
	URL     string
}

type Formatter struct {
	output io.Writer
}

func NewFormatter() *Formatter {
	return &Formatter{
		output: os.Stdout,
	}
}

func (f *Formatter) SetOutput(w io.Writer) {
	f.output = w
}

func (f *Formatter) PrintVersionTable(items []VersionItem) {
	t := table.NewWriter()
	t.SetOutputMirror(f.output)
	t.SetStyle(table.StyleRounded)
	t.AppendHeader(table.Row{"VERSION", "SIZE", "DATE", "COMMIT", "DOWNLOAD_URL"})
	for _, item := range items {
		t.AppendRow(table.Row{item.Version, item.Size, item.Date, item.Commit, item.URL})
	}
	t.Render()
}

func (f *Formatter) PrintJSON(items []VersionItem) {
	enc := json.NewEncoder(f.output)
	enc.SetIndent("", "  ")
	enc.Encode(items)
}

func (f *Formatter) FormatSize(bytes int64) string {
	if bytes < 1024 {
		return fmt.Sprintf("%d B", bytes)
	}
	if bytes < 1024*1024 {
		return fmt.Sprintf("%.1f KB", float64(bytes)/1024)
	}
	if bytes < 1024*1024*1024 {
		return fmt.Sprintf("%.1f MB", float64(bytes)/(1024*1024))
	}
	return fmt.Sprintf("%.2f GB", float64(bytes)/(1024*1024*1024))
}
