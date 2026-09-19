// Human-readable provider names for display only.
//
// Provider names and types are routing identifiers (config keys, "provider/
// model" selectors, filter values), so they stay lowercase everywhere they are
// matched or sent back to the gateway. providerLabel only changes what the
// dashboard renders: registered provider types map to their brand spelling
// (see run/providers.go and docs/providers/*.mdx titles), and any other name
// gets its first letter capitalized.

const PROVIDER_LABELS = {
  anthropic: "Anthropic",
  azure: "Azure OpenAI",
  bailian: "Alibaba Model Studio",
  bedrock: "Amazon Bedrock",
  "bedrock-mantle": "Amazon Bedrock Mantle",
  chatgpt: "ChatGPT",
  chutes: "Chutes",
  cohere: "Cohere",
  deepseek: "DeepSeek",
  elevenlabs: "ElevenLabs",
  fireworks: "Fireworks AI",
  gemini: "Google Gemini",
  groq: "Groq",
  hetzner: "Hetzner",
  kilo: "Kilo Code",
  kimicode: "Kimi Code",
  llamacpp: "llama.cpp",
  llmd: "llm-d",
  meta: "Meta",
  minimax: "MiniMax",
  ollama: "Ollama",
  openai: "OpenAI",
  "opencode-go": "OpenCode Go",
  openrouter: "OpenRouter",
  oracle: "Oracle GenAI",
  sglang: "SGLang",
  vertex: "Google Vertex AI",
  vllm: "vLLM",
  xai: "xAI",
  xiaomi: "Xiaomi MiMo",
  zai: "Z.ai",
};

export function providerLabel(name) {
  const raw = String(name || "").trim();
  if (!raw) return "";
  const key = raw.toLowerCase().replaceAll("_", "-");
  if (Object.prototype.hasOwnProperty.call(PROVIDER_LABELS, key)) {
    return PROVIDER_LABELS[key];
  }
  return raw.charAt(0).toUpperCase() + raw.slice(1);
}
