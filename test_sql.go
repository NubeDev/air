package main

import (
	"context"
	"fmt"
	"log"

	"github.com/NubeDev/air/internal/config"
	"github.com/NubeDev/air/internal/llm"
	"github.com/ollama/ollama/api"
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

	// Test SQL generation
	ctx := context.Background()

	prompt := "Write a SQL query to select all users from the users table"
	schema := "CREATE TABLE users (id INT PRIMARY KEY, name VARCHAR(100), email VARCHAR(100));"

	fullPrompt := fmt.Sprintf(`You are a SQL expert. Generate a SQL query based on the following request and schema.

Request: %s

Schema:
%s

Generate only the SQL query, no explanations or markdown formatting.`, prompt, schema)

	req := llm.GenerateRequest{
		Model:  "sqlcoder:7b",
		Prompt: fullPrompt,
		Stream: false,
		Options: &api.Options{
			Temperature: 0.1,
			TopP:        0.9,
		},
	}

	resp, err := client.GenerateText(ctx, req)
	if err != nil {
		log.Fatalf("SQL generation failed: %v", err)
	}

	fmt.Printf("✅ SQL generation successful:\n%s\n", resp.Response)
}
