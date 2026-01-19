package main

import (
    "fmt"
    "log"
    "net/http"

    "github.com/deeplooplabs/ai-gateway/gateway"
    "github.com/deeplooplabs/ai-gateway/hook"
    "github.com/deeplooplabs/ai-gateway/model"
    "github.com/deeplooplabs/ai-gateway/provider"
)

func main() {
    // Create providers
    openAIProvider := provider.NewHTTPProvider("https://api.openai.com", "your-api-key")

    // Setup model registry
    registry := model.NewMapModelRegistry()
    registry.Register("gpt-4", openAIProvider, "")
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
    fmt.Println("AI Gateway listening on :8080")
    log.Fatal(http.ListenAndServe(":8080", gw))
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
