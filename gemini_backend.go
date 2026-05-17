package main

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

type GeminiBackend struct {
	client *genai.Client
	model  string
	apiKey string
}

func NewGeminiBackend(model, apiKey string) *GeminiBackend {
	if model == "" {
		model = "gemini-1.5-pro"
	}
	return &GeminiBackend{
		model:  model,
		apiKey: apiKey,
	}
}

func (b *GeminiBackend) Initialize(ctx context.Context) error {
	client, err := genai.NewClient(ctx, option.WithAPIKey(b.apiKey))
	if err != nil {
		return fmt.Errorf("failed to create Gemini client: %w", err)
	}
	b.client = client
	return nil
}

func (b *GeminiBackend) Process(ctx context.Context, systemPrompt string, userPrompt string) error {
	model := b.client.GenerativeModel(b.model)
	model.SystemInstruction = genai.NewUserContent(genai.Text(systemPrompt))

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
	model.Tools = []*genai.Tool{bashTool}

	session := model.StartChat()

	// Initial message
	iter := session.SendMessageStream(ctx, genai.Text(userPrompt))

	for {
		resp, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return fmt.Errorf("error during Gemini stream: %w", err)
		}

		if len(resp.Candidates) == 0 {
			continue
		}

		candidate := resp.Candidates[0]
		if candidate.Content == nil {
			continue
		}

		// Collect function calls to handle them after the stream or if they appear
		var funcResponses []genai.Part

		for _, part := range candidate.Content.Parts {
			switch p := part.(type) {
			case genai.Text:
				fmt.Print(string(p))
			case genai.FunctionCall:
				if p.Name == "bash" {
					cmdArg, ok := p.Args["command"].(string)
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

					funcResponses = append(funcResponses, genai.FunctionResponse{
						Name: "bash",
						Response: map[string]any{
							"output": responseContent,
						},
					})
				}
			}
		}

		if len(funcResponses) > 0 {
			iter = session.SendMessageStream(ctx, funcResponses...)
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
	if b.client != nil {
		return b.client.Close()
	}
	return nil
}
