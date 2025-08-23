package main

import (
	"context"
	"fmt"
	"image/color"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"go.rumenx.com/chess/ai"
	"go.rumenx.com/chess/engine"
)

// GameMode represents play configuration.
type GameMode int

const (
	HumanVsHuman GameMode = iota
	HumanVsAI
)

// uiGame implements ebiten.Game
// Keep logic minimal – demo only.
type uiGame struct {
	g            *engine.Game
	mode         GameMode
	aiEngine     ai.Engine
	aiPending    bool
	aiCtxCancel  context.CancelFunc
	selected     *engine.Square
	legalTargets map[engine.Square]bool
	legalMoves   map[engine.Square]engine.Move
	lastMove     *engine.Move

	movesSAN []string
	// track if last action was an undo (affects messaging)
	lastUndone    bool
	statusText    string
	difficulty    ai.Difficulty
	evalScore     *int
	msg           string
	msgUntil      time.Time
	playerColor   engine.Color
	whiteAtBottom bool
	// piece cache for vector or image drawings
	pieceCache    map[string]*ebiten.Image
	imageBaseDir  string // directory containing piece assets (svg or png)
	imageCheckRun bool   // only check presence once
	cursorX       int
	cursorY       int
	// svg rasterization
	rasterTool   string // selected external tool (rsvg-convert or inkscape)
	rasterWarned bool   // logged missing tool once
	wasMouseDown bool   // for edge-trigger mouse click detection
}

const (
	boardPixels = 640
	squareSize  = boardPixels / 8
	panelWidth  = 240
	windowW     = boardPixels + panelWidth
	windowH     = boardPixels
)

func newUIGame() *uiGame {
	g := engine.NewGame()
	ug := &uiGame{
		g:             g,
		mode:          HumanVsAI,
		aiEngine:      ai.NewMinimaxAI(ai.DifficultyMedium),
		difficulty:    ai.DifficultyMedium,
		legalTargets:  map[engine.Square]bool{},
		legalMoves:    map[engine.Square]engine.Move{},
		playerColor:   engine.White,
		whiteAtBottom: true,
		pieceCache:    map[string]*ebiten.Image{},
		imageBaseDir:  "examples/gui/assets/pieces",
	}
	ug.detectRasterTool()
	return ug
}

func (u *uiGame) Layout(outsideWidth, outsideHeight int) (int, int) { return windowW, windowH }

func (u *uiGame) Update() error {
	// Handle quit keys
	if ebiten.IsKeyPressed(ebiten.KeyEscape) || ebiten.IsKeyPressed(ebiten.KeyQ) {
		return ebiten.Termination
	}

	// Timed message expiration
	if !u.msgUntil.IsZero() && time.Now().After(u.msgUntil) {
		u.msg = ""
	}

	// Input handling
	u.cursorX, u.cursorY = ebiten.CursorPosition()
	u.handleKeys()
	u.handleMouse()

	// If AI move pending, poll (goroutine will set lastMove when done)
	if u.mode == HumanVsAI && !u.aiPending {
		if u.g.ActiveColor() == u.aiColor() {
			u.startAIMove()
		}
	}

	return nil
}

func (u *uiGame) handleKeys() {
	if ebiten.IsKeyPressed(ebiten.KeyN) {
		u.resetGame(u.playerColor)
	}
	if ebiten.IsKeyPressed(ebiten.KeyU) { // undo last move
		u.handleUndo()
	}
	if ebiten.IsKeyPressed(ebiten.KeySpace) {
		u.cycleDifficulty()
	}
	if ebiten.IsKeyPressed(ebiten.KeyA) {
		if u.mode == HumanVsAI {
			u.mode = HumanVsHuman
		} else {
			u.mode = HumanVsAI
		}
		u.flashMsg("Mode: " + u.modeString())
		u.selected = nil
	}
	if ebiten.IsKeyPressed(ebiten.KeyE) {
		s := u.g.Evaluate()
		u.evalScore = &s
	}
	if ebiten.IsKeyPressed(ebiten.KeyF) {
		u.whiteAtBottom = !u.whiteAtBottom
	}
	// style toggle removed (always attempt images)
	if len(u.g.MoveHistory()) == 0 { // allow choosing color before first move only
		if ebiten.IsKeyPressed(ebiten.KeyW) && u.playerColor != engine.White {
			u.playerColor = engine.White
			u.whiteAtBottom = true
			u.resetGame(u.playerColor)
		}
		if ebiten.IsKeyPressed(ebiten.KeyB) && u.playerColor != engine.Black {
			u.playerColor = engine.Black
			u.whiteAtBottom = false
			u.resetGame(u.playerColor)
		}
	}
}

func (u *uiGame) handleMouse() {
	pressed := ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)
	if !pressed {
		u.wasMouseDown = false
		return
	}
	// Only act on transition (just pressed)
	if u.wasMouseDown {
		return
	}
	u.wasMouseDown = true
	x, y := ebiten.CursorPosition()
	// Panel clicks
	if x >= boardPixels {
		relX := x - boardPixels
		relY := y
		// Mode toggle button at y 8-28
		if relY >= 8 && relY < 28 && relX >= 8 && relX < 8+120 {
			if u.mode == HumanVsAI {
				u.mode = HumanVsHuman
			} else {
				u.mode = HumanVsAI
			}
			u.flashMsg("Mode: " + u.modeString())
			return
		}
		// Undo button (y 32-52)
		if relY >= 32 && relY < 52 && relX >= 8 && relX < 8+120 {
			u.handleUndo()
			return
		}
		// color selection boxes (moved up after removing style button)
		if relY >= 56 && relY < 76 && len(u.g.MoveHistory()) == 0 && relX >= 8 && relX < 8+90 {
			if u.playerColor != engine.White {
				u.playerColor = engine.White
				u.whiteAtBottom = true
				u.resetGame(u.playerColor)
			}
			return
		} else if relY >= 81 && relY < 101 && len(u.g.MoveHistory()) == 0 && relX >= 8 && relX < 8+90 {
			if u.playerColor != engine.Black {
				u.playerColor = engine.Black
				u.whiteAtBottom = false
				u.resetGame(u.playerColor)
			}
			return
		}
		// difficulty list starting at new y 135, each 22px height
		if relY >= 135 && relY < 135+5*22 && relX >= 8 && relX < 8+120 {
			idx := (relY - 135) / 22
			order := []ai.Difficulty{ai.DifficultyBeginner, ai.DifficultyEasy, ai.DifficultyMedium, ai.DifficultyHard, ai.DifficultyExpert}
			if idx >= 0 && idx < len(order) {
				if u.difficulty != order[idx] {
					u.difficulty = order[idx]
					u.aiEngine = ai.NewMinimaxAI(u.difficulty)
					u.flashMsg("Difficulty: " + u.difficultyLabel())
				}
			}
			return
		}
		_ = relX // reserved for possible future use
		return
	}
	// Board clicks
	if x < 0 || x >= boardPixels || y < 0 || y >= boardPixels {
		return
	}
	file := x / squareSize
	var rank int
	if u.whiteAtBottom {
		rank = 7 - (y / squareSize)
	} else {
		rank = y / squareSize
	}
	sq := engine.Square(rank*8 + file)

	if u.selected == nil {
		p := u.g.Board().GetPiece(sq)
		if p.IsEmpty() || p.Color != u.g.ActiveColor() {
			return
		}
		u.selected = &sq
		u.computeLegalTargets()
		return
	}
	if *u.selected == sq { // deselect
		u.selected = nil
		u.legalTargets = map[engine.Square]bool{}
		return
	}
	if u.legalTargets[sq] { // perform move
		// Prefer using precomputed legal move (handles promotions, castling etc.)
		if mv, ok := u.legalMoves[sq]; ok {
			// Auto-queen promotion if needed (simplified)
			if mv.Piece.Type == engine.Pawn && (sq.Rank() == 7 || sq.Rank() == 0) && mv.Type != engine.Promotion {
				mv.Type = engine.Promotion
				mv.Promotion = engine.Queen
			}
			u.applyMove(mv)
		} else {
			// Fallback parse
			notation := u.selected.String() + sq.String()
			mv, err := u.g.ParseMove(notation)
			if err == nil && u.g.IsLegalMove(mv) {
				// auto queen if reaches end rank
				if mv.Piece.Type == engine.Pawn && (sq.Rank() == 7 || sq.Rank() == 0) && mv.Type != engine.Promotion {
					mv.Type = engine.Promotion
					mv.Promotion = engine.Queen
				}
				u.applyMove(mv)
			}
		}
	}
	u.selected = nil
	u.legalTargets = map[engine.Square]bool{}
}

func (u *uiGame) Draw(screen *ebiten.Image) {
	u.drawBoard(screen)
	u.drawHighlights(screen)
	u.drawPieces(screen)
	u.drawPanel(screen)
}

func (u *uiGame) drawBoard(screen *ebiten.Image) {
	light := color.RGBA{0xEE, 0xD9, 0xB6, 0xFF}
	dark := color.RGBA{0xB5, 0x88, 0x63, 0xFF}
	for vrank := 0; vrank < 8; vrank++ { // visual rank top->bottom
		for file := 0; file < 8; file++ {
			c := light
			if (vrank+file)%2 == 1 {
				c = dark
			}
			rect := ebiten.NewImage(squareSize, squareSize)
			rect.Fill(c)
			y := vrank * squareSize
			screen.DrawImage(rect, &ebiten.DrawImageOptions{GeoM: translate(file*squareSize, y)})
		}
	}
}

func (u *uiGame) drawPieces(screen *ebiten.Image) {
	board := u.g.Board()
	for sq := engine.Square(0); sq < 64; sq++ {
		p := board.GetPiece(sq)
		if p.IsEmpty() {
			continue
		}
		file := sq.File()
		rank := sq.Rank()
		vx, vy := file, rank
		if u.whiteAtBottom {
			vy = 7 - rank
		}
		x := vx * squareSize
		y := vy * squareSize
		img := u.pieceImage(p)
		screen.DrawImage(img, &ebiten.DrawImageOptions{GeoM: translate(x, y)})
	}
}

func (u *uiGame) drawHighlights(screen *ebiten.Image) {
	if u.lastMove != nil {
		u.highlightSquare(screen, u.lastMove.From, color.RGBA{0x66, 0xFF, 0x66, 0x55})
		u.highlightSquare(screen, u.lastMove.To, color.RGBA{0x66, 0xFF, 0x66, 0x55})
	}
	if u.selected != nil {
		u.highlightSquare(screen, *u.selected, color.RGBA{0x33, 0x66, 0xFF, 0x66})
		for tgt := range u.legalTargets {
			u.highlightCircle(screen, tgt, color.RGBA{0x33, 0x66, 0xFF, 0xAA})
		}
	}
}

func (u *uiGame) highlightSquare(screen *ebiten.Image, sq engine.Square, c color.Color) {
	file := sq.File()
	rank := sq.Rank()
	vrank := rank
	if u.whiteAtBottom {
		vrank = 7 - rank
	}
	x := file * squareSize
	y := vrank * squareSize
	o := ebiten.NewImage(squareSize, squareSize)
	o.Fill(c)
	screen.DrawImage(o, &ebiten.DrawImageOptions{GeoM: translate(x, y)})
}

func (u *uiGame) highlightCircle(screen *ebiten.Image, sq engine.Square, c color.Color) {
	file := sq.File()
	rank := sq.Rank()
	vrank := rank
	if u.whiteAtBottom {
		vrank = 7 - rank
	}
	x := file*squareSize + squareSize/2 - 8
	y := vrank*squareSize + squareSize/2 - 8
	circ := ebiten.NewImage(16, 16)
	circ.Fill(c)
	screen.DrawImage(circ, &ebiten.DrawImageOptions{GeoM: translate(x, y)})
}

func (u *uiGame) drawPanel(screen *ebiten.Image) {
	x0 := boardPixels
	panel := ebiten.NewImage(panelWidth, windowH)
	panel.Fill(color.RGBA{0x22, 0x22, 0x22, 0xFF})
	screen.DrawImage(panel, &ebiten.DrawImageOptions{GeoM: translate(x0, 0)})

	// status strings now built later in infoLines
	var eval string
	if u.evalScore != nil {
		eval = evalString(*u.evalScore)
	}
	// Hover detection helper
	hover := func(x, y, w, h int) bool {
		return u.cursorX >= boardPixels+x && u.cursorX < boardPixels+x+w && u.cursorY >= y && u.cursorY < y+h
	}
	// Mode toggle
	drawSelectableBox(screen, x0+8, 8, 120, 20, u.modeString(), hover(8, 8, 120, 20))
	drawSelectableBox(screen, x0+8, 32, 120, 20, "Undo (U)", hover(8, 32, 120, 20))
	ebitenutil.DebugPrintAt(screen, "Color:", x0+8, 58)
	drawSelectableBox(screen, x0+8, 56, 90, 20, "White", (u.playerColor == engine.White && len(u.g.MoveHistory()) == 0) || hover(8, 56, 90, 20))
	drawSelectableBox(screen, x0+8, 81, 90, 20, "Black", (u.playerColor == engine.Black && len(u.g.MoveHistory()) == 0) || hover(8, 81, 90, 20))
	if len(u.g.MoveHistory()) > 0 {
		ebitenutil.DebugPrintAt(screen, "(locked)", x0+104, 64)
	}
	// Difficulty list
	ebitenutil.DebugPrintAt(screen, "Difficulty:", x0+8, 115)
	diffs := []ai.Difficulty{ai.DifficultyBeginner, ai.DifficultyEasy, ai.DifficultyMedium, ai.DifficultyHard, ai.DifficultyExpert}
	for i, d := range diffs {
		sel := d == u.difficulty
		label := u.difficultyCustomLabel(d)
		drawSelectableBox(screen, x0+8, 135+i*22, 120, 20, label, sel || hover(8, 135+i*22, 120, 20))
	}
	// Info section starting below selectors
	infoY := 135 + len(diffs)*22 + 16
	infoLines := []string{"Status: " + u.g.Status().String(), "Turn: " + u.g.ActiveColor().String(), "Player: " + u.playerColor.String(), "Diff: " + u.difficultyLabel(), "Moves: " + stringFromInt(len(u.g.MoveHistory()))}
	if u.evalScore != nil {
		infoLines = append(infoLines, eval)
	}
	if u.msg != "" {
		infoLines = append(infoLines, "Msg: "+u.msg)
	}
	for i, l := range infoLines {
		ebitenutil.DebugPrintAt(screen, l, x0+8, infoY+i*14)
	}
	// SAN list (last 14 entries)
	sanStartY := infoY + len(infoLines)*14 + 12
	ebitenutil.DebugPrintAt(screen, "SAN (latest):", x0+8, sanStartY)
	maxShow := 14
	san := u.movesSAN
	if len(san) > maxShow {
		san = san[len(san)-maxShow:]
	}
	for i, mv := range san {
		ebitenutil.DebugPrintAt(screen, mv, x0+8, sanStartY+14*(i+1))
	}
	// Help at bottom
	ebitenutil.DebugPrintAt(screen, "Keys: N=new A=mode F=flip E=eval U=undo", x0+8, windowH-40)
	ebitenutil.DebugPrintAt(screen, "Click to select pieces & buttons", x0+8, windowH-24)
}

func (u *uiGame) computeLegalTargets() {
	u.legalTargets = map[engine.Square]bool{}
	u.legalMoves = map[engine.Square]engine.Move{}
	if u.selected == nil {
		return
	}
	all := u.g.GetAllLegalMoves()
	for _, mv := range all {
		if mv.From == *u.selected {
			u.legalTargets[mv.To] = true
			u.legalMoves[mv.To] = mv
		}
	}
}

func (u *uiGame) applyMove(mv engine.Move) {
	if err := u.g.MakeMove(mv); err != nil {
		u.flashMsg("Illegal move")
		return
	}
	u.lastMove = &mv
	u.movesSAN = u.g.GenerateSAN()
	u.evalScore = nil
	u.lastUndone = false
}

// handleUndo attempts to undo the last move (single ply) if available.
// It prevents undoing while an AI move is pending.
func (u *uiGame) handleUndo() {
	if u.aiPending {
		u.flashMsg("AI thinking")
		return
	}
	// In HumanVsAI mode, if it's player's turn then last move was AI's; undo twice to revert player's last move.
	undoCount := 1
	if u.mode == HumanVsAI && u.g.ActiveColor() == u.playerColor {
		undoCount = 2
	}
	undone := 0
	for i := 0; i < undoCount; i++ {
		mv, err := u.g.UndoMove()
		if err != nil {
			if undone == 0 {
				u.flashMsg("No move to undo")
			}
			break
		}
		undone++
		_ = mv
	}
	if undone > 0 {
		u.movesSAN = u.g.GenerateSAN()
		u.lastMove = nil
		u.evalScore = nil
		u.lastUndone = true
		if undone == 2 {
			u.flashMsg("Undid your last move")
		} else {
			u.flashMsg("Undid move")
		}
	}
}

func (u *uiGame) startAIMove() {
	if u.aiPending {
		return
	}
	u.aiPending = true
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	u.aiCtxCancel = cancel
	gameCopy := u.g // using game directly; assumption: no concurrent human move until AI done
	go func() {
		mv, err := u.aiEngine.GetBestMove(ctx, gameCopy)
		if err == nil {
			_ = u.g.MakeMove(mv)
			u.lastMove = &mv
			u.movesSAN = u.g.GenerateSAN()
		}
		u.aiPending = false
		cancel()
	}()
}

func (u *uiGame) cycleDifficulty() {
	order := []ai.Difficulty{ai.DifficultyBeginner, ai.DifficultyEasy, ai.DifficultyMedium, ai.DifficultyHard, ai.DifficultyExpert}
	idx := 0
	for i, d := range order {
		if d == u.difficulty {
			idx = (i + 1) % len(order)
			break
		}
	}
	u.difficulty = order[idx]
	u.aiEngine = ai.NewMinimaxAI(u.difficulty)
	u.flashMsg("Difficulty: " + u.difficultyLabel())
}

func (u *uiGame) resetGame(color engine.Color) {
	u.g = engine.NewGame()
	u.selected = nil
	u.lastMove = nil
	u.movesSAN = nil
	u.evalScore = nil
	u.aiPending = false
	u.flashMsg("New game (" + color.String() + ")")
}

func (u *uiGame) flashMsg(m string) { u.msg = m; u.msgUntil = time.Now().Add(2 * time.Second) }

func (u *uiGame) modeString() string {
	if u.mode == HumanVsAI {
		return "Human vs AI"
	}
	return "Human vs Human"
}

func (u *uiGame) difficultyLabel() string {
	switch u.difficulty {
	case ai.DifficultyBeginner:
		return "harmless"
	case ai.DifficultyEasy:
		return "easy"
	case ai.DifficultyMedium:
		return "normal"
	case ai.DifficultyHard:
		return "hard"
	case ai.DifficultyExpert:
		return "godlike"
	default:
		return u.difficulty.String()
	}
}

// label helper for arbitrary difficulty (used in panel list)
func (u *uiGame) difficultyCustomLabel(d ai.Difficulty) string {
	switch d {
	case ai.DifficultyBeginner:
		return "harmless"
	case ai.DifficultyEasy:
		return "easy"
	case ai.DifficultyMedium:
		return "normal"
	case ai.DifficultyHard:
		return "hard"
	case ai.DifficultyExpert:
		return "godlike"
	default:
		return d.String()
	}
}

func (u *uiGame) aiColor() engine.Color {
	if u.mode != HumanVsAI {
		return engine.None
	}
	if u.playerColor == engine.White {
		return engine.Black
	}
	return engine.White
}

// pieceChar maps an engine.Piece to an ASCII representation.
// pieceImage returns (and caches) a stylized piece image built from primitives.
func (u *uiGame) pieceImage(p engine.Piece) *ebiten.Image {
	key := "piece:" + p.Color.String() + ":" + p.Type.String()
	if img, ok := u.pieceCache[key]; ok {
		return img
	}
	// Attempt to rasterize SVG if available
	if u.imageAssetsAvailable() {
		// Try pre-rasterized PNG first (bundled for environments without raster tools)
		pngName := fmt.Sprintf("%s_%s.png", pieceColorCode(p.Color), pieceTypeCode(p.Type))
		pngPath := filepath.Join(u.imageBaseDir, pngName)
		if fimg, _, err := ebitenutil.NewImageFromFile(pngPath); err == nil {
			w, h := fimg.Size()
			if w != squareSize || h != squareSize {
				canvas := ebiten.NewImage(squareSize, squareSize)
				op := &ebiten.DrawImageOptions{}
				scaleX := float64(squareSize) / float64(w)
				scaleY := float64(squareSize) / float64(h)
				s := scaleX
				if scaleY < s {
					s = scaleY
				}
				op.GeoM.Scale(s, s)
				op.GeoM.Translate(float64(squareSize)/2-float64(w)*s/2, float64(squareSize)/2-float64(h)*s/2)
				canvas.DrawImage(fimg, op)
				fimg = canvas
			}
			u.pieceCache[key] = fimg
			return fimg
		}
		// Else attempt SVG rasterization
		svgName := fmt.Sprintf("%s_%s.svg", pieceColorCode(p.Color), pieceTypeCode(p.Type))
		svgPath := filepath.Join(u.imageBaseDir, svgName)
		if _, err := os.Stat(svgPath); err == nil && u.rasterTool != "" {
			if img := rasterizeSVGWithTool(u.rasterTool, svgPath, squareSize, squareSize); img != nil {
				u.pieceCache[key] = img
				return img
			}
		} else if err == nil && u.rasterTool == "" && !u.rasterWarned {
			log.Println("[gui] SVG files present but no rasterization tool found (install 'rsvg-convert' via 'brew install librsvg' or 'inkscape'). Falling back to vector shapes.")
			u.rasterWarned = true
		}
	}
	// Fallback vector drawing
	img := ebiten.NewImage(squareSize, squareSize)
	img.Fill(color.RGBA{0, 0, 0, 0})
	baseLight := color.RGBA{0xF6, 0xF6, 0xF6, 0xFF}
	outlineLight := color.RGBA{0x33, 0x33, 0x33, 0xFF}
	baseDark := color.RGBA{0x22, 0x22, 0x22, 0xFF}
	outlineDark := color.RGBA{0xEE, 0xEE, 0xEE, 0xFF}
	fillCol := baseLight
	lineCol := outlineLight
	if p.Color == engine.Black {
		fillCol = baseDark
		lineCol = outlineDark
	}
	fillRect := func(x, y, w, h int, c color.Color) {
		r := ebiten.NewImage(w, h)
		r.Fill(c)
		img.DrawImage(r, &ebiten.DrawImageOptions{GeoM: translate(x, y)})
	}
	fillCircle := func(cx, cy, r int, c color.Color) {
		for yy := -r; yy <= r; yy++ {
			for xx := -r; xx <= r; xx++ {
				if xx*xx+yy*yy <= r*r {
					px := cx + xx
					py := cy + yy
					if px >= 0 && py >= 0 && px < squareSize && py < squareSize {
						img.Set(px, py, c)
					}
				}
			}
		}
	}
	switch p.Type {
	case engine.Pawn:
		fillCircle(squareSize/2, squareSize/3, squareSize/6, fillCol)
		fillRect(squareSize/2-5, squareSize/3, 10, squareSize/2, fillCol)
		fillRect(squareSize/2-12, squareSize-18, 24, 6, fillCol)
	case engine.Rook:
		fillRect(squareSize/4, squareSize/4, squareSize/2, squareSize/2+8, fillCol)
		for i := 0; i < 4; i++ {
			fillRect(squareSize/4+i*(squareSize/8), squareSize/4-6, squareSize/10, 6, fillCol)
		}
		fillRect(squareSize/4-6, squareSize-20, squareSize/2+12, 6, fillCol)
	case engine.Knight:
		fillRect(squareSize/3, squareSize/3, squareSize/3+6, squareSize/2+6, fillCol)
		fillCircle(squareSize/2+8, squareSize/3+4, squareSize/6+2, fillCol)
		fillRect(squareSize/3-8, squareSize-20, squareSize/2+24, 6, fillCol)
	case engine.Bishop:
		fillCircle(squareSize/2, squareSize/3, squareSize/6+2, fillCol)
		fillRect(squareSize/2-5, squareSize/3, 10, squareSize/2, fillCol)
		fillCircle(squareSize/2, squareSize/2+6, squareSize/4, fillCol)
		fillRect(squareSize/2-12, squareSize-20, 24, 6, fillCol)
	case engine.Queen:
		fillRect(squareSize/3, squareSize/3, squareSize/3, squareSize/2+6, fillCol)
		for i := 0; i < 5; i++ {
			fillCircle(squareSize/3+i*(squareSize/15)+6, squareSize/3-6, squareSize/12, fillCol)
		}
		fillCircle(squareSize/2, squareSize/2+2, squareSize/3, fillCol)
		fillRect(squareSize/3-8, squareSize-20, squareSize/3+16, 6, fillCol)
	case engine.King:
		fillRect(squareSize/3, squareSize/3, squareSize/3, squareSize/2+8, fillCol)
		fillRect(squareSize/2-4, squareSize/4, 8, squareSize/5, fillCol)
		fillRect(squareSize/2-14, squareSize/4+8, 28, 6, fillCol)
		fillRect(squareSize/3-8, squareSize-20, squareSize/3+16, 6, fillCol)
	}
	// border
	for x := 0; x < squareSize; x++ {
		img.Set(x, 0, lineCol)
		img.Set(x, squareSize-1, lineCol)
	}
	for y := 0; y < squareSize; y++ {
		img.Set(0, y, lineCol)
		img.Set(squareSize-1, y, lineCol)
	}
	u.pieceCache[key] = img
	return img
}

func (u *uiGame) imageAssetsAvailable() bool {
	if u.imageCheckRun {
		return u.imageBaseDir != ""
	}
	u.imageCheckRun = true
	if st, err := os.Stat(u.imageBaseDir); err == nil && st.IsDir() {
		return true
	}
	u.imageBaseDir = "" // mark missing
	return false
}

func pieceColorCode(c engine.Color) string {
	if c == engine.White {
		return "w"
	}
	if c == engine.Black {
		return "b"
	}
	return ""
}
func pieceTypeCode(t engine.PieceType) string {
	switch t {
	case engine.Pawn:
		return "p"
	case engine.Knight:
		return "n"
	case engine.Bishop:
		return "b"
	case engine.Rook:
		return "r"
	case engine.Queen:
		return "q"
	case engine.King:
		return "k"
	}
	return "?"
}

// detectRasterTool selects an installed external SVG raster tool if available.
func (u *uiGame) detectRasterTool() {
	if _, err := exec.LookPath("rsvg-convert"); err == nil {
		u.rasterTool = "rsvg-convert"
		return
	}
	if _, err := exec.LookPath("inkscape"); err == nil {
		u.rasterTool = "inkscape"
		return
	}
}

// rasterizeSVGWithTool rasterizes an SVG to a temporary PNG using the given tool and loads it.
func rasterizeSVGWithTool(tool, path string, width, height int) *ebiten.Image {
	outPNG := path + fmt.Sprintf(".%dx%d.tmp.png", width, height)
	var cmd *exec.Cmd
	switch tool {
	case "rsvg-convert":
		cmd = exec.Command("rsvg-convert", "-w", fmt.Sprintf("%d", width), "-h", fmt.Sprintf("%d", height), path)
		outFile, err := os.Create(outPNG)
		if err != nil {
			return nil
		}
		defer outFile.Close()
		cmd.Stdout = outFile
	case "inkscape":
		cmd = exec.Command("inkscape", path, "--export-type=png", "--export-filename", outPNG, "--export-width", fmt.Sprintf("%d", width), "--export-height", fmt.Sprintf("%d", height))
	default:
		return nil
	}
	if err := cmd.Run(); err != nil {
		return nil
	}
	img, _, err := ebitenutil.NewImageFromFile(outPNG)
	if err != nil {
		return nil
	}
	_ = os.Remove(outPNG)
	return img
}

// drawSelectableBox draws a simple selectable rectangle with label.
func drawSelectableBox(screen *ebiten.Image, x, y, w, h int, label string, selected bool) {
	bg := color.RGBA{60, 60, 60, 0xFF}
	if selected {
		bg = color.RGBA{90, 120, 200, 0xFF}
	}
	img := ebiten.NewImage(w, h)
	img.Fill(bg)
	screen.DrawImage(img, &ebiten.DrawImageOptions{GeoM: translate(x, y)})
	ebitenutil.DebugPrintAt(screen, label, x+6, y+3)
}

// --- helpers
func translate(x, y int) ebiten.GeoM {
	var g ebiten.GeoM
	g.Translate(float64(x), float64(y))
	return g
}

func stringFromInt(v int) string { return fmtInt(v) }

// local tiny int->string (avoid fmt for tight loops, though not critical here)
func fmtInt(v int) string {
	if v == 0 {
		return "0"
	}
	neg := v < 0
	if neg {
		v = -v
	}
	buf := [12]byte{}
	i := len(buf)
	for v > 0 {
		i--
		buf[i] = byte('0' + v%10)
		v /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}

func evalString(cp int) string {
	// Convert centipawns to approximate score with sign indicator
	return "Eval: " + fmtInt(cp) + " cp"
}

func main() {
	g := newUIGame()
	ebiten.SetWindowSize(windowW, windowH)
	ebiten.SetWindowTitle("go-chess GUI Demo")
	if err := ebiten.RunGame(g); err != nil && err != ebiten.Termination {
		log.Fatal(err)
	}
}
