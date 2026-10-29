# Changelog

## [1.1.0] - 2026-10-28

### Security

- Require Go 1.26.10 or newer. Go 1.25 and older are end of life and do not include the October 2026 toolchain fixes.
- Raise `golang.org/x/net` to v0.60.0 and `golang.org/x/crypto` to v0.57.0. Gin stays on v1.12.0 and `github.com/quic-go/quic-go` stays on v0.59.1. Older copies are reachable from HTTP handlers.
- The server build does not import `github.com/andybalholm/brotli`. Versions before v1.2.7 panic on crafted Brotli input, so any future import must be v1.2.7 or newer.

### Changed

- Default chat and LLM models are now OpenAI `gpt-5.4-mini`, Anthropic `claude-sonnet-5`, Google `gemini-3.8-flash`, and xAI `grok-4.6`. DeepSeek stays on `deepseek-chat`. Older ids still work when set explicitly.
- CI tests Go 1.26.10 and 1.27.3. Container builds use `golang:1.26.10-alpine`.

## [1.0.5] - 2025-08-10

### Added

- Evaluation_after / evaluation_diff (and *_cp variants) on AI move & hint endpoints.
- Material & mobility analysis fields.
- PGN: Variant & Annotator tags, dynamic player names, SetUp/FEN for custom starts.
- SAN improvements: promotion notation, en passant, check (+), mate (#), disambiguation tests (N1d2, N4f3, Raxd1).

### Changed

- Minimal SAN disambiguation logic clarified.
- Health endpoint version extracted to constant.

### Future

- Additional disambiguation edge cases.
- Enhanced evaluation (mobility weighting, PST tables).
- WebSocket structured event stream.

<!-- Previous release history trimmed for brevity in this upstream patch. -->
