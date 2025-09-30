package main

import (
	"context"
	"fmt"
	"log"

	"github.com/NubeDev/air/internal/config"
	"github.com/NubeDev/air/internal/llm"
)

func main() {
	// Load config
	cfg, err := config.Load("data/config-dev.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Create Ollama client
	client, err := llm.NewOllamaClient(cfg.Models.Ollama)
	if err != nil {
		log.Fatalf("Failed to create Ollama client: %v", err)
	}

	// Test health
	ctx := context.Background()
	if err := client.Health(ctx); err != nil {
		log.Fatalf("Ollama health check failed: %v", err)
	}

	fmt.Println("✅ Ollama client is working!")

	// Test chat completion
	req := llm.ChatRequest{
		Model: "llama3",
		Messages: []llm.Message{
			{Role: "user", Content: "Hello! How are you?"},
		},
		Stream: false,
	}

	resp, err := client.ChatCompletion(ctx, req)
	if err != nil {
		log.Fatalf("Chat completion failed: %v", err)
	}

	fmt.Printf("✅ Chat completion successful: %s\n", resp.Message.Content)
}
