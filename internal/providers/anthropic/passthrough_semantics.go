package anthropic

import "github.com/nexusrun/nexus_aigateway/internal/providers"

var passthroughSemanticEnricher = providers.NewSemanticEnricher("anthropic", map[string]providers.PassthroughEndpointSemantics{
	"/messages":         {Operation: "anthropic.messages", GenAIOperation: "chat", AuditPath: "/v1/messages"},
	"/messages/batches": {Operation: "anthropic.messages_batches", AuditPath: "/v1/messages/batches"},
})
