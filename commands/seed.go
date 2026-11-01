package commands

import (
	"embed"
)

//go:embed seed/cv.yaml
//go:embed seed/cv.json
var seedFiles embed.FS
