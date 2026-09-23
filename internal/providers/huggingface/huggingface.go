// Package huggingface provides Hugging Face Inference Providers integration.
package huggingface

import (
	"net/http"

	"github.com/nexusrun/nexus_aigateway/internal/core"
	"github.com/nexusrun/nexus_aigateway/internal/llmclient"
	"github.com/nexusrun/nexus_aigateway/internal/providers"
	"github.com/nexusrun/nexus_aigateway/internal/providers/openai"
)

const defaultBaseURL = "https://router.huggingface.co/v1"

// Registration provides factory registration for Hugging Face Inference Providers.
var Registration = providers.Registration{
	Type: "huggingface",
	New:  New,
	Discovery: providers.DiscoveryConfig{
		DefaultBaseURL: defaultBaseURL,
	},
}

// Provider implements the OpenAI-compatible Hugging Face router API.
type Provider struct {
	*openai.ChatCompatible
}

var _ core.Provider = (*Provider)(nil)

// New creates a Hugging Face Inference Providers client.
func New(cfg providers.ProviderConfig, opts providers.ProviderOptions) core.Provider {
	return &Provider{openai.NewChatCompatible(cfg.APIKey, opts, openai.CompatibleProviderConfig{
		ProviderName: "huggingface",
		BaseURL:      providers.ResolveBaseURL(cfg.BaseURL, defaultBaseURL),
	})}
}

// NewWithHTTPClient creates a Hugging Face provider with a custom HTTP client.
func NewWithHTTPClient(apiKey, baseURL string, httpClient *http.Client, hooks llmclient.Hooks) *Provider {
	return &Provider{openai.NewChatCompatibleWithHTTPClient(apiKey, httpClient, hooks, openai.CompatibleProviderConfig{
		ProviderName: "huggingface",
		BaseURL:      providers.ResolveBaseURL(baseURL, defaultBaseURL),
	})}
}
