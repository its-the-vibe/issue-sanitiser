package main

import (
	"context"
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

	ctx := context.Background()

	// Start the client
	if err := client.Start(ctx); err != nil {
		log.Fatalf("Failed to start Copilot client: %v", err)
	}
	defer client.Stop()

	// Create a session with the agent description as system prompt
	session, err := client.CreateSession(ctx, &copilot.SessionConfig{
		Model:     "gpt-4.1",
		Streaming: true,
		SystemMessage: &copilot.SystemMessageConfig{
			Content: agentDescription,
			Mode:    "replace",
		},
		AvailableTools: []string{
			"github-mcp-server-issue_read",
			"github-mcp-server-search_code",
			"github-mcp-server-get_file_contents",
			"web_search",
			"bash",
		},
		OnPermissionRequest: copilot.PermissionHandler.ApproveAll,
	})
	if err != nil {
		log.Fatalf("Failed to create session: %v", err)
	}
	defer session.Destroy()

	// Set up event handler to collect and display the response
	done := make(chan bool)
	var response strings.Builder

	session.On(func(event copilot.SessionEvent) {
		switch d := event.Data.(type) {
		case *copilot.AssistantMessageData:
			if d.Content != "" {
				// Stream the content as it arrives
				fmt.Print(d.Content)
				response.WriteString(d.Content)
			}
		case *copilot.SessionIdleData:
			close(done)
		case *copilot.SessionErrorData:
			fmt.Fprintf(os.Stderr, "\nError: %v\n", d.Message)
			close(done)
		}
	})

	// Send the issue URL to the agent
	fmt.Printf("Analyzing issue: %s\n\n", issueURL)
	_, err = session.Send(ctx, copilot.MessageOptions{
		Prompt: fmt.Sprintf("Please sanitize this GitHub issue: %s", issueURL),
	})
	if err != nil {
		log.Fatalf("Failed to send message: %v", err)
	}

	// Wait for the agent to finish processing
	<-done

	fmt.Println("\n\n✅ Issue sanitisation complete!")
}
