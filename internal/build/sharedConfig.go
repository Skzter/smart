package build

import "embed"

// SharedEmbedConfigs is an embedded filesystem containing the suproxy configuration files.
// It is used to load the suproxy configuration at runtime.
//
//go:embed configs/shared.msgpack
var SharedEmbedConfigs embed.FS
