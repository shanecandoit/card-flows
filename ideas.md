# Ideas & Design Vision

This document captures the long-term vision, design philosophy, and speculative ideas for Card Flows.

## Design Vision

To transition an infinite canvas workflow tool from a technical IDE to an accessible interface for Excel-level users, we focus on **structured data flow** and **visual logic** over syntax.

### Structural Paradigm: Logic Blocks

Instead of raw Python scripts, provide pre-configured blocks that mirror Excel functions:

- **Data Source**: Import CSV, Google Sheets.
- **Text Transform**: Uppercase, Trim, Find/Replace.
- **Logic**: If/Else, Filter, Sort.
- **AI Action**: Summarize, Categorize, Extract.
- **Output**: Export to Sheet, Email, Slack.

### The "Scratch" Approach for Python/AI

Abstract complexity using form-based inputs rather than code editors.

- **Python Wrappers**: Select "Group By" and "Sum" from dropdowns instead of writing code.
- **LLM Calls**: Structured fields for Input, Task (Summarize/Translate), and Tone (Slider).

### Progressive Disclosure

Keep the initial canvas clean. Hide advanced settings (Temperature, Top-P, specific Python libraries) inside a "Settings" gear on each card.

---

## Spatial Logic: The Y-X Alignment

Organizing the canvas by intent and time creates a predictable mental model.

- **Vertical Flow (Y-Axis)**: Chronological execution (Top = Input, Bottom = Output).
- **Horizontal Flow (X-Axis)**: Logic branching.
  - **Left Side**: The "Happy Path".
  - **Right Side**: Error handling and edge cases.
- **Swimlanes**: Visual separation of primary workflow from exception zones.

## Data-First Notebook Interface

Bridge the gap between Jupyter and Excel by keeping data visible.

- **Persistent Table**: Every selected card populates a "Data Preview" at the bottom.
- **Reactive Updates**: Visual "ripple effect" down connectors during real-time re-calculation.
- **Split-View Dashboards**: Toggle between plumbing (Canvas) and presentation (Dashboard).

## Context-Aware Assistance

- **Schema-Aware Suggestions**: Suggest "Format Date" if output is a Date column.
- **In-Line Documentation**: Sidebar shows "Prompt Recipes" tailored to detected data.
- **Auto-Join**: Automatically suggest "Join" when overlapping schemas are detected.

---

## Portability & Deployment

### The .FLOW Bundle
A zipped directory containing:

- `manifest.json`: Node coordinates and logic.
- `/scripts`: Python files for auditing.
- `/cache`: Last run data (Parquet/CSV).
- `/logs`: Execution history.

### Feature Tiering

- **Local Private**: Pure Go binary, local interpreter, data stays on disk.
- **Cloud Freemium**: Remote execution, API management, shared workspaces, automated snapshotting.

### Deployment Model

The Go backend can run locally or in a container, with Starlark scripts executed in a sandboxed environment.
