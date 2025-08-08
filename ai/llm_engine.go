// Package ai provides LLM-powered AI implementations using external providers.
package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"go.rumenx.com/chess/engine"
)

// LLMProvider represents the supported LLM providers.
type LLMProvider string

const (
	// ProviderOpenAI uses OpenAI GPT models.
	ProviderOpenAI LLMProvider = "openai"
	// ProviderAnthropic uses Anthropic Claude models.
	ProviderAnthropic LLMProvider = "anthropic"
	// ProviderGemini uses Google Gemini models.
	ProviderGemini LLMProvider = "gemini"
	// ProviderXAI uses xAI Grok models.
	ProviderXAI LLMProvider = "xai"
	// ProviderDeepSeek uses DeepSeek models.
	ProviderDeepSeek LLMProvider = "deepseek"
)

// LLMConfig represents configuration for LLM-powered AI.
type LLMConfig struct {
	Provider    LLMProvider `json:"provider"`
	APIKey      string      `json:"api_key"`
	Model       string      `json:"model"`
	Endpoint    string      `json:"endpoint"`
	Difficulty  Difficulty  `json:"difficulty"`
	Personality string      `json:"personality"`
	ChatEnabled bool        `json:"chat_enabled"`
}

// LLMAIEngine implements an AI engine powered by Large Language Models.
type LLMAIEngine struct {
	config     LLMConfig
	httpClient *http.Client
	context    []ChatMessage
}

// ChatMessage represents a message in the conversation.
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OpenAIRequest represents an OpenAI API request.
type OpenAIRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	Temperature float64       `json:"temperature"`
	MaxTokens   int           `json:"max_tokens"`
}

// OpenAIResponse represents an OpenAI API response.
type OpenAIResponse struct {
	Choices []struct {
		Message ChatMessage `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// AnthropicRequest represents an Anthropic API request.
type AnthropicRequest struct {
	Model     string        `json:"model"`
	Messages  []ChatMessage `json:"messages"`
	MaxTokens int           `json:"max_tokens"`
	System    string        `json:"system,omitempty"`
}

// AnthropicResponse represents an Anthropic API response.
type AnthropicResponse struct {
	Content []struct {
		Text string `json:"text"`
	} `json:"content"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// GeminiRequest represents a Gemini API request.
type GeminiRequest struct {
	Contents []struct {
		Parts []struct {
			Text string `json:"text"`
		} `json:"parts"`
		Role string `json:"role,omitempty"`
	} `json:"contents"`
	GenerationConfig *struct {
		Temperature *float64 `json:"temperature,omitempty"`
		MaxTokens   *int     `json:"maxOutputTokens,omitempty"`
	} `json:"generationConfig,omitempty"`
}

// GeminiResponse represents a Gemini API response.
type GeminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// NewLLMAIEngine creates a new LLM-powered AI engine.
func NewLLMAIEngine(cfg LLMConfig) (*LLMAIEngine, error) {
	if cfg.APIKey == "" && cfg.Provider != ProviderDeepSeek {
		return nil, fmt.Errorf("API key is required for provider %s", cfg.Provider)
	}

	// Set default endpoints
	if cfg.Endpoint == "" {
		switch cfg.Provider {
		case ProviderOpenAI:
			cfg.Endpoint = "https://api.openai.com/v1/chat/completions"
		case ProviderAnthropic:
			cfg.Endpoint = "https://api.anthropic.com/v1/messages"
		case ProviderGemini:
			cfg.Endpoint = "https://generativelanguage.googleapis.com/v1beta/models"
		case ProviderXAI:
			cfg.Endpoint = "https://api.x.ai/v1/chat/completions"
		case ProviderDeepSeek:
			cfg.Endpoint = "https://api.deepseek.com/v1/chat/completions"
		}
	}

	// Set default models
	if cfg.Model == "" {
		switch cfg.Provider {
		case ProviderOpenAI:
			cfg.Model = "gpt-3.5-turbo"
		case ProviderAnthropic:
			cfg.Model = "claude-3-haiku-20240307"
		case ProviderGemini:
			cfg.Model = "gemini-1.5-flash"
		case ProviderXAI:
			cfg.Model = "grok-beta"
		case ProviderDeepSeek:
			cfg.Model = "deepseek-chat"
		}
	}

	if cfg.Personality == "" {
		cfg.Personality = "a friendly but competitive chess player"
	}

	return &LLMAIEngine{
		config: cfg,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		context: make([]ChatMessage, 0),
	}, nil
}

// GetBestMove returns the best move using LLM analysis.
func (ai *LLMAIEngine) GetBestMove(ctx context.Context, game *engine.Game) (engine.Move, error) {
	// Generate prompt for the LLM
	prompt := ai.generateChessPrompt(game)

	// Ask the LLM for a move
	response, err := ai.askLLM(ctx, prompt, ai.getSystemPrompt())
	if err != nil {
		// Fallback to RandomAI if LLM fails
		randomAI := NewRandomAI()
		randomAI.SetDifficulty(ai.config.Difficulty)
		return randomAI.GetBestMove(ctx, game)
	}

	// Parse the move from LLM response
	move, err := ai.parseMoveFromResponse(response, game)
	if err != nil {
		// Fallback to RandomAI if parsing fails
		randomAI := NewRandomAI()
		randomAI.SetDifficulty(ai.config.Difficulty)
		return randomAI.GetBestMove(ctx, game)
	}

	// Add this interaction to context for future moves
	ai.addToContext("user", prompt)
	ai.addToContext("assistant", response)

	return move, nil
}

// GetDifficulty returns the current difficulty level.
func (ai *LLMAIEngine) GetDifficulty() Difficulty {
	return ai.config.Difficulty
}

// SetDifficulty sets the difficulty level.
func (ai *LLMAIEngine) SetDifficulty(difficulty Difficulty) {
	ai.config.Difficulty = difficulty
}

// Chat provides conversational interaction with the AI.
func (ai *LLMAIEngine) Chat(ctx context.Context, message string, game *engine.Game) (string, error) {
	if !ai.config.ChatEnabled {
		return "Chat is disabled for this AI opponent.", nil
	}

	// Create a chess-aware chat prompt
	chatPrompt := ai.generateChatPrompt(message, game)

	response, err := ai.askLLM(ctx, chatPrompt, ai.getChatSystemPrompt())
	if err != nil {
		return "Sorry, I'm having trouble responding right now.", err
	}

	return response, nil
}

// ReactToMove generates a reaction to a player's move.
func (ai *LLMAIEngine) ReactToMove(ctx context.Context, move engine.Move, game *engine.Game) (string, error) {
	if !ai.config.ChatEnabled {
		return "", nil
	}

	prompt := ai.generateReactionPrompt(move, game)

	response, err := ai.askLLM(ctx, prompt, ai.getReactionSystemPrompt())
	if err != nil {
		return "", err
	}

	// Don't return empty reactions
	response = strings.TrimSpace(response)
	if response == "" || strings.ToLower(response) == "no reaction" {
		return "", nil
	}

	return response, nil
}

// GetProvider returns the LLM provider being used.
func (ai *LLMAIEngine) GetProvider() LLMProvider {
	return ai.config.Provider
}

// askLLM sends a request to the configured LLM provider.
func (ai *LLMAIEngine) askLLM(ctx context.Context, message, systemPrompt string) (string, error) {
	switch ai.config.Provider {
	case ProviderOpenAI, ProviderXAI, ProviderDeepSeek:
		return ai.askOpenAICompatible(ctx, message, systemPrompt)
	case ProviderAnthropic:
		return ai.askAnthropic(ctx, message, systemPrompt)
	case ProviderGemini:
		return ai.askGemini(ctx, message, systemPrompt)
	default:
		return "", fmt.Errorf("unsupported provider: %s", ai.config.Provider)
	}
}

// askOpenAICompatible sends a request to OpenAI-compatible APIs.
func (ai *LLMAIEngine) askOpenAICompatible(ctx context.Context, message, systemPrompt string) (string, error) {
	messages := []ChatMessage{
		{Role: "system", Content: systemPrompt},
	}

	// Add conversation history
	messages = append(messages, ai.context...)
	messages = append(messages, ChatMessage{Role: "user", Content: message})

	request := OpenAIRequest{
		Model:       ai.config.Model,
		Messages:    messages,
		Temperature: ai.getTemperatureForDifficulty(),
		MaxTokens:   200,
	}

	reqBody, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", ai.config.Endpoint, bytes.NewBuffer(reqBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+ai.config.APIKey)

	resp, err := ai.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var response OpenAIResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if response.Error != nil {
		return "", fmt.Errorf("API error: %s", response.Error.Message)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no response from API")
	}

	return response.Choices[0].Message.Content, nil
}

// askAnthropic sends a request to Anthropic's API.
func (ai *LLMAIEngine) askAnthropic(ctx context.Context, message, systemPrompt string) (string, error) {
	messages := []ChatMessage{}

	// Add conversation history
	messages = append(messages, ai.context...)
	messages = append(messages, ChatMessage{Role: "user", Content: message})

	request := AnthropicRequest{
		Model:     ai.config.Model,
		Messages:  messages,
		MaxTokens: 200,
		System:    systemPrompt,
	}

	reqBody, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", ai.config.Endpoint, bytes.NewBuffer(reqBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", ai.config.APIKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	resp, err := ai.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var response AnthropicResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if response.Error != nil {
		return "", fmt.Errorf("API error: %s", response.Error.Message)
	}

	if len(response.Content) == 0 {
		return "", fmt.Errorf("no response from API")
	}

	return response.Content[0].Text, nil
}

// askGemini sends a request to Google's Gemini API.
func (ai *LLMAIEngine) askGemini(ctx context.Context, message, systemPrompt string) (string, error) {
	// Combine system prompt with message for Gemini
	fullPrompt := fmt.Sprintf("%s\n\n%s", systemPrompt, message)

	request := GeminiRequest{
		Contents: []struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
			Role string `json:"role,omitempty"`
		}{
			{
				Parts: []struct {
					Text string `json:"text"`
				}{
					{Text: fullPrompt},
				},
				Role: "user",
			},
		},
		GenerationConfig: &struct {
			Temperature *float64 `json:"temperature,omitempty"`
			MaxTokens   *int     `json:"maxOutputTokens,omitempty"`
		}{
			Temperature: &[]float64{ai.getTemperatureForDifficulty()}[0],
			MaxTokens:   &[]int{200}[0],
		},
	}

	reqBody, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/%s:generateContent?key=%s", ai.config.Endpoint, ai.config.Model, ai.config.APIKey)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := ai.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var response GeminiResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if response.Error != nil {
		return "", fmt.Errorf("API error: %s", response.Error.Message)
	}

	if len(response.Candidates) == 0 || len(response.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no response from API")
	}

	return response.Candidates[0].Content.Parts[0].Text, nil
}

// getSystemPrompt returns the system prompt for chess move generation.
func (ai *LLMAIEngine) getSystemPrompt() string {
	personalityContext := ""
	if ai.config.Personality != "" {
		personalityContext = fmt.Sprintf("You have the personality of: %s. ", ai.config.Personality)
	}

	difficultyContext := ""
	switch ai.config.Difficulty {
	case DifficultyBeginner:
		difficultyContext = "Play at a beginner level, occasionally making suboptimal moves. "
	case DifficultyEasy:
		difficultyContext = "Play at an easy level with basic tactical awareness. "
	case DifficultyMedium:
		difficultyContext = "Play at a medium level with good tactical and some strategic understanding. "
	case DifficultyHard:
		difficultyContext = "Play at a hard level with strong tactical and strategic play. "
	case DifficultyExpert:
		difficultyContext = "Play at an expert level with excellent tactical and strategic understanding. "
	}

	return fmt.Sprintf(`You are a chess AI opponent. %s%sYour task is to analyze the chess position and suggest the best move in algebraic notation.

IMPORTANT RULES:
1. Respond ONLY with a valid chess move in algebraic notation (e.g., "e4", "Nf3", "O-O", "Qxd5")
2. Do not include explanations, commentary, or extra text
3. Ensure the move is legal in the current position
4. Consider the difficulty level in your move selection
5. If you cannot determine a good move, respond with "random"

The move should be in standard algebraic notation:
- Pawn moves: e4, d5, exd5
- Piece moves: Nf3, Bd3, Qh5
- Castling: O-O (kingside), O-O-O (queenside)
- Captures: Nxf3, Qxd5
- Promotion: e8=Q
- Check: Nf3+
- Checkmate: Qh7#`, personalityContext, difficultyContext)
}

// getChatSystemPrompt returns the system prompt for chat interactions.
func (ai *LLMAIEngine) getChatSystemPrompt() string {
	personalityContext := ""
	if ai.config.Personality != "" {
		personalityContext = fmt.Sprintf("You have the personality of: %s. ", ai.config.Personality)
	}

	return fmt.Sprintf(`You are a chess AI opponent that can chat with players. %sYou are knowledgeable about chess, friendly, and engaging. You can:

1. Discuss chess strategy and tactics
2. Comment on the current game position
3. Share chess knowledge and tips
4. React to impressive or interesting moves
5. Provide encouragement and maintain good sportsmanship

Keep responses conversational, helpful, and appropriate for a chess game setting.`, personalityContext)
}

// getReactionSystemPrompt returns the system prompt for move reactions.
func (ai *LLMAIEngine) getReactionSystemPrompt() string {
	return `You are a chess AI that reacts to moves made by your opponent. Provide brief, engaging reactions that show your chess understanding. Examples:

- For good moves: "Nice move!", "I didn't see that coming!", "Clever!"
- For brilliant moves: "Wow! Brilliant!", "Outstanding!", "That's a beautiful move!"
- For mistakes: "Interesting choice...", "Hmm, that gives me an opportunity"
- For blunders: "Thank you for that!", "I'll take advantage of that"

Keep reactions short (1-10 words), appropriate, and show chess personality. If the move is ordinary, respond with "no reaction" or leave empty.`
}

// generateChessPrompt creates a prompt for chess move generation.
func (ai *LLMAIEngine) generateChessPrompt(game *engine.Game) string {
	board := game.Board()

	// Create a simple board representation
	boardString := ai.boardToString(board)

	// Get move history
	moveHistory := game.MoveHistory()
	historyString := ""
	if len(moveHistory) > 0 {
		recent := moveHistory
		if len(recent) > 10 { // Only show last 10 moves
			recent = recent[len(recent)-10:]
		}

		moves := make([]string, len(recent))
		for i, move := range recent {
			moves[i] = move.String()
		}
		historyString = fmt.Sprintf("Recent moves: %s\n", strings.Join(moves, " "))
	}

	activeColor := "White"
	if game.ActiveColor() == engine.Black {
		activeColor = "Black"
	}

	return fmt.Sprintf(`Current chess position:

%s

%sActive color: %s

Provide your move in algebraic notation:`, boardString, historyString, activeColor)
}

// generateChatPrompt creates a prompt for chat interactions.
func (ai *LLMAIEngine) generateChatPrompt(message string, game *engine.Game) string {
	board := game.Board()
	boardString := ai.boardToString(board)

	return fmt.Sprintf(`Player says: "%s"

Current position:
%s

Respond as a chess AI opponent:`, message, boardString)
}

// generateReactionPrompt creates a prompt for move reactions.
func (ai *LLMAIEngine) generateReactionPrompt(move engine.Move, game *engine.Game) string {
	return fmt.Sprintf(`The player just played: %s

Provide a brief reaction to this move (or respond with "no reaction" if it's not noteworthy):`, move.String())
}

// boardToString converts a chess board to a string representation.
func (ai *LLMAIEngine) boardToString(board *engine.Board) string {
	var sb strings.Builder

	sb.WriteString("  a b c d e f g h\n")
	for rank := 7; rank >= 0; rank-- {
		sb.WriteString(fmt.Sprintf("%d ", rank+1))
		for file := 0; file < 8; file++ {
			square := engine.Square(rank*8 + file)
			piece := board.GetPiece(square)

			if piece.IsEmpty() {
				sb.WriteString(". ")
			} else {
				symbol := ai.pieceToSymbol(piece)
				sb.WriteString(symbol + " ")
			}
		}
		sb.WriteString(fmt.Sprintf("%d\n", rank+1))
	}
	sb.WriteString("  a b c d e f g h")

	return sb.String()
}

// pieceToSymbol converts a piece to its symbol representation.
func (ai *LLMAIEngine) pieceToSymbol(piece engine.Piece) string {
	symbols := map[engine.PieceType]string{
		engine.King:   "K",
		engine.Queen:  "Q",
		engine.Rook:   "R",
		engine.Bishop: "B",
		engine.Knight: "N",
		engine.Pawn:   "P",
	}

	symbol := symbols[piece.Type]
	if piece.Color == engine.Black {
		symbol = strings.ToLower(symbol)
	}

	return symbol
}

// parseMoveFromResponse extracts a chess move from LLM response.
func (ai *LLMAIEngine) parseMoveFromResponse(response string, game *engine.Game) (engine.Move, error) {
	// Clean the response
	response = strings.TrimSpace(response)
	response = strings.Trim(response, "\"'")

	// If the LLM says "random", use random move
	if strings.ToLower(response) == "random" {
		randomAI := NewRandomAI()
		return randomAI.GetBestMove(context.Background(), game)
	}

	// Try to parse as algebraic notation
	lines := strings.Split(response, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Remove common prefixes/suffixes
		line = strings.TrimPrefix(line, "Move: ")
		line = strings.TrimPrefix(line, "Best move: ")
		line = strings.TrimSuffix(line, ".")
		line = strings.TrimSuffix(line, "!")
		line = strings.TrimSuffix(line, "?")

		// Try to parse this as a move
		move, err := game.ParseMove(line)
		if err == nil && game.IsLegalMove(move) {
			return move, nil
		}
	}

	// If we can't parse any move, return error
	return engine.Move{}, fmt.Errorf("could not parse move from response: %s", response)
}

// addToContext adds a message to the conversation context.
func (ai *LLMAIEngine) addToContext(role, content string) {
	ai.context = append(ai.context, ChatMessage{
		Role:    role,
		Content: content,
	})

	// Keep context manageable (last 10 messages)
	if len(ai.context) > 10 {
		ai.context = ai.context[len(ai.context)-10:]
	}
}

// getTemperatureForDifficulty returns appropriate temperature for difficulty level.
func (ai *LLMAIEngine) getTemperatureForDifficulty() float64 {
	switch ai.config.Difficulty {
	case DifficultyBeginner:
		return 1.2 // High creativity, more mistakes
	case DifficultyEasy:
		return 0.9 // Some creativity
	case DifficultyMedium:
		return 0.7 // Balanced
	case DifficultyHard:
		return 0.5 // More focused
	case DifficultyExpert:
		return 0.3 // Very focused
	default:
		return 0.7
	}
}

// NewLLMAIFromEnv creates an LLM AI engine from environment variables.
func NewLLMAIFromEnv(provider string, difficulty Difficulty) (*LLMAIEngine, error) {
	var apiKey string
	var envVar string

	switch LLMProvider(provider) {
	case ProviderOpenAI:
		envVar = "OPENAI_API_KEY"
	case ProviderAnthropic:
		envVar = "ANTHROPIC_API_KEY"
	case ProviderGemini:
		envVar = "GEMINI_API_KEY"
	case ProviderXAI:
		envVar = "XAI_API_KEY"
	case ProviderDeepSeek:
		envVar = "DEEPSEEK_API_KEY"
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}

	apiKey = os.Getenv(envVar)
	if apiKey == "" {
		return nil, fmt.Errorf("environment variable %s is required for provider %s", envVar, provider)
	}

	cfg := LLMConfig{
		Provider:    LLMProvider(provider),
		APIKey:      apiKey,
		Difficulty:  difficulty,
		ChatEnabled: true,
		Personality: "a friendly but competitive chess player",
	}

	return NewLLMAIEngine(cfg)
}
