// Package main is the entry point for the LLM gateway server.
package main

import (
	"context"
	"os"

	"github.com/nexusrun/nexus_aigateway/run"
)

// @title          NEXUS AI Gateway
// @version        1.0
// @description    AI gateway routing requests to multiple LLM providers (OpenAI, Anthropic, Gemini, Groq, Fireworks AI, Meta, OpenRouter, Kilo AI, DeepSeek, Z.ai, xAI, MiniMax, Xiaomi MiMo, OpenCode Go, Oracle, Ollama, Bailian). Drop-in OpenAI-compatible API.
// @BasePath       /
// @schemes        http
// @securityDefinitions.apikey BearerAuth
// @in             header
// @name           Authorization
func main() {
	err := run.Run(context.Background(), run.Options{
		ProductName:          "NEXUS AI Gateway",
		ConfigureSwaggerDocs: configureSwaggerDocs,
	})
	if code := run.ExitCode(err); code != 0 {
		os.Exit(code)
	}
}
