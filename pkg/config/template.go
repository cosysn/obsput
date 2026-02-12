package config

import "embed"

//go:embed template.yaml
var FS embed.FS

const DefaultConfig = `configs: []
`
