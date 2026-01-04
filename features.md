# Features

This document outlines the planned and implemented features for the Card Flows engine.

## Visual Interface: Snap-to-Grid Connectors

Excel users understand cell dependencies. Use "Snap-to-Grid" connectors where outputs and inputs are clearly labeled data types (Number, Text, Table).

- Connectors should validate data types automatically (e.g., preventing a Table output from connecting to a Text-only input).
- Use color-coding: Blue for Data, Purple for AI, Green for Logic.

## Execution and Feedback Loop

Excel users rely on immediate recalculation.

- **Ghost Data**: Show a preview of the first 3 rows of data directly under each card.
- **Caching**: Visually indicate cached states with a "Green Check" icon. If a previous card changes, downstream cards turn "Yellow" to indicate they are out of date.
- **Run Logs**: Replace terminal logs with a "History" sidebar showing a table of previous runs, timestamps, and success/fail status.

## Technical Implementation for Accessibility

- **WASM/Pyodide**: Run Python directly in the browser so users do not need to manage environments or installs.
- **Schema Mapping**: When connecting two cards, show a "Mapping" modal where users drag-and-drop column names from the source to the destination.

---

## Technical Architecture (Go/Starlark)

### 1. The 2.5D UI Engine (Frontend)

Go is excellent for high-performance cross-platform GUIs. Instead of a flat 2D canvas, use a 2.5D approach where depth (Z-axis) represents layering or nested logic.

- **Engine Choice**: Using **Ebitengine** for the "game-like" 2.5D feel (smooth zooming, sprites, shadows).
- **2.5D Implementation**: Render nodes as isometric sprites or tilted quads. Zooming out allows small metadata to pop out as 3D tooltips.
- **The Infinite Canvas**: Uses a virtual coordinate system mapping to the screen's center.

### 2. The Starlark Logic Engine (Backend)

Starlark is a Go-native dialect of Python. It is deterministic, thread-safe, and perfect for embedding logic.

- **Package**: `go.starlark.net/starlark`.
- **The "Wrapper" Pattern**: Every UI block corresponds to a Starlark function.
- **Custom Built-ins**: Inject Go functions for "Heavy Lifting" (CSV loading, LLM calls).
- **Deterministic Runs**: Prevents recursion or non-deterministic loops, ensuring the UI stays responsive.

### 3. Architecture & Data Structures

Uses a **Reactive DAG (Directed Acyclic Graph)**.

#### Node Structure
```go
type Node struct {
    ID          string
    Type        string          // "filter", "ai", "script"
    Position    [2]float64      // X, Y (Time/Logic axis)
    Script      string          // Starlark code
    Inputs      []string        // IDs of upstream nodes
    LastRunHash string          // Content-addressable cache key
}
```

#### Large Table Handling

- **Columnar Storage**: Use `[]interface{}` per column.
- **Virtual Scrolling**: Only render visible rows.
- **Lazy Loading**: Stream chunks into the UI.

### 4. Caching & Storage (The .FLOW Zip)

Use a **Content-Addressable Storage (CAS)** system.

- **Hashing**: `Hash(Node.Script + Input_Data_Hashes) = CacheKey`.
- **Zip Structure**:
  - `manifest.json`: Graph layout and node settings.
  - `/data/`: `.csv` or `.parquet` files named by `CacheKey`.
  - `/runs/`: JSON logs for execution history.

---

## Workflow Logic

### Node Classification

- **Entry Nodes**: File pickers, API webhooks.
- **Transform Nodes**: Filter, Map, Sort (Starlark).
- **AI Nodes**: Extraction, Summarization (LLM wrappers).
- **Synthesis Nodes**: Joining data streams.
- **Terminal Nodes**: Dashboard widgets, Exporters.

### Edge Directionality

- **Primary Flow**: Downward.
- **Logic Branching**: Lateral.
- **Validation**: Prevention of circular dependencies during `OnConnect`.

---

## The Four-View Architecture

1.  **Canvas View**: The default 2.5D workspace with tactile blocks.
2.  **Dashboard View**: High-level presentation; only "Visual" nodes (charts, tables) are rendered.
3.  **Table/JSON View**: Split-pane inspection at the bottom.
4.  **Graph View**: High-altitude 2D map for identifying bottlenecks.

### Implementation Details
```go
type Edge struct {
    FromID string
    ToID   string
    Schema string // Metadata about columns
}
```

- **View State Management**: A `ViewState` enum controls the `Draw()` loop without stopping the Starlark runtime.
- **Trace-back**: Starlark snippets can highlight related rows in upstream nodes from a table selection.
