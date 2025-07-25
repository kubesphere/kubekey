package config

import "embed"

// Swagger embeds the swagger-ui directory containing the OpenAPI/Swagger documentation UI
// This allows serving the Swagger UI directly from the binary without needing external files
//
//go:embed swagger-ui
var Swagger embed.FS
