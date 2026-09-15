package zai

import "github.com/nexusrun/nexus_aigateway/internal/providers"

var passthroughSemanticEnricher = providers.NewOpenAICompatibleSemanticEnricher("zai")
