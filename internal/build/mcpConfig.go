package build

import "embed"

// MCPEmbedConfigs is an embedded filesystem containing the mcp configuration files.
// It is used to load the mcp configuration at runtime.
//
//go:embed configs/mcp.msgpack
var MCPEmbedConfigs embed.FS
