<p align="center">
  <img alt="NEXUS AI Gateway logo" src="docs/nexus-ai-logo.png" width="220">
</p>
[![Deploy to NEXUS AI](https://nexusai.run/deploy-button.svg)](https://nexusai.run/deploy?repo=https://github.com/nexusrun/nexus_aigateway)
<h1 align="center">
  NEXUS AI Gateway - FREE AI gateway
</h1>

</p>


<p align="center">
  NEXUS AI Gateway is the control plane for your AI stack: one fast, resource-efficient gateway for routing requests across every model provider you use.
</p>

<p>
  Spend less. Stay in control.
</p>
<p>
  <strong>Spend smarter</strong> reduce unnecessary model calls with response caching, track usage and costs in one place, and get more from every token with prompt compression and intelligent routing.
</p>
<p>
  <strong>Operate with confidence</strong> keep your applications reliable as providers, models, and workloads change. NEXUS AI Gateway brings the performance, failover, observability, and controls you need to run AI in production.
</p>
<p>
  Built to be the last AI gateway you need: fast at the edge, efficient by design, and ready for whatever comes next.
</p>

## Demo

<p align="center">
  <a href="./docs/AI-Gateway.gif">
    <img src="./docs/AI-Gateway.gif" alt="NEXUS AI Gateway demo" width="100%">
  </a>
</p>

## Local Deployment

Run NEXUS AI Gateway locally with SQLite and the provider of your choice. You
need Go, Node.js/npm, and at least one provider API key.

```bash
git clone https://github.com/nexusrun/nexus_aigateway.git
cd nexus_aigateway
cp .env.template .env
```

Edit `.env` and configure local storage plus a provider key. For example:

```env
STORAGE_TYPE=sqlite
GROQ_API_KEY=your_key_here
```

Build the dashboard and start the gateway:

```bash
make frontend
make docs
make run
```

The local services are available at:

- Dashboard: `http://localhost:8080/admin/dashboard`
- Product documentation: `http://localhost:8080/docs`
- Models: `http://localhost:8080/v1/models`
- OpenAI-compatible API: `http://localhost:8080/v1`

`make docs` exports the product documentation into `bin/docs`. Production
Docker images build and serve the same documentation automatically at
`/docs`; Swagger remains available separately at `/swagger/index.html`.

Send a test request using a model returned by `GET /v1/models`:

```bash
curl http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer change-me" \
  -d '{
    "model": "groq/llama-3.3-70b-versatile",
    "messages": [{"role": "user", "content": "Hello"}]
  }'
```

## Quick Start

**Step 1:** Prepare the NEXUS AI deployment

Clone the repository, create a protected environment file, and authenticate the NEXUS AI CLI:

```bash
git clone https://github.com/nexusrun/nexus_aigateway.git
cd nexus_aigateway
cp .env.template .env
nexus auth login
```

Configure provider keys, database settings, and dashboard credentials in NEXUS AI deployment environment variables or the protected `.env` file.

ℹ️ See [`.env.template`](./.env.template) for the complete list of environment variables, including all available providers.

**Step 2:** Deploy to NEXUS AI

Push the current source to the repository that NEXUS AI will build, then deploy it through the NEXUS AI CLI:

```bash
git push origin main

nexus deploy source \
  --repo https://github.com/nexusrun/nexus_aigateway.git \
  --name aigateway \
  --branch main \
  --provider docker \
  --services postgresql,redis \
  --dockerfile Dockerfile.nexus \
  --env-file .env \
  --environment PRODUCTION \
  --wait \
  --json
```

Keep provider keys, database credentials, and `ADMIN_SESSION_SECRET` in NEXUS AI environment settings or a protected `.env` file. Do not commit secrets. Verify the deployed source and environment after each build:

```bash
nexus deploy get aigateway --json
nexus deploy status aigateway
nexus deploy logs aigateway --lines 200
```

**Step 3:** Open the dashboard

```text
https://aigateway.nexusai.run/
```

The product documentation is served by the same gateway at:

```text
https://aigateway.nexusai.run/docs
```

**Step 4:** Make an API call

```bash
curl https://aigateway.nexusai.run/v1/responses \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $NEXUS_AI_API_KEY" \
  -d '{
    "model": "gpt-5-chat-latest",
    "input": "Hello!"
  }'
```

## NEXUS AI Gateway and official SDKs

NEXUS AI Gateway accepts requests in two compatible formats:

- OpenAI-compatible at `/v1`
- Anthropic-compatible at `/v1/messages`

The official SDKs therefore work unchanged. Configure their base URLs as follows:

- OpenAI SDK: `https://aigateway.nexusai.run/v1`
- Anthropic SDK: `https://aigateway.nexusai.run` (the SDK appends `/v1/messages`)

## Use NEXUS AI Gateway with coding tools

NEXUS AI Gateway works with Claude Code, Codex, Cursor, and other coding agents.
Configure the tool to use your gateway URL and a gateway API key, then choose a
model returned by `GET /v1/models`.

For a local gateway, the common OpenAI-compatible endpoint is:

```text
Base URL: http://localhost:8080/v1
API key:  your NEXUS AI Gateway key
```

### Claude Code

Claude Code uses the Anthropic Messages API. Set the gateway URL and token in
your shell, then start Claude Code normally:

```bash
export ANTHROPIC_BASE_URL=http://localhost:8080
export ANTHROPIC_AUTH_TOKEN=your-gateway-key
claude
```

The managed gateway route is `POST /v1/messages`. To pin Claude Code directly
to the Anthropic passthrough, use `http://localhost:8080/p/anthropic` instead.
See the [Claude Code integration guide](https://aigateway.nexusai.run/docs/guides/claude-code?utm_source=readme).

### Codex

Codex uses the OpenAI Responses API. Add a custom provider to
`~/.codex/config.toml`:

```toml
model_provider = "aigateway"
model = "your-model-id"

[model_providers.aigateway]
name = "NEXUS AI Gateway"
base_url = "http://localhost:8080/v1"
env_key = "AIGATEWAY_API_KEY"
wire_api = "responses"
```

Then export the gateway key and run Codex:

```bash
export AIGATEWAY_API_KEY=your-gateway-key
codex
```

See the [Codex integration guide](https://aigateway.nexusai.run/docs/guides/codex?utm_source=readme)
for ChatGPT subscription routing and provider-specific notes.

### Cursor

Cursor can route OpenAI BYOK requests through NEXUS AI Gateway:

1. Open **Cursor Settings → Models**.
2. Set **OpenAI API Key** to a dedicated gateway key.
3. Enable **Override OpenAI Base URL** and enter
   `https://your-gateway.example.com/v1`.
4. Select a model exposed by the gateway.

Cursor requires a publicly reachable HTTPS gateway URL. Use a dedicated managed
API key rather than the gateway master key. Cursor-hosted models, Tab, and some
other Cursor features do not route through the OpenAI base URL override.
See the [Cursor integration guide](https://aigateway.nexusai.run/docs/guides/cursor?utm_source=readme).

### Other coding agents and clients

OpenCode, Cline, Roo Code, OpenAI-compatible SDKs, and similar clients generally
use the same settings:

```text
Base URL: http://localhost:8080/v1
API key:  your NEXUS AI Gateway key
Model:    an ID returned by GET /v1/models
```

See the guide for [OpenCode and other agents](https://aigateway.nexusai.run/docs/guides/opencode-and-other-agents?utm_source=readme),
or the [API endpoints reference](https://aigateway.nexusai.run/docs/advanced/api-endpoints?utm_source=readme)
for client-specific behavior.

## List of Supported LLM Providers

- OpenAI
- Anthropic
- xAI (Grok)
- Google Gemini
- Cohere
- Cerebras Inference
- Cloudflare Workers AI
- Hugging Face Inference Providers
- Vertex AI
- DeepSeek
- Groq
- Fireworks AI
- Meta (Muse Spark)
- OpenRouter
- Z.ai
- Alibaba Cloud Model Studio (Bailian)
- Kilo AI
- MiniMax
- Xiaomi MiMo
- OpenCode Go
- Azure OpenAI
- Oracle
- Ollama
- SGLang
- vLLM
- llm-d
- Amazon Bedrock Runtime and Bedrock Mantle
- ChatGPT (the Codex backend) and Claude
- ElevenLabs (text-to-speech and speech-to-text)
- All OpenAI-compatible providers

See the [Providers Overview](https://aigateway.nexusai.run/docs/providers/overview?utm_source=readme) for the full
per-provider feature matrix.

---

## API docs

- [API Endpoints](https://aigateway.nexusai.run/docs/advanced/api-endpoints?utm_source=readme)
- [Admin API Endpoints](https://aigateway.nexusai.run/docs/advanced/admin-endpoints?utm_source=readme)

---

## Gateway Configuration

NEXUS AI Gateway resolves configuration in the following order, with each source
overriding those to its left:

[Good defaults](https://aigateway.nexusai.run/docs/about/technical-philosophy#good-defaults) → [`config.yaml`](./config/config.example.yaml) → [`.env`](./.env.template) → exported environment variables

See the [Configuration reference](https://aigateway.nexusai.run/docs/advanced/configuration?utm_source=readme)
for the full list of settings.

---

## Features

NEXUS AI Gateway gives your applications one consistent layer for routing,
governance, and observability across the AI providers you use.

### Route every request with confidence

- [Virtual models](https://aigateway.nexusai.run/docs/features/virtual-models?utm_source=readme) - stable model aliases, provider-qualified selectors, and load balancing across targets
- [Failover and resilience](https://aigateway.nexusai.run/docs/features/failover?utm_source=readme) - automatic provider failover with [retries and circuit breakers](https://aigateway.nexusai.run/docs/advanced/resilience?utm_source=readme)
- [Session keeping](https://aigateway.nexusai.run/docs/features/session-keeping?utm_source=readme) - keep related requests on the same target and provider key so provider-side prompt caches stay warm
- [Provider key rotation](https://aigateway.nexusai.run/docs/providers/key-rotation?utm_source=readme) - distribute traffic across multiple keys to increase rate-limit headroom

### Control cost and access

- [Response caching](https://aigateway.nexusai.run/docs/features/cache?utm_source=readme) - exact and semantic caching to avoid repeat upstream requests
- [Cost tracking](https://aigateway.nexusai.run/docs/features/cost-tracking?utm_source=readme) - per-request estimates, usage analytics, and spending breakdowns
- [Budgets](https://aigateway.nexusai.run/docs/features/budgets?utm_source=readme) - enforce spend limits per user, team, or API key
- [Rate limits](https://aigateway.nexusai.run/docs/features/rate-limits?utm_source=readme) - cap requests, tokens, and concurrency by user path, provider, or model
- [User paths and access control](https://aigateway.nexusai.run/docs/features/user-path?utm_source=readme) - scope API keys, model access, budgets, usage, and audit logs across teams and users
- [Labelling](https://aigateway.nexusai.run/docs/features/labelling?utm_source=readme) - tag requests by HTTP header or API key and break down usage by label
- [Usage API](https://aigateway.nexusai.run/docs/advanced/usage-api?utm_source=readme) - let clients inspect their own usage, budget, and rate-limit headroom

### Use the APIs your applications already speak

- [OpenAI-compatible APIs](https://aigateway.nexusai.run/docs/advanced/api-endpoints?utm_source=readme) - Chat Completions, Responses, Conversations, embeddings, audio, and image endpoints
- [Anthropic Messages API](https://aigateway.nexusai.run/docs/advanced/anthropic-messages-api?utm_source=readme) - route native Anthropic requests while preserving provider-specific message state
- [Provider passthrough](https://aigateway.nexusai.run/docs/features/passthrough-api?utm_source=readme) - access provider-native APIs under `/p/{provider}/...` with gateway authentication and tracking
- [Provider replay state](https://aigateway.nexusai.run/docs/advanced/extra-content?utm_source=readme) - preserve thinking blocks, thought signatures, and other provider state across turns and providers
- [Audio and image APIs](https://aigateway.nexusai.run/docs/advanced/audio-api?utm_source=readme) - text-to-speech, transcription, image generation, and image editing with the same access controls and usage tracking

### Extend and operate the platform

- [MCP gateway](https://aigateway.nexusai.run/docs/features/mcp-gateway?utm_source=readme) - aggregate MCP servers behind one authenticated endpoint
- [Guardrails](https://aigateway.nexusai.run/docs/advanced/guardrails?utm_source=readme) - enforce request and response policies at the gateway
- [Plugins](https://aigateway.nexusai.run/docs/advanced/plugins?utm_source=readme) - add guardrails, filters, header edits, and routing strategies through one plugin contract
- [Workflows](https://aigateway.nexusai.run/docs/advanced/workflows?utm_source=readme) - apply versioned per-request policies for caching, budgets, audit logs, guardrails, and failover
- [Observability](https://aigateway.nexusai.run/docs/guides/prometheus-metrics?utm_source=readme) - Prometheus metrics, [OpenTelemetry](https://aigateway.nexusai.run/docs/guides/opentelemetry?utm_source=readme) traces, audit logs, and live request streaming
- [Playground](https://aigateway.nexusai.run/docs/features/playground?utm_source=readme) - test models and virtual models from the dashboard and inspect request and response JSON
- [Admin API and Swagger UI](https://aigateway.nexusai.run/docs/advanced/admin-endpoints?utm_source=readme) - manage the gateway programmatically and explore its API surface

### NEXUS AI Gateway Pro

- [Pro capabilities](https://nexusai.run/ai-gateway) - prompt compression, intelligent routing, and SSO for teams that need deeper optimization and control


## Roadmap

See the [roadmap and product documentation](http://nexusai.run/docs) for upcoming NEXUS AI Gateway releases.

For the Pro version, visit [NEXUS AI Gateway Pro](https://nexusai.run/ai-gateway).

## Sponsors
