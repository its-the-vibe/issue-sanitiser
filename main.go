package main

import (
	"context"
	_ "embed"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

//go:embed .github/agents/issue-sanitiser-command.agent.md
var agentDescription string

func main() {
	// Define CLI flags
	backendType := flag.String("backend", "copilot", "LLM backend to use (copilot or gemini)")
	geminiAPIKey := flag.String("gemini-api-key", os.Getenv("GEMINI_API_KEY"), "Gemini API key (defaults to GEMINI_API_KEY env var)")
	modelName := flag.String("model", "", "Specific model name to use (optional)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <github-issue-url>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Example: %s https://github.com/owner/repo/issues/123\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	// Check if an issue URL was provided
	args := flag.Args()
	if len(args) < 1 {
		flag.Usage()
		os.Exit(1)
	}

	issueURL := args[0]

	// Validate the URL is a GitHub issue
	if !strings.Contains(issueURL, "github.com") || !strings.Contains(issueURL, "/issues/") {
		fmt.Println("Error: Please provide a valid GitHub issue URL")
		fmt.Println("Example: https://github.com/owner/repo/issues/123")
		os.Exit(1)
	}

	var backend LLMBackend

	switch strings.ToLower(*backendType) {
	case "copilot":
		backend = NewCopilotBackend(*modelName)
	case "gemini":
		if *geminiAPIKey == "" {
			log.Fatal("Error: Gemini API key is required when using the Gemini backend. Use --gemini-api-key or set GEMINI_API_KEY environment variable.")
		}
		backend = NewGeminiBackend(*modelName, *geminiAPIKey)
	default:
		log.Fatalf("Error: Unsupported backend type: %s", *backendType)
	}

	ctx := context.Background()

	// Initialize the backend
	if err := backend.Initialize(ctx); err != nil {
		log.Fatalf("Failed to initialize backend: %v", err)
	}
	defer backend.Close()

	// Send the issue URL to the agent
	fmt.Printf("Analyzing issue with %s backend: %s\n\n", *backendType, issueURL)
	err := backend.Process(ctx, agentDescription, fmt.Sprintf("Please sanitize this GitHub issue: %s", issueURL))
	if err != nil {
		log.Fatalf("Failed to process issue: %v", err)
	}

	fmt.Println("\n\n✅ Issue sanitisation complete!")
}
