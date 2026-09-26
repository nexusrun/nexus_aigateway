package openai

import (
	"encoding/json"
	"testing"

	"github.com/nexusrun/nexus_aigateway/internal/core"
)

func TestAdaptChatRequest_ToolsOnReasoningModels(t *testing.T) {
	const tools = `"tools":[{"type":"function","function":{"name":"list_files","parameters":{"type":"object","properties":{}}}}]`
	tests := []struct {
		name string
		body string
		want string // reasoning_effort sent upstream; "" means absent
	}{
		{"gpt-5.6 with tools gets none", `{"model":"gpt-5.6-luna",` + tools + `}`, `"none"`},
		{"bare gpt-5.6", `{"model":"gpt-5.6",` + tools + `}`, `"none"`},
		{"gpt-6 with tools gets none", `{"model":"gpt-6-sol",` + tools + `}`, `"none"`},
		{"empty nested effort counts as unset", `{"model":"gpt-6-luna","reasoning":{},` + tools + `}`, `"none"`},
		{"explicit flat effort is kept", `{"model":"gpt-5.6-luna","reasoning_effort":"low",` + tools + `}`, `"low"`},
		{"explicit nested effort is kept", `{"model":"gpt-5.6-luna","reasoning":{"effort":"high"},` + tools + `}`, `"high"`},
		{"no tools, no change", `{"model":"gpt-5.6-luna"}`, ``},
		{"gpt-6-astra cannot use none", `{"model":"gpt-6-astra",` + tools + `}`, ``},
		{"older gpt-5 accepts tools as is", `{"model":"gpt-5.5",` + tools + `}`, ``},
		{"unrelated prefix", `{"model":"gpt-60",` + tools + `}`, ``},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req core.ChatRequest
			if err := json.Unmarshal([]byte(tt.body), &req); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			adapted, err := adaptChatRequest(&req)
			if err != nil {
				t.Fatalf("adaptChatRequest: %v", err)
			}
			body, err := json.Marshal(adapted)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			var raw map[string]json.RawMessage
			if err := json.Unmarshal(body, &raw); err != nil {
				t.Fatalf("unmarshal body: %v", err)
			}
			if got := string(raw["reasoning_effort"]); got != tt.want {
				t.Errorf("reasoning_effort = %q, want %q (body %s)", got, tt.want, body)
			}
			if _, ok := raw["reasoning"]; ok {
				t.Errorf("nested reasoning must not reach OpenAI chat: %s", body)
			}
		})
	}
}
