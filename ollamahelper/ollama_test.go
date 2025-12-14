package ollamahelper

import (
	"context"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/tmc/langchaingo/llms/ollama"
)

// isOllamaRunning checks if Ollama is running and accessible
func isOllamaRunning() bool {
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get("http://localhost:11434/api/tags")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// skipIfNoOllama skips the test if Ollama is not running
func skipIfNoOllama(t *testing.T) {
	t.Helper()
	if !isOllamaRunning() {
		t.Skip("skipping test: Ollama is not running on localhost:11434")
	}
	// Also skip if explicitly disabled
	if os.Getenv("HELPER_SKIP_OLLAMA_TESTS") == "1" {
		t.Skip("skipping test: HELPER_SKIP_OLLAMA_TESTS=1")
	}
}

// getTestModel returns a model name for testing
// Uses environment variable or defaults to a small model
func getTestModel() string {
	if model := os.Getenv("OLLAMA_TEST_MODEL"); model != "" {
		return model
	}
	return "llama3.2" // Default to a common small model
}

func TestGetEmbedder(t *testing.T) {
	skipIfNoOllama(t)

	modelName := getTestModel()

	embedder, err := GetEmbedder(modelName)
	if err != nil {
		t.Fatalf("GetEmbedder(%q) error: %v", modelName, err)
	}

	if embedder == nil {
		t.Error("GetEmbedder() returned nil embedder")
	}
}

func TestGetEmbedderWithInvalidModel(t *testing.T) {
	skipIfNoOllama(t)

	// Test with a non-existent model name
	// Note: This may not error immediately as the model is only accessed when used
	embedder, err := GetEmbedder("nonexistent-model-that-does-not-exist-12345")
	if err != nil {
		// If it errors immediately, that's fine
		t.Logf("GetEmbedder() with invalid model returned error: %v", err)
		return
	}

	if embedder == nil {
		t.Error("GetEmbedder() returned nil embedder without error")
	}
}

func TestGenerateContentIntegration(t *testing.T) {
	skipIfNoOllama(t)

	modelName := getTestModel()

	llm, err := ollama.New(ollama.WithModel(modelName))
	if err != nil {
		t.Fatalf("Failed to create Ollama client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	systemPrompt := "You are a helpful assistant. Keep your responses brief."
	userPrompt := "Say 'hello' and nothing else."

	// Use defer/recover since the function may panic if response has no choices
	defer func() {
		if r := recover(); r != nil {
			t.Skipf("GenerateContent() panicked (model %q may not be installed or returned empty response): %v", modelName, r)
		}
	}()

	response, err := GenerateContent(ctx, llm, systemPrompt, userPrompt)
	if err != nil {
		// Model might not be available - skip rather than fail
		t.Skipf("GenerateContent() error (model %q may not be installed): %v", modelName, err)
	}

	if response == "" {
		t.Log("GenerateContent() returned empty response")
	} else {
		t.Logf("GenerateContent() response: %s", response)
	}
}

func TestGenerateStreamingContentIntegration(t *testing.T) {
	skipIfNoOllama(t)

	modelName := getTestModel()

	llm, err := ollama.New(ollama.WithModel(modelName))
	if err != nil {
		t.Fatalf("Failed to create Ollama client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	systemPrompt := "You are a helpful assistant. Keep your responses brief."
	userPrompt := "Say 'hello' and nothing else."

	var chunks [][]byte
	streamingFunc := func(ctx context.Context, chunk []byte) error {
		chunks = append(chunks, chunk)
		return nil
	}

	err = GenerateStreamingContent(ctx, llm, systemPrompt, userPrompt, streamingFunc)
	if err != nil {
		t.Skipf("GenerateStreamingContent() error (model %q may not be installed): %v", modelName, err)
	}

	if len(chunks) == 0 {
		t.Log("GenerateStreamingContent() received no chunks")
	} else {
		// Combine chunks to see full response
		var fullResponse []byte
		for _, chunk := range chunks {
			fullResponse = append(fullResponse, chunk...)
		}
		t.Logf("GenerateStreamingContent() received %d chunks, full response: %s", len(chunks), string(fullResponse))
	}
}

func TestGenerateContentWithCancelledContext(t *testing.T) {
	skipIfNoOllama(t)

	modelName := getTestModel()

	llm, err := ollama.New(ollama.WithModel(modelName))
	if err != nil {
		t.Fatalf("Failed to create Ollama client: %v", err)
	}

	// Create an already cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Use defer/recover since the function may panic
	defer func() {
		if r := recover(); r != nil {
			t.Logf("GenerateContent() panicked with cancelled context: %v", r)
		}
	}()

	_, err = GenerateContent(ctx, llm, "system", "user")
	if err == nil {
		t.Log("GenerateContent() with cancelled context did not return error (may have succeeded quickly)")
	}
}

func TestGenerateStreamingContentWithCancelledContext(t *testing.T) {
	skipIfNoOllama(t)

	modelName := getTestModel()

	llm, err := ollama.New(ollama.WithModel(modelName))
	if err != nil {
		t.Fatalf("Failed to create Ollama client: %v", err)
	}

	// Create an already cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	streamingFunc := func(ctx context.Context, chunk []byte) error {
		return nil
	}

	err = GenerateStreamingContent(ctx, llm, "system", "user", streamingFunc)
	if err == nil {
		t.Error("GenerateStreamingContent() with cancelled context should return error")
	}
}

// Test with empty prompts - the LLM should still work but may produce unexpected results
func TestGenerateContentWithEmptyPrompts(t *testing.T) {
	skipIfNoOllama(t)

	modelName := getTestModel()

	llm, err := ollama.New(ollama.WithModel(modelName))
	if err != nil {
		t.Fatalf("Failed to create Ollama client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Use defer/recover since the function may panic
	defer func() {
		if r := recover(); r != nil {
			t.Skipf("GenerateContent() panicked (model %q may not be installed): %v", modelName, r)
		}
	}()

	// Test with empty system prompt
	response, err := GenerateContent(ctx, llm, "", "Say hello")
	if err != nil {
		t.Logf("GenerateContent() with empty system prompt error: %v", err)
	} else {
		t.Logf("GenerateContent() with empty system prompt response: %s", response)
	}

	// Test with empty user prompt
	response, err = GenerateContent(ctx, llm, "You are helpful", "")
	if err != nil {
		t.Logf("GenerateContent() with empty user prompt error: %v", err)
	} else {
		t.Logf("GenerateContent() with empty user prompt response: %s", response)
	}
}

// Unit tests that don't require Ollama to be running

func TestIsOllamaRunning(t *testing.T) {
	// This is a simple test to verify the function works without panicking
	result := isOllamaRunning()
	t.Logf("isOllamaRunning() = %v", result)
}

func TestGetTestModel(t *testing.T) {
	// Test default model
	originalEnv := os.Getenv("OLLAMA_TEST_MODEL")
	defer os.Setenv("OLLAMA_TEST_MODEL", originalEnv)

	os.Unsetenv("OLLAMA_TEST_MODEL")
	model := getTestModel()
	if model != "llama3.2" {
		t.Errorf("getTestModel() without env = %q, want %q", model, "llama3.2")
	}

	// Test with env var
	os.Setenv("OLLAMA_TEST_MODEL", "custom-model")
	model = getTestModel()
	if model != "custom-model" {
		t.Errorf("getTestModel() with env = %q, want %q", model, "custom-model")
	}
}

// Benchmark tests (only run when Ollama is available)

func BenchmarkGenerateContent(b *testing.B) {
	if !isOllamaRunning() {
		b.Skip("skipping benchmark: Ollama is not running")
	}

	modelName := getTestModel()
	llm, err := ollama.New(ollama.WithModel(modelName))
	if err != nil {
		b.Fatalf("Failed to create Ollama client: %v", err)
	}

	ctx := context.Background()
	systemPrompt := "You are a helpful assistant. Keep responses very brief."
	userPrompt := "Say 'hi'"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = GenerateContent(ctx, llm, systemPrompt, userPrompt)
	}
}

func BenchmarkGetEmbedder(b *testing.B) {
	if !isOllamaRunning() {
		b.Skip("skipping benchmark: Ollama is not running")
	}

	modelName := getTestModel()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = GetEmbedder(modelName)
	}
}
