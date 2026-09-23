// Package cloudflare provides Cloudflare Workers AI integration for the gateway.
package cloudflare

import (
	"net/http"

	"github.com/nexusrun/nexus_aigateway/internal/core"
	"github.com/nexusrun/nexus_aigateway/internal/llmclient"
	"github.com/nexusrun/nexus_aigateway/internal/providers"
	"github.com/nexusrun/nexus_aigateway/internal/providers/openai"
)

// Registration provides factory registration for Cloudflare Workers AI.
// Cloudflare's OpenAI-compatible URL is account-scoped, so base_url is required.
var Registration = providers.Registration{
	Type: "cloudflare",
	New:  New,
	Discovery: providers.DiscoveryConfig{
		RequireBaseURL: true,
	},
}

// Provider implements the OpenAI-compatible Cloudflare Workers AI API.
type Provider struct {
	*openai.ChatCompatible
}

var _ core.Provider = (*Provider)(nil)

// New creates a Cloudflare Workers AI provider.
func New(cfg providers.ProviderConfig, opts providers.ProviderOptions) core.Provider {
	return &Provider{openai.NewChatCompatible(cfg.APIKey, opts, openai.CompatibleProviderConfig{
		ProviderName: "cloudflare",
		BaseURL:      cfg.BaseURL,
	})}
}

// NewWithHTTPClient creates a Cloudflare Workers AI provider with a custom HTTP client.
func NewWithHTTPClient(apiKey, baseURL string, httpClient *http.Client, hooks llmclient.Hooks) *Provider {
	return &Provider{openai.NewChatCompatibleWithHTTPClient(apiKey, httpClient, hooks, openai.CompatibleProviderConfig{
		ProviderName: "cloudflare",
		BaseURL:      baseURL,
	})}
}
