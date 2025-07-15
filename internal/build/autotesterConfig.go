package build

import "embed"

// AutotesterEmbedConfigs is an embedded filesystem containing the autotester configuration files.
// It is used to load the autotester configuration at runtime.
//
//go:embed configs/autotester.msgpack
var AutotesterEmbedConfigs embed.FS
