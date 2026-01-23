package main

import (
	"context"
	"encoding/json"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/deeplooplabs/ai-gateway/gateway"
	"github.com/deeplooplabs/ai-gateway/hook"
	"github.com/deeplooplabs/ai-gateway/model"
	"github.com/deeplooplabs/ai-gateway/provider"
	"github.com/deeplooplabs/ai-gateway/provider/openai"
)

func main() {
	// Create providers using the new API

	// Example 1: Standard provider (BaseURL doesn't include /v1)
	openAIProvider := provider.NewHTTPProviderWithBaseURLAndPath(
		os.Getenv("OPENAI_BASE_URL"),
		os.Getenv("OPENAI_API_KEY"),
		"/v1",
	)

	// Example 2: Provider with BasePath (when BaseURL already includes /v1)
	// For providers like SiliconFlow where the base URL is "https://api.siliconflow.cn/v1"
	// you need to strip "/v1" from the endpoint to avoid duplication:
	//
	// siliconFlowProvider := provider.NewHTTPProviderWithBaseURLAndPath(
	// 	os.Getenv("SILICONFLOW_BASE_URL"),  // e.g., "https://api.siliconflow.cn/v1"
	// 	os.Getenv("SILICONFLOW_API_KEY"),
	// 	"/v1",  // Strip /v1 from endpoint
	// )

	// Setup model registry with new API
	registry := model.NewMapModelRegistry()
	registry.RegisterWithOptions("gpt-4o", openAIProvider,
		model.WithModelRewrite("deepseek-ai/DeepSeek-V3.2"),
		model.WithPreferredAPI(provider.APITypeChatCompletions),
	)
	registry.Register("gpt-3.5-turbo", openAIProvider)

	// Example: Register SiliconFlow models with BasePath
	// siliconFlowProvider := provider.NewHTTPProviderWithBaseURLAndPath(
	// 	"https://api.siliconflow.cn/v1",
	// 	os.Getenv("SILICONFLOW_API_KEY"),
	// 	"/v1",  // BasePath to strip from endpoint
	// )
	// registry.RegisterWithOptions("Qwen/Qwen2.5-7B-Instruct", siliconFlowProvider,
	// 	model.WithPreferredAPI(provider.APITypeChatCompletions),
	// )

	// Create hooks
	hooks := hook.NewRegistry()
	hooks.Register(&LoggingHook{}, &AuthenticateHook{}, &ErrorHook{})

	// Create gateway
	gw := gateway.New(
		gateway.WithModelRegistry(registry),
		gateway.WithHooks(hooks),
		gateway.WithCORS(gateway.DefaultCORSConfig()),
	)

	// Start server
	slog.Info("AI Gateway listening on :8083")
	log.Fatal(http.ListenAndServe(":8083", gw))
}

type AuthenticateHook struct{}

func (h *AuthenticateHook) Name() string {
	return "authenticate"
}

func (h *AuthenticateHook) Authenticate(ctx context.Context, apiKey string) (bool, string, error) {
	splits := strings.Split(apiKey, ":")
	if len(splits) < 1 {
		return false, "", nil
	}
	var jwt, teamId string
	jwt = splits[0]
	if len(splits) == 2 {
		teamId = splits[1]
	}
	return jwt != "" && teamId != "", teamId, nil
}

var _ hook.AuthenticationHook = new(AuthenticateHook)

// LoggingHook logs all requests
type LoggingHook struct{}

func (h *LoggingHook) BeforeRequest(ctx context.Context, req *openai.ChatCompletionRequest) error {
	slog.InfoContext(ctx, "[Hook] BeforeRequest", "request", jsonString(req))
	return nil
}

func (h *LoggingHook) AfterRequest(ctx context.Context, req *openai.ChatCompletionRequest, resp *openai.ChatCompletionResponse) error {
	tenantID, ok := ctx.Value("tenant_id").(string)
	if !ok {
		tenantID = "unknown"
	}
	slog.InfoContext(ctx, "[Hook] AfterRequest", "request", jsonString(req), "response", jsonString(resp), "tenant_id", tenantID)
	return nil
}

func jsonString(v interface{}) string {
	s, _ := json.Marshal(v)
	return string(s)
}

func (h *LoggingHook) Name() string {
	return "logging"
}

var _ hook.RequestHook = new(LoggingHook)

type ErrorHook struct{}

func (h *ErrorHook) Name() string {
	return "error"
}

func (h *ErrorHook) OnError(ctx context.Context, err error) {
	slog.ErrorContext(ctx, "[Hook] OnError", "error", err)
}

var _ hook.ErrorHook = new(ErrorHook)
