package output

import (
	"encoding/json"
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
