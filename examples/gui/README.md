# go-chess GUI Demo (Ebiten)

Lightweight integrated graphical demo showcasing the core engine without the HTTP API.

## Features

- Interactive board (click to select & move)
- Legal move highlighting
- Last move highlight
- Human vs Human or Human vs AI (Minimax) toggle
- Difficulty cycle (Beginner->Easy->Medium->Hard->Expert)
- Evaluation score (centipawns, from White POV)
- New Game reset

## Run

```bash
make run-gui
```

## Controls

- Left click: select piece / destination
- Space: cycle AI difficulty
- A: toggle Human vs AI mode
- N: new game
- E: evaluate position (updates score display)
- Esc / Q: quit

## Notes

- AI search uses a simple evaluation + shallow depth (per existing Minimax implementation).
- GUI excluded from coverage metrics (example only).
- Assets fall back to letter rendering if piece images missing.
