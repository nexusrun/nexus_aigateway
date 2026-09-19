// Pure-logic tests for the provider display label helper.
import test from "node:test";
import assert from "node:assert/strict";

import { providerLabel } from "../src/lib/utils/providerLabel.js";

test("providerLabel maps registered types to their brand spelling", () => {
  for (const [input, want] of [
    ["openai", "OpenAI"],
    ["anthropic", "Anthropic"],
    ["xai", "xAI"],
    ["vllm", "vLLM"],
    ["llamacpp", "llama.cpp"],
    ["deepseek", "DeepSeek"],
    ["opencode_go", "OpenCode Go"],
    ["bedrock-mantle", "Amazon Bedrock Mantle"],
    [" OpenAI ", "OpenAI"],
  ]) {
    assert.equal(providerLabel(input), want, input);
  }
});

test("providerLabel capitalizes custom provider names", () => {
  assert.equal(providerLabel("prod-openai"), "Prod-openai");
  assert.equal(providerLabel("alpha"), "Alpha");
  assert.equal(providerLabel("Already"), "Already");
});

test("providerLabel returns empty for empty input", () => {
  assert.equal(providerLabel(""), "");
  assert.equal(providerLabel(null), "");
  assert.equal(providerLabel(undefined), "");
});
