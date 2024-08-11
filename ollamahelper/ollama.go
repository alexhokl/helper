package ollamahelper

import (
	"context"
	"fmt"

	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
)

/// GenerateStreamingContent generates a streaming response from the given model, with the given system and user prompts
func GenerateStreamingContent(parentContext context.Context, llm *ollama.LLM, systemPrompt string, userPrompt string, streamingFunc func(ctx context.Context, chunk []byte) error) error {
	_, err := llm.GenerateContent(
		parentContext,
		[]llms.MessageContent{
			{
				Role: llms.ChatMessageTypeSystem,
				Parts: []llms.ContentPart{llms.TextContent{Text: systemPrompt}},
			},
			{
				Role: llms.ChatMessageTypeHuman,
				Parts: []llms.ContentPart{llms.TextContent{Text: userPrompt}},
			},
		},
		llms.WithStreamingFunc(streamingFunc),
	)
	return err
}

/// GenerateContent generates a response from the given model, with the given system and user prompts
func GenerateContent(parentContext context.Context, llm *ollama.LLM, systemPrompt string, userPrompt string) (string, error) {
	response, err := llm.GenerateContent(
		parentContext,
		[]llms.MessageContent{
			{
				Role:  llms.ChatMessageTypeSystem,
				Parts: []llms.ContentPart{llms.TextContent{Text: systemPrompt}},
			},
			{
				Role:  llms.ChatMessageTypeHuman,
				Parts: []llms.ContentPart{llms.TextContent{Text: userPrompt}},
			},
		},
	)
	return response.Choices[0].Content, err
}

/// GetEmbedder returns an embedding model for the given model name
func GetEmbedder(modelName string) (embeddings.Embedder, error) {
	embedClient, err := ollama.New(
		ollama.WithModel(modelName),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to load LLM model: %w", err)
	}
	embedder, err := embeddings.NewEmbedder(embedClient)
	if err != nil {
		return nil, fmt.Errorf("unable to create embedder: %w", err)
	}
	return embedder, nil
}
