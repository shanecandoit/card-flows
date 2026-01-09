package input

import (
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// PortInfo is a minimal view of a card port used by the input system.
type PortInfo struct {
	Name    string
	Type    string
	IsInput bool
}

// Host defines the callbacks the input system needs from the main game.
type Host interface {
	ScreenToWorld(sx, sy float64) (float64, float64)
	IsMouseOver(mx, my int) bool
	RequestScreenshot()
	RunEngine()
	SaveState(filename string) error
	GetCardAt(wx, wy float64) interface{}
	AddTextCardHandle(wx, wy float64) interface{}
	DeleteCardHandle(card interface{})
	DuplicateCardHandle(card interface{})
	IsInputPortConnected(cardID, portName string) bool
	GetCardID(card interface{}) string
	GetCardTitle(card interface{}) string
	GetCardBounds(card interface{}) (x, y, w, h float64)
	SetCardBounds(card interface{}, x, y, w, h float64)
	GetCornerAt(card interface{}, wx, wy, zoom float64) int
	GetPortAt(card interface{}, wx, wy, zoom float64) *PortInfo
	GetOutputPortPosition(card interface{}, portName string) (float64, float64)
	CheckActionButton(card interface{}, wx, wy float64) string // returns "delete", "duplicate", or ""
	ApplyPan(dx, dy float64)
	RegisterSubscription(fromID, toID, toPort string)
	UnregisterSubscription(fromID, toID, toPort string)
	PropagateTextByID(cardID string)
}

type InputSystem struct {
	host Host

	// Exposed state for other packages (main) to read
	ActiveCard     interface{}
	IsHot          bool
	ResizingCard   interface{}
	ResizingCorner int
	EditingCard    interface{}
	DragOffsetX    float64
	DragOffsetY    float64

	DraggingArrow bool
	DragStartCard interface{}
	DragStartPort string

	HoveredPortCard interface{}
	HoveredPortInfo *PortInfo

	// Internal state
	isPanning  bool
	lastMouseX int
	lastMouseY int

	lastClickTime int64
	lastClickPos  [2]int
}

func NewInputSystem(h Host) *InputSystem {
	return &InputSystem{host: h}
}

func (is *InputSystem) Update() {
	mx, my := ebiten.CursorPosition()
	wx, wy := is.host.ScreenToWorld(float64(mx), float64(my))
	overUI := is.host.IsMouseOver(mx, my)

	is.handleControlKeys()
	is.handleZoom()

	if is.handleTextEditing(wx, wy) {
		return
	}

	is.handleWiring(mx, my, wx, wy)
	if is.DraggingArrow {
		return
	}

	is.handleMouseInteraction(mx, my, wx, wy, overUI)
	is.handlePanning(mx, my, overUI)
}

func (is *InputSystem) handleControlKeys() {
	// --- Screenshot ---
	if inpututil.IsKeyJustPressed(ebiten.KeyF12) {
		is.host.RequestScreenshot()
	}

	// --- Run Engine ---
	if inpututil.IsKeyJustPressed(ebiten.KeyF5) {
		is.host.RunEngine()
	}

	// --- Save State ---
	if ebiten.IsKeyPressed(ebiten.KeyControl) && inpututil.IsKeyJustPressed(ebiten.KeyS) {
		_ = is.host.SaveState("state.yaml")
	}
}

func (is *InputSystem) handleZoom() {
	_, dy := ebiten.Wheel()

	if ebiten.IsKeyPressed(ebiten.KeyEqual) || ebiten.IsKeyPressed(ebiten.KeyKPAdd) {
		dy += 0.1
	}
	if ebiten.IsKeyPressed(ebiten.KeyMinus) || ebiten.IsKeyPressed(ebiten.KeyKPSubtract) {
		dy -= 0.1
	}

	if dy != 0 {
		// delegate to host via changing camera externally; input system only signals
		// No-op here; host Zoom handled elsewhere via UI callbacks
		_ = dy
	}
}

func (is *InputSystem) handleTextEditing(wx, wy float64) bool {
	if is.EditingCard == nil {
		return false
	}

	// Append input chars via host: host will update the card text
	// For simplicity, no per-char callback here; host is responsible for read/write

	// Handle Backspace and Enter via inpututil as before; commit via host
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) ||
		(inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) && is.host.GetCardAt(wx, wy) != is.EditingCard) {

		// Commit editing
		if id := is.host.GetCardID(is.EditingCard); id != "" {
			is.host.PropagateTextByID(id)
			is.host.RunEngine()
		}
		is.EditingCard = nil
		return true
	}

	return true
}

func (is *InputSystem) handleMouseInteraction(mx, my int, wx, wy float64, overUI bool) {
	// Double-click and click handling simplified: delegate card creation and deletion to host
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) && !ebiten.IsKeyPressed(ebiten.KeySpace) && !overUI {
		now := time.Now().UnixMilli()
		isDoubleClick := false
		if now-is.lastClickTime < 350 {
			dx := mx - is.lastClickPos[0]
			dy := my - is.lastClickPos[1]
			if dx*dx+dy*dy < 1000 { // ~31px
				isDoubleClick = true
			}
		}
		is.lastClickTime = now
		is.lastClickPos = [2]int{mx, my}

		if isDoubleClick {
			card := is.host.GetCardAt(wx, wy)
			if card != nil {
				// Check action buttons first (delete/duplicate)
				action := is.host.CheckActionButton(card, wx, wy)
				if action == "delete" {
					is.host.DeleteCardHandle(card)
					return
				} else if action == "duplicate" {
					is.host.DuplicateCardHandle(card)
					return
				}

				// Start editing for text cards
				is.EditingCard = card
			} else {
				newCard := is.host.AddTextCardHandle(wx, wy)
				is.EditingCard = newCard
			}
			return
		}

		// Single click logic - check action buttons first
		if card := is.host.GetCardAt(wx, wy); card != nil {
			action := is.host.CheckActionButton(card, wx, wy)
			if action == "delete" {
				is.host.DeleteCardHandle(card)
				return
			} else if action == "duplicate" {
				is.host.DuplicateCardHandle(card)
				return
			}

			// Check if clicking on a resize corner
			corner := is.host.GetCornerAt(card, wx, wy, 1.0)
			if corner != -1 {
				is.ResizingCard = card
				is.ResizingCorner = corner
				return
			}

			// set active for dragging
			is.ActiveCard = card
			is.IsHot = true
			// compute drag offsets using host bounds
			if x, y, _, _ := is.host.GetCardBounds(card); true {
				is.DragOffsetX = wx - x
				is.DragOffsetY = wy - y
			}
		} else {
			// Clicked on empty space - clear active card so panning can start
			is.ActiveCard = nil
			is.IsHot = false
		}
	} else if is.ResizingCard != nil {
		is.handleResizing(wx, wy)
	} else if is.ActiveCard != nil {
		is.handleDragging(wx, wy)
	}
}

func (is *InputSystem) handleResizing(wx, wy float64) {
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		card := is.ResizingCard
		if card == nil {
			return
		}
		ax, ay, aw, ah := is.host.GetCardBounds(card)

		switch is.ResizingCorner {
		case 0: // TL
			newW := (ax + aw) - wx
			newH := (ay + ah) - wy
			if newW > 10 {
				ax = wx
				aw = newW
			}
			if newH > 10 {
				ay = wy
				ah = newH
			}
		case 1: // TR
			newW := wx - ax
			newH := (ay + ah) - wy
			if newW > 10 {
				aw = newW
			}
			if newH > 10 {
				ay = wy
				ah = newH
			}
		case 2: // BL
			newW := (ax + aw) - wx
			newH := wy - ay
			if newW > 10 {
				ax = wx
				aw = newW
			}
			if newH > 10 {
				ah = newH
			}
		case 3: // BR
			newW := wx - ax
			newH := wy - ay
			if newW > 10 {
				aw = newW
			}
			if newH > 10 {
				ah = newH
			}
		}

		is.host.SetCardBounds(card, ax, ay, aw, ah)
	} else {
		// commit
		is.ResizingCard = nil
	}
}

func (is *InputSystem) handleDragging(wx, wy float64) {
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		card := is.ActiveCard
		if card == nil {
			return
		}
		_, _, w, h := is.host.GetCardBounds(card)
		newX := wx - is.DragOffsetX
		newY := wy - is.DragOffsetY
		is.host.SetCardBounds(card, newX, newY, w, h)
	} else {
		// release
		is.ActiveCard = nil
		is.IsHot = false
	}
}

func (is *InputSystem) handlePanning(mx, my int, overUI bool) {
	isPanButtonHeld := ebiten.IsMouseButtonPressed(ebiten.MouseButtonMiddle) ||
		(ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) && is.ActiveCard == nil && is.ResizingCard == nil && !overUI)

	if !is.isPanning {
		if isPanButtonHeld {
			is.isPanning = true
			is.lastMouseX, is.lastMouseY = mx, my
		}
	} else {
		if isPanButtonHeld {
			dx := float64(mx - is.lastMouseX)
			dy := float64(my - is.lastMouseY)
			is.host.ApplyPan(dx, dy)
			is.lastMouseX, is.lastMouseY = mx, my
		} else {
			is.isPanning = false
		}
	}
}

func (is *InputSystem) handleWiring(mx, my int, wx, wy float64) {
	// Simplified wiring handling: start drag when clicking an output; on release, ask host to connect
	if is.DraggingArrow {
		is.HoveredPortCard = nil
		is.HoveredPortInfo = nil
		// Find hovered input port via host
		// Host is expected to return port info
	}

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		if card := is.host.GetCardAt(wx, wy); card != nil {
			portInfo := is.host.GetPortAt(card, wx, wy, 1.0)
			if portInfo != nil && !portInfo.IsInput {
				is.DraggingArrow = true
				is.DragStartCard = card
				is.DragStartPort = portInfo.Name
				return
			}
		}
	}

	if is.DraggingArrow && inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		defer func() {
			is.DraggingArrow = false
			is.DragStartCard = nil
			is.DragStartPort = ""
			is.HoveredPortCard = nil
			is.HoveredPortInfo = nil
		}()

		// connection resolution is delegated to host; simplified here
	}
}

// GetCardList is a convenience delegator; host may implement a richer API.
func (is *InputSystem) GetCardList() []interface{} {
	// Try to call Host.GetCards via type assertion if available
	type getter interface{ GetCards() []interface{} }
	if g, ok := is.host.(getter); ok {
		return g.GetCards()
	}
	return nil
}
