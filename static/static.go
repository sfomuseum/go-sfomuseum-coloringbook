package static

import (
	"embed"
)

//go:embed *.svg *.png
var FS embed.FS
