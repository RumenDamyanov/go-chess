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

	"go.rumenx.com/chess/ai"
	"go.rumenx.com/chess/chat"
	"go.rumenx.com/chess/config"
	"go.rumenx.com/chess/engine"
)

// GameResponse represents a game in API responses.
type GameResponse struct {
	ID          int            `json:"id"`
	Status      string         `json:"status"`
	ActiveColor string         `json:"active_color"`
	AIColor     string         `json:"ai_color,omitempty"` // Which color the AI plays
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
	Notation  string `json:"notation,omitempty"`
}

// AIRequest represents an AI move request.
// AIRequest represents an AI move request.
type AIRequest struct {
	Level    string `json:"level"`    // beginner, easy, medium, hard, expert
	Engine   string `json:"engine"`   // random, minimax, llm
	Provider string `json:"provider"` // openai, anthropic, gemini, xai, deepseek (for LLM engine)
}

// GameCreateRequest represents a game creation request.
type GameCreateRequest struct {
	AIColor string `json:"ai_color,omitempty"` // "white", "black", or empty for default (black)
}

// GameMetadata stores additional game information.
type GameMetadata struct {
	AIColor   string    `json:"ai_color"`
	CreatedAt time.Time `json:"created_at"`
}

// ChatRequest represents a chat message request.
type ChatRequest struct {
	Message  string `json:"message"`
	Provider string `json:"provider,omitempty"` // LLM provider to use (openai, anthropic, gemini, xai)
	APIKey   string `json:"api_key,omitempty"`  // Custom API key for this request
}

// Enhanced ChatResponse represents a chat message response.
type ChatResponse struct {
	Response    string                 `json:"response"`
	Provider    string                 `json:"provider"`
	GameContext map[string]interface{} `json:"game_context,omitempty"`
	Suggestions []string               `json:"suggestions,omitempty"`
}

// ErrorResponse represents an error response.
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// Server represents the chess API server.
// Server represents the HTTP API server.
type Server struct {
	config       *config.Config
	logger       *zap.Logger
	games        map[int]*engine.Game
	gameMetadata map[int]*GameMetadata
	gamesMux     sync.RWMutex
	nextID       int
	upgrader     websocket.Upgrader
	chatService  *chat.ChatService
}

// NewServer creates a new API server.
func NewServer(cfg *config.Config) *Server {
	logger, _ := zap.NewProduction()

	// Initialize chat service
	chatService, err := chat.NewChatService(logger)
	if err != nil {
		logger.Error("Failed to create chat service", zap.Error(err))
		// Continue without chat service for now
	}

	return &Server{
		config:       cfg,
		logger:       logger,
		games:        make(map[int]*engine.Game),
		gameMetadata: make(map[int]*GameMetadata),
		nextID:       1,
		chatService:  chatService,
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
		api.POST("/games/:id/ai-hint", s.getAIHint)

		// Chat functionality
		api.POST("/games/:id/chat", s.chatWithAI)
		api.POST("/games/:id/react", s.getAIReaction)
		api.POST("/chat", s.generalChat) // General chat for demos

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

	// Parse request body for AI color preference
	var req GameCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// If no body or invalid JSON, use defaults
		req.AIColor = "black" // Default: AI plays black
	}

	// Validate AI color
	if req.AIColor != "white" && req.AIColor != "black" {
		req.AIColor = "black" // Default to black if invalid
	}

	game := engine.NewGame()
	gameID := s.nextID
	s.nextID++

	s.games[gameID] = game
	s.gameMetadata[gameID] = &GameMetadata{
		AIColor:   req.AIColor,
		CreatedAt: time.Now(),
	}

	response := s.gameToResponse(gameID, game)

	s.logger.Info("Created new game",
		zap.Int("game_id", gameID),
		zap.String("ai_color", req.AIColor))
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
	var notation string
	if req.Notation != "" {
		// Use provided notation (for castling moves like "O-O")
		notation = req.Notation
	} else {
		// Construct notation from from/to coordinates
		notation = req.From + req.To
		if req.Promotion != "" {
			notation += req.Promotion
		}
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
	game, gameExists := s.games[gameID]
	metadata, metadataExists := s.gameMetadata[gameID]
	s.gamesMux.RUnlock()

	if !gameExists {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "game_not_found"})
		return
	}

	// Get AI color from metadata, default to black if not found
	aiColor := "black"
	if metadataExists && metadata.AIColor != "" {
		aiColor = metadata.AIColor
	}

	// Validate that it's the AI's turn
	currentColor := game.ActiveColor().String()
	if currentColor != aiColor {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "not_ai_turn",
			Message: fmt.Sprintf("It's not the AI's turn to move (AI plays %s, current turn: %s)", aiColor, currentColor),
		})
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

// getAIHint gets a move suggestion from the AI without making the move.
func (s *Server) getAIHint(c *gin.Context) {
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
	default:
		difficulty = ai.DifficultyMedium
	}

	// Create AI engine
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

	// Get the best move suggestion (without making it)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	bestMove, err := aiEngine.GetBestMove(ctx, game)
	if err != nil {
		// Fallback to a random legal move if AI engine fails
		legalMoves := game.GetAllLegalMoves()
		if len(legalMoves) == 0 {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "no_legal_moves"})
			return
		}
		moveIndex := int(time.Now().UnixNano()) % len(legalMoves)
		bestMove = legalMoves[moveIndex]
	}

	// Return the hint without making the move
	hintResponse := map[string]interface{}{
		"from":        bestMove.From.String(),
		"to":          bestMove.To.String(),
		"explanation": fmt.Sprintf("AI suggests moving from %s to %s", bestMove.From.String(), bestMove.To.String()),
		"level":       req.Level,
		"engine":      req.Engine,
	}

	c.JSON(http.StatusOK, hintResponse)
}

// getLegalMoves gets all legal moves for the current position.
func (s *Server) getLegalMoves(c *gin.Context) {
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

	// Generate all legal moves for the current position
	legalMoves := s.generateAllLegalMoves(game)

	var moveResponses []MoveResponse
	for _, move := range legalMoves {
		moveResponses = append(moveResponses, s.moveToResponse(move))
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"legal_moves": moveResponses,
		"count":       len(moveResponses),
	})
}

// generateAllLegalMoves generates all legal moves for the current position.
// This uses the engine's optimized move generation logic.
func (s *Server) generateAllLegalMoves(game *engine.Game) []engine.Move {
	return game.GetAllLegalMoves()
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

	// Get AI color from metadata
	aiColor := "black" // Default
	if metadata, exists := s.gameMetadata[id]; exists {
		aiColor = metadata.AIColor
	}

	// Get creation time from metadata
	createdAt := time.Now().UTC()
	if metadata, exists := s.gameMetadata[id]; exists {
		createdAt = metadata.CreatedAt
	}

	return GameResponse{
		ID:          id,
		Status:      game.Status().String(),
		ActiveColor: game.ActiveColor().String(),
		AIColor:     aiColor,
		Board:       game.Board().String(),
		MoveCount:   game.MoveCount(),
		MoveHistory: moves,
		CreatedAt:   createdAt,
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
	Provider string `json:"provider,omitempty"` // LLM provider to use
	APIKey   string `json:"api_key,omitempty"`  // Custom API key for this request
}

// ReactionResponse represents the AI's reaction to a move
type ReactionResponse struct {
	Reaction    string                 `json:"reaction"`
	Provider    string                 `json:"provider"`
	GameContext map[string]interface{} `json:"game_context,omitempty"`
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

	// Check if chat service is available
	if s.chatService == nil {
		c.JSON(503, gin.H{"error": "Chat service unavailable"})
		return
	}

	// Create enhanced move context from current game state
	var moveContext *chat.MoveContext
	if game != nil {
		moveHistory := game.MoveHistory()
		var lastMoveStr string
		var capturedPiece string

		if len(moveHistory) > 0 {
			lastMove := moveHistory[len(moveHistory)-1]
			lastMoveStr = lastMove.String()
			if !lastMove.Captured.IsEmpty() {
				capturedPiece = lastMove.Captured.String()
			}
		}

		// Get legal moves
		legalMoves := game.GetAllLegalMoves()
		legalMoveStrs := make([]string, len(legalMoves))
		for i, legalMove := range legalMoves {
			legalMoveStrs[i] = legalMove.String()
		}

		moveContext = &chat.MoveContext{
			LastMove:      lastMoveStr,
			MoveCount:     len(moveHistory),
			CurrentPlayer: game.ActiveColor().String(),
			GameStatus:    game.Status().String(),
			Position:      game.ToFEN(), // Use real FEN now
			LegalMoves:    legalMoveStrs,
			InCheck:       game.Status() == engine.Check,
			CapturedPiece: capturedPiece,
		}
	}

	// Create chat request for the service
	chatReq := chat.ChatRequest{
		GameID:   gameID,
		Message:  req.Message,
		UserID:   "player", // Default user ID
		MoveData: moveContext,
		Provider: req.Provider, // Pass through custom provider
		APIKey:   req.APIKey,   // Pass through custom API key
	}

	// Generate chat response using the chat service
	ctx := context.Background()
	response, err := s.chatService.Chat(ctx, chatReq)
	if err != nil {
		s.logger.Error("Failed to get chat response", zap.Error(err))
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to get AI response: %v", err)})
		return
	}

	c.JSON(200, ChatResponse{
		Response:    response.Message,
		Provider:    response.Personality, // Use the provider that was actually used
		GameContext: response.GameContext,
		Suggestions: response.Suggestions,
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

	// Check if chat service is available
	if s.chatService == nil {
		c.JSON(503, gin.H{"error": "Chat service unavailable"})
		return
	}

	// Parse the move to validate it
	_, err = game.ParseMove(req.Move)
	if err != nil {
		c.JSON(400, gin.H{"error": fmt.Sprintf("Invalid move format: %v", err)})
		return
	}

	// Generate reaction using the enhanced ReactToMove method
	ctx := context.Background()
	response, err := s.chatService.ReactToMove(ctx, gameID, req.Move, game, req.Provider, req.APIKey)
	if err != nil {
		s.logger.Error("Failed to get move reaction", zap.Error(err))
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to get AI reaction: %v", err)})
		return
	}

	c.JSON(200, ReactionResponse{
		Reaction:    response.Message,
		Provider:    response.Personality,
		GameContext: response.GameContext,
	})
}

// generalChat handles general chat messages without game context
func (s *Server) generalChat(c *gin.Context) {
	if s.chatService == nil {
		c.JSON(500, gin.H{"error": "Chat service not available"})
		return
	}

	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request body"})
		return
	}

	// Create chat request for general conversation
	chatReq := chat.ChatRequest{
		GameID:   0, // No game context
		Message:  req.Message,
		UserID:   "demo-user",
		MoveData: nil,          // No move context
		Provider: req.Provider, // Pass through custom provider
		APIKey:   req.APIKey,   // Pass through custom API key
	}

	// Generate response using the chat service
	ctx := context.Background()
	response, err := s.chatService.Chat(ctx, chatReq)
	if err != nil {
		s.logger.Error("Failed to get chat response", zap.Error(err))
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to get AI response: %v", err)})
		return
	}

	c.JSON(200, ChatResponse{
		Response:    response.Message,
		Provider:    response.Personality,
		GameContext: response.GameContext,
		Suggestions: response.Suggestions,
	})
}
