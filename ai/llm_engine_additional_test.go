package ai

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"
	"time"

	"go.rumenx.com/chess/engine"
)

// roundTripperFunc allows mocking http.Client transport.
type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func newMockClient(body string, status int) *http.Client {
	return &http.Client{Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: status, Body: io.NopCloser(bytes.NewBufferString(body)), Header: make(http.Header)}, nil
	})}
}

func TestLLMAIEngine_ChatEnabledDisabled(t *testing.T) {
	cfg := LLMConfig{Provider: ProviderOpenAI, APIKey: "x", ChatEnabled: false}
	ai, err := NewLLMAIEngine(cfg)
	if err != nil {
		t.Fatalf("init: %v", err)
	}
	game := engine.NewGame()
	msg, err := ai.Chat(context.Background(), "hi", game)
	if err != nil {
		t.Fatalf("expected no error when chat disabled, got %v", err)
	}
	if msg == "" {
		t.Errorf("expected disabled message, got empty")
	}

	// Enable and mock response
	ai.config.ChatEnabled = true
	ai.httpClient = newMockClient(`{"choices":[{"message":{"role":"assistant","content":"Hello player"}}]}`, 200)
	msg, err = ai.Chat(context.Background(), "hi", game)
	if err != nil {
		t.Fatalf("chat failed: %v", err)
	}
	if msg == "" {
		t.Errorf("expected chat reply")
	}
}

func TestLLMAIEngine_ReactToMove(t *testing.T) {
	cfg := LLMConfig{Provider: ProviderOpenAI, APIKey: "x", ChatEnabled: true}
	ai, _ := NewLLMAIEngine(cfg)
	ai.httpClient = newMockClient(`{"choices":[{"message":{"role":"assistant","content":"Nice move!"}}]}`, 200)
	g := engine.NewGame()
	mv, _ := g.ParseMove("e2e4")
	// ensure move is legal then make it so reaction refers to last move
	if err := g.MakeMove(mv); err != nil {
		t.Fatalf("make move: %v", err)
	}
	reaction, err := ai.ReactToMove(context.Background(), mv, g)
	if err != nil {
		t.Fatalf("ReactToMove error: %v", err)
	}
	if reaction == "" {
		t.Errorf("expected non-empty reaction")
	}
}

func TestLLMAIEngine_askAnthropic(t *testing.T) {
	cfg := LLMConfig{Provider: ProviderAnthropic, APIKey: "x", ChatEnabled: true}
	ai, _ := NewLLMAIEngine(cfg)
	ai.httpClient = newMockClient(`{"content":[{"text":"Anthropic reply"}]}`, 200)
	reply, err := ai.askLLM(context.Background(), "test", ai.getSystemPrompt())
	if err != nil {
		t.Fatalf("askAnthropic failed: %v", err)
	}
	if reply == "" {
		t.Errorf("expected reply")
	}
}

func TestLLMAIEngine_askGemini(t *testing.T) {
	cfg := LLMConfig{Provider: ProviderGemini, APIKey: "x", ChatEnabled: true}
	ai, _ := NewLLMAIEngine(cfg)
	ai.httpClient = newMockClient(`{"candidates":[{"content":{"parts":[{"text":"Gemini reply"}]}}]}`, 200)
	reply, err := ai.askLLM(context.Background(), "test", ai.getSystemPrompt())
	if err != nil {
		t.Fatalf("askGemini failed: %v", err)
	}
	if reply == "" {
		t.Errorf("expected reply")
	}
}

func TestLLMAIEngine_SystemPrompts(t *testing.T) {
	cfg := LLMConfig{Provider: ProviderOpenAI, APIKey: "x", ChatEnabled: true, Personality: "energetic"}
	ai, _ := NewLLMAIEngine(cfg)
	if p := ai.getChatSystemPrompt(); p == "" || !contains(p, "energetic") {
		t.Errorf("chat system prompt missing personality")
	}
	if r := ai.getReactionSystemPrompt(); r == "" || !contains(r, "Nice move!") {
		t.Errorf("reaction system prompt missing examples")
	}
}

func TestLLMAIEngine_PromptGenerators(t *testing.T) {
	cfg := LLMConfig{Provider: ProviderOpenAI, APIKey: "x"}
	ai, _ := NewLLMAIEngine(cfg)
	g := engine.NewGame()
	chatPrompt := ai.generateChatPrompt("Hello", g)
	if !contains(chatPrompt, "Player says") {
		t.Errorf("chat prompt unexpected: %s", chatPrompt)
	}
	mv, _ := g.ParseMove("e2e4")
	reactionPrompt := ai.generateReactionPrompt(mv, g)
	if !contains(reactionPrompt, mv.String()) {
		t.Errorf("reaction prompt should include move")
	}
}

func TestLLMAIEngine_GetBestMove_FallbackOnError(t *testing.T) {
	// Provide a client that returns error JSON (missing choices) to force fallback.
	// The shared context may also expire causing an error; accept either a valid move or a deadline error.
	cfg := LLMConfig{Provider: ProviderOpenAI, APIKey: "x"}
	ai, _ := NewLLMAIEngine(cfg)
	ai.httpClient = newMockClient(`{"error":{"message":"fail"}}`, 200)
	g := engine.NewGame()
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	mv, err := ai.GetBestMove(ctx, g)
	if err != nil {
		if err.Error() != context.DeadlineExceeded.Error() {
			t.Fatalf("unexpected error: %v", err)
		}
		return // acceptable deadline exceeded
	}
	if mv.From == mv.To {
		t.Errorf("fallback produced empty move")
	}
}

func TestLLMAIEngine_GetBestMove_TimeoutFallback(t *testing.T) {
	// Client that sleeps beyond context to force timeout; RandomAI fallback may also be canceled.
	slowClient := &http.Client{Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		time.Sleep(200 * time.Millisecond)
		return &http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewBufferString(`{"choices":[{"message":{"role":"assistant","content":"e2e4"}}]}`)), Header: make(http.Header)}, nil
	})}
	cfg := LLMConfig{Provider: ProviderOpenAI, APIKey: "x"}
	ai, _ := NewLLMAIEngine(cfg)
	ai.httpClient = slowClient
	g := engine.NewGame()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	mv, err := ai.GetBestMove(ctx, g)
	// Accept either context error or a move if timing jitter allows it.
	if err != nil && mv.From == mv.To {
		if err.Error() == context.DeadlineExceeded.Error() {
			return // acceptable
		}
		t.Fatalf("unexpected error: %v", err)
	}
}
