package build

import "embed"

//go:embed configs/autotester.msgpack
var AutotesterEmbedConfigs embed.FS
