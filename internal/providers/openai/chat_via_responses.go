package openai

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/goccy/go-json"

	"github.com/nexusrun/nexus_aigateway/internal/core"
	"github.com/nexusrun/nexus_aigateway/internal/llmclient"
)

// gpt-6-astra reasons by default and rejects both function tools and
// reasoning_effort "none" on /v1/chat/completions, so a Chat Completions
// request with tools cannot reach it there. The gateway sends those requests
// to /v1/responses and translates the answer back to Chat Completions.

// needsResponsesForTools reports whether a chat request must go through
// /v1/responses.
func needsResponsesForTools(req *core.ChatRequest) bool {
	if req == nil || len(req.Tools) == 0 {
		return false
	}
	rest, ok := strings.CutPrefix(strings.ToLower(strings.TrimSpace(req.Model)), "gpt-6-astra")
	return ok && (rest == "" || rest[0] == '-' || rest[0] == '.')
}

// ChatCompletion serves tool requests for gpt-6-astra through /v1/responses
// and everything else through Chat Completions.
func (p *Provider) ChatCompletion(ctx context.Context, req *core.ChatRequest) (*core.ChatResponse, error) {
	if !needsResponsesForTools(req) {
		return p.CompatibleProvider.ChatCompletion(ctx, req)
	}
	adapted, err := p.adaptedChatRequest(req)
	if err != nil {
		return nil, err
	}
	body, err := chatToResponsesBody(adapted, false)
	if err != nil {
		return nil, err
	}
	var raw json.RawMessage
	if err := p.Do(ctx, llmclient.Request{
		Method:    http.MethodPost,
		Endpoint:  "/responses",
		Operation: llmclient.OperationChat,
		Model:     req.Model,
		Body:      body,
	}, &raw); err != nil {
		return nil, err
	}
	return responsesToChatResponse(raw, req.Model)
}

// StreamChatCompletion is the streaming counterpart of ChatCompletion.
func (p *Provider) StreamChatCompletion(ctx context.Context, req *core.ChatRequest) (io.ReadCloser, error) {
	if !needsResponsesForTools(req) {
		return p.CompatibleProvider.StreamChatCompletion(ctx, req)
	}
	adapted, err := p.adaptedChatRequest(req)
	if err != nil {
		return nil, err
	}
	streamReq := adapted.WithStreaming()
	body, err := chatToResponsesBody(streamReq, true)
	if err != nil {
		return nil, err
	}
	stream, err := p.client.DoStream(ctx, p.prepareRequest(llmclient.Request{
		Method:    http.MethodPost,
		Endpoint:  "/responses",
		Operation: llmclient.OperationChat,
		Model:     req.Model,
		Body:      body,
	}))
	if err != nil {
		return nil, err
	}
	includeUsage := streamReq.StreamOptions != nil && streamReq.StreamOptions.IncludeUsage
	return newResponsesChatStream(stream, req.Model, includeUsage), nil
}

// ---------------------------------------------------------------------------
// Request: Chat Completions -> Responses

type viaResponsesMessage struct {
	Role       string          `json:"role"`
	Content    json.RawMessage `json:"content"`
	ToolCallID string          `json:"tool_call_id"`
	ToolCalls  []struct {
		ID       string `json:"id"`
		Function struct {
			Name      string `json:"name"`
			Arguments string `json:"arguments"`
		} `json:"function"`
	} `json:"tool_calls"`
}

// Fields dropped on the way to /v1/responses: the gateway routing hint,
// options Responses has no equivalent for, and temperature, which reasoning
// models reject on Chat Completions too.
var viaResponsesDropped = map[string]bool{"provider": true, "stream_options": true, "temperature": true}

func unsupportedViaResponses(field string) error {
	return core.NewInvalidRequestError(fmt.Sprintf(
		"%s is not supported for gpt-6-astra tool calls: the gateway serves them through /v1/responses", field), nil)
}

func chatToResponsesBody(req *core.ChatRequest, stream bool) (map[string]any, error) {
	encoded, err := json.Marshal(req)
	if err != nil {
		return nil, core.NewInvalidRequestError("failed to encode chat request: "+err.Error(), err)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &fields); err != nil {
		return nil, core.NewInvalidRequestError("failed to decode chat request: "+err.Error(), err)
	}
	body := map[string]any{"model": req.Model, "store": false}
	if stream {
		body["stream"] = true
	}
	names := make([]string, 0, len(fields))
	for name := range fields {
		names = append(names, name)
	}
	sort.Strings(names) // deterministic error for requests with several unsupported fields
	for _, name := range names {
		raw := fields[name]
		switch name {
		case "model", "stream":
		case "messages":
			input, err := messagesToResponsesInput(raw)
			if err != nil {
				return nil, err
			}
			body["input"] = input
		case "tools":
			tools, err := toolsToResponses(raw)
			if err != nil {
				return nil, err
			}
			body["tools"] = tools
		case "tool_choice":
			choice, err := toolChoiceToResponses(raw)
			if err != nil {
				return nil, err
			}
			body["tool_choice"] = choice
		case "max_tokens", "max_completion_tokens":
			body["max_output_tokens"] = raw
		case "reasoning_effort":
			body["reasoning"] = map[string]json.RawMessage{"effort": raw}
		case "response_format":
			format, err := responseFormatToResponses(raw)
			if err != nil {
				return nil, err
			}
			body["text"] = map[string]any{"format": format}
		case "parallel_tool_calls", "top_p", "user", "service_tier", "metadata", "store",
			"prompt_cache_key", "safety_identifier":
			body[name] = raw
		case "n":
			if strings.TrimSpace(string(raw)) != "1" {
				return nil, unsupportedViaResponses("n")
			}
		default:
			if !viaResponsesDropped[name] {
				return nil, unsupportedViaResponses(name)
			}
		}
	}
	return body, nil
}

func messagesToResponsesInput(raw json.RawMessage) ([]any, error) {
	var messages []viaResponsesMessage
	if err := json.Unmarshal(raw, &messages); err != nil {
		return nil, core.NewInvalidRequestError("invalid messages: "+err.Error(), err)
	}
	input := make([]any, 0, len(messages))
	for _, m := range messages {
		switch m.Role {
		case "tool":
			output, err := contentText(m.Content)
			if err != nil {
				return nil, err
			}
			input = append(input, map[string]any{"type": "function_call_output", "call_id": m.ToolCallID, "output": output})
		case "assistant":
			if content, err := contentToResponses(m.Content, "output_text"); err != nil {
				return nil, err
			} else if content != nil {
				input = append(input, map[string]any{"role": "assistant", "content": content})
			}
			for _, call := range m.ToolCalls {
				input = append(input, map[string]any{
					"type":      "function_call",
					"call_id":   call.ID,
					"name":      call.Function.Name,
					"arguments": call.Function.Arguments,
				})
			}
		case "system", "developer", "user":
			content, err := contentToResponses(m.Content, "input_text")
			if err != nil {
				return nil, err
			}
			if content == nil {
				content = ""
			}
			input = append(input, map[string]any{"role": m.Role, "content": content})
		default:
			return nil, unsupportedViaResponses(fmt.Sprintf("message role %q", m.Role))
		}
	}
	return input, nil
}

// contentToResponses converts chat message content; nil means no content.
func contentToResponses(raw json.RawMessage, textType string) (any, error) {
	if isNullJSON(raw) {
		return nil, nil
	}
	var text string
	if err := json.Unmarshal(raw, &text); err == nil {
		return text, nil
	}
	var parts []map[string]any
	if err := json.Unmarshal(raw, &parts); err != nil {
		return nil, core.NewInvalidRequestError("message content must be a string or an array of parts", err)
	}
	out := make([]any, 0, len(parts))
	for _, part := range parts {
		switch part["type"] {
		case "text":
			out = append(out, map[string]any{"type": textType, "text": part["text"]})
		case "image_url":
			image, _ := part["image_url"].(map[string]any)
			url, _ := image["url"].(string)
			detail, _ := image["detail"].(string)
			if detail == "" {
				detail = "auto"
			}
			out = append(out, map[string]any{"type": "input_image", "image_url": url, "detail": detail})
		case "file":
			file, _ := part["file"].(map[string]any)
			converted := map[string]any{"type": "input_file"}
			for _, key := range []string{"file_id", "file_data", "filename"} {
				if value, ok := file[key]; ok {
					converted[key] = value
				}
			}
			out = append(out, converted)
		default:
			return nil, unsupportedViaResponses(fmt.Sprintf("content part %q", part["type"]))
		}
	}
	return out, nil
}

// contentText flattens tool output to the string Responses expects.
func contentText(raw json.RawMessage) (string, error) {
	if isNullJSON(raw) {
		return "", nil
	}
	var text string
	if err := json.Unmarshal(raw, &text); err == nil {
		return text, nil
	}
	var parts []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}
	if err := json.Unmarshal(raw, &parts); err != nil {
		return "", core.NewInvalidRequestError("tool message content must be a string or text parts", err)
	}
	var b strings.Builder
	for _, part := range parts {
		if part.Type != "text" {
			return "", unsupportedViaResponses(fmt.Sprintf("tool content part %q", part.Type))
		}
		b.WriteString(part.Text)
	}
	return b.String(), nil
}

func toolsToResponses(raw json.RawMessage) ([]any, error) {
	var tools []struct {
		Type     string `json:"type"`
		Function struct {
			Name        string          `json:"name"`
			Description string          `json:"description,omitempty"`
			Parameters  json.RawMessage `json:"parameters,omitempty"`
			Strict      *bool           `json:"strict,omitempty"`
		} `json:"function"`
	}
	if err := json.Unmarshal(raw, &tools); err != nil {
		return nil, core.NewInvalidRequestError("invalid tools: "+err.Error(), err)
	}
	out := make([]any, 0, len(tools))
	for _, tool := range tools {
		if tool.Type != "function" {
			return nil, unsupportedViaResponses(fmt.Sprintf("tool type %q", tool.Type))
		}
		// Chat tools are non-strict unless they say otherwise; Responses
		// defaults to strict, which rejects many ordinary schemas.
		strict := tool.Function.Strict != nil && *tool.Function.Strict
		converted := map[string]any{"type": "function", "name": tool.Function.Name, "strict": strict}
		if tool.Function.Description != "" {
			converted["description"] = tool.Function.Description
		}
		if !isNullJSON(tool.Function.Parameters) {
			converted["parameters"] = tool.Function.Parameters
		}
		out = append(out, converted)
	}
	return out, nil
}

func toolChoiceToResponses(raw json.RawMessage) (any, error) {
	var mode string
	if err := json.Unmarshal(raw, &mode); err == nil {
		return mode, nil
	}
	var choice struct {
		Type     string `json:"type"`
		Function struct {
			Name string `json:"name"`
		} `json:"function"`
	}
	if err := json.Unmarshal(raw, &choice); err != nil || choice.Type != "function" {
		return nil, unsupportedViaResponses("this tool_choice")
	}
	return map[string]any{"type": "function", "name": choice.Function.Name}, nil
}

func responseFormatToResponses(raw json.RawMessage) (any, error) {
	var format struct {
		Type       string                     `json:"type"`
		JSONSchema map[string]json.RawMessage `json:"json_schema"`
	}
	if err := json.Unmarshal(raw, &format); err != nil {
		return nil, core.NewInvalidRequestError("invalid response_format: "+err.Error(), err)
	}
	if format.Type != "json_schema" {
		return map[string]any{"type": format.Type}, nil
	}
	converted := map[string]any{"type": "json_schema"}
	for key, value := range format.JSONSchema {
		converted[key] = value
	}
	return converted, nil
}

func isNullJSON(raw json.RawMessage) bool {
	trimmed := strings.TrimSpace(string(raw))
	return trimmed == "" || trimmed == "null"
}

// ---------------------------------------------------------------------------
// Response: Responses -> Chat Completions

type viaResponsesResult struct {
	ID        string `json:"id"`
	CreatedAt int64  `json:"created_at"`
	Model     string `json:"model"`
	Status    string `json:"status"`
	Error     *struct {
		Message string `json:"message"`
	} `json:"error"`
	IncompleteDetails *struct {
		Reason string `json:"reason"`
	} `json:"incomplete_details"`
	Output []struct {
		Type      string `json:"type"`
		CallID    string `json:"call_id"`
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
		Content   []struct {
			Type    string `json:"type"`
			Text    string `json:"text"`
			Refusal string `json:"refusal"`
		} `json:"content"`
	} `json:"output"`
	Usage *viaResponsesUsage `json:"usage"`
}

type viaResponsesUsage struct {
	InputTokens        int `json:"input_tokens"`
	OutputTokens       int `json:"output_tokens"`
	TotalTokens        int `json:"total_tokens"`
	InputTokensDetails *struct {
		CachedTokens int `json:"cached_tokens"`
	} `json:"input_tokens_details"`
	OutputTokensDetails *struct {
		ReasoningTokens int `json:"reasoning_tokens"`
	} `json:"output_tokens_details"`
}

func (u *viaResponsesUsage) chat() map[string]any {
	if u == nil {
		return map[string]any{"prompt_tokens": 0, "completion_tokens": 0, "total_tokens": 0}
	}
	usage := map[string]any{
		"prompt_tokens":     u.InputTokens,
		"completion_tokens": u.OutputTokens,
		"total_tokens":      u.TotalTokens,
	}
	if u.InputTokensDetails != nil {
		usage["prompt_tokens_details"] = map[string]any{"cached_tokens": u.InputTokensDetails.CachedTokens}
	}
	if u.OutputTokensDetails != nil {
		usage["completion_tokens_details"] = map[string]any{"reasoning_tokens": u.OutputTokensDetails.ReasoningTokens}
	}
	return usage
}

func viaResponsesFinishReason(sawToolCalls bool, incompleteReason string) string {
	switch {
	case incompleteReason == "max_output_tokens":
		return "length"
	case incompleteReason == "content_filter":
		return "content_filter"
	case sawToolCalls:
		return "tool_calls"
	default:
		return "stop"
	}
}

func responsesToChatResponse(raw json.RawMessage, requestModel string) (*core.ChatResponse, error) {
	var result viaResponsesResult
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, core.NewProviderError("openai", http.StatusBadGateway, "failed to parse Responses result: "+err.Error(), err)
	}
	if result.Status == "failed" {
		message := "OpenAI response failed"
		if result.Error != nil && result.Error.Message != "" {
			message = result.Error.Message
		}
		return nil, core.NewProviderError("openai", http.StatusBadGateway, message, nil)
	}
	var text, refusal strings.Builder
	var toolCalls []map[string]any
	for _, item := range result.Output {
		switch item.Type {
		case "message":
			for _, part := range item.Content {
				text.WriteString(part.Text)
				refusal.WriteString(part.Refusal)
			}
		case "function_call":
			toolCalls = append(toolCalls, map[string]any{
				"id":       item.CallID,
				"type":     "function",
				"function": map[string]any{"name": item.Name, "arguments": item.Arguments},
			})
		}
	}
	message := map[string]any{"role": "assistant", "content": nil}
	if text.Len() > 0 {
		message["content"] = text.String()
	}
	if refusal.Len() > 0 {
		message["refusal"] = refusal.String()
	}
	if len(toolCalls) > 0 {
		message["tool_calls"] = toolCalls
	}
	incomplete := ""
	if result.IncompleteDetails != nil {
		incomplete = result.IncompleteDetails.Reason
	}
	model := result.Model
	if model == "" {
		model = requestModel
	}
	encoded, err := json.Marshal(map[string]any{
		"id":      result.ID,
		"object":  "chat.completion",
		"created": result.CreatedAt,
		"model":   model,
		"choices": []any{map[string]any{
			"index":         0,
			"message":       message,
			"finish_reason": viaResponsesFinishReason(len(toolCalls) > 0, incomplete),
		}},
		"usage": result.Usage.chat(),
	})
	if err != nil {
		return nil, core.NewProviderError("openai", http.StatusBadGateway, "failed to encode chat response", err)
	}
	var resp core.ChatResponse
	if err := json.Unmarshal(encoded, &resp); err != nil {
		return nil, core.NewProviderError("openai", http.StatusBadGateway, "failed to build chat response", err)
	}
	return &resp, nil
}

// ---------------------------------------------------------------------------
// Streaming: Responses events -> Chat Completions chunks

type responsesChatStream struct {
	reader *io.PipeReader
	body   io.ReadCloser
}

func newResponsesChatStream(body io.ReadCloser, model string, includeUsage bool) io.ReadCloser {
	reader, writer := io.Pipe()
	go convertResponsesChatStream(body, writer, model, includeUsage)
	return &responsesChatStream{reader: reader, body: body}
}

func (s *responsesChatStream) Read(p []byte) (int, error) { return s.reader.Read(p) }

func (s *responsesChatStream) Close() error {
	_ = s.reader.Close()
	return s.body.Close()
}

type responsesChatStreamState struct {
	model        string
	includeUsage bool
	id           string
	created      int64
	roleSent     bool
	finished     bool
	calls        map[string]int // Responses item id -> chat tool call index
}

type viaResponsesEvent struct {
	Type     string              `json:"type"`
	Delta    string              `json:"delta"`
	ItemID   string              `json:"item_id"`
	Message  string              `json:"message"`
	Response *viaResponsesResult `json:"response"`
	Item     *struct {
		ID     string `json:"id"`
		Type   string `json:"type"`
		CallID string `json:"call_id"`
		Name   string `json:"name"`
	} `json:"item"`
}

func convertResponsesChatStream(body io.ReadCloser, out *io.PipeWriter, model string, includeUsage bool) {
	defer func() { _ = body.Close() }()
	state := &responsesChatStreamState{model: model, includeUsage: includeUsage, created: time.Now().Unix(), calls: map[string]int{}}
	scanner := bufio.NewScanner(body)
	scanner.Buffer(make([]byte, 0, 64*1024), 8*1024*1024)
	for scanner.Scan() {
		line := strings.TrimRight(scanner.Text(), "\r")
		data, ok := strings.CutPrefix(line, "data:")
		if !ok {
			continue
		}
		if err := state.consume(out, strings.TrimSpace(data)); err != nil {
			finishResponsesChatStream(out, err.Error())
			return
		}
	}
	if err := scanner.Err(); err != nil {
		finishResponsesChatStream(out, "failed to read OpenAI stream: "+err.Error())
		return
	}
	if !state.finished {
		finishResponsesChatStream(out, "OpenAI stream ended before the response completed")
		return
	}
	finishResponsesChatStream(out, "")
}

// finishResponsesChatStream writes an optional error chunk, then [DONE].
func finishResponsesChatStream(out *io.PipeWriter, errMessage string) {
	if errMessage != "" {
		if err := writeViaResponsesChunk(out, map[string]any{"error": map[string]any{
			"type": core.ErrorTypeProvider, "message": errMessage, "param": nil, "code": nil,
		}}); err != nil {
			_ = out.CloseWithError(err)
			return
		}
	}
	if _, err := io.WriteString(out, "data: [DONE]\n\n"); err != nil {
		_ = out.CloseWithError(err)
		return
	}
	_ = out.Close()
}

func (s *responsesChatStreamState) consume(out io.Writer, raw string) error {
	if raw == "" || raw == "[DONE]" || s.finished {
		return nil
	}
	var event viaResponsesEvent
	if err := json.Unmarshal([]byte(raw), &event); err != nil {
		return errors.New("failed to parse OpenAI stream event")
	}
	switch event.Type {
	case "response.created":
		if event.Response != nil {
			s.id = event.Response.ID
			if event.Response.CreatedAt > 0 {
				s.created = event.Response.CreatedAt
			}
			if event.Response.Model != "" {
				s.model = event.Response.Model
			}
		}
		return s.writeDelta(out, map[string]any{"role": "assistant", "content": ""})
	case "response.output_text.delta":
		return s.writeDelta(out, map[string]any{"content": event.Delta})
	case "response.refusal.delta":
		return s.writeDelta(out, map[string]any{"refusal": event.Delta})
	case "response.output_item.added":
		if event.Item == nil || event.Item.Type != "function_call" {
			return nil
		}
		index := len(s.calls)
		s.calls[event.Item.ID] = index
		return s.writeDelta(out, map[string]any{"tool_calls": []any{map[string]any{
			"index":    index,
			"id":       event.Item.CallID,
			"type":     "function",
			"function": map[string]any{"name": event.Item.Name, "arguments": ""},
		}}})
	case "response.function_call_arguments.delta":
		index, ok := s.calls[event.ItemID]
		if !ok {
			return nil
		}
		return s.writeDelta(out, map[string]any{"tool_calls": []any{map[string]any{
			"index":    index,
			"function": map[string]any{"arguments": event.Delta},
		}}})
	case "response.completed", "response.incomplete":
		incomplete := ""
		var usage *viaResponsesUsage
		if event.Response != nil {
			usage = event.Response.Usage
			if event.Response.IncompleteDetails != nil {
				incomplete = event.Response.IncompleteDetails.Reason
			}
		}
		return s.writeFinish(out, viaResponsesFinishReason(len(s.calls) > 0, incomplete), usage)
	case "response.failed":
		message := "OpenAI response failed"
		if event.Response != nil && event.Response.Error != nil && event.Response.Error.Message != "" {
			message = event.Response.Error.Message
		}
		return errors.New(message)
	case "error":
		if event.Message == "" {
			return errors.New("OpenAI stream error")
		}
		return errors.New(event.Message)
	default:
		return nil
	}
}

func (s *responsesChatStreamState) writeDelta(out io.Writer, delta map[string]any) error {
	if !s.roleSent {
		delta["role"] = "assistant"
		s.roleSent = true
	}
	return writeViaResponsesChunk(out, s.chunk([]any{map[string]any{"index": 0, "delta": delta, "finish_reason": nil}}, nil))
}

func (s *responsesChatStreamState) writeFinish(out io.Writer, finish string, usage *viaResponsesUsage) error {
	s.finished = true
	if err := writeViaResponsesChunk(out, s.chunk([]any{map[string]any{"index": 0, "delta": map[string]any{}, "finish_reason": finish}}, nil)); err != nil {
		return err
	}
	if !s.includeUsage {
		return nil
	}
	return writeViaResponsesChunk(out, s.chunk([]any{}, usage.chat()))
}

func (s *responsesChatStreamState) chunk(choices []any, usage map[string]any) map[string]any {
	if s.id == "" {
		s.id = "chatcmpl-" + strconv.FormatInt(s.created, 10)
	}
	chunk := map[string]any{
		"id":      s.id,
		"object":  "chat.completion.chunk",
		"created": s.created,
		"model":   s.model,
		"choices": choices,
	}
	if usage != nil {
		chunk["usage"] = usage
	}
	return chunk
}

func writeViaResponsesChunk(out io.Writer, chunk map[string]any) error {
	body, err := json.Marshal(chunk)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(out, "data: %s\n\n", body)
	return err
}
