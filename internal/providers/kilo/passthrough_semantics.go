package kilo

import "github.com/nexusrun/nexus_aigateway/internal/providers"

var passthroughSemanticEnricher = providers.NewSemanticEnricher("kilo", map[string]providers.PassthroughEndpointSemantics{
	"/chat/completions": {Operation: "kilo.chat_completions", GenAIOperation: "chat", AuditPath: "/v1/chat/completions"},
})
