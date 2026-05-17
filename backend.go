package main

import (
	"context"
)

// LLMBackend defines the interface for different LLM providers.
type LLMBackend interface {
	// Initialize the backend.
	Initialize(ctx context.Context) error
	// Process the issue using the provided prompt and system instructions.
	Process(ctx context.Context, systemPrompt string, userPrompt string) error
	// Close any resources.
	Close() error
}
