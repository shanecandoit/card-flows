package main

import "image/color"

const (
	// --- Camera & View ---
	DefaultCameraX    = 400.0
	DefaultCameraY    = 200.0
	DefaultCameraZoom = 1.0
	ZoomLimitMin      = 0.1
	ZoomLimitMax      = 10.0
	ZoomSpeed         = 0.1
	CameraLimitMin    = -200.0

	// --- Grid & Background ---
	GridSizeSmall = 50.0
	GridSizeLarge = 100.0 // Adjusted from 100 in background.go to be a multiple
	SnapGridSmall = 10.0
	SnapGridLarge = 50.0

	// --- Cards ---
	DefaultCardWidth  = 200.0
	DefaultCardHeight = 100.0
	MinCardSize       = 50.0
	MaxCardSize       = 500.0
	HeaderHeight      = 50.0
	FooterHeight      = 30.0
	PortSize          = 10.0
	ShadowOffset      = 5.0
	BorderThickness   = 3.0
	BorderOffset      = 2.0
	CornerRadius      = 6.0
	CornerThreshold   = 15.0
	CardPaddingX      = 10.0
	CardPaddingY      = 8.0

	CardActionButtonWidth  = 30.0
	CardActionButtonHeight = 20.0
	DuplicateOffset        = 20.0
	// PortPanelWidth is now calculated as 1/3 of card width (not a constant)

	// --- Input ---
	DoubleClickThreshold = 500 // ms
	DoubleClickDistance  = 25  // px squared (5px)
	CursorBlinkRate      = 500 // ms

	// --- UI ---
	ButtonWidth   = 30.0
	ButtonHeight  = 30.0
	ButtonPadding = 10.0
	ButtonMargin  = 10.0
)

var (
	// --- Colors ---
	ColorBackground  = color.RGBA{30, 30, 35, 255}
	ColorGrid        = color.RGBA{32, 32, 10, 10}
	ColorGridBlocked = color.RGBA{20, 20, 25, 255}
	ColorOriginCross = color.RGBA{255, 100, 100, 150}
	ColorShadow      = color.RGBA{0, 0, 0, 100}
	ColorCardDefault = color.RGBA{45, 45, 50, 255}
	ColorCardHover   = color.RGBA{0, 120, 255, 255}
	ColorCardActive  = color.RGBA{50, 205, 50, 255}
	ColorCardHot     = color.RGBA{255, 140, 0, 255}
	ColorPortBody    = color.RGBA{150, 150, 150, 255}
	ColorPortBodyDim = color.RGBA{110, 110, 110, 180}
	ColorPortDot     = color.RGBA{0, 0, 0, 255}
	// Port label colors
	ColorPortLabel           = color.RGBA{220, 220, 220, 255}
	ColorPortLabelDim        = color.RGBA{150, 150, 150, 160}
	ColorPortDotDim          = color.RGBA{0, 0, 0, 120}
	ColorDivider             = color.RGBA{0, 0, 0, 50}
	ColorCornerHandle        = color.RGBA{255, 255, 255, 200}
	ColorButtonBackground    = color.RGBA{60, 60, 70, 200}
	ColorCardActionDelete    = color.RGBA{220, 50, 50, 255}
	ColorCardActionDuplicate = color.RGBA{80, 160, 240, 255}
	ColorArrowDefault        = color.RGBA{200, 200, 200, 255}
	ColorArrowDrag           = color.RGBA{255, 255, 100, 180}
	ColorPortHighlight       = color.RGBA{100, 200, 255, 255}
	ColorPortActive          = color.RGBA{255, 200, 50, 255}
	ColorPortHover           = color.RGBA{150, 255, 150, 255}
)
