package main

import (
	"context"
	"fmt"

	copilot "github.com/github/copilot-sdk/go"
)

type CopilotBackend struct {
	client *copilot.Client
	model  string
}

func NewCopilotBackend(model string) *CopilotBackend {
	if model == "" {
		model = "claude-haiku-4.5"
	}
	return &CopilotBackend{
		model: model,
	}
}

func (b *CopilotBackend) Initialize(ctx context.Context) error {
	b.client = copilot.NewClient(&copilot.ClientOptions{
		LogLevel: "error",
	})

	if err := b.client.Start(ctx); err != nil {
		return fmt.Errorf("failed to start Copilot client: %w", err)
	}
	return nil
}

func (b *CopilotBackend) Process(ctx context.Context, systemPrompt string, userPrompt string) error {
	session, err := b.client.CreateSession(ctx, &copilot.SessionConfig{
		Model:     b.model,
		Streaming: true,
		SystemMessage: &copilot.SystemMessageConfig{
			Content: systemPrompt,
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
		return fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Disconnect()

	done := make(chan bool)
	var sessionErr error

	session.On(func(event copilot.SessionEvent) {
		switch d := event.Data.(type) {
		case *copilot.AssistantMessageData:
			if d.Content != "" {
				fmt.Print(d.Content)
			}
		case *copilot.SessionIdleData:
			close(done)
		case *copilot.SessionErrorData:
			sessionErr = fmt.Errorf("%s", d.Message)
			close(done)
		}
	})

	_, err = session.Send(ctx, copilot.MessageOptions{
		Prompt: userPrompt,
	})
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	<-done
	return sessionErr
}

func (b *CopilotBackend) Close() error {
	if b.client != nil {
		b.client.Stop()
	}
	return nil
}
