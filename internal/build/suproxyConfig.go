package build

import "embed"

// SuproxyEmbedConfigs is an embedded filesystem containing the suproxy configuration files.
// It is used to load the suproxy configuration at runtime.
//
//go:embed configs/suproxy.msgpack
var SuproxyEmbedConfigs embed.FS
