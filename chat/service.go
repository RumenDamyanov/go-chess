// Package chat provides AI chatbot functionality for chess gameplay interactions.
// It integrates with the go-chatbot package to provide conversational AI
// that can discuss chess moves, strategies, and provide friendly commentary.
package chat

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	gochatbot "go.rumenx.com/chatbot"
	"go.rumenx.com/chatbot/config"
	"go.rumenx.com/chatbot/models"
	"go.rumenx.com/chess/engine"
	"go.uber.org/zap"
)

// ChatbotClient is a minimal interface the underlying chatbot must satisfy.
type ChatbotClient interface {
	Ask(ctx context.Context, prompt string) (string, error)
}

// chatbotAdapter wraps the underlying gochatbot.Chatbot to satisfy ChatbotClient.
type chatbotAdapter struct{ base *gochatbot.Chatbot }

func (a *chatbotAdapter) Ask(ctx context.Context, prompt string) (string, error) {
	return a.base.Ask(ctx, prompt)
}

// ChatService represents the chess AI chatbot service.
type ChatService struct {
	chatbot       ChatbotClient
	config        *config.Config
	logger        *zap.Logger
	conversations map[int]*Conversation // gameID -> conversation
	mu            sync.RWMutex
}

// Conversation represents a chat conversation for a specific game.
type Conversation struct {
	GameID    int                    `json:"game_id"`
	Messages  []Message              `json:"messages"`
	Context   map[string]interface{} `json:"context"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

// Message represents a single chat message.
type Message struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"` // "user", "ai", "system"
	Content   string                 `json:"content"`
	GameState map[string]interface{} `json:"game_state,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// ChatRequest represents a request to the chat service.
type ChatRequest struct {
	GameID   int          `json:"game_id"`
	Message  string       `json:"message"`
	UserID   string       `json:"user_id,omitempty"`
	MoveData *MoveContext `json:"move_data,omitempty"`
	Provider string       `json:"provider,omitempty"` // Override default provider
	APIKey   string       `json:"api_key,omitempty"`  // Custom API key for this request
}

// ChatResponse represents a response from the chat service.
type ChatResponse struct {
	Message     string                 `json:"message"`
	MessageID   string                 `json:"message_id"`
	Personality string                 `json:"personality"`
	GameContext map[string]interface{} `json:"game_context,omitempty"`
	Suggestions []string               `json:"suggestions,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
}

// MoveContext provides context about recent moves for the AI.
type MoveContext struct {
	LastMove      string   `json:"last_move"`
	MoveCount     int      `json:"move_count"`
	CurrentPlayer string   `json:"current_player"`
	GameStatus    string   `json:"game_status"`
	Position      string   `json:"position"`                 // FEN notation
	LegalMoves    []string `json:"legal_moves"`              // Available legal moves
	InCheck       bool     `json:"in_check"`                 // Whether current player is in check
	CapturedPiece string   `json:"captured_piece,omitempty"` // Last captured piece
}

// NewChatService creates a new chat service instance.
func NewChatService(logger *zap.Logger) (*ChatService, error) {
	// Create chatbot configuration
	cfg := &config.Config{
		Model: "openai", // Default to OpenAI, can be overridden by env
		OpenAI: config.OpenAIConfig{
			APIKey: os.Getenv("OPENAI_API_KEY"),
			Model:  "gpt-4o-mini", // More cost-effective for chat
		},
		Anthropic: config.AnthropicConfig{
			APIKey: os.Getenv("ANTHROPIC_API_KEY"),
			Model:  "claude-3-haiku-20240307",
		},
		Gemini: config.GeminiConfig{
			APIKey: os.Getenv("GEMINI_API_KEY"),
			Model:  "gemini-1.5-flash",
		},
		XAI: config.XAIConfig{
			APIKey: os.Getenv("XAI_API_KEY"),
			Model:  "grok-1.5",
		},
		// Chess-specific prompt configuration
		Prompt: `You are a friendly AI chess opponent and coach. You are enthusiastic about chess and enjoy discussing games, moves, and strategies. Your personality traits:

- Encouraging and positive, especially towards beginners
- Knowledgeable about chess openings, tactics, and strategies
- Able to provide constructive feedback on moves
- Enjoy making chess-related jokes and observations
- Respectful and sportsmanlike
- Can explain chess concepts in simple terms

Guidelines for responses:
- Keep responses concise (1-3 sentences typically)
- Focus on chess-related topics
- Be encouraging about good moves and gentle about mistakes
- Use chess terminology appropriately
- Suggest improvements when relevant
- React to the current game situation
- Maintain a friendly, conversational tone

You can discuss:
- Chess moves and their quality
- Opening theory and strategies
- Tactical patterns and combinations
- Endgame techniques
- General chess improvement tips
- Chess history and famous games
- Encouragement and motivation

Avoid:
- Long lectures unless specifically asked
- Overly technical analysis unless requested
- Non-chess topics (politely redirect)
- Negative or discouraging language
- Giving away the best moves directly (let them discover)`,
		Language: "en",
		Tone:     "friendly",
		MessageFiltering: config.MessageFilteringConfig{
			Instructions: []string{
				"Stay focused on chess-related topics",
				"Be encouraging and positive",
				"Avoid spoiling the game by giving direct answers",
				"Keep responses conversational and fun",
			},
			Profanities:        []string{}, // Use default filter
			AggressionPatterns: []string{}, // Use default filter
			LinkPattern:        `https?://[\w\.-]+`,
		},
	}

	// Detect which AI provider to use based on available API keys
	if cfg.OpenAI.APIKey != "" {
		cfg.Model = "openai"
	} else if cfg.Anthropic.APIKey != "" {
		cfg.Model = "anthropic"
	} else if cfg.Gemini.APIKey != "" {
		cfg.Model = "gemini"
	} else if cfg.XAI.APIKey != "" {
		cfg.Model = "xai"
	} else {
		// Fallback to free model for development
		cfg.Model = "free"
		logger.Warn("No AI API keys found, using free model for chat")
	}

	// Create model based on configuration
	model, err := models.NewFromConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create AI model: %w", err)
	}

	// Create chatbot instance
	chatbot, err := gochatbot.New(cfg, gochatbot.WithModel(model))
	if err != nil {
		return nil, fmt.Errorf("failed to create chatbot: %w", err)
	}

	service := &ChatService{
		chatbot:       &chatbotAdapter{base: chatbot},
		config:        cfg,
		logger:        logger,
		conversations: make(map[int]*Conversation),
	}

	logger.Info("Chat service initialized", zap.String("model", cfg.Model))
	return service, nil
}

// createCustomChatbot creates a chatbot instance with custom API key and provider.
func (cs *ChatService) createCustomChatbot(provider, apiKey string) (ChatbotClient, error) {
	if provider == "" {
		return nil, fmt.Errorf("provider is required")
	}
	if apiKey == "" {
		return nil, fmt.Errorf("API key is required")
	}

	// Create custom configuration
	cfg := &config.Config{
		Model:  provider,
		Prompt: cs.config.Prompt, // Use same prompt as default
	}

	// Set API key based on provider
	switch strings.ToLower(provider) {
	case "openai":
		cfg.OpenAI = config.OpenAIConfig{
			APIKey: apiKey,
			Model:  "gpt-4o-mini",
		}
	case "anthropic":
		cfg.Anthropic = config.AnthropicConfig{
			APIKey: apiKey,
			Model:  "claude-3-haiku-20240307",
		}
	case "gemini":
		cfg.Gemini = config.GeminiConfig{
			APIKey: apiKey,
			Model:  "gemini-1.5-flash",
		}
	case "xai":
		cfg.XAI = config.XAIConfig{
			APIKey: apiKey,
			Model:  "grok-1.5",
		}
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}

	// Create model and chatbot
	model, err := models.NewFromConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create custom model: %w", err)
	}

	chatbot, err := gochatbot.New(cfg, gochatbot.WithModel(model))
	if err != nil {
		return nil, fmt.Errorf("failed to create custom chatbot: %w", err)
	}

	return &chatbotAdapter{base: chatbot}, nil
}

// StartConversation creates a new conversation for a game.
func (cs *ChatService) StartConversation(gameID int) *Conversation {
	cs.mu.Lock()
	conversation := &Conversation{
		GameID:    gameID,
		Messages:  make([]Message, 0),
		Context:   make(map[string]interface{}),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	cs.conversations[gameID] = conversation
	cs.mu.Unlock()

	// Add welcome message
	welcomeMsg := cs.generateWelcomeMessage()
	cs.addMessage(conversation, "ai", welcomeMsg, nil)

	cs.logger.Info("Started new conversation", zap.Int("game_id", gameID))
	return conversation
}

// Chat processes a chat message and returns AI response.
func (cs *ChatService) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	// Get or create conversation
	conversation, exists := cs.conversations[req.GameID]
	if !exists {
		conversation = cs.StartConversation(req.GameID)
	}

	// Add user message to conversation
	messageID := cs.addMessage(conversation, "user", req.Message, req.MoveData)

	// Build context for AI
	contextualMessage := cs.buildContextualMessage(req.Message, conversation, req.MoveData)

	// Get chatbot instance (custom or default)
	chatbot, err := cs.createCustomChatbot(req.Provider, req.APIKey)
	if err != nil {
		cs.logger.Error("Failed to create custom chatbot", zap.Error(err))
		chatbot = cs.chatbot // Fallback to default
	}

	// Get AI response
	response, err := chatbot.Ask(ctx, contextualMessage)
	if err != nil {
		cs.logger.Error("Failed to get AI response", zap.Error(err))
		return nil, fmt.Errorf("failed to get AI response: %w", err)
	}

	// Clean up response (remove any unwanted formatting)
	cleanResponse := cs.cleanResponse(response)

	// Add AI response to conversation
	cs.addMessage(conversation, "ai", cleanResponse, nil)

	// Generate suggestions for follow-up
	suggestions := cs.generateSuggestions(conversation, req.MoveData)

	return &ChatResponse{
		Message:     cleanResponse,
		MessageID:   messageID,
		Personality: "friendly_chess_coach",
		GameContext: cs.buildGameContext(req.MoveData),
		Suggestions: suggestions,
		Timestamp:   time.Now(),
	}, nil
}

// ReactToMove generates an AI reaction to a chess move.
func (cs *ChatService) ReactToMove(ctx context.Context, gameID int, move string, gameState *engine.Game, provider, apiKey string) (*ChatResponse, error) {
	// Get or create conversation
	conversation, exists := cs.conversations[gameID]
	if !exists {
		conversation = cs.StartConversation(gameID)
	}

	// Build enhanced move context
	legalMoves := gameState.GetAllLegalMoves()
	legalMoveStrs := make([]string, len(legalMoves))
	for i, legalMove := range legalMoves {
		legalMoveStrs[i] = legalMove.String()
	}

	moveData := &MoveContext{
		LastMove:      move,
		MoveCount:     len(gameState.MoveHistory()),
		CurrentPlayer: gameState.ActiveColor().String(),
		GameStatus:    gameState.Status().String(),
		Position:      gameState.ToFEN(), // Use real FEN
		LegalMoves:    legalMoveStrs,
		InCheck:       gameState.Status() == engine.Check,
	}

	// Generate contextual reaction prompt
	reactionPrompt := cs.buildMoveReactionPrompt(move, moveData)

	// Get chatbot instance (custom or default)
	chatbot, err := cs.createCustomChatbot(provider, apiKey)
	if err != nil {
		cs.logger.Error("Failed to create custom chatbot", zap.Error(err))
		chatbot = cs.chatbot // Fallback to default
	}

	// Get AI reaction
	reaction, err := chatbot.Ask(ctx, reactionPrompt)
	if err != nil {
		cs.logger.Error("Failed to get AI reaction", zap.Error(err))
		return nil, fmt.Errorf("failed to get AI reaction: %w", err)
	}

	// Clean response
	cleanReaction := cs.cleanResponse(reaction)

	// Add reaction to conversation
	cs.addMessage(conversation, "ai", cleanReaction, moveData)

	return &ChatResponse{
		Message:     cleanReaction,
		MessageID:   fmt.Sprintf("reaction_%d_%d", gameID, time.Now().Unix()),
		Personality: "observant_chess_coach",
		GameContext: cs.buildGameContext(moveData),
		Timestamp:   time.Now(),
	}, nil
}

// GetConversation returns the conversation for a game.
func (cs *ChatService) GetConversation(gameID int) *Conversation {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	return cs.conversations[gameID]
}

// GetConversationHistory returns the message history for a game.
func (cs *ChatService) GetConversationHistory(gameID int) []Message {
	cs.mu.RLock()
	conversation := cs.conversations[gameID]
	cs.mu.RUnlock()
	if conversation == nil {
		return []Message{}
	}
	return conversation.Messages
}

// ClearConversation removes the conversation for a game.
func (cs *ChatService) ClearConversation(gameID int) {
	cs.mu.Lock()
	delete(cs.conversations, gameID)
	cs.mu.Unlock()
	cs.logger.Info("Cleared conversation", zap.Int("game_id", gameID))
}

// Helper methods

func (cs *ChatService) generateWelcomeMessage() string {
	welcomeMessages := []string{
		"Hello! I'm your AI chess companion. Ready for a great game? 😊",
		"Welcome to our chess match! I'm excited to play and chat with you. 🎯",
		"Hi there! Let's have some fun with chess. Feel free to ask me anything about the game! ♟️",
		"Greetings, chess friend! I'm here to play, chat, and maybe share some chess wisdom. 🤔",
		"Hello! Ready to make some great moves? I love discussing chess strategy and tactics! ⚡",
	}

	// Simple random selection (could be improved with proper randomization)
	index := int(time.Now().Unix()) % len(welcomeMessages)
	return welcomeMessages[index]
}

func (cs *ChatService) addMessage(conversation *Conversation, msgType, content string, moveData *MoveContext) string {
	messageID := fmt.Sprintf("%s_%d_%d", msgType, conversation.GameID, time.Now().UnixNano())

	var gameState map[string]interface{}
	if moveData != nil {
		gameState = map[string]interface{}{
			"last_move":      moveData.LastMove,
			"move_count":     moveData.MoveCount,
			"current_player": moveData.CurrentPlayer,
			"game_status":    moveData.GameStatus,
			"position":       moveData.Position,
		}
	}

	message := Message{
		ID:        messageID,
		Type:      msgType,
		Content:   content,
		GameState: gameState,
		Timestamp: time.Now(),
	}

	cs.mu.Lock()
	conversation.Messages = append(conversation.Messages, message)
	conversation.UpdatedAt = time.Now()
	cs.mu.Unlock()

	return messageID
}

func (cs *ChatService) buildContextualMessage(userMessage string, conversation *Conversation, moveData *MoveContext) string {
	var contextBuilder strings.Builder

	// Add game context if available
	if moveData != nil {
		positionPreview := moveData.Position
		if len(positionPreview) > 20 {
			positionPreview = positionPreview[:20] + "..."
		}
		contextBuilder.WriteString(fmt.Sprintf("[Game Context: Move %d, %s to play, Position: %s]",
			moveData.MoveCount, moveData.CurrentPlayer, positionPreview))
		if moveData.LastMove != "" {
			contextBuilder.WriteString(fmt.Sprintf(" [Last move: %s]", moveData.LastMove))
		}
		contextBuilder.WriteString("\n\n")
	}

	// Add recent conversation context (last 3 messages)
	recentMessages := conversation.Messages
	if len(recentMessages) > 6 { // Keep last 3 exchanges (6 messages)
		recentMessages = recentMessages[len(recentMessages)-6:]
	}

	if len(recentMessages) > 0 {
		contextBuilder.WriteString("[Recent conversation:\n")
		for _, msg := range recentMessages {
			if msg.Type == "user" {
				contextBuilder.WriteString(fmt.Sprintf("Human: %s\n", msg.Content))
			} else if msg.Type == "ai" {
				contextBuilder.WriteString(fmt.Sprintf("Assistant: %s\n", msg.Content))
			}
		}
		contextBuilder.WriteString("]\n\n")
	}

	// Add current user message
	contextBuilder.WriteString(fmt.Sprintf("Human: %s", userMessage))

	return contextBuilder.String()
}

func (cs *ChatService) buildMoveReactionPrompt(move string, moveData *MoveContext) string {
	return fmt.Sprintf(`[Game Context: Move %d, %s just played %s, Status: %s]

Please give a brief, encouraging reaction to this chess move. Consider:
- Is this a good opening move, tactical shot, or strategic decision?
- Should you congratulate, encourage, or gently suggest improvements?
- Keep it conversational and positive
- 1-2 sentences maximum

The move played was: %s`,
		moveData.MoveCount, moveData.CurrentPlayer, move, moveData.GameStatus, move)
}

func (cs *ChatService) cleanResponse(response string) string {
	// Remove common AI response artifacts
	cleaned := strings.TrimSpace(response)

	// Remove any bracketed context that might leak through
	if strings.HasPrefix(cleaned, "[") {
		if idx := strings.Index(cleaned, "]"); idx != -1 && idx < 50 {
			cleaned = strings.TrimSpace(cleaned[idx+1:])
		}
	}

	// Remove "Assistant:" or "AI:" prefixes if present
	prefixes := []string{"Assistant:", "AI:", "Bot:", "Chatbot:"}
	for _, prefix := range prefixes {
		if strings.HasPrefix(cleaned, prefix) {
			cleaned = strings.TrimSpace(cleaned[len(prefix):])
		}
	}

	// Ensure response isn't too long for chat (max 280 chars)
	if len(cleaned) > 280 {
		// Find last complete sentence within limit
		if idx := strings.LastIndex(cleaned[:280], "."); idx > 100 {
			cleaned = cleaned[:idx+1]
		} else if idx := strings.LastIndex(cleaned[:280], "!"); idx > 100 {
			cleaned = cleaned[:idx+1]
		} else if idx := strings.LastIndex(cleaned[:280], "?"); idx > 100 {
			cleaned = cleaned[:idx+1]
		} else {
			cleaned = cleaned[:277] + "..."
		}
	}

	return cleaned
}

func (cs *ChatService) generateSuggestions(_ *Conversation, moveData *MoveContext) []string {
	suggestions := []string{
		"What do you think about this position?",
		"Any tips for improvement?",
		"What's your favorite opening?",
		"How would you rate my play so far?",
	}

	// Add context-specific suggestions
	if moveData != nil {
		if moveData.MoveCount < 10 {
			suggestions = append(suggestions, "Tell me about this opening")
		} else if moveData.MoveCount > 30 {
			suggestions = append(suggestions, "How's my endgame technique?")
		} else {
			suggestions = append(suggestions, "Any tactical opportunities here?")
		}
	}

	// Return max 3 suggestions
	if len(suggestions) > 3 {
		return suggestions[:3]
	}
	return suggestions
}

func (cs *ChatService) buildGameContext(moveData *MoveContext) map[string]interface{} {
	if moveData == nil {
		return nil
	}

	context := map[string]interface{}{
		"move_count":     moveData.MoveCount,
		"current_player": moveData.CurrentPlayer,
		"game_status":    moveData.GameStatus,
		"game_phase":     cs.determineGamePhase(moveData.MoveCount),
		"position_fen":   moveData.Position,
	}

	// Add optional fields if available
	if moveData.LastMove != "" {
		context["last_move"] = moveData.LastMove
	}
	if len(moveData.LegalMoves) > 0 {
		context["legal_moves_count"] = len(moveData.LegalMoves)
		context["sample_legal_moves"] = moveData.LegalMoves[:min(5, len(moveData.LegalMoves))] // Show first 5 moves
	}
	if moveData.InCheck {
		context["in_check"] = true
	}
	if moveData.CapturedPiece != "" {
		context["captured_piece"] = moveData.CapturedPiece
	}

	return context
}

// Helper function for min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (cs *ChatService) determineGamePhase(moveCount int) string {
	if moveCount < 15 {
		return "opening"
	} else if moveCount < 35 {
		return "middlegame"
	} else {
		return "endgame"
	}
}

// SetChatbotForTesting allows injection of a mock chatbot in tests.
func (cs *ChatService) SetChatbotForTesting(c ChatbotClient) { cs.chatbot = c }
