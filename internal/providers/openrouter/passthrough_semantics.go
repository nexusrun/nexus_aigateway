package openrouter

import "github.com/nexusrun/nexus_aigateway/internal/providers"

var passthroughSemanticEnricher = providers.NewOpenAICompatibleSemanticEnricher("openrouter")
