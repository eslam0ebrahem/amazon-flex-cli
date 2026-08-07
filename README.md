# Amazon Flex CLI & Dashboard

A powerful, high-performance Go-based Command Line Interface (CLI) and Terminal User Interface (TUI) for managing Amazon Flex deliveries. 

This tool directly interfaces with the Amazon Flex (Rabbit) APIs to authenticate, fetch itineraries, and securely spoof GPS coordinates for pickups and simulations. The built-in TUI Dashboard completely replaces the need to run commands one-by-one by providing an interactive, beautifully formatted routing interface.

## ✨ Features

- **Interactive Dashboard (TUI)**: A gorgeous Terminal UI built with Bubble Tea (`charmbracelet/bubbletea`), featuring custom styling, dynamic resizing, and alert modals.
- **Smart Grouping**: Packages are automatically grouped by Recipient & Phone Number, cleaning up your view and showing exactly how many packages go to each stop.
- **Arabic Text Support**: Seamlessly renders and shapes Arabic text (RTL) directly in the terminal, fixing standard terminal font corruption issues.
- **Deep Data Drilldown**: Press `Enter` on any stop to pop open a detailed overlay showing package dimensions, exact weights, routing windows, order IDs, and even raw product image URLs.
- **Integrated Actions**:
  - `[p]` Pickup: Automatically spoofs a GPS location in a 3-meter radius and sends a secure pickup request to the Amazon API.
  - `[S]` Simulate: Safely simulate a pickup action.
- **In-App Token Refresh**: Seamlessly refresh expired Bearer tokens directly from the dashboard without having to drop back into the shell.

## 🚀 Installation

Ensure you have [Go](https://go.dev/dl/) installed on your machine.

```bash
# 1. Clone the repository
git clone <your-repo-url>
cd amazonFlex/flex-cli

# 2. Download Go dependencies
go mod tidy

# 3. Build the CLI binary
go build -o flexcli
```

## 🛠️ Usage

### 1. Authentication
Before using the API, you must log in to generate an active Bearer token. The token and session information are saved securely in the local `DB/` folder.
```bash
./flexcli login <your-email> <your-password>
```

### 2. The Dashboard (TUI)
The dashboard is the main entry point for managing your active route.
```bash
./flexcli dashboard
```

**Dashboard Keybindings:**
- `Up/Down` or `j/k` : Navigate packages
- `Enter` : Open expanded package details
- `r` : Refresh the itinerary from Amazon Flex servers
- `T` : Refresh API Bearer Token
- `1` / `2` / `3` : Filter view by `Pending` / `Picked Up` / `All`
- `/` : Open search bar (Search by ID, Address, Phone, or Name)
- `p` : Send a **Pickup** request for the selected package(s)
- `S` : **Simulate** a pickup request
- `Esc` : Close details overlay or cancel search
- `q` or `Ctrl+C` : Quit application

### 3. Basic CLI Commands
If you prefer not to use the interactive dashboard, standard commands are still supported:
```bash
# Fetch and save latest itinerary to DB/
./flexcli packages

# List all packages in a simple terminal output
./flexcli list

# Record a manual pickup (requires GPS spoofing math)
./flexcli pickup <scannable_id> <lat> <lon>
```

## 📂 Project Structure

- `cmd/`: Cobra CLI commands (e.g. `dashboard.go`, `login.go`, `packages.go`)
- `pkg/api/`: Amazon Flex API HTTP interaction payloads (Itinerary, Actions, Tokens)
- `pkg/config/`: Configuration schemas, headers, and token I/O logic
- `pkg/itinerary/`: Structs and parsers for unpacking complex Amazon Rabbit JSON
- `pkg/ui/`: TUI helper functions (like `ui.FixArabic` and JSON physical stat formatters)
- `DB/`: (Git-ignored) Stores your generated `flex_token.json` and `flex_itinerary.json`

## ⚠️ Disclaimer

This software intercepts and interacts with internal Amazon APIs. It is meant for educational and research purposes only. The user assumes all responsibility for adhering to Amazon Flex's Terms of Service.
