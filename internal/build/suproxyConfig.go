package build

import "embed"

//go:embed configs/suproxy.msgpack
var SuproxyEmbedConfigs embed.FS
