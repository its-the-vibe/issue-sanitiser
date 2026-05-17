package main

import (
	"context"
	"fmt"
	"os/exec"

	"google.golang.org/genai"
)

type GeminiBackend struct {
	client *genai.Client
	model  string
	apiKey string
}

func NewGeminiBackend(model, apiKey string) *GeminiBackend {
	if model == "" {
		model = "gemini-2.5-flash"
		// model = "gemini-flash-latest"
	}
	return &GeminiBackend{
		model:  model,
		apiKey: apiKey,
	}
}

func (b *GeminiBackend) Initialize(ctx context.Context) error {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  b.apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return fmt.Errorf("failed to create Gemini client: %w", err)
	}
	b.client = client
	return nil
}

func (b *GeminiBackend) Process(ctx context.Context, systemPrompt string, userPrompt string) error {
	// Define the bash tool
	bashTool := &genai.Tool{
		FunctionDeclarations: []*genai.FunctionDeclaration{
			{
				Name:        "bash",
				Description: "Execute a bash command in the local environment. Use this to interact with GitHub CLI (gh) or search the codebase.",
				Parameters: &genai.Schema{
					Type: genai.TypeObject,
					Properties: map[string]*genai.Schema{
						"command": {
							Type:        genai.TypeString,
							Description: "The bash command to execute",
						},
					},
					Required: []string{"command"},
				},
			},
		},
	}

	chat, err := b.client.Chats.Create(ctx, b.model, &genai.GenerateContentConfig{
		SystemInstruction: &genai.Content{
			Parts: []*genai.Part{{Text: systemPrompt}},
		},
		Tools: []*genai.Tool{bashTool},
	}, nil)
	if err != nil {
		return fmt.Errorf("failed to create Gemini chat: %w", err)
	}

	// Initial message
	messageParts := []genai.Part{{Text: userPrompt}}

	for {
		iter := chat.SendMessageStream(ctx, messageParts...)
		messageParts = nil // Clear parts for next iteration
		hasToolCall := false

		for resp, err := range iter {
			if err != nil {
				return fmt.Errorf("error during Gemini stream: %w", err)
			}

			if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil {
				continue
			}

			candidate := resp.Candidates[0]
			for _, part := range candidate.Content.Parts {
				if part.Text != "" {
					fmt.Print(part.Text)
				}
				if part.FunctionCall != nil {
					fmt.Printf("\n[Tool Call] %s\n", part.FunctionCall.Name)
					fmt.Printf("Arguments: %v\n", part.FunctionCall.Args)
					hasToolCall = true
					fc := part.FunctionCall
					if fc.Name == "bash" {
						cmdArg, ok := fc.Args["command"].(string)
						if !ok {
							return fmt.Errorf("invalid command argument for bash tool")
						}

						// Execute bash command
						output, err := executeBash(cmdArg)
						var responseContent string
						if err != nil {
							responseContent = fmt.Sprintf("Error executing command: %v\nOutput: %s", err, output)
						} else {
							responseContent = output
						}

						// Send function response in a separate message
						messageParts = append(messageParts, genai.Part{
							FunctionResponse: &genai.FunctionResponse{
								ID:   fc.ID,
								Name: "bash",
								Response: map[string]any{
									"output": responseContent,
								},
							},
						})
					}
				}
			}
		}

		// Only continue the loop if there were tool calls; otherwise, we're done
		if !hasToolCall || len(messageParts) == 0 {
			break
		}
	}

	return nil
}

func executeBash(command string) (string, error) {
	cmd := exec.Command("bash", "-c", command)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

func (b *GeminiBackend) Close() error {
	return nil
}
