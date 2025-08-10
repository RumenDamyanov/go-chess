# Changelog

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
