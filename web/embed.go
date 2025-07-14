package web

import (
	"embed"
)

//go:embed all:dist/assets
var DistFS embed.FS

//go:embed auth_config.json
var Auth0Config embed.FS
