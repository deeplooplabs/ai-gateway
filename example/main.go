package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/deeplooplabs/ai-gateway/gateway"
	"github.com/deeplooplabs/ai-gateway/hook"
	"github.com/deeplooplabs/ai-gateway/model"
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
	hooks.Register(&LoggingHook{})

	// Create gateway
	gw := gateway.New(
		gateway.WithModelRegistry(registry),
		gateway.WithHooks(hooks),
	)

	// Start server
	fmt.Println("AI Gateway listening on :8083")
	log.Fatal(http.ListenAndServe(":8083", gw))
}

// LoggingHook logs all requests
type LoggingHook struct{}

func (h *LoggingHook) Name() string {
	return "logging"
}

func (h *LoggingHook) BeforeRequest(ctx any, req any) error {
	fmt.Printf("[Hook] BeforeRequest: model=%v\n", req)
	return nil
}

func (h *LoggingHook) AfterRequest(ctx any, req any, resp any) error {
	fmt.Printf("[Hook] AfterRequest\n")
	return nil
}
