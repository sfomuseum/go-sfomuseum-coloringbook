package static

import (
	"embed"
)

//go:embed *.svg *.png jar/*.jar
var FS embed.FS
