// Package cerebras provides Cerebras Inference API integration for the gateway.
package cerebras

import (
	"net/http"

	"github.com/nexusrun/nexus_aigateway/internal/core"
	"github.com/nexusrun/nexus_aigateway/internal/llmclient"
	"github.com/nexusrun/nexus_aigateway/internal/providers"
	"github.com/nexusrun/nexus_aigateway/internal/providers/openai"
)

const defaultBaseURL = "https://api.cerebras.ai/v1"

// Registration provides factory registration for Cerebras Inference.
var Registration = providers.Registration{
	Type: "cerebras",
	New:  New,
	Discovery: providers.DiscoveryConfig{
		DefaultBaseURL: defaultBaseURL,
	},
}

// Provider implements the OpenAI-compatible Cerebras Inference API.
type Provider struct {
	*openai.ChatCompatible
}

var _ core.Provider = (*Provider)(nil)

// New creates a Cerebras provider.
func New(cfg providers.ProviderConfig, opts providers.ProviderOptions) core.Provider {
	return &Provider{openai.NewChatCompatible(cfg.APIKey, opts, openai.CompatibleProviderConfig{
		ProviderName: "cerebras",
		BaseURL:      providers.ResolveBaseURL(cfg.BaseURL, defaultBaseURL),
	})}
}

// NewWithHTTPClient creates a Cerebras provider with a custom HTTP client.
func NewWithHTTPClient(apiKey, baseURL string, httpClient *http.Client, hooks llmclient.Hooks) *Provider {
	return &Provider{openai.NewChatCompatibleWithHTTPClient(apiKey, httpClient, hooks, openai.CompatibleProviderConfig{
		ProviderName: "cerebras",
		BaseURL:      providers.ResolveBaseURL(baseURL, defaultBaseURL),
	})}
}
