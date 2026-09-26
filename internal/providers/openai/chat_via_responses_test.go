package openai

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/nexusrun/nexus_aigateway/internal/core"
	"github.com/nexusrun/nexus_aigateway/internal/llmclient"
)

const viaResponsesTools = `"tools":[{"type":"function","function":{"name":"list_files","description":"List files","parameters":{"type":"object","properties":{}}}}]`

func viaResponsesProvider(t *testing.T, handler http.HandlerFunc) *Provider {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	provider := NewWithHTTPClient("test-api-key", nil, llmclient.Hooks{})
	provider.SetBaseURL(server.URL)
	return provider
}

func decodeChatRequest(t *testing.T, body string) *core.ChatRequest {
	t.Helper()
	var req core.ChatRequest
	if err := json.Unmarshal([]byte(body), &req); err != nil {
		t.Fatalf("unmarshal chat request: %v", err)
	}
	return &req
}

func TestNeedsResponsesForTools(t *testing.T) {
	tests := []struct {
		body string
		want bool
	}{
		{`{"model":"gpt-6-astra",` + viaResponsesTools + `}`, true},
		{`{"model":"gpt-6-astra-2026-09-01",` + viaResponsesTools + `}`, true},
		{`{"model":"gpt-6-astra"}`, false},
		{`{"model":"gpt-6-sol",` + viaResponsesTools + `}`, false},
		{`{"model":"gpt-6-astral",` + viaResponsesTools + `}`, false},
	}
	for _, tt := range tests {
		if got := needsResponsesForTools(decodeChatRequest(t, tt.body)); got != tt.want {
			t.Errorf("needsResponsesForTools(%s) = %v, want %v", tt.body, got, tt.want)
		}
	}
}

func TestChatCompletion_AstraToolsGoThroughResponses(t *testing.T) {
	var upstream map[string]any
	provider := viaResponsesProvider(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/responses" {
			t.Fatalf("path = %s, want /responses", r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		if err := json.Unmarshal(body, &upstream); err != nil {
			t.Fatalf("upstream body: %v", err)
		}
		_, _ = w.Write([]byte(`{
			"id": "resp_1", "created_at": 1700000000, "model": "gpt-6-astra", "status": "completed",
			"output": [
				{"type": "reasoning", "id": "rs_1", "summary": []},
				{"type": "message", "role": "assistant", "content": [{"type": "output_text", "text": "Listing now."}]},
				{"type": "function_call", "id": "fc_1", "call_id": "call_9", "name": "list_files", "arguments": "{}"}
			],
			"usage": {"input_tokens": 40, "output_tokens": 12, "total_tokens": 52,
				"input_tokens_details": {"cached_tokens": 8}, "output_tokens_details": {"reasoning_tokens": 5}}
		}`))
	})

	req := decodeChatRequest(t, `{
		"model": "gpt-6-astra",
		"max_tokens": 900,
		"temperature": 0.2,
		"reasoning": {"effort": "high"},
		"tool_choice": {"type": "function", "function": {"name": "list_files"}},
		"messages": [
			{"role": "system", "content": "SYS"},
			{"role": "user", "content": [{"type": "text", "text": "look"}, {"type": "image_url", "image_url": {"url": "data:image/png;base64,AA"}}]},
			{"role": "assistant", "content": null, "tool_calls": [{"id": "call_1", "type": "function", "function": {"name": "list_files", "arguments": "{}"}}]},
			{"role": "tool", "tool_call_id": "call_1", "content": "a.ts"}
		],
		`+viaResponsesTools+`}`)
	resp, err := provider.ChatCompletion(context.Background(), req)
	if err != nil {
		t.Fatalf("ChatCompletion: %v", err)
	}

	for key, want := range map[string]string{
		"model":             `"gpt-6-astra"`,
		"store":             `false`,
		"max_output_tokens": `900`,
		"reasoning":         `{"effort":"high"}`,
		"tool_choice":       `{"name":"list_files","type":"function"}`,
		"tools":             `[{"description":"List files","name":"list_files","parameters":{"properties":{},"type":"object"},"strict":false,"type":"function"}]`,
		"input": `[{"content":"SYS","role":"system"},` +
			`{"content":[{"text":"look","type":"input_text"},{"detail":"auto","image_url":"data:image/png;base64,AA","type":"input_image"}],"role":"user"},` +
			`{"arguments":"{}","call_id":"call_1","name":"list_files","type":"function_call"},` +
			`{"call_id":"call_1","output":"a.ts","type":"function_call_output"}]`,
	} {
		got, _ := json.Marshal(upstream[key])
		if string(got) != want {
			t.Errorf("upstream %s = %s, want %s", key, got, want)
		}
	}
	for _, key := range []string{"temperature", "messages", "max_tokens", "reasoning_effort"} {
		if _, ok := upstream[key]; ok {
			t.Errorf("upstream body should not contain %s", key)
		}
	}

	choice := resp.Choices[0]
	if choice.FinishReason != "tool_calls" || choice.Message.Content != "Listing now." {
		t.Errorf("choice = %+v", choice)
	}
	if len(choice.Message.ToolCalls) != 1 || choice.Message.ToolCalls[0].ID != "call_9" || choice.Message.ToolCalls[0].Function.Name != "list_files" {
		t.Errorf("tool calls = %+v", choice.Message.ToolCalls)
	}
	if resp.Usage.PromptTokens != 40 || resp.Usage.CompletionTokens != 12 || resp.Usage.TotalTokens != 52 {
		t.Errorf("usage = %+v", resp.Usage)
	}
	if resp.Usage.PromptTokensDetails == nil || resp.Usage.PromptTokensDetails.CachedTokens != 8 {
		t.Errorf("prompt token details = %+v", resp.Usage.PromptTokensDetails)
	}
}

func TestChatCompletion_OtherRequestsStayOnChatCompletions(t *testing.T) {
	for _, body := range []string{
		`{"model":"gpt-6-astra","messages":[{"role":"user","content":"hi"}]}`,
		`{"model":"gpt-6-sol","messages":[{"role":"user","content":"hi"}],` + viaResponsesTools + `}`,
	} {
		provider := viaResponsesProvider(t, func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/chat/completions" {
				t.Fatalf("path = %s, want /chat/completions", r.URL.Path)
			}
			_, _ = w.Write([]byte(`{"id":"c","object":"chat.completion","model":"m","choices":[{"index":0,"message":{"role":"assistant","content":"hi"},"finish_reason":"stop"}]}`))
		})
		if _, err := provider.ChatCompletion(context.Background(), decodeChatRequest(t, body)); err != nil {
			t.Fatalf("ChatCompletion(%s): %v", body, err)
		}
	}
}

func TestChatCompletion_AstraRejectsUnsupportedFields(t *testing.T) {
	provider := viaResponsesProvider(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("no upstream call expected")
	})
	_, err := provider.ChatCompletion(context.Background(), decodeChatRequest(t,
		`{"model":"gpt-6-astra","stop":["END"],"messages":[{"role":"user","content":"hi"}],`+viaResponsesTools+`}`))
	if err == nil || !strings.Contains(err.Error(), "stop is not supported") {
		t.Fatalf("err = %v, want unsupported stop", err)
	}
}

func TestChatCompletion_AstraLengthAndFailure(t *testing.T) {
	provider := viaResponsesProvider(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"id":"r","status":"incomplete","incomplete_details":{"reason":"max_output_tokens"},
			"output":[{"type":"message","content":[{"type":"output_text","text":"partial"}]}]}`))
	})
	req := decodeChatRequest(t, `{"model":"gpt-6-astra","messages":[{"role":"user","content":"hi"}],`+viaResponsesTools+`}`)
	resp, err := provider.ChatCompletion(context.Background(), req)
	if err != nil || resp.Choices[0].FinishReason != "length" || resp.Model != "gpt-6-astra" {
		t.Fatalf("resp = %+v, err = %v", resp, err)
	}

	failing := viaResponsesProvider(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"id":"r","status":"failed","error":{"message":"boom"},"output":[]}`))
	})
	if _, err := failing.ChatCompletion(context.Background(), req); err == nil || !strings.Contains(err.Error(), "boom") {
		t.Fatalf("err = %v, want boom", err)
	}
}

func TestStreamChatCompletion_AstraToolsGoThroughResponses(t *testing.T) {
	events := []string{
		`{"type":"response.created","response":{"id":"resp_7","created_at":1700000000,"model":"gpt-6-astra"}}`,
		`{"type":"response.output_item.added","item":{"id":"rs_1","type":"reasoning"}}`,
		`{"type":"response.output_text.delta","delta":"Hel"}`,
		`{"type":"response.output_text.delta","delta":"lo"}`,
		`{"type":"response.output_item.added","item":{"id":"fc_1","type":"function_call","call_id":"call_a","name":"list_files"}}`,
		`{"type":"response.function_call_arguments.delta","item_id":"fc_1","delta":"{\"dir\":"}`,
		`{"type":"response.function_call_arguments.delta","item_id":"fc_1","delta":"\"src\"}"}`,
		`{"type":"response.completed","response":{"id":"resp_7","status":"completed","usage":{"input_tokens":10,"output_tokens":5,"total_tokens":15}}}`,
	}
	var upstream map[string]any
	provider := viaResponsesProvider(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/responses" {
			t.Fatalf("path = %s, want /responses", r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &upstream)
		w.Header().Set("Content-Type", "text/event-stream")
		for _, event := range events {
			var typed struct {
				Type string `json:"type"`
			}
			_ = json.Unmarshal([]byte(event), &typed)
			_, _ = io.WriteString(w, "event: "+typed.Type+"\ndata: "+event+"\n\n")
		}
	})
	req := decodeChatRequest(t, `{"model":"gpt-6-astra","stream_options":{"include_usage":true},"messages":[{"role":"user","content":"hi"}],`+viaResponsesTools+`}`)
	stream, err := provider.StreamChatCompletion(context.Background(), req)
	if err != nil {
		t.Fatalf("StreamChatCompletion: %v", err)
	}
	raw, err := io.ReadAll(stream)
	_ = stream.Close()
	if err != nil {
		t.Fatalf("read stream: %v", err)
	}
	if upstream["stream"] != true || upstream["stream_options"] != nil {
		t.Errorf("upstream stream fields = %v / %v", upstream["stream"], upstream["stream_options"])
	}

	var content, args, finish, callID string
	var usage map[string]any
	lines := strings.Split(strings.TrimSpace(string(raw)), "\n\n")
	if lines[len(lines)-1] != "data: [DONE]" {
		t.Fatalf("stream should end with [DONE]: %q", raw)
	}
	for _, line := range lines[:len(lines)-1] {
		var chunk struct {
			ID      string         `json:"id"`
			Object  string         `json:"object"`
			Usage   map[string]any `json:"usage"`
			Choices []struct {
				Delta struct {
					Content   string `json:"content"`
					ToolCalls []struct {
						ID       string `json:"id"`
						Function struct {
							Arguments string `json:"arguments"`
						} `json:"function"`
					} `json:"tool_calls"`
				} `json:"delta"`
				FinishReason *string `json:"finish_reason"`
			} `json:"choices"`
		}
		if err := json.Unmarshal([]byte(strings.TrimPrefix(line, "data: ")), &chunk); err != nil {
			t.Fatalf("chunk %q: %v", line, err)
		}
		if chunk.ID != "resp_7" || chunk.Object != "chat.completion.chunk" {
			t.Errorf("chunk id/object = %s/%s", chunk.ID, chunk.Object)
		}
		if chunk.Usage != nil {
			usage = chunk.Usage
		}
		for _, choice := range chunk.Choices {
			content += choice.Delta.Content
			for _, call := range choice.Delta.ToolCalls {
				if call.ID != "" {
					callID = call.ID
				}
				args += call.Function.Arguments
			}
			if choice.FinishReason != nil {
				finish = *choice.FinishReason
			}
		}
	}
	if content != "Hello" || callID != "call_a" || args != `{"dir":"src"}` || finish != "tool_calls" {
		t.Errorf("content=%q callID=%q args=%q finish=%q", content, callID, args, finish)
	}
	if usage == nil || usage["total_tokens"] != float64(15) {
		t.Errorf("usage = %v", usage)
	}
}

func TestStreamChatCompletion_AstraStreamEndsEarly(t *testing.T) {
	provider := viaResponsesProvider(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, "data: {\"type\":\"response.failed\",\"response\":{\"error\":{\"message\":\"boom\"}}}\n\n")
	})
	stream, err := provider.StreamChatCompletion(context.Background(), decodeChatRequest(t,
		`{"model":"gpt-6-astra","messages":[{"role":"user","content":"hi"}],`+viaResponsesTools+`}`))
	if err != nil {
		t.Fatalf("StreamChatCompletion: %v", err)
	}
	raw, _ := io.ReadAll(stream)
	_ = stream.Close()
	if !strings.Contains(string(raw), `"message":"boom"`) || !strings.HasSuffix(string(raw), "data: [DONE]\n\n") {
		t.Fatalf("stream = %q", raw)
	}
}
