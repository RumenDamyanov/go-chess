// Package api provides HTTP API endpoints for the chess engine.
// It includes RESTful API handlers for game management, move execution,
// and real-time game updates via WebSocket connections.
package api

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"

	"github.com/rumendamyanov/go-chess/ai"
	"github.com/rumendamyanov/go-chess/config"
	"github.com/rumendamyanov/go-chess/engine"
)

// GameResponse represents a game in API responses.
type GameResponse struct {
	ID          int            `json:"id"`
	Status      string         `json:"status"`
	ActiveColor string         `json:"active_color"`
	Board       string         `json:"board"`
	MoveCount   int            `json:"move_count"`
	MoveHistory []MoveResponse `json:"move_history"`
	CreatedAt   time.Time      `json:"created_at"`
}

// MoveResponse represents a move in API responses.
type MoveResponse struct {
	From      string `json:"from"`
	To        string `json:"to"`
	Type      string `json:"type"`
	Piece     string `json:"piece"`
	Captured  string `json:"captured,omitempty"`
	Promotion string `json:"promotion,omitempty"`
	Notation  string `json:"notation"`
}

// MoveRequest represents a move request.
type MoveRequest struct {
	From      string `json:"from"`
	To        string `json:"to"`
	Promotion string `json:"promotion,omitempty"`
}

// AIRequest represents an AI move request.
type AIRequest struct {
	Level    string `json:"level"`    // beginner, easy, medium, hard, expert
	Engine   string `json:"engine"`   // random, minimax, llm
	Provider string `json:"provider"` // openai, anthropic, gemini, xai, deepseek (for LLM engine)
}

// ChatRequest represents a chat message request.
type ChatRequest struct {
	Message  string `json:"message"`
	Provider string `json:"provider,omitempty"` // LLM provider to use
}

// ChatResponse represents a chat message response.
type ChatResponse struct {
	Response string `json:"response"`
	Provider string `json:"provider"`
}

// ErrorResponse represents an error response.
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// Server represents the chess API server.
type Server struct {
	config   *config.Config
	logger   *zap.Logger
	games    map[int]*engine.Game
	gamesMux sync.RWMutex
	nextID   int
	upgrader websocket.Upgrader
}

// NewServer creates a new API server.
func NewServer(cfg *config.Config) *Server {
	logger, _ := zap.NewProduction()

	return &Server{
		config: cfg,
		logger: logger,
		games:  make(map[int]*engine.Game),
		nextID: 1,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins for demo purposes
			},
		},
	}
}

// SetupRoutes sets up the API routes.
func (s *Server) SetupRoutes(r *gin.Engine) {
	// Enable CORS for development
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	api := r.Group("/api")
	{
		// Game management
		api.POST("/games", s.createGame)
		api.GET("/games/:id", s.getGame)
		api.DELETE("/games/:id", s.deleteGame)
		api.GET("/games", s.listGames)

		// Game actions
		api.POST("/games/:id/moves", s.makeMove)
		api.GET("/games/:id/moves", s.getMoveHistory)
		api.POST("/games/:id/ai-move", s.getAIMove)

		// Chat functionality
		api.POST("/games/:id/chat", s.chatWithAI)
		api.POST("/games/:id/react", s.getAIReaction)

		// Game analysis
		api.GET("/games/:id/legal-moves", s.getLegalMoves)
		api.POST("/games/:id/fen", s.loadFromFEN)
		api.GET("/games/:id/analysis", s.analyzePosition)
	}

	// WebSocket endpoint
	r.GET("/ws/games/:id", s.handleWebSocket)

	// Health check
	r.GET("/health", s.health)
}

// createGame creates a new chess game.
func (s *Server) createGame(c *gin.Context) {
	s.gamesMux.Lock()
	defer s.gamesMux.Unlock()

	game := engine.NewGame()
	gameID := s.nextID
	s.nextID++

	s.games[gameID] = game

	response := s.gameToResponse(gameID, game)

	s.logger.Info("Created new game", zap.Int("game_id", gameID))
	c.JSON(http.StatusCreated, response)
}

// getGame retrieves a specific game.
func (s *Server) getGame(c *gin.Context) {
	gameID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_game_id"})
		return
	}

	s.gamesMux.RLock()
	game, exists := s.games[gameID]
	s.gamesMux.RUnlock()

	if !exists {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "game_not_found"})
		return
	}

	response := s.gameToResponse(gameID, game)
	c.JSON(http.StatusOK, response)
}

// deleteGame deletes a specific game.
func (s *Server) deleteGame(c *gin.Context) {
	gameID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_game_id"})
		return
	}

	s.gamesMux.Lock()
	defer s.gamesMux.Unlock()

	if _, exists := s.games[gameID]; !exists {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "game_not_found"})
		return
	}

	delete(s.games, gameID)

	s.logger.Info("Deleted game", zap.Int("game_id", gameID))
	c.JSON(http.StatusNoContent, nil)
}

// listGames lists all active games.
func (s *Server) listGames(c *gin.Context) {
	s.gamesMux.RLock()
	defer s.gamesMux.RUnlock()

	var games []GameResponse
	for id, game := range s.games {
		games = append(games, s.gameToResponse(id, game))
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"games": games,
		"count": len(games),
	})
}

// makeMove makes a move in a game.
func (s *Server) makeMove(c *gin.Context) {
	gameID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_game_id"})
		return
	}

	var req MoveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_request", Message: err.Error()})
		return
	}

	s.gamesMux.Lock()
	game, exists := s.games[gameID]
	s.gamesMux.Unlock()

	if !exists {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "game_not_found"})
		return
	}

	// Parse the move
	notation := req.From + req.To
	if req.Promotion != "" {
		notation += req.Promotion
	}

	move, err := game.ParseMove(notation)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_move", Message: err.Error()})
		return
	}

	// Make the move
	if err := game.MakeMove(move); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "illegal_move", Message: err.Error()})
		return
	}

	s.logger.Info("Move made", zap.Int("game_id", gameID), zap.String("move", move.String()))

	response := s.gameToResponse(gameID, game)
	c.JSON(http.StatusOK, response)
}

// getMoveHistory retrieves the move history of a game.
func (s *Server) getMoveHistory(c *gin.Context) {
	gameID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_game_id"})
		return
	}

	s.gamesMux.RLock()
	game, exists := s.games[gameID]
	s.gamesMux.RUnlock()

	if !exists {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "game_not_found"})
		return
	}

	history := game.MoveHistory()
	moves := make([]MoveResponse, len(history))

	for i, move := range history {
		moves[i] = s.moveToResponse(move)
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"moves": moves,
		"count": len(moves),
	})
}

// getAIMove gets a move suggestion from the AI.
func (s *Server) getAIMove(c *gin.Context) {
	gameID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_game_id"})
		return
	}

	var req AIRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Level = "medium"  // Default level
		req.Engine = "random" // Default engine
	}

	s.gamesMux.RLock()
	game, exists := s.games[gameID]
	s.gamesMux.RUnlock()

	if !exists {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "game_not_found"})
		return
	}

	// Parse difficulty
	var difficulty ai.Difficulty
	switch req.Level {
	case "beginner":
		difficulty = ai.DifficultyBeginner
	case "easy":
		difficulty = ai.DifficultyEasy
	case "medium":
		difficulty = ai.DifficultyMedium
	case "hard":
		difficulty = ai.DifficultyHard
	case "expert":
		difficulty = ai.DifficultyExpert
	default:
		difficulty = ai.DifficultyMedium
	}

	// Create AI engine based on type
	var aiEngine ai.Engine

	switch req.Engine {
	case "llm":
		// Use LLM AI if configured and provider specified
		if s.config.LLMAI.Enabled && req.Provider != "" && s.config.HasValidLLMProvider(req.Provider) {
			llmEngine, err := ai.NewLLMAIFromEnv(req.Provider, difficulty)
			if err != nil {
				s.logger.Warn("Failed to create LLM AI engine, falling back to random", zap.Error(err))
				aiEngine = ai.NewRandomAI()
			} else {
				aiEngine = llmEngine
			}
		} else {
			// Fallback to random if LLM not available
			aiEngine = ai.NewRandomAI()
		}
	case "minimax":
		aiEngine = ai.NewMinimaxAI(difficulty)
	default:
		aiEngine = ai.NewRandomAI()
	}

	aiEngine.SetDifficulty(difficulty)

	// Set timeout
	timeout := 30 * time.Second

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Get AI move
	move, err := aiEngine.GetBestMove(ctx, game)
	if err != nil {
		s.logger.Error("AI move generation failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "ai_move_failed"})
		return
	}

	// Convert move to response format
	moveResp := s.moveToResponse(move)

	c.JSON(http.StatusOK, map[string]interface{}{
		"move":     moveResp,
		"notation": move.String(),
		"level":    req.Level,
		"engine":   req.Engine,
		"provider": req.Provider,
	})
}

// getLegalMoves gets all legal moves for the current position.
func (s *Server) getLegalMoves(c *gin.Context) {
	gameID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_game_id"})
		return
	}

	s.gamesMux.RLock()
	_, exists := s.games[gameID]
	s.gamesMux.RUnlock()

	if !exists {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "game_not_found"})
		return
	}

	// This is a placeholder - we would need to implement move generation
	// For now, return an empty list
	c.JSON(http.StatusOK, map[string]interface{}{
		"moves": []MoveResponse{},
		"count": 0,
	})
}

// loadFromFEN loads a game position from FEN notation.
func (s *Server) loadFromFEN(c *gin.Context) {
	gameID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_game_id"})
		return
	}

	var req struct {
		FEN string `json:"fen"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_request", Message: err.Error()})
		return
	}

	s.gamesMux.Lock()
	_, exists := s.games[gameID]
	s.gamesMux.Unlock()

	if !exists {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "game_not_found"})
		return
	}

	// TODO: Implement FEN loading
	c.JSON(http.StatusNotImplemented, ErrorResponse{Error: "not_implemented", Message: "FEN loading not yet implemented"})
}

// analyzePosition analyzes the current position.
func (s *Server) analyzePosition(c *gin.Context) {
	gameID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_game_id"})
		return
	}

	s.gamesMux.RLock()
	game, exists := s.games[gameID]
	s.gamesMux.RUnlock()

	if !exists {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "game_not_found"})
		return
	}

	// Basic position analysis
	analysis := map[string]interface{}{
		"status":       game.Status().String(),
		"active_color": game.ActiveColor().String(),
		"move_count":   game.MoveCount(),
		"evaluation":   0.0, // TODO: Implement position evaluation
	}

	c.JSON(http.StatusOK, analysis)
}

// handleWebSocket handles WebSocket connections for real-time game updates.
func (s *Server) handleWebSocket(c *gin.Context) {
	gameID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_game_id"})
		return
	}

	s.gamesMux.RLock()
	game, exists := s.games[gameID]
	s.gamesMux.RUnlock()

	if !exists {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "game_not_found"})
		return
	}

	conn, err := s.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		s.logger.Error("WebSocket upgrade failed", zap.Error(err))
		return
	}
	defer conn.Close()

	// Send initial game state
	response := s.gameToResponse(gameID, game)
	if err := conn.WriteJSON(response); err != nil {
		s.logger.Error("Failed to send initial game state", zap.Error(err))
		return
	}

	// Keep connection alive and handle messages
	for {
		var msg map[string]interface{}
		if err := conn.ReadJSON(&msg); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				s.logger.Error("WebSocket error", zap.Error(err))
			}
			break
		}

		// Echo the message back (placeholder for game update handling)
		if err := conn.WriteJSON(msg); err != nil {
			s.logger.Error("Failed to send WebSocket message", zap.Error(err))
			break
		}
	}
}

// health returns the health status of the API.
func (s *Server) health(c *gin.Context) {
	s.gamesMux.RLock()
	gameCount := len(s.games)
	s.gamesMux.RUnlock()

	c.JSON(http.StatusOK, map[string]interface{}{
		"status":     "healthy",
		"timestamp":  time.Now().UTC(),
		"version":    "1.0.0",
		"game_count": gameCount,
	})
}

// Helper methods

// gameToResponse converts a game to API response format.
func (s *Server) gameToResponse(id int, game *engine.Game) GameResponse {
	history := game.MoveHistory()
	moves := make([]MoveResponse, len(history))

	for i, move := range history {
		moves[i] = s.moveToResponse(move)
	}

	return GameResponse{
		ID:          id,
		Status:      game.Status().String(),
		ActiveColor: game.ActiveColor().String(),
		Board:       game.Board().String(),
		MoveCount:   game.MoveCount(),
		MoveHistory: moves,
		CreatedAt:   time.Now().UTC(), // TODO: Store actual creation time
	}
}

// moveToResponse converts a move to API response format.
func (s *Server) moveToResponse(move engine.Move) MoveResponse {
	response := MoveResponse{
		From:     move.From.String(),
		To:       move.To.String(),
		Type:     move.Type.String(),
		Piece:    move.Piece.String(),
		Notation: move.String(),
	}

	if !move.Captured.IsEmpty() {
		response.Captured = move.Captured.String()
	}

	if move.Promotion != engine.Empty {
		response.Promotion = move.Promotion.String()
	}

	return response
}

// ReactionRequest represents a request for AI reaction to a move
type ReactionRequest struct {
	Move     string `json:"move"`
	Provider string `json:"provider,omitempty"`
}

// ReactionResponse represents the AI's reaction to a move
type ReactionResponse struct {
	Reaction string `json:"reaction"`
	Provider string `json:"provider"`
}

// chatWithAI handles chat requests with the AI
func (s *Server) chatWithAI(c *gin.Context) {
	gameIDStr := c.Param("id")
	gameID, err := strconv.Atoi(gameIDStr)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid game ID"})
		return
	}

	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request body"})
		return
	}

	s.gamesMux.RLock()
	game, exists := s.games[gameID]
	s.gamesMux.RUnlock()

	if !exists {
		c.JSON(404, gin.H{"error": "Game not found"})
		return
	}

	// Use default provider if not specified
	provider := req.Provider
	if provider == "" {
		provider = s.config.LLMAI.DefaultProvider
		if provider == "" {
			provider = "openai"
		}
	}

	// Create LLM engine config
	llmConfig := ai.LLMConfig{
		Provider:    ai.LLMProvider(provider),
		APIKey:      s.config.LLMAI.Providers[provider].APIKey,
		Model:       s.config.LLMAI.Providers[provider].Model,
		Endpoint:    s.config.LLMAI.Providers[provider].Endpoint,
		Difficulty:  ai.DifficultyEasy, // Use easy difficulty for chat
		Personality: "friendly",
		ChatEnabled: true,
	}

	llmEngine, err := ai.NewLLMAIEngine(llmConfig)
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to create AI engine: %v", err)})
		return
	}

	// Generate chat response
	ctx := context.Background()
	response, err := llmEngine.Chat(ctx, req.Message, game)
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to get AI response: %v", err)})
		return
	}

	c.JSON(200, ChatResponse{
		Response: response,
		Provider: provider,
	})
}

// getAIReaction handles requests for AI reactions to moves
func (s *Server) getAIReaction(c *gin.Context) {
	gameIDStr := c.Param("id")
	gameID, err := strconv.Atoi(gameIDStr)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid game ID"})
		return
	}

	var req ReactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request body"})
		return
	}

	s.gamesMux.RLock()
	game, exists := s.games[gameID]
	s.gamesMux.RUnlock()

	if !exists {
		c.JSON(404, gin.H{"error": "Game not found"})
		return
	}

	// Use default provider if not specified
	provider := req.Provider
	if provider == "" {
		provider = s.config.LLMAI.DefaultProvider
		if provider == "" {
			provider = "openai"
		}
	}

	// Parse the move
	move, err := game.ParseMove(req.Move)
	if err != nil {
		c.JSON(400, gin.H{"error": fmt.Sprintf("Invalid move format: %v", err)})
		return
	}

	// Create LLM engine config
	llmConfig := ai.LLMConfig{
		Provider:    ai.LLMProvider(provider),
		APIKey:      s.config.LLMAI.Providers[provider].APIKey,
		Model:       s.config.LLMAI.Providers[provider].Model,
		Endpoint:    s.config.LLMAI.Providers[provider].Endpoint,
		Difficulty:  ai.DifficultyEasy, // Use easy difficulty for reactions
		Personality: "entertaining",
		ChatEnabled: true,
	}

	llmEngine, err := ai.NewLLMAIEngine(llmConfig)
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to create AI engine: %v", err)})
		return
	}

	// Generate reaction to the move
	ctx := context.Background()
	reaction, err := llmEngine.ReactToMove(ctx, move, game)
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to get AI reaction: %v", err)})
		return
	}

	c.JSON(200, ReactionResponse{
		Reaction: reaction,
		Provider: provider,
	})
}
