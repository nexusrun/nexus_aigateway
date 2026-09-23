package run

import (
	"github.com/nexusrun/nexus_aigateway/config"
	"github.com/nexusrun/nexus_aigateway/internal/observability"
	"github.com/nexusrun/nexus_aigateway/internal/providers"
	"github.com/nexusrun/nexus_aigateway/internal/providers/anthropic"
	"github.com/nexusrun/nexus_aigateway/internal/providers/azure"
	"github.com/nexusrun/nexus_aigateway/internal/providers/bailian"
	"github.com/nexusrun/nexus_aigateway/internal/providers/bedrock"
	"github.com/nexusrun/nexus_aigateway/internal/providers/bedrockmantle"
	"github.com/nexusrun/nexus_aigateway/internal/providers/cerebras"
	"github.com/nexusrun/nexus_aigateway/internal/providers/chatgpt"
	"github.com/nexusrun/nexus_aigateway/internal/providers/chutes"
	"github.com/nexusrun/nexus_aigateway/internal/providers/cloudflare"
	"github.com/nexusrun/nexus_aigateway/internal/providers/cohere"
	"github.com/nexusrun/nexus_aigateway/internal/providers/deepseek"
	"github.com/nexusrun/nexus_aigateway/internal/providers/elevenlabs"
	"github.com/nexusrun/nexus_aigateway/internal/providers/fireworks"
	"github.com/nexusrun/nexus_aigateway/internal/providers/gemini"
	"github.com/nexusrun/nexus_aigateway/internal/providers/groq"
	"github.com/nexusrun/nexus_aigateway/internal/providers/hetzner"
	"github.com/nexusrun/nexus_aigateway/internal/providers/huggingface"
	"github.com/nexusrun/nexus_aigateway/internal/providers/kilo"
	"github.com/nexusrun/nexus_aigateway/internal/providers/kimicode"
	"github.com/nexusrun/nexus_aigateway/internal/providers/llamacpp"
	"github.com/nexusrun/nexus_aigateway/internal/providers/llmd"
	"github.com/nexusrun/nexus_aigateway/internal/providers/meta"
	"github.com/nexusrun/nexus_aigateway/internal/providers/minimax"
	"github.com/nexusrun/nexus_aigateway/internal/providers/ollama"
	"github.com/nexusrun/nexus_aigateway/internal/providers/openai"
	"github.com/nexusrun/nexus_aigateway/internal/providers/opencodego"
	"github.com/nexusrun/nexus_aigateway/internal/providers/openrouter"
	"github.com/nexusrun/nexus_aigateway/internal/providers/oracle"
	"github.com/nexusrun/nexus_aigateway/internal/providers/sglang"
	"github.com/nexusrun/nexus_aigateway/internal/providers/vertex"
	"github.com/nexusrun/nexus_aigateway/internal/providers/vllm"
	"github.com/nexusrun/nexus_aigateway/internal/providers/xai"
	"github.com/nexusrun/nexus_aigateway/internal/providers/xiaomi"
	"github.com/nexusrun/nexus_aigateway/internal/providers/zai"
)

// defaultProviderFactory builds the provider factory with every provider type
// the standard gateway ships with.
func defaultProviderFactory(cfg *config.Config) *providers.ProviderFactory {
	factory := providers.NewProviderFactory()

	if cfg.Metrics.Enabled {
		factory.SetHooks(observability.NewPrometheusHooks())
	}

	factory.Add(openai.Registration)
	factory.Add(openrouter.Registration)
	factory.Add(azure.Registration)
	factory.Add(bailian.Registration)
	factory.Add(oracle.Registration)
	factory.Add(anthropic.Registration)
	factory.Add(bedrock.Registration)
	factory.Add(bedrockmantle.Registration)
	factory.Add(chatgpt.Registration)
	factory.Add(chutes.Registration)
	factory.Add(cerebras.Registration)
	factory.Add(cloudflare.Registration)
	factory.Add(cohere.Registration)
	factory.Add(deepseek.Registration)
	factory.Add(elevenlabs.Registration)
	factory.Add(fireworks.Registration)
	factory.Add(gemini.Registration)
	factory.Add(vertex.Registration)
	factory.Add(groq.Registration)
	factory.Add(hetzner.Registration)
	factory.Add(huggingface.Registration)
	factory.Add(kilo.Registration)
	factory.Add(kimicode.Registration)
	factory.Add(llamacpp.Registration)
	factory.Add(llmd.Registration)
	factory.Add(meta.Registration)
	factory.Add(minimax.Registration)
	factory.Add(ollama.Registration)
	factory.Add(opencodego.Registration)
	factory.Add(sglang.Registration)
	factory.Add(vllm.Registration)
	factory.Add(xai.Registration)
	factory.Add(xiaomi.Registration)
	factory.Add(zai.Registration)

	return factory
}
