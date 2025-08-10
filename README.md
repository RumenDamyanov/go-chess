# go-chess

[![CI](https://github.com/rumendamyanov/go-chess/actions/workflows/ci.yml/badge.svg)](https://github.com/rumendamyanov/go-chess/actions/workflows/ci.yml)
[![CodeQL](https://github.com/rumendamyanov/go-chess/actions/workflows/github-code-scanning/codeql/badge.svg)](https://github.com/rumendamyanov/go-chess/actions/workflows/github-code-scanning/codeql)
[![Dependabot](https://github.com/rumendamyanov/go-chess/actions/workflows/dependabot/dependabot-updates/badge.svg)](https://github.com/rumendamyanov/go-chess/actions/workflows/dependabot/dependabot-updates)
[![codecov](https://codecov.io/gh/rumendamyanov/go-chess/graph/badge.svg)](https://codecov.io/gh/rumendamyanov/go-chess)
[![Go Report Card](https://goreportcard.com/badge/go.rumenx.com/chess?5)](https://goreportcard.com/report/go.rumenx.com/chess)
[![Go Reference](https://pkg.go.dev/badge/go.rumenx.com/chess.svg)](https://pkg.go.dev/go.rumenx.com/chess)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/rumendamyanov/go-chess/blob/master/LICENSE.md)

> 📖 **Documentation**: [📚 Complete Wiki](https://github.com/RumenDamyanov/go-chess/wiki) · [🚀 Quick Start](https://github.com/RumenDamyanov/go-chess/wiki/Quick-Start-Guide) · [📋 API Reference](https://github.com/RumenDamyanov/go-chess/wiki/API-Reference) · [🤖 LLM AI Guide](https://github.com/RumenDamyanov/go-chess/wiki/LLM-AI-Guide)

**go-chess** is a modern, AI-powered chess engine and API library written in Go. It provides a complete chess implementation with move validation, game state management, AI opponent capabilities, and a RESTful API for easy integration with frontend applications.

**What makes go-chess special:**

🧠 **Real Chess Intelligence**: Unlike many chess APIs that use placeholder data or basic move validation, go-chess provides genuine chess AI understanding through:

- Complete FEN notation support for real board positions
- Legal move generation and validation
- Check/checkmate detection with proper game state tracking
- Rich game context for AI decision making

🤖 **Advanced LLM Integration**: Connect with leading AI providers using your own API keys:

- OpenAI GPT-4, Anthropic Claude, Google Gemini, xAI Grok, DeepSeek
- Per-request API key configuration for cost control
- Real board analysis instead of placeholder responses
- Strategic commentary based on actual positions

Designed for both educational purposes and production use, it demonstrates best practices in Go development while remaining simple and practical.

## About

### Project Inspiration

This project showcases modern Go development practices and serves as a demonstration of building a complete, production-ready chess engine. It's designed to be educational yet practical, providing a solid foundation for chess applications while maintaining clean, idiomatic Go code.

## ✨ Key Features

### Core Chess Engine

• **Complete Rule Implementation**: Full chess rules including castling, en passant, pawn promotion
• **Move Validation**: Legal move checking with check/checkmate detection
• **FEN Notation Support**: Complete Forsyth-Edwards Notation for position export/import
• **Game State Management**: Real-time FEN generation, game history tracking
• **AI Integration**: Pluggable AI system with multiple difficulty levels
• **Position Analysis**: Board evaluation, threat detection, piece mobility analysis
• **Legal Move Generation**: Fast legal move computation for AI analysis

### 🚀 Advanced Features

• **RESTful API**: Complete HTTP API for frontend integration
• **WebSocket Support**: Real-time game updates and move streaming
• **AI Opponents**: Multiple AI algorithms with configurable difficulty
• **Game Persistence**: Save and load games in standard formats (PGN, FEN)
• **Analysis Engine**: Position evaluation and move suggestions

### 🛠️ Technical Excellence

• **High Test Coverage**: Comprehensive unit and integration tests
• **Static Analysis**: golangci-lint, CodeQL security scanning
• **CI/CD Pipeline**: Automated testing, coverage reporting, and quality checks
• **Clean Architecture**: Modular design with clear separation of concerns
• **Documentation**: Extensive API documentation with examples

### 🤖 LLM-Powered AI Integration ✨

• **Multiple Provider Support**: OpenAI GPT-4, Anthropic Claude, Google Gemini, xAI Grok, DeepSeek
• **Custom API Keys**: Per-request API key support for any LLM provider
• **Chess Intelligence**: AI understands real game state via FEN notation and legal moves
• **Rich Game Context**: AI sees legal moves, check status, captured pieces, and game history
• **Conversational AI**: Chat with your AI opponent about moves and strategy
• **Move Reactions**: AI provides entertaining commentary on specific moves
• **Difficulty-Based Personalities**: Different AI behaviors based on skill level
• **Fallback Mechanism**: Gracefully falls back to traditional AI if LLM fails
• **Real-time Analysis**: AI provides position evaluation and strategic insights

## 📚 Documentation

> **📖 Complete documentation available in our [GitHub Wiki](https://github.com/RumenDamyanov/go-chess/wiki)**

### 🚀 Quick Navigation

• **[🚀 Quick Start Guide](https://github.com/RumenDamyanov/go-chess/wiki/Quick-Start-Guide)** - Get up and running in 5 minutes
• **[📋 API Reference](https://github.com/RumenDamyanov/go-chess/wiki/API-Reference)** - Complete HTTP API documentation
• **[🤖 LLM AI Guide](https://github.com/RumenDamyanov/go-chess/wiki/LLM-AI-Guide)** - Advanced AI integration with ChatGPT, Claude, etc.
• **[🔧 Basic Usage](https://github.com/RumenDamyanov/go-chess/wiki/Basic-Usage)** - Fundamental concepts and patterns
• **[⚡ Advanced Usage](https://github.com/RumenDamyanov/go-chess/wiki/Advanced-Usage)** - Production deployment and optimization
• **[🔧 Troubleshooting](https://github.com/RumenDamyanov/go-chess/wiki/Troubleshooting)** - Common issues and solutions
• **[❓ FAQ](https://github.com/RumenDamyanov/go-chess/wiki/FAQ)** - Frequently asked questions

### 📖 More Guides

• [Installation Guide](https://github.com/RumenDamyanov/go-chess/wiki/Installation-Guide) - Detailed installation instructions
• [Docker Deployment](https://github.com/RumenDamyanov/go-chess/wiki/Docker-Deployment) - Container deployment and orchestration
• [Chess Engine Basics](https://github.com/RumenDamyanov/go-chess/wiki/Chess-Engine-Basics) - Understanding the core engine
• [Frontend Integration](https://github.com/RumenDamyanov/go-chess/wiki/Frontend-Integration) - Building chess UIs
• [js-chess Demo](https://github.com/RumenDamyanov/js-chess) - Live JavaScript frontend showcase using go-chess backend
• [Game Formats](https://github.com/RumenDamyanov/go-chess/wiki/Game-Formats) - Working with PGN and FEN notation
• [Examples](https://github.com/RumenDamyanov/go-chess/wiki/Examples) - Real-world usage examples

## Supported AI Engines

| AI Engine | Description | Difficulty Levels | Performance | Special Features |
|-----------|-------------|------------------|-------------|------------------|
| Random | Simple random move selection | Beginner | Fast | - |
| Minimax | Classic minimax algorithm | Easy - Medium | Moderate | Alpha-beta pruning |
| **LLM-Powered** | **Advanced AI using Large Language Models** | **All levels** | **Variable** | **🤖 Chat, Reactions, Strategy** |
| - OpenAI GPT-4 | Premium AI with excellent chess understanding | Expert | Excellent | Balanced analysis, helpful explanations |
| - Anthropic Claude | Detailed analytical AI with educational focus | Expert | Excellent | In-depth move analysis, teaching mode |
| - Google Gemini | Fast and efficient LLM with good chess knowledge | Hard - Expert | Very Good | Quick responses, solid play |
| - xAI Grok | Creative AI with entertaining commentary | Medium - Hard | Good | Humorous reactions, creative explanations |
| - DeepSeek | Cost-effective AI with solid chess capabilities | Medium - Expert | Good | Budget-friendly, reliable performance |

## 🧠 Enhanced Chess Intelligence & Chat Features

### Real Chess AI Understanding

The AI integration now provides genuine chess intelligence powered by Large Language Models:

- **Real Board Analysis**: AI sees actual game positions via FEN notation, not placeholder data
- **Legal Move Awareness**: AI knows all available legal moves in the current position
- **Check Detection**: AI understands when kings are in check and responds appropriately
- **Game Context**: AI tracks move history, captured pieces, and game progression
- **Strategic Commentary**: AI provides meaningful analysis based on actual position evaluation

### Flexible API Key Management

Use your own API keys for maximum control and cost efficiency:

- **Per-Request Keys**: Specify different API keys for each request
- **Multi-Provider Support**: Switch between OpenAI, Anthropic, Gemini, xAI seamlessly
- **Environment Fallback**: Set default keys via environment variables
- **Cost Control**: Use your preferred provider billing and rate limits

### Enhanced MoveContext

Every AI interaction includes rich game context:

```json
{
  "position": "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1",
  "legal_moves": ["a7a6", "a7a5", "b7b6", "b7b5", "c7c6", "c7c5", "d7d6", "d7d5", "e7e6", "e7e5", "f7f6", "f7f5", "g7g6", "g7g5", "h7h6", "h7h5", "b8a6", "b8c6", "g8f6", "g8h6"],
  "in_check": false,
  "current_player": "black",
  "move_count": 1,
  "last_move": "e2e4",
  "game_status": "active"
}
```

## 🏗️ Project Structure

```text
go-chess/
├── engine/              # Core chess engine
│   ├── board.go         # Board representation
│   ├── game.go          # Game logic and rules
│   ├── board_test.go    # Board tests
│   └── game_test.go     # Game tests
├── ai/                  # AI implementations
│   ├── engine.go        # AI interfaces and implementations
│   └── engine_test.go   # AI tests
├── api/                 # HTTP API server
│   └── server.go        # REST API and WebSocket handlers
├── config/              # Configuration management
│   └── config.go        # Environment-based config
├── examples/            # Example applications
│   ├── cli/             # Command-line interface
│   └── api-server/      # HTTP API server
├── scripts/             # Deployment and automation scripts
│   └── docker-deploy.sh # Docker deployment automation
├── .github/             # GitHub workflows
│   ├── workflows/       # CI/CD pipelines
│   └── dependabot.yml   # Dependency automation
├── Dockerfile           # Multi-stage container build
├── Dockerfile.cli       # CLI container variant
├── docker-compose.yml   # Container orchestration
├── .dockerignore        # Docker build optimization
├── main.go              # Main demonstration app
├── go.mod               # Go module definition
├── Makefile             # Build automation with Docker support
├── README.md            # Project documentation
├── CONTRIBUTING.md      # Contribution guidelines
├── SECURITY.md          # Security policy
├── CHANGELOG.md         # Version history
├── LICENSE.md           # MIT license
└── .env.example         # Environment configuration
```

## 🛠️ Technical Stack

- **Language**: Go 1.23+ (latest features and performance improvements)
- **LLM Integration**: [go-chatbot](https://github.com/RumenDamyanov/go-chatbot) v1.0.2 with multi-provider support
- **Containerization**: Docker with multi-stage builds and security hardening
- **Orchestration**: Docker Compose with health checks and auto-restart
- **Web Framework**: Gin (HTTP API)
- **WebSocket**: Gorilla WebSocket
- **Testing**: Standard Go testing + comprehensive test suite (engine: passing, API: passing, chat: passing)
- **Build System**: Make with Docker integration and automation
- **CI/CD**: GitHub Actions with automated testing and security scanning
- **Code Quality**: golangci-lint, CodeQL, Gosec
- **Documentation**: Extensive API documentation and wiki

## Installation

```bash
go get go.rumenx.com/chess
```

## 🐳 Docker Support

**go-chess** includes comprehensive Docker support for easy deployment and development.

### Quick Start with Docker

```bash
# Build and run with docker-compose
docker-compose up --build

# Or use Make commands
make docker-build
make docker-run

# Development environment
make docker-dev
```

### Docker Features

- **Multi-stage builds** for optimized production images
- **Security hardening** with non-root user and minimal Alpine base
- **Health checks** for container monitoring
- **Environment configuration** with .env support
- **Development mode** with auto-restart
- **Production-ready** deployment automation

### Docker Commands

```bash
# Build Docker image
make docker-build

# Run container in background
make docker-run

# Stop container
make docker-stop

# Start with docker-compose
make docker-compose-up

# Stop docker-compose services
make docker-compose-down

# Development environment with live reload
make docker-dev
```

### Manual Docker Usage

```bash
# Build the image
docker build -t go-chess .

# Run the container
docker run -d --name go-chess -p 8080:8080 \
  -e CHESS_HOST=0.0.0.0 \
  go-chess

# View logs
docker logs go-chess

# Stop container
docker stop go-chess
```

### Docker Compose

```yaml
# docker-compose.yml example
version: '3.8'
services:
  chess-server:
    build: .
    ports:
      - "8080:8080"
    environment:
      - CHESS_HOST=0.0.0.0
      - CHESS_PORT=8080
    healthcheck:
      test: ["CMD", "wget", "--quiet", "--tries=1", "--spider", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
    restart: unless-stopped
```

### Production Deployment

Use the deployment automation script:

```bash
# Full deployment with build and run
./scripts/docker-deploy.sh build

# Development mode with auto-restart
./scripts/docker-deploy.sh dev

# View container logs
./scripts/docker-deploy.sh logs

# Clean up containers and images
./scripts/docker-deploy.sh clean
```

### Basic Usage

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/rumendamyanov/go-chess"
    "github.com/rumendamyanov/go-chess/ai"
    "github.com/rumendamyanov/go-chess/engine"
)

func main() {
    // Create a new chess game
    game := engine.NewGame()

    // Create an AI opponent
    aiPlayer := ai.NewMinimaxAI(ai.DifficultyMedium)

    // Make a move
    move, err := game.ParseMove("e2e4")
    if err != nil {
        log.Fatal(err)
    }

    if err := game.MakeMove(move); err != nil {
        log.Fatal(err)
    }

    fmt.Println("Move made:", move.String())
    fmt.Println("Board state:")
    fmt.Println(game.Board().String())

    // Export current position to FEN notation
    fen := game.ToFEN()
    fmt.Println("Position in FEN:", fen)
    // Output: rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1

    // Get AI response
    ctx := context.Background()
    aiMove, err := aiPlayer.GetBestMove(ctx, game)
    if err != nil {
        log.Fatal(err)
    }

    if err := game.MakeMove(aiMove); err != nil {
        log.Fatal(err)
    }

    fmt.Println("AI played:", aiMove.String())
    fmt.Println("New FEN position:", game.ToFEN())
}
```

### HTTP API Server

```go
package main

import (
    "log"

    "github.com/gin-gonic/gin"
    "github.com/rumendamyanov/go-chess/api"
    "github.com/rumendamyanov/go-chess/config"
)

func main() {
    // Create configuration
    cfg := config.Default()

    // Create API server
    server := api.NewServer(cfg)

    // Setup routes
    r := gin.Default()
    server.SetupRoutes(r)

    // Start server
    log.Println("Starting chess API server on :8080")
    if err := r.Run(":8080"); err != nil {
        log.Fatal("Failed to start server:", err)
    }
}
```

## 🎮 API Endpoints

### Game Management

• `POST /api/games` - Create a new game
• `GET /api/games/{id}` - Get game state
• `DELETE /api/games/{id}` - Delete a game

### Game Actions

• `POST /api/games/{id}/moves` - Make a move
• `GET /api/games/{id}/moves` - Get move history
• `POST /api/games/{id}/ai-move` - Get AI move suggestion

### 🤖 LLM AI Features

• `POST /api/games/{id}/chat` - Chat with your AI opponent
• `POST /api/games/{id}/react` - Get AI reaction to a move

### Game Analysis

• `GET /api/games/{id}/analysis` - Get position analysis
• `GET /api/games/{id}/legal-moves` - Get all legal moves
• `POST /api/games/{id}/fen` - Load position from FEN

### Example API Usage

```bash
# Create a new game
curl -X POST http://localhost:8080/api/games

# Make a move
curl -X POST http://localhost:8080/api/games/1/moves \
  -H "Content-Type: application/json" \
  -d '{"from": "e2", "to": "e4"}'

# Get AI move suggestion
curl -X POST http://localhost:8080/api/games/1/ai-move \
  -H "Content-Type: application/json" \
  -d '{"difficulty": "medium"}'
```

### 🤖 LLM AI Usage Examples

```bash
# Request a move from GPT-4 with custom API key
curl -X POST http://localhost:8080/api/games/1/ai-move \
  -H "Content-Type: application/json" \
  -d '{
    "engine": "llm",
    "provider": "openai",
    "api_key": "your-openai-api-key",
    "level": "expert"
  }'

# Chat with your AI opponent using Claude
curl -X POST http://localhost:8080/api/games/1/chat \
  -H "Content-Type: application/json" \
  -d '{
    "message": "What do you think about my opening?",
    "provider": "anthropic",
    "api_key": "your-anthropic-api-key"
  }'

# Get AI reaction to a brilliant move using xAI Grok
curl -X POST http://localhost:8080/api/games/1/react \
  -H "Content-Type: application/json" \
  -d '{
    "move": "Qh5",
    "provider": "xai",
    "api_key": "your-xai-api-key"
  }'

# Use different providers for different AI personalities
curl -X POST http://localhost:8080/api/games/1/chat \
  -H "Content-Type: application/json" \
  -d '{
    "message": "That was unexpected!",
    "provider": "gemini",
    "api_key": "your-gemini-api-key"
  }'

# Chat without game context (general chess discussion)
curl -X POST http://localhost:8080/api/chat \
  -H "Content-Type: application/json" \
  -d '{
    "message": "Explain the Sicilian Defense",
    "provider": "openai"
  }'
```

### Enhanced API Response Examples

**Chat Response with Rich Game Context:**

```json
{
  "response": "Excellent opening! The King's Pawn opening controls the center and develops quickly. I'm considering Nc6 to challenge your central control.",
  "provider": "anthropic",
  "game_context": {
    "position": "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1",
    "legal_moves": ["a7a6", "a7a5", "b7b6", "b7b5", "c7c6", "c7c5", "d7d6", "d7d5", "e7e6", "e7e5", "f7f6", "f7f5", "g7g6", "g7g5", "h7h6", "h7h5", "b8a6", "b8c6", "g8f6", "g8h6"],
    "in_check": false,
    "current_player": "black",
    "move_count": 1
  },
  "suggestions": [
    "Consider developing your knights first",
    "Control the center with your pawns",
    "Castle early for king safety"
  ]
}
```

## 🔧 Advanced Features Implementation

### Real-time Game Updates

Real-time updates are provided by the built-in WebSocket endpoint exposed by the API server at:

```text
GET /ws/games/:id
```

Each connected client receives JSON payloads with current FEN, move history, and status after state changes. No separate websocket package is required—`api.Server` configures the handler internally. Example (JavaScript):

```javascript
const ws = new WebSocket(`ws://localhost:8080/ws/games/${gameId}`);
ws.onmessage = ev => {
  const update = JSON.parse(ev.data);
  console.log('Game update', update);
};
```

For CLI debugging you can use websocat:

```bash
websocat ws://localhost:8080/ws/games/1
```

### AI Configuration

The core repository currently ships with Random, Minimax, and LLM-backed engines. Additional placeholders (AlphaBeta, MonteCarlo, etc.) shown in earlier docs are not yet implemented. Configure engines as needed, e.g. when wiring custom routing or selection logic.

```go
import "go.rumenx.com/chess/ai"

engines := map[string]ai.Engine{
  "random":  ai.NewRandomAI(),
  "minimax": ai.NewMinimaxAI(ai.DifficultyMedium),
  // LLM engine is created on demand based on request/provider config
}
```

### Game Persistence

PGN export and FEN load are handled via engine methods and API endpoints (e.g. `/api/games/{id}/pgn`, `/api/games/{id}/fen`). Use:

```go
fen := game.ToFEN()
// To load: engine.NewGameFromFEN(fen) or API POST /api/games/{id}/fen
```

No separate `persistence` package is currently included—older docs referenced a future module.

## Testing

```bash
# Run all tests (core + examples)
go test ./...
make test

# Coverage (core packages only – examples excluded to avoid dilution)
make test-coverage         # uses filtered package list

# Full coverage including examples (may lower % because examples are illustrative)
make test-coverage-all

# Benchmarks
go test -bench=. ./...
make bench

# Build example binaries
make build-examples

# Clean build artifacts
make clean
```

Coverage strategy: examples are intentionally excluded from the primary badge to focus on engine/API/chat/config quality. Use `test-coverage-all` for a holistic view.

## Build & Development

```bash
# Build main application (CLI by default)
make build

# Build specific examples
make build-cli
make build-server

# Run examples
make run-cli
make run-server

# Docker workflow
make docker-build      # Build Docker image
make docker-run        # Run container
make docker-dev        # Development mode
make docker-stop       # Stop container

# Development tools
make fmt               # Format code
make vet               # Vet code
make lint              # Run linter
make help              # Show all available targets
```

## Static Analysis

```bash
# Run linter
golangci-lint run

# Security scan
gosec ./...

# Dependency check
go mod verify
```

## 🔧 Configuration

Environment variables and configuration options:

```bash
# Server configuration
export CHESS_PORT=8080
export CHESS_HOST=localhost

# AI configuration
export CHESS_AI_TIMEOUT=30s
export CHESS_AI_DEFAULT_DIFFICULTY=medium

# LLM Provider API Keys (use your own for better performance)
export OPENAI_API_KEY=your-openai-key
export ANTHROPIC_API_KEY=your-anthropic-key
export GEMINI_API_KEY=your-gemini-key
export XAI_API_KEY=your-xai-key

# Logging
export CHESS_LOG_LEVEL=info
export CHESS_LOG_FORMAT=json
```

## 🆕 Recent Enhancements

**✨ What's New in the Latest Version:**

🧠 **Real FEN Support**: Complete Forsyth-Edwards Notation implementation

- Export any board position to standard FEN format
- Real-time position tracking for AI analysis
- Comprehensive castling rights, en passant, and move clock tracking

🤖 **Enhanced AI Integration**: Genuine chess intelligence instead of placeholders

- Rich MoveContext with legal moves and check status
- Custom API keys for cost control and better rate limits
- Multi-provider support with graceful fallbacks
- Strategic AI commentary based on actual board positions

🔧 **Improved Developer Experience**: Better testing and documentation

- All test suites passing with comprehensive coverage
- Enhanced API responses with game context
- Production-ready deployment with Docker
- Detailed examples and enhanced documentation

## Configuration

Environment variables and configuration options:

```bash
# Server configuration
export CHESS_PORT=8080
export CHESS_HOST=localhost

# AI configuration
export CHESS_AI_TIMEOUT=30s
export CHESS_AI_DEFAULT_DIFFICULTY=medium

# Logging
export CHESS_LOG_LEVEL=info
export CHESS_LOG_FORMAT=json
```

## Frontend Integration

> 🎮 **Live Demo**: Check out [js-chess](https://github.com/RumenDamyanov/js-chess) - a complete JavaScript frontend showcase that uses go-chess as its backend, featuring interactive chess gameplay with AI opponents and real-time chat.

The API is designed to work seamlessly with frontend applications. Example integration with a JavaScript chess UI:

```javascript
// Create a new game
const response = await fetch('/api/games', { method: 'POST' });
const game = await response.json();

// Make a move
await fetch(`/api/games/${game.id}/moves`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ from: 'e2', to: 'e4' })
});

// Connect to WebSocket for real-time updates
const ws = new WebSocket(`ws://localhost:8080/ws/games/${game.id}`);
ws.onmessage = (event) => {
    const update = JSON.parse(event.data);
    updateBoard(update.board);
};
```

## Performance Characteristics

- **Move Generation**: ~50,000 moves/second on modern hardware
- **Position Evaluation**: ~10,000 positions/second
- **FEN Generation**: Real-time position export in standard Forsyth-Edwards Notation
- **Memory Usage**: <10MB for typical game states
- **AI Response Time**: 50ms-5s depending on difficulty level and LLM provider
- **Build Time**: <5 seconds for full project, ~7 seconds for Docker image
- **Test Coverage**: All packages passing with comprehensive test suites
- **Core Coverage**: ~81% (engine, ai, api, chat, config) with examples excluded
- **Docker Image Size**: ~15MB (multi-stage Alpine-based build)
- **Container Startup**: <2 seconds with health checks
- **LLM Integration**: Sub-second response times with proper API keys

## 🧪 Testing & Quality Assurance

### Testing Strategy

- **Unit Tests**: Core engine and AI logic
- **Benchmark Tests**: Performance measurement
- **Integration Tests**: API endpoint testing
- **Security Tests**: CodeQL and Gosec vulnerability scanning

### CI/CD Pipeline

- **Automated Testing**: On every push/PR
- **Code Quality**: Static analysis and linting
- **Security Scanning**: CodeQL and Gosec
- **Dependency Updates**: Automated with Dependabot
- **Coverage Reporting**: Codecov integration

## 🎯 Learning Outcomes & Best Practices

This project demonstrates:

- **Clean Architecture**: Separation of concerns with modular design
- **Go Best Practices**: Idiomatic Go code following community standards
- **Testing Excellence**: Comprehensive test coverage and benchmarking
- **API Design**: RESTful and WebSocket APIs for modern applications
- **DevOps**: CI/CD automation, monitoring, and quality checks
- **Security**: Vulnerability scanning and secure coding practices
- **Documentation**: Clear, comprehensive documentation and examples

## 🏆 Production Readiness

The project includes:

- ✅ **Error Handling**: Comprehensive error handling and logging
- ✅ **Configuration**: Environment-based configuration management
- ✅ **Security**: Input validation, rate limiting, and vulnerability scanning
- ✅ **Performance**: Optimized algorithms and efficient data structures
- ✅ **Monitoring**: Health checks and performance metrics
- ✅ **Documentation**: Extensive API documentation and user guides
- ✅ **Testing**: Automated testing and deployment pipelines
- ✅ **Containerization**: Docker support with multi-stage builds and security hardening
- ✅ **Orchestration**: Docker Compose with health checks and auto-restart
- ✅ **Deployment**: Automated deployment scripts and Make targets

## Security & Best Practices

• Input validation for all move commands
• Rate limiting for API endpoints
• Secure WebSocket connections
• Comprehensive error handling
• Structured logging with context

## Example Projects

The `examples/` directory contains complete example applications:
• **CLI Game** (`examples/cli`) – minimal interactive CLI
• **API Server** (`examples/api-server`) – standalone HTTP server
• **Minimal Server** (`examples/minimal-server`) – smallest runnable demo
• **Test Server** (`examples/test-server`) – utility server for integration tests

```bash
# Run CLI example
go run examples/cli/main.go

# Run API server example
go run examples/api-server/main.go

# Run minimal server
go run examples/minimal-server/minimal_server.go

# Run test server (used internally)
go run examples/test-server/test_server.go
```

Note: Some earlier documentation referenced a tournament and dedicated websocket demo—those have not been merged yet.

## Contributing

See [CONTRIBUTING.md](https://github.com/RumenDamyanov/go-chess/blob/master/CONTRIBUTING.md) for guidelines on how to contribute to this project.

## Code of Conduct

This project adheres to a Code of Conduct to ensure a welcoming environment for all contributors. See [CODE_OF_CONDUCT.md](https://github.com/RumenDamyanov/go-chess/blob/master/CODE_OF_CONDUCT.md) for details.

## Security

For security vulnerabilities, please see our [Security Policy](https://github.com/RumenDamyanov/go-chess/blob/master/SECURITY.md).

## Changelog

See [CHANGELOG.md](https://github.com/RumenDamyanov/go-chess/blob/master/CHANGELOG.md) for a detailed history of changes and releases.

## License

MIT License. See [LICENSE.md](https://github.com/RumenDamyanov/go-chess/blob/master/LICENSE.md) for details.
