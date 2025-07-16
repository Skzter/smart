package web

import (
	"embed"
)

// DistFS is an embedded filesystem containing the static assets for the web application.
// It is used to serve the frontend assets at runtime.
//
//go:embed all:dist
var DistFS embed.FS

// Auth0Config is an embedded filesystem containing the Auth0 configuration file.
// It is used to serve the Auth0 configuration at runtime.
//
//go:embed auth_config.json
var Auth0Config embed.FS
