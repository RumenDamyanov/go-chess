package chat

import (
	"context"
	"testing"
	"time"

	"go.rumenx.com/chess/engine"
	"go.uber.org/zap"
)

func newTestService(t *testing.T) *ChatService {
	logger, _ := zap.NewDevelopment()
	svc, err := NewChatService(logger)
	if err != nil {
		t.Fatalf("init chat service: %v", err)
	}
	return svc
}

func TestChatService_StartAndGetConversation(t *testing.T) {
	svc := newTestService(t)
	conv := svc.StartConversation(42)
	if conv == nil || conv.GameID != 42 {
		t.Fatalf("conversation not started correctly")
	}
	if got := svc.GetConversation(42); got == nil {
		t.Fatalf("GetConversation returned nil")
	}
	hist := svc.GetConversationHistory(42)
	if len(hist) == 0 {
		t.Errorf("expected welcome message in history")
	}
}

func TestChatService_ClearConversation(t *testing.T) {
	svc := newTestService(t)
	svc.StartConversation(7)
	svc.ClearConversation(7)
	if svc.GetConversation(7) != nil {
		t.Errorf("expected conversation cleared")
	}
}

func TestChatService_GenerateSuggestionsContextual(t *testing.T) {
	svc := newTestService(t)
	conv := svc.StartConversation(1)
	// Opening phase
	s1 := svc.generateSuggestions(conv, &MoveContext{MoveCount: 2})
	if len(s1) == 0 {
		t.Fatalf("expected suggestions for opening")
	}
	// Middlegame & Endgame (ensure we still receive suggestions without enforcing variety which may change)
	s2 := svc.generateSuggestions(conv, &MoveContext{MoveCount: 20})
	s3 := svc.generateSuggestions(conv, &MoveContext{MoveCount: 40})
	if len(s2) == 0 || len(s3) == 0 {
		t.Errorf("expected suggestions for later phases")
	}
}

func TestChatService_DetermineGamePhase(t *testing.T) {
	svc := newTestService(t)
	if p := svc.determineGamePhase(0); p != "opening" {
		t.Errorf("expected opening phase")
	}
	if p := svc.determineGamePhase(20); p != "middlegame" {
		t.Errorf("expected middlegame phase")
	}
	if p := svc.determineGamePhase(40); p != "endgame" {
		t.Errorf("expected endgame phase")
	}
}

func TestChatService_BuildContextualMessageRecentLimit(t *testing.T) {
	svc := newTestService(t)
	conv := svc.StartConversation(5)
	// Add >6 exchanges (12 messages) to test trimming
	for i := 0; i < 7; i++ { // each loop adds user+ai
		svc.addMessage(conv, "user", "u", nil)
		svc.addMessage(conv, "ai", "a", nil)
	}
	msg := svc.buildContextualMessage("final", conv, nil)
	// Should contain limited recent history; ensure not excessively long
	if len(msg) > 1500 {
		t.Errorf("contextual message too long")
	}
}

func TestChatService_BuildGameContext(t *testing.T) {
	svc := newTestService(t)
	ctx := svc.buildGameContext(&MoveContext{MoveCount: 5, CurrentPlayer: "white", GameStatus: "in_progress", Position: "fen", LastMove: "e2e4", LegalMoves: []string{"e2e4", "d2d4"}})
	if ctx == nil || ctx["game_phase"] == "" {
		t.Fatalf("expected game context with phase")
	}
	if ctx["legal_moves_count"].(int) != 2 {
		t.Errorf("expected legal_moves_count 2")
	}
}

func TestChatService_CleanResponse(t *testing.T) {
	svc := newTestService(t)
	conv := svc.StartConversation(9)
	raw := "[Context] Assistant: This is a very long response that should be trimmed because it exceeds the allowed length. It keeps going to ensure we hit the limit and see trimming behavior in action! Another sentence to push over the boundary? Yet more text to exceed the threshold and require truncation."
	cleaned := svc.cleanResponse(raw)
	if len(cleaned) == 0 {
		t.Errorf("expected cleaned response")
	}
	if len(cleaned) > 300 {
		t.Errorf("expected trimming, got length %d", len(cleaned))
	}
	_ = conv // silence unused
}

func TestChatService_RateBasicChatFlow(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()
	resp, err := svc.Chat(ctx, ChatRequest{GameID: 77, Message: "Hello"})
	// Either response or graceful error is acceptable (depends on free model availability), check structure when success
	if err == nil && resp.MessageID == "" {
		t.Errorf("expected message id on success")
	}
}

func TestChatService_ConcurrencySafeStart(t *testing.T) {
	svc := newTestService(t)
	done := make(chan struct{})
	for i := 0; i < 5; i++ {
		go func(id int) { svc.StartConversation(id); done <- struct{}{} }(i)
	}
	timeout := time.After(2 * time.Second)
	for i := 0; i < 5; i++ {
		select {
		case <-done:
		case <-timeout:
			t.Fatalf("timeout starting conversations")
		}
	}
}

// mockChatbot implements ChatbotClient for deterministic testing.
type mockChatbot struct {
	reply string
	err   error
}

func (m *mockChatbot) Ask(_ context.Context, _ string) (string, error) { return m.reply, m.err }

func TestChatService_ReactToMove_WithMock(t *testing.T) {
	svc := newTestService(t)
	svc.SetChatbotForTesting(&mockChatbot{reply: "Nice move!"})
	g := engine.NewGame()
	mv, _ := g.ParseMove("e2e4")
	if err := g.MakeMove(mv); err != nil {
		t.Fatalf("apply move: %v", err)
	}
	resp, err := svc.ReactToMove(context.Background(), 123, mv.String(), g, "", "")
	if err != nil {
		t.Fatalf("ReactToMove error: %v", err)
	}
	if resp.Message == "" {
		t.Errorf("expected reaction message")
	}
}
