package main

import (
	_ "embed"
	"fmt"
	"log"
	"os"
	"strings"

	copilot "github.com/github/copilot-sdk/go"
)

//go:embed .github/agents/issue-sanitiser-command.agent.md
var agentDescription string

func main() {
	// Check if an issue URL was provided
	if len(os.Args) < 2 {
		fmt.Println("Usage: issue-sanitiser <github-issue-url>")
		fmt.Println("Example: issue-sanitiser https://github.com/owner/repo/issues/123")
		os.Exit(1)
	}

	issueURL := os.Args[1]

	// Validate the URL is a GitHub issue
	if !strings.Contains(issueURL, "github.com") || !strings.Contains(issueURL, "/issues/") {
		fmt.Println("Error: Please provide a valid GitHub issue URL")
		fmt.Println("Example: https://github.com/owner/repo/issues/123")
		os.Exit(1)
	}

	// Create Copilot client
	client := copilot.NewClient(&copilot.ClientOptions{
		LogLevel: "error",
	})

	// Start the client
	if err := client.Start(); err != nil {
		log.Fatalf("Failed to start Copilot client: %v", err)
	}
	defer client.Stop()

	// Create a session with the agent description as system prompt
	session, err := client.CreateSession(&copilot.SessionConfig{
		Model:     "gpt-4.1",
		Streaming: true,
		SystemMessage: &copilot.SystemMessageConfig{
			Content: agentDescription,
		},
	})
	if err != nil {
		log.Fatalf("Failed to create session: %v", err)
	}
	defer session.Destroy()

	// Set up event handler to collect and display the response
	done := make(chan bool)
	var response strings.Builder

	session.On(func(event copilot.SessionEvent) {
		if event.Type == "assistant.message" {
			if event.Data.Content != nil {
				// Stream the content as it arrives
				fmt.Print(*event.Data.Content)
				response.WriteString(*event.Data.Content)
			}
		}
		if event.Type == "session.idle" {
			close(done)
		}
		if event.Type == "error" {
			fmt.Fprintf(os.Stderr, "\nError: %v\n", event.Data)
			close(done)
		}
	})

	// Send the issue URL to the agent
	fmt.Printf("Analyzing issue: %s\n\n", issueURL)
	_, err = session.Send(copilot.MessageOptions{
		Prompt: fmt.Sprintf("Please sanitize this GitHub issue: %s", issueURL),
	})
	if err != nil {
		log.Fatalf("Failed to send message: %v", err)
	}

	// Wait for the agent to finish processing
	<-done

	fmt.Println("\n\n✅ Issue sanitisation complete!")
}
