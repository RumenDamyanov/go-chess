# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Initial chess engine implementation
- Complete board representation and move validation
- AI opponents with configurable difficulty levels
- RESTful API for game management
- WebSocket support for real-time updates
- CLI example application
- Comprehensive test suite
- CI/CD pipeline with GitHub Actions
- Code coverage reporting with Codecov
- Security scanning with CodeQL and Gosec
- Automated dependency updates with Dependabot

### Core Features
- Full chess rule implementation including castling, en passant, and pawn promotion
- Multiple AI algorithms: Random, Minimax with alpha-beta pruning
- Game state management with move history and status tracking
- FEN notation support (planned)
- PGN export/import (planned)

### API Endpoints
- `POST /api/games` - Create new game
- `GET /api/games/{id}` - Get game state
- `POST /api/games/{id}/moves` - Make move
- `POST /api/games/{id}/ai-move` - Get AI move suggestion
- `GET /api/games/{id}/legal-moves` - Get legal moves
- `GET /api/games/{id}/analysis` - Position analysis
- WebSocket endpoint for real-time updates

### Development
- Clean architecture with separated concerns
- Comprehensive unit and integration tests
- Benchmarks for performance testing
- Static analysis with golangci-lint
- Security scanning and vulnerability detection
- Automated code quality checks

## [1.0.0] - 2025-08-02

### Added
- Initial release of go-chess
- Basic chess engine with move validation
- Simple AI opponents
- HTTP API for frontend integration
- CLI interface for interactive play
- Complete documentation and examples
