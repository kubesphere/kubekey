package config

import "embed"

// Swagger embeds the swagger-ui directory containing the OpenAPI/Swagger documentation UI
// This allows serving the Swagger UI directly from the binary without needing external files
//
//go:embed swagger-ui
var Swagger embed.FS

// WebUI embeds the web directory containing the static web UI assets
// This allows serving the web UI directly from the binary without needing external files
//
//go:embed all:web
var WebUI embed.FS
