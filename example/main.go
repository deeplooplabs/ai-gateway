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
	"github.com/deeplooplabs/ai-gateway/openai"
	"github.com/deeplooplabs/ai-gateway/provider"
)

func main() {
	// Create providers
	openAIProvider := provider.NewHTTPProvider(os.Getenv("OPENAI_BASE_URL"), os.Getenv("OPENAI_API_KEY"))

	// Setup model registry
	registry := model.NewMapModelRegistry()
	registry.Register("gpt-4o", openAIProvider, "deepseek-ai/DeepSeek-V3.2")
	registry.Register("gpt-3.5-turbo", openAIProvider, "")

	// Create hooks
	hooks := hook.NewRegistry()
	hooks.Register(&LoggingHook{}, &AuthenticateHook{})

	// Create gateway
	gw := gateway.New(
		gateway.WithModelRegistry(registry),
		gateway.WithHooks(hooks),
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
	tenantID := ctx.Value("tenant_id").(string)
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
