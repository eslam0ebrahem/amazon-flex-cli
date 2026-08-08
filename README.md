# Amazon Flex CLI & Dashboard

A powerful, high-performance Go-based Command Line Interface (CLI) and Terminal User Interface (TUI) for managing Amazon Flex deliveries.

This tool directly interfaces with the Amazon Flex (Rabbit) APIs to authenticate, fetch itineraries, and securely spoof GPS coordinates for pickups, updates, and simulations. The built-in TUI Dashboard completely replaces the need to run commands one-by-one by providing an interactive, beautifully formatted routing interface.

## ✨ Features

- **Interactive Dashboard (TUI)**: A gorgeous Terminal UI built with Bubble Tea (`charmbracelet/bubbletea`), featuring custom styling, dynamic resizing, and alert modals.
- **Smart Grouping**: Packages are automatically grouped by Recipient & Phone Number, cleaning up your view and showing exactly how many packages go to each stop.
- **Advanced State Updates**: Support for multiple API update states such as Picked Up, Business Closed, Address Not Found, Rescheduled, and Out of Delivery Time.
- **Arabic Text Support**: Seamlessly renders and shapes Arabic text (RTL) directly in the terminal, fixing standard terminal font corruption issues.
- **Deep Data Drilldown**: Press `Enter` on any stop to pop open a detailed overlay showing package dimensions, exact weights, routing windows, order IDs, and even raw product image URLs.
- **Integrated Actions**:
  - `[p]` Pickup: Automatically spoofs a GPS location in a 3-meter radius and sends a secure pickup request to the Amazon API.
  - `[b/n/C/o]` Failures: Easily mark packages as Business Closed, No Address, Customer Rescheduled, or Out of Time.
  - `[S]` Simulate: Safely simulate a pickup action.
- **In-App Token Refresh**: Seamlessly refresh expired Bearer tokens directly from the dashboard without having to drop back into the shell.

## 🚀 Installation

Ensure you have [Go](https://go.dev/dl/) installed on your machine.

```bash
# 1. Clone the repository
git clone https://github.com/eslam0ebrahem/amazon-flex-cli.git
cd amazon-flex-cli

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
- `1` / `2` / `3` : Filter view by `Pending` / `Picked Up` / `All`
- `/` : Open search bar (Search by ID, Address, Phone, or Name)
- `p` : Send a **Pickup** request (`PICKED_UP`)
- `b` : Mark as **Business Closed** (`BUSINESS_CLOSED`)
- `n` : Mark as **Address Not Found** (`ADDRESS_NOT_FOUND`)
- `C` : Mark as **Customer Rescheduled** (`RESCHEDULED_BY_CUSTOMER`)
- `o` : Mark as **Out of Delivery Time** (`OUT_OF_DELIVERY_TIME`)
- `c` : View/Copy customer Phone Number
- `S` : **Simulate** an action
- `Esc` : Close details overlay or cancel search
- `q` or `Ctrl+C` : Quit application

### 3. Basic CLI Commands
If you prefer not to use the interactive dashboard, standard commands are still supported:

```bash
# General / Drivers
./flexcli phone                  # Get driver phone numbers
./flexcli assignments            # List upcoming schedule assignments
./flexcli tours                  # List driver tours
./flexcli instant-offers         # Interact with Instant Offers
./flexcli active                 # View active device sessions
./flexcli nonce                  # Generate an App Attestation Nonce

# Itinerary & Packages
./flexcli packages               # Fetch and save latest itinerary to DB/
./flexcli list                   # List all packages in a simple terminal output

# Updates
./flexcli pickup <scannable_id> <lat> <lon>   # Record a manual pickup 
./flexcli update <scannable_id> <state_key>   # Update package with a state key (pickup, closed, no-address, etc.)
```

## 📂 Project Structure

- `cmd/`: Cobra CLI commands (e.g. `dashboard.go`, `update_package.go`, `login.go`)
- `pkg/api/`: Amazon Flex API HTTP interaction payloads (Itinerary, Actions, Device, Tokens)
- `pkg/config/`: Configuration schemas, headers, and token I/O logic
- `pkg/itinerary/`: Structs and parsers for unpacking complex Amazon Rabbit JSON
- `pkg/ui/`: TUI helper functions (like `ui.FixArabic` and JSON physical stat formatters)
- `DB/`: (Git-ignored) Stores your generated tokens and itineraries

## ⚠️ Disclaimer

This software intercepts and interacts with internal Amazon APIs. It is meant for educational and research purposes only. The user assumes all responsibility for adhering to Amazon Flex's Terms of Service.
