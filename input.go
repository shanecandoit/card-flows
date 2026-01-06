package main

import (
	"math"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type InputSystem struct {
	game *Game

	// Panning state
	isPanning  bool
	lastMouseX int
	lastMouseY int

	// Card Dragging state
	activeCard  *Card
	dragOffsetX float64
	dragOffsetY float64
	isHot       bool

	// Resizing state
	resizingCard   *Card
	resizingCorner int

	// Double-click detection
	lastClickTime int64
	lastClickPos  [2]int

	// Editing state
	editingCard *Card
}

func NewInputSystem(g *Game) *InputSystem {
	return &InputSystem{
		game: g,
	}
}

func (is *InputSystem) Update() {
	g := is.game
	mx, my := ebiten.CursorPosition()
	wx, wy := g.screenToWorld(float64(mx), float64(my))
	overUI := is.game.ui.IsMouseOver(mx, my)

	is.handleControlKeys()
	is.handleZoom()

	if is.handleTextEditing(wx, wy) {
		return
	}

	is.handleMouseInteraction(mx, my, wx, wy, overUI)
	is.handlePanning(mx, my, overUI)
}

func (is *InputSystem) handleControlKeys() {
	g := is.game
	// --- Screenshot ---
	if inpututil.IsKeyJustPressed(ebiten.KeyF1) {
		g.screenshotRequested = true
	}

	// --- Save State ---
	if ebiten.IsKeyPressed(ebiten.KeyControl) && inpututil.IsKeyJustPressed(ebiten.KeyS) {
		err := SaveState(g, "state.yaml")
		if err != nil {
			// In a real app we'd show a UI notification
		}
	}
}

func (is *InputSystem) handleZoom() {
	g := is.game
	_, dy := ebiten.Wheel()

	// Keyboard Zooming
	if ebiten.IsKeyPressed(ebiten.KeyEqual) || ebiten.IsKeyPressed(ebiten.KeyKPAdd) {
		dy += 0.1
	}
	if ebiten.IsKeyPressed(ebiten.KeyMinus) || ebiten.IsKeyPressed(ebiten.KeyKPSubtract) {
		dy -= 0.1
	}

	if dy != 0 {
		zoomSpeed := ZoomSpeed
		newZoom := g.camera.Zoom * math.Pow(1+zoomSpeed, dy)
		if newZoom > ZoomLimitMin && newZoom < ZoomLimitMax {
			g.camera.Zoom = newZoom
		}
	}
}

func (is *InputSystem) handleTextEditing(wx, wy float64) bool {
	g := is.game
	if is.editingCard == nil {
		return false
	}

	// Capture characters
	is.editingCard.Text = string(ebiten.AppendInputChars([]rune(is.editingCard.Text)))

	// Handle Backspace
	if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) && len(is.editingCard.Text) > 0 {
		is.editingCard.Text = is.editingCard.Text[:len(is.editingCard.Text)-1]
	}

	// Handle Enter or Click Outside to Commit
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) ||
		(inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) && g.getCardAt(wx, wy) != is.editingCard) {

		is.editingCard.IsEditing = false
		is.editingCard.IsCommit = true
		is.editingCard = nil
		return true
	}

	return true // In editing mode, we block most other interactions
}

func (is *InputSystem) handleMouseInteraction(mx, my int, wx, wy float64, overUI bool) {
	g := is.game

	// --- Mouse Click / Double Click Handling ---
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) && !ebiten.IsKeyPressed(ebiten.KeySpace) && !overUI {
		now := time.Now().UnixMilli()
		isDoubleClick := false
		if now-is.lastClickTime < DoubleClickThreshold {
			dx := mx - is.lastClickPos[0]
			dy := my - is.lastClickPos[1]
			if dx*dx+dy*dy < DoubleClickDistance {
				isDoubleClick = true
			}
		}
		is.lastClickTime = now
		is.lastClickPos = [2]int{mx, my}

		if isDoubleClick {
			card := g.getCardAt(wx, wy)
			if card != nil {
				// Action buttons logic
				if is.handleActionButtons(card, wx, wy) {
					return
				}
				// Start editing
				is.editingCard = card
				card.IsEditing = true
			} else {
				// Create card
				newCard := g.AddTextCard(wx, wy)
				is.editingCard = newCard
				newCard.IsEditing = true
			}
			return
		}

		// Single click logic
		if card := g.getCardAt(wx, wy); card != nil {
			if is.handleActionButtons(card, wx, wy) {
				return
			}
		}

		// Resizing check
		for i := len(g.cards) - 1; i >= 0; i-- {
			card := g.cards[i]
			corner := card.GetCornerAt(wx, wy, g.camera.Zoom)
			if corner != -1 {
				is.resizingCard = card
				is.resizingCorner = corner
				return
			}
		}

		// Dragging check
		if card := g.getCardAt(wx, wy); card != nil {
			is.activeCard = card
			is.dragOffsetX = wx - card.X
			is.dragOffsetY = wy - card.Y
			is.isHot = true
		}
	} else if is.resizingCard != nil {
		is.handleResizing(wx, wy)
	} else if is.activeCard != nil {
		is.handleDragging(wx, wy)
	}
}

func (is *InputSystem) handleActionButtons(card *Card, wx, wy float64) bool {
	// X button (Delete)
	if wx >= card.X+(card.Width-CardActionButtonWidth-5) && wx <= card.X+card.Width-5 &&
		wy >= card.Y+5 && wy <= card.Y+5+CardActionButtonHeight {
		is.game.DeleteCard(card)
		return true
	}
	// ++ button (Duplicate)
	if wx >= card.X+(card.Width-2*CardActionButtonWidth-10) && wx <= card.X+(card.Width-CardActionButtonWidth-10) &&
		wy >= card.Y+5 && wy <= card.Y+5+CardActionButtonHeight {
		is.game.DuplicateCard(card)
		return true
	}
	return false
}

func (is *InputSystem) handleResizing(wx, wy float64) {
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		card := is.resizingCard
		minSize := MinCardSize
		maxSize := MaxCardSize

		swx := math.Round(wx/SnapGridLarge) * SnapGridLarge
		swy := math.Round(wy/SnapGridLarge) * SnapGridLarge

		switch is.resizingCorner {
		case 0: // TL
			diffX := card.X - swx
			diffY := card.Y - swy
			newW := card.Width + diffX
			newH := card.Height + diffY
			if newW >= minSize && newW <= maxSize {
				card.X = swx
				card.Width = newW
			}
			if newH >= minSize && newH <= maxSize {
				card.Y = swy
				card.Height = newH
			}
		case 1: // TR
			newW := swx - card.X
			diffY := card.Y - swy
			newH := card.Height + diffY
			if newW >= minSize && newW <= maxSize {
				card.Width = newW
			}
			if newH >= minSize && newH <= maxSize {
				card.Y = swy
				card.Height = newH
			}
		case 2: // BL
			diffX := card.X - swx
			newW := card.Width + diffX
			newH := swy - card.Y
			if newW >= minSize && newW <= maxSize {
				card.X = swx
				card.Width = newW
			}
			if newH >= minSize && newH <= maxSize {
				card.Height = newH
			}
		case 3: // BR
			newW := swx - card.X
			newH := swy - card.Y
			if newW >= minSize && newW <= maxSize {
				card.Width = newW
			}
			if newH >= minSize && newH <= maxSize {
				card.Height = newH
			}
		}
	} else {
		is.resizingCard.Width = math.Round(is.resizingCard.Width/SnapGridLarge) * SnapGridLarge
		is.resizingCard.Height = math.Round(is.resizingCard.Height/SnapGridLarge) * SnapGridLarge
		is.resizingCard = nil
	}
}

func (is *InputSystem) handleDragging(wx, wy float64) {
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		is.isHot = false
		newX := wx - is.dragOffsetX
		newY := wy - is.dragOffsetY

		is.activeCard.X = math.Round(newX/SnapGridSmall) * SnapGridSmall
		is.activeCard.Y = math.Round(newY/SnapGridSmall) * SnapGridSmall
	} else {
		is.activeCard.X = math.Round(is.activeCard.X/SnapGridLarge) * SnapGridLarge
		is.activeCard.Y = math.Round(is.activeCard.Y/SnapGridLarge) * SnapGridLarge

		if is.activeCard.X < 0 {
			is.activeCard.X = 0
		}
		if is.activeCard.Y < 0 {
			is.activeCard.Y = 0
		}

		is.activeCard = nil
		is.isHot = false
	}
}

func (is *InputSystem) handlePanning(mx, my int, overUI bool) {
	g := is.game
	isPanButtonHeld := ebiten.IsMouseButtonPressed(ebiten.MouseButtonMiddle) ||
		(ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) && is.activeCard == nil && is.resizingCard == nil && !overUI)

	if !is.isPanning {
		if isPanButtonHeld {
			shouldStartPan := false
			if ebiten.IsMouseButtonPressed(ebiten.MouseButtonMiddle) {
				shouldStartPan = true
			} else if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
				if ebiten.IsKeyPressed(ebiten.KeySpace) {
					shouldStartPan = true
				} else {
					mxWorld, myWorld := g.screenToWorld(float64(mx), float64(my))
					if g.getCardAt(mxWorld, myWorld) == nil {
						shouldStartPan = true
					}
				}
			}

			if shouldStartPan {
				is.isPanning = true
				is.lastMouseX, is.lastMouseY = mx, my
			}
		}
	} else {
		if isPanButtonHeld {
			dx := float64(mx - is.lastMouseX)
			dy := float64(my - is.lastMouseY)

			g.camera.X -= dx / g.camera.Zoom
			g.camera.Y -= dy / g.camera.Zoom

			if g.camera.X < CameraLimitMin {
				g.camera.X = CameraLimitMin
			}
			if g.camera.Y < CameraLimitMin {
				g.camera.Y = CameraLimitMin
			}

			is.lastMouseX, is.lastMouseY = mx, my
		} else {
			is.isPanning = false
		}
	}
}
