package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"flexcli/pkg/api"
	"flexcli/pkg/config"
	"flexcli/pkg/itinerary"
	"flexcli/pkg/ui"
	"flexcli/pkg/utils"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/go-resty/resty/v2"
	"github.com/spf13/cobra"
)

// ─── Palette ──────────────────────────────────────────────────────────────────
var (
	cBlack   = lipgloss.Color("#0C0C0C")
	cDark    = lipgloss.Color("#141414")
	cSurface = lipgloss.Color("#1C1C1C")
	cBorder  = lipgloss.Color("#2C2C2C")
	cMuted   = lipgloss.Color("#505050")
	cDim     = lipgloss.Color("#888888")
	cWhite   = lipgloss.Color("#EFEFEF")

	cOrange = lipgloss.Color("#FF9900")
	cAmber  = lipgloss.Color("#F59E0B")
	cGreen  = lipgloss.Color("#22C55E")
	cRed    = lipgloss.Color("#EF4444")
	cBlue   = lipgloss.Color("#3B82F6")
	cCyan   = lipgloss.Color("#22D3EE")
)

// ─── Status helpers ───────────────────────────────────────────────────────────

// statusIcon returns a one-char ASCII icon for the status (no ANSI — safe for table cells).
func statusIcon(s string) string {
	switch strings.ToUpper(s) {
	case "PENDING_PICKUP":
		return "○"
	case "PICKED_UP":
		return "●"
	case "DELIVERED":
		return "✔"
	case "UNDELIVERABLE", "FAILED", "REJECTED":
		return "✖"
	default:
		return "·"
	}
}

// statusLabel returns a SHORT plain-text label for the status (no ANSI — safe for table cells).
// Keeping labels short prevents ANSI-width mismatch from clipping cells.
func statusLabel(s string) string {
	switch strings.ToUpper(s) {
	case "PENDING_PICKUP":
		return "○ Pending"
	case "PICKED_UP":
		return "● Picked Up"
	case "DELIVERED":
		return "✔ Delivered"
	case "UNDELIVERABLE":
		return "✖ Undeliverable"
	case "FAILED", "REJECTED":
		return "✖ " + s
	default:
		if s == "" {
			return "—"
		}
		return s
	}
}

// statusColor returns the lipgloss color for a status (used in overlays, not table cells).
func statusColor(s string) lipgloss.Color {
	switch strings.ToUpper(s) {
	case "PENDING_PICKUP":
		return cAmber
	case "PICKED_UP", "DELIVERED":
		return cGreen
	case "UNDELIVERABLE", "FAILED", "REJECTED":
		return cRed
	default:
		return cBlue
	}
}

// formatTimeWindow converts unix epoch to a human-readable window.
func formatTimeWindow(start, end float64) string {
	if start == 0 {
		return "—"
	}
	s := time.Unix(int64(start), 0).Format("15:04")
	e := time.Unix(int64(end), 0).Format("15:04")
	return s + " – " + e
}

// ─── View states ──────────────────────────────────────────────────────────────

type viewState int

const (
	vsMain    viewState = iota // table is active
	vsSearch                   // search input focused
	vsDetail                   // package detail panel open
	vsResult                   // action result (pickup / simulate / refresh)
)

// ─── Enriched package (deduplicated, time-windowed) ──────────────────────────

type enrichedPkg struct {
	*itinerary.PackageDetails
	GroupedPackages []*itinerary.PackageDetails
	PackageCount    int
	TimeStart       float64
	TimeEnd         float64
	HasImages       bool
	Idx             int // display index (1-based, stable)
}

// ─── Model ────────────────────────────────────────────────────────────────────

type dashModel struct {
	state  viewState
	width  int
	height int

	// Data — deduplicated, enriched
	allPkgs      []*enrichedPkg
	displayPkgs  []*enrichedPkg // after filter + search
	seenIDs      map[string]bool

	// View filter: "PENDING_PICKUP" | "PICKED_UP" | "ALL"
	activeFilter string // default: PENDING_PICKUP

	// Counts for header bar
	countPending int
	countPickedUp int
	countTotal   int

	// Table
	table table.Model

	// Search
	searchInput textinput.Model
	searchQuery string

	// Detail / result overlay
	detailPkg     *enrichedPkg
	resultTitle   string
	resultContent string
	resultIsError bool
}

// ─── Init ─────────────────────────────────────────────────────────────────────

func newDashModel() dashModel {
	ti := textinput.New()
	ti.Placeholder = "search by ID, name, or address…"
	ti.Prompt = "  / "
	ti.PromptStyle = lipgloss.NewStyle().Foreground(cOrange)
	ti.TextStyle = lipgloss.NewStyle().Foreground(cWhite)
	ti.PlaceholderStyle = lipgloss.NewStyle().Foreground(cMuted)

	m := dashModel{
		state:        vsMain,
		width:        160,
		height:       40,
		activeFilter: "PENDING_PICKUP", // default: show what still needs doing
		seenIDs:      map[string]bool{},
		searchInput:  ti,
	}
	m = m.loadData()
	return m
}

func (m dashModel) Init() tea.Cmd { return nil }

// ─── Data loading & dedup ─────────────────────────────────────────────────────

func (m dashModel) loadData() dashModel {
	data := itinerary.LoadItinerary()
	if !data.Exists() {
		return m
	}

	seen := map[string]bool{}
	var pkgs []*enrichedPkg
	idx := 0

	for _, act := range itinerary.ExtractActivities(data) {
		tw := act.Get("timeWindow")
		twStart := tw.Get("startDate").Float()
		twEnd := tw.Get("endDate").Float()

		for _, op := range act.Get("operations").Array() {
			for _, tr := range op.Get("transportRequests").Array() {
				// If timeWindow wasn't on activity, try operation's transport request
				if twStart == 0 {
					twStart = tr.Get("pickupWindow.startDate").Float()
					twEnd = tr.Get("pickupWindow.endDate").Float()
				}
				if twStart == 0 {
					twStart = tr.Get("deliveryWindow.startDate").Float()
					twEnd = tr.Get("deliveryWindow.endDate").Float()
				}
				if twStart == 0 {
					twStart = tr.Get("timeWindow.startDate").Float()
					twEnd = tr.Get("timeWindow.endDate").Float()
				}

				trScannable  := tr.Get("scannableId").String()
				trackingId   := tr.Get("clientMetaData.trackingId").String()
				extObjId     := tr.Get("clientMetaData.externalObjectId").String()
				slamTracking := tr.Get("labels.SLAM.details.trackingId.text").String()
				slamAltExec  := tr.Get("labels.SLAM.details.alternateExecutionId.text").String()
				fallbacks    := []string{trScannable, trackingId, extObjId, slamTracking, slamAltExec}

				for _, item := range tr.Get("transportItems").Array() {
					scannable := ""
					for _, s := range append([]string{item.Get("scannableId").String()}, fallbacks...) {
						if s != "" {
							scannable = s
							break
						}
					}
					if scannable == "" || seen[scannable] {
						continue
					}
					seen[scannable] = true

					det, _ := itinerary.GetDetails(scannable)
					if det == nil {
						continue
					}

					// Pull time window from details JSON if still missing
					if twStart == 0 {
						twStart = item.Get("timeWindow.startDate").Float()
						twEnd = item.Get("timeWindow.endDate").Float()
					}

					idx++
					pkgs = append(pkgs, &enrichedPkg{
						PackageDetails: det,
						TimeStart:      twStart,
						TimeEnd:        twEnd,
						HasImages:      len(det.Images) > 0,
						Idx:            idx,
					})
				}
			}
		}
	}

	// Group by Recipient + Phone
	groupsMap := make(map[string]*enrichedPkg)
	var groups []*enrichedPkg

	for _, p := range pkgs {
		key := p.RecipientName + "|" + p.RecipientPhone
		if g, exists := groupsMap[key]; exists {
			g.GroupedPackages = append(g.GroupedPackages, p.PackageDetails)
			g.PackageCount++
			if p.HasImages {
				g.HasImages = true
			}
			if g.Status != p.Status {
				g.Status = "Mixed"
			}
		} else {
			p.GroupedPackages = []*itinerary.PackageDetails{p.PackageDetails}
			p.PackageCount = 1
			groupsMap[key] = p
			groups = append(groups, p)
		}
	}

	// Sort: pending first, then picked up; within group sort by scannable ID
	sort.Slice(groups, func(i, j int) bool {
		si := groups[i].Status
		sj := groups[j].Status
		if si != sj {
			// PENDING_PICKUP before everything else
			if si == "PENDING_PICKUP" {
				return true
			}
			if sj == "PENDING_PICKUP" {
				return false
			}
		}
		return groups[i].ScannableId < groups[j].ScannableId
	})

	// Re-assign sequential indices after sort
	for i, p := range groups {
		p.Idx = i + 1
	}

	m.allPkgs = groups
	m.seenIDs = seen

	// Recount using the raw pkgs slice since m.countTotal should count raw packages, not groups
	m.countTotal = len(pkgs)
	m.countPending = 0
	m.countPickedUp = 0
	for _, p := range pkgs {
		switch p.Status {
		case "PENDING_PICKUP":
			m.countPending++
		case "PICKED_UP":
			m.countPickedUp++
		}
	}

	m = m.applyFilter()
	return m
}

func (m dashModel) applyFilter() dashModel {
	var filtered []*enrichedPkg
	for _, p := range m.allPkgs {
		switch m.activeFilter {
		case "ALL":
			filtered = append(filtered, p)
		default:
			if p.Status == m.activeFilter {
				filtered = append(filtered, p)
			}
		}
	}

	// Search
	q := strings.ToLower(m.searchQuery)
	if q != "" {
		var matched []*enrichedPkg
		for _, p := range filtered {
			if strings.Contains(strings.ToLower(p.ScannableId), q) ||
				strings.Contains(strings.ToLower(p.RecipientName), q) ||
				strings.Contains(strings.ToLower(p.AddressName), q) ||
				strings.Contains(strings.ToLower(p.City), q) ||
				strings.Contains(strings.ToLower(p.RecipientPhone), q) {
				matched = append(matched, p)
			}
		}
		filtered = matched
	}

	m.displayPkgs = filtered
	m = m.rebuildTable()
	return m
}

// ─── Table ────────────────────────────────────────────────────────────────────

func (m dashModel) rebuildTable() dashModel {
	w := m.width
	if w < 80 {
		w = 80
	}

	// Column widths — based on what a delivery driver actually needs to scan
	numW    := 4  // row number
	idW     := 15 // scannable ID (AEG... prefix, last digits matter)
	statusW := 16 // PENDING_PICKUP / PICKED_UP
	recipW  := 20 // recipient name
	imgW    := 5  // 📷 indicator
	phoneW := 15
	pkgsW := 8
	mapsW := 33

	// Distribute remaining width to address
	addrW := w - (numW + idW + statusW + recipW + phoneW + mapsW + pkgsW + imgW + 23)
	if addrW < 18 {
		addrW = 18
	}

	cols := []table.Column{
		{Title: " # ", Width: numW},
		{Title: "Scannable ID", Width: idW},
		{Title: "Status", Width: statusW},
		{Title: "Recipient", Width: recipW},
		{Title: "Phone", Width: phoneW},
		{Title: "Address", Width: addrW},
		{Title: "Maps", Width: mapsW},
		{Title: "Pkgs", Width: pkgsW},
		{Title: " 📷 ", Width: imgW},
	}

	tableH := m.height - 7
	if tableH < 5 {
		tableH = 5
	}

	t := table.New(
		table.WithColumns(cols),
		table.WithFocused(m.state == vsMain),
		table.WithHeight(tableH),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(cBorder).
		BorderBottom(true).
		Bold(true).
		Foreground(cOrange).
		Background(cDark)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("#000000")).
		Background(cOrange).
		Bold(true)
	s.Cell = s.Cell.Foreground(cWhite)
	t.SetStyles(s)

	var rows []table.Row
	for _, p := range m.displayPkgs {
		// Scannable ID: show last (idW-1) runes — the unique digits at the end of AEG...
		shortID := p.ScannableId
		rID := []rune(shortID)
		if len(rID) > idW-1 {
			shortID = string(rID[len(rID)-(idW-1):])
		}

		// Recipient: shape Arabic FIRST, then truncate by rune so multi-byte chars don't miscount.
		recip := ui.FixArabic(p.RecipientName)
		reciprunes := []rune(recip)
		if len(reciprunes) > recipW-1 {
			recip = string(reciprunes[:recipW-2]) + "…"
		}

		// Address: shape Arabic first, append city if not already present, then truncate by rune.
		addr := ui.FixArabic(p.AddressName)
		if p.City != "" && !strings.Contains(strings.ToLower(addr), strings.ToLower(p.City)) {
			addr = ui.FixArabic(p.City) + ", " + addr
		}
		addrRunes := []rune(addr)
		if len(addrRunes) > addrW-1 {
			addr = string(addrRunes[:addrW-2]) + "…"
		}

		mapsLink := "—"
		if p.Latitude != 0 && p.Longitude != 0 {
			mapsLink = fmt.Sprintf("maps.google.com/?q=%.4f,%.4f", p.Latitude, p.Longitude)
		}

		// Image indicator: plain ASCII — no ANSI, so table width math stays correct.
		imgCell := "  —"
		if p.HasImages {
			imgCell = "  [+]"
		}

		// Status: plain text label — NO lipgloss ANSI inside table cells.
		// ANSI escape codes inflate byte-length, breaking the table's cell-clipping logic.
		status := statusLabel(p.Status)

		phone := p.RecipientPhone
		if phone == "" {
			phone = "—"
		}

		pkgsCount := fmt.Sprintf("%d Pkgs", p.PackageCount)

		rows = append(rows, table.Row{
			fmt.Sprintf(" %d", p.Idx),
			shortID,
			status,
			recip,
			phone,
			addr,
			mapsLink,
			pkgsCount,
			imgCell,
		})
	}
	t.SetRows(rows)

	// Preserve cursor
	cur := m.table.Cursor()
	if cur < len(rows) {
		t.SetCursor(cur)
	}
	m.table = t
	return m
}

// ─── Update ───────────────────────────────────────────────────────────────────

func (m dashModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m = m.rebuildTable()
		return m, nil

	case tea.KeyMsg:

		// ── Search mode ───────────────────────────────────────────────────
		if m.state == vsSearch {
			switch msg.String() {
			case "esc":
				m.searchQuery = ""
				m.searchInput.SetValue("")
				m.searchInput.Blur()
				m.state = vsMain
				m = m.applyFilter()
				m = m.rebuildTable()
			case "enter":
				m.searchQuery = m.searchInput.Value()
				m.searchInput.Blur()
				m.state = vsMain
				m = m.applyFilter()
				m = m.rebuildTable()
			default:
				m.searchInput, cmd = m.searchInput.Update(msg)
				m.searchQuery = m.searchInput.Value()
				m = m.applyFilter()
				m = m.rebuildTable()
				return m, cmd
			}
			return m, nil
		}

		// ── Overlay dismiss ───────────────────────────────────────────────
		if m.state == vsDetail || m.state == vsResult {
			switch msg.String() {
			case "esc", "q", "enter", "backspace":
				m.state = vsMain
				m = m.rebuildTable()
			case "p":
				if m.state == vsDetail && m.detailPkg != nil {
					m = m.doPickup(m.detailPkg)
				}
			case "S":
				if m.state == vsDetail && m.detailPkg != nil {
					m = m.doSimulate(m.detailPkg)
				}
			case "c":
				// Copy phone to clipboard hint — just display it prominently
				if m.state == vsDetail && m.detailPkg != nil && m.detailPkg.RecipientPhone != "" {
					m.resultTitle = "📞  Phone Number"
					m.resultContent = lipgloss.NewStyle().
						Foreground(cCyan).Bold(true).
						Render(m.detailPkg.RecipientPhone)
					m.resultIsError = false
					m.state = vsResult
				}
			}
			return m, nil
		}

		// ── Main table mode ───────────────────────────────────────────────
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "enter":
			if len(m.displayPkgs) > 0 {
				m.detailPkg = m.displayPkgs[m.table.Cursor()]
				m.state = vsDetail
				m = m.rebuildTable()
			}

		case "/":
			m.state = vsSearch
			m.searchInput.SetValue(m.searchQuery)
			m.searchInput.Focus()
			return m, textinput.Blink

		case "esc":
			// Clear search
			m.searchQuery = ""
			m.searchInput.SetValue("")
			m = m.applyFilter()
			m = m.rebuildTable()

		// Filter shortcuts — the three views that matter
		case "1":
			m.activeFilter = "PENDING_PICKUP"
			m.table.SetCursor(0)
			m = m.applyFilter()
			m = m.rebuildTable()

		case "2":
			m.activeFilter = "PICKED_UP"
			m.table.SetCursor(0)
			m = m.applyFilter()
			m = m.rebuildTable()

		case "3", "a":
			m.activeFilter = "ALL"
			m.table.SetCursor(0)
			m = m.applyFilter()
			m = m.rebuildTable()

		case "r":
			m = m.doRefresh()

		case "T":
			m = m.doTokenRefresh()

		case "p":
			if len(m.displayPkgs) > 0 {
				m = m.doPickup(m.displayPkgs[m.table.Cursor()])
			}

		case "S":
			if len(m.displayPkgs) > 0 {
				m = m.doSimulate(m.displayPkgs[m.table.Cursor()])
			}
		}
	}

	if m.state == vsMain {
		m.table, cmd = m.table.Update(msg)
	}
	return m, cmd
}

// ─── Actions ──────────────────────────────────────────────────────────────────

// doTokenRefresh refreshes the bearer token in-place without leaving the TUI.
// api.Refresh() returns {"access_token":"...", ...}. config.SaveTokens() wraps it
// as raw_response.access_token which GetBearerToken() now reads via Path 2.
func (m dashModel) doTokenRefresh() dashModel {
	m.resultTitle = "⟳  Refreshing Token…"
	m.resultContent = lipgloss.NewStyle().Foreground(cDim).Render("Contacting Amazon auth servers…")
	m.resultIsError = false
	m.state = vsResult

	newData, err := api.Refresh()
	if err != nil {
		m.resultTitle = "✖  Token Refresh Failed"
		m.resultContent = lipgloss.NewStyle().Foreground(cRed).Render(err.Error()) +
			"\n\n" + lipgloss.NewStyle().Foreground(cDim).Render(
			"Refresh token may also be expired.\nExit and run:  flexcli login <email> <password>")
		m.resultIsError = true
		return m
	}

	if saveErr := config.SaveTokens(newData, "", ""); saveErr != nil {
		m.resultTitle = "✖  Token Save Failed"
		m.resultContent = lipgloss.NewStyle().Foreground(cRed).Render(saveErr.Error())
		m.resultIsError = true
		return m
	}

	// Token saved — immediately re-fetch the itinerary with the new token.
	m = m.doRefresh()
	if !m.resultIsError {
		m.resultTitle = "✔  Token Refreshed + Itinerary Updated"
	}
	return m
}

func (m dashModel) doRefresh() dashModel {
	token := config.GetBearerToken()
	if token == "" {
		m.resultTitle = "⚠  Not Logged In"
		m.resultContent = lipgloss.NewStyle().Foreground(cRed).
			Render("Run  flexcli login <email> <password>  first.")
		m.resultIsError = true
		m.state = vsResult
		return m
	}

	// NOTE: We intentionally do NOT call api.Refresh() here.
	// api.Refresh() returns a flat {access_token:"…"} response, but config.SaveTokens()
	// + config.GetBearerToken() expect the deep login response structure
	// (raw_response.response.success.tokens.bearer.access_token).
	// Saving the refresh response corrupts the token file and makes every subsequent
	// GetBearerToken() call return "". Use `flexcli refresh` + `flexcli login` from
	// the terminal to rotate tokens.
	data, err := api.FetchItinerary(0, 0)
	
	// Simulate the resty response structure for the error handling below
	// Because the user's code expects `resp.IsError()` and `resp.String()` 
	// we will handle the error directly.
	if err != nil {
		m.resultTitle = "✖  Network Error"
		m.resultContent = lipgloss.NewStyle().Foreground(cRed).Render(err.Error())
		m.resultIsError = true
		m.state = vsResult
		return m
	}
	if err != nil {
		body := err.Error()
		if len(body) > 400 {
			body = body[:400] + "…"
		}
		
		hint := ""
		if strings.Contains(body, "UnknownOperationException") || strings.Contains(body, "404") || strings.Contains(body, "401") || strings.Contains(body, "403") {
			hint = "\n\n" + lipgloss.NewStyle().Foreground(cDim).Render(
				"Token expired — press  T  to refresh it in-place,\n"+
				"or exit and run:  flexcli login <email> <pass>")
		}
		
		m.resultTitle = "✖  API Error"
		m.resultContent = lipgloss.NewStyle().Foreground(cRed).Render(body) + hint
		m.resultIsError = true
		m.state = vsResult
		return m
	}

	itinerary.SaveItinerary(data)
	config.EnsureDbDir()
	m = m.loadData()
	m.resultTitle = "✔  Itinerary Refreshed"
	m.resultContent = lipgloss.NewStyle().Foreground(cGreen).Render(
		fmt.Sprintf("Loaded %d unique packages  ·  %d pending  ·  %d picked up",
			m.countTotal, m.countPending, m.countPickedUp),
	)
	m.resultIsError = false
	m.state = vsResult
	return m
}

func (m dashModel) doPickup(p *enrichedPkg) dashModel {
	// Read token directly — do NOT go through api.RecordPickup which internally
	// calls config.GetBearerToken() again. If the token file was ever written in
	// a non-standard format, that internal call would return "" and fail silently.
	token := config.GetBearerToken()
	if token == "" {
		m.resultTitle = "✖  Pickup Failed — " + p.ScannableId
		m.resultContent = lipgloss.NewStyle().Foreground(cRed).
			Render("Not logged in — run  flexcli login <email> <password>  first.")
		m.resultIsError = true
		m.state = vsResult
		return m
	}

	lat, lon := p.Latitude, p.Longitude
	if lat == 0 {
		lat, lon = 30.13384, 31.72604
	}
	newLat, newLon := utils.GenerateGPSInCircle(lat, lon, 3.0)
	nowEpoch := float64(time.Now().UnixNano()) / 1e9
	payload := api.BuildPickupPayload(p.ScannableId, p.TrId, p.ItemId, newLat, newLon, nowEpoch)

	dimStyle := lipgloss.NewStyle().Foreground(cDim)
	gpsLine := dimStyle.Render(fmt.Sprintf("GPS spoofed → %.6f, %.6f  |  TR: %s", newLat, newLon, p.TrId))

	// Send pickup directly — same pattern as cmd/pickup.go
	pickupClient := resty.New().SetTimeout(30 * time.Second)
	resp, err := pickupClient.R().
		SetHeader("x-amz-access-token", token).
		SetHeader("User-Agent", config.RabbitUserAgent).
		SetHeader("Accept", "application/json").
		SetHeader("Content-Type", "application/json").
		SetBody(payload).
		SetResult(&map[string]interface{}{}).
		Post(config.RecordActionsEndpoint)

	if err != nil {
		m.resultTitle = "✖  Pickup Failed — " + p.ScannableId
		m.resultContent = lipgloss.NewStyle().Foreground(cRed).Render(err.Error()) +
			"\n\n" + gpsLine
		m.resultIsError = true
		m.state = vsResult
		return m
	}
	if resp.IsError() {
		body := resp.String()
		if len(body) > 400 {
			body = body[:400] + "…"
		}
		m.resultTitle = fmt.Sprintf("✖  Pickup Failed [HTTP %d] — %s", resp.StatusCode(), p.ScannableId)
		m.resultContent = lipgloss.NewStyle().Foreground(cRed).Render(body) +
			"\n\n" + gpsLine
		m.resultIsError = true
		m.state = vsResult
		return m
	}

	resultData := *resp.Result().(*map[string]interface{})
	b, _ := json.MarshalIndent(resultData, "", "  ")
	respStr := string(b)
	if len(respStr) > 500 {
		respStr = respStr[:500] + "\n…"
	}
	m.resultTitle = "✔  Pickup Sent — " + p.ScannableId
	m.resultContent = gpsLine + "\n\n" +
		lipgloss.NewStyle().Foreground(cDim).Render(respStr)
	m.resultIsError = false
	m.state = vsResult
	return m
}

func (m dashModel) doSimulate(p *enrichedPkg) dashModel {
	lat, lon := p.Latitude, p.Longitude
	if lat == 0 {
		lat, lon = 30.13384, 31.72604
	}
	newLat, newLon := utils.GenerateGPSInCircle(lat, lon, 3.0)
	nowEpoch := float64(time.Now().UnixNano()) / 1e9
	payload := api.BuildPickupPayload(p.ScannableId, p.TrId, p.ItemId, newLat, newLon, nowEpoch)

	b, _ := json.MarshalIndent(payload, "", "  ")
	preview := string(b)
	if len(preview) > 700 {
		preview = preview[:700] + "\n… (truncated)"
	}

	m.resultTitle = "◎  Simulation — " + p.ScannableId
	m.resultContent = lipgloss.NewStyle().Foreground(cAmber).
		Render(fmt.Sprintf("Spoofed GPS → %.6f, %.6f\n", newLat, newLon)) +
		lipgloss.NewStyle().Foreground(cDim).Render(preview)
	m.resultIsError = false
	m.state = vsResult
	return m
}

// ─── View ─────────────────────────────────────────────────────────────────────

func (m dashModel) View() string {
	w := m.width
	if w == 0 {
		w = 120
	}
	parts := []string{
		m.renderTopBar(w),
		m.renderProgressBar(w),
	}

	switch m.state {
	case vsDetail:
		parts = append(parts, m.renderDetailPanel(w))
	case vsResult:
		parts = append(parts, m.renderResultPanel(w))
	default:
		parts = append(parts, m.table.View())
	}

	parts = append(parts, m.renderFooter(w))
	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

// ─── Top bar ──────────────────────────────────────────────────────────────────

func (m dashModel) renderTopBar(w int) string {
	// Left: brand
	brand := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#000000")).
		Background(cOrange).
		Padding(0, 2).
		Render("FLEX CLI")

	// Middle: filter tabs
	filters := []struct {
		label  string
		key    string
		value  string
		count  int
	}{
		{"Pending", "1", "PENDING_PICKUP", m.countPending},
		{"Picked Up", "2", "PICKED_UP", m.countPickedUp},
		{"All", "3", "ALL", m.countTotal},
	}

	var tabParts []string
	for _, f := range filters {
		countStr := fmt.Sprintf(" (%d)", f.count)
		label := f.key + "  " + f.label + countStr
		if m.activeFilter == f.value {
			col := cOrange
			if f.value == "PICKED_UP" {
				col = cGreen
			} else if f.value == "ALL" {
				col = cBlue
			}
			tabParts = append(tabParts, lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#000000")).
				Background(col).
				Padding(0, 2).
				Render(label))
		} else {
			tabParts = append(tabParts, lipgloss.NewStyle().
				Foreground(cDim).
				Background(cSurface).
				Padding(0, 2).
				Render(label))
		}
	}
	tabs := strings.Join(tabParts, " ")

	// Right: search state + clock
	var rightParts []string
	if m.state == vsSearch || m.searchQuery != "" {
		sq := m.searchInput.View()
		if m.searchQuery != "" && m.state != vsSearch {
			sq = "\"" + m.searchQuery + "\""
		}
		rightParts = append(rightParts, lipgloss.NewStyle().
			Foreground(cCyan).Background(cDark).Padding(0, 1).
			Render("🔍 "+sq))
	}
	clock := lipgloss.NewStyle().
		Foreground(cOrange).Bold(true).Background(cDark).Padding(0, 2).
		Render(time.Now().Format("15:04:05"))
	rightParts = append(rightParts, clock)
	right := strings.Join(rightParts, "")

	// Assemble
	leftW := lipgloss.Width(brand) + 1 + lipgloss.Width(tabs)
	rightW := lipgloss.Width(right)
	gapW := w - leftW - rightW - 2
	if gapW < 0 {
		gapW = 0
	}
	fill := lipgloss.NewStyle().Background(cDark).Render(strings.Repeat(" ", gapW))

	return lipgloss.JoinHorizontal(lipgloss.Top,
		brand,
		lipgloss.NewStyle().Background(cDark).Render(" "),
		tabs,
		fill,
		right,
	)
}

// ─── Progress bar ─────────────────────────────────────────────────────────────

// renderProgressBar shows a visual pickup progress and contextual info.
func (m dashModel) renderProgressBar(w int) string {
	if m.countTotal == 0 {
		return lipgloss.NewStyle().Background(cDark).Width(w).
			Foreground(cMuted).Padding(0, 2).
			Render("No packages loaded — press r to fetch from API")
	}

	// Progress fraction
	done := m.countPickedUp
	total := m.countTotal
	pct := 0
	if total > 0 {
		pct = (done * 100) / total
	}

	// Draw a simple block-based bar
	barW := 20
	filled := (done * barW) / total
	if filled > barW {
		filled = barW
	}
	bar := lipgloss.NewStyle().Foreground(cGreen).Render(strings.Repeat("█", filled)) +
		lipgloss.NewStyle().Foreground(cMuted).Render(strings.Repeat("░", barW-filled))

	progress := fmt.Sprintf("%d/%d picked up (%d%%)", done, total, pct)
	leftStr := lipgloss.NewStyle().Foreground(cDim).Render("Progress ") +
		bar + "  " +
		lipgloss.NewStyle().Foreground(cWhite).Bold(true).Render(progress)

	// Right: showing X of Y
	showing := fmt.Sprintf("Showing %d of %d", len(m.displayPkgs), m.countTotal)
	rightStr := lipgloss.NewStyle().Foreground(cMuted).Render(showing)

	gapW := w - lipgloss.Width(leftStr) - lipgloss.Width(rightStr) - 4
	if gapW < 0 {
		gapW = 0
	}
	fill := strings.Repeat(" ", gapW)

	line := "  " + leftStr + fill + rightStr
	return lipgloss.NewStyle().Background(cDark).Width(w).Render(line) + "\n" +
		lipgloss.NewStyle().Foreground(cBorder).Render(strings.Repeat("─", w))
}

// ─── Detail Panel ─────────────────────────────────────────────────────────────

func (m dashModel) renderDetailPanel(w int) string {
	p := m.detailPkg
	if p == nil {
		return ""
	}

	panelW := w - 4
	if panelW > 140 {
		panelW = 140
	}
	if panelW < 60 {
		panelW = 60
	}
	innerW := panelW - 4

	kStyle := lipgloss.NewStyle().Foreground(cOrange).Bold(true).Width(16)
	vStyle := lipgloss.NewStyle().Foreground(cWhite)
	dimV   := lipgloss.NewStyle().Foreground(cDim)
	secHdr := func(label string) string {
		return lipgloss.NewStyle().
			Foreground(cDim).Bold(true).
			Render("  " + strings.ToUpper(label)) + "\n"
	}
	row := func(k, v string) string {
		return kStyle.Render(k) + "  " + vStyle.Render(v) + "\n"
	}

	// ── Title ─────────────────────────────────────────────────────────────────
	statusCol := statusColor(p.Status)
	titleLine := lipgloss.NewStyle().Bold(true).Foreground(cOrange).
		Render("Package  "+p.ScannableId) +
		"   " +
		lipgloss.NewStyle().Foreground(statusCol).Bold(true).
			Render(statusIcon(p.Status)+" "+p.Status)

	sep := lipgloss.NewStyle().Foreground(cBorder).Render(strings.Repeat("─", innerW))

	body := titleLine + "\n" + sep + "\n\n"

	// ── Recipient & Contact ───────────────────────────────────────────────────
	body += secHdr("Customer")
	body += row("Name", ui.FixArabic(p.RecipientName))
	if p.RecipientPhone != "" {
		phoneStr := vStyle.Render(p.RecipientPhone) + "  " +
			dimV.Render("(press c to copy)")
		body += kStyle.Render("Phone") + "  " + phoneStr + "\n"
	}

	// ── Address ───────────────────────────────────────────────────────────────
	body += "\n" + secHdr("Address")
	body += row("Street", ui.FixArabic(p.AddressName))
	body += row("City", ui.FixArabic(p.City)+" "+p.Postal)
	if p.Latitude != 0 {
		mapsURL := fmt.Sprintf("maps.google.com/?q=%.5f,%.5f", p.Latitude, p.Longitude)
		body += kStyle.Render("Maps") + "  " +
			lipgloss.NewStyle().Foreground(cCyan).Render(mapsURL) + "\n"
	}

	for idx, gp := range p.GroupedPackages {
		if len(p.GroupedPackages) > 1 {
			body += "\n" + lipgloss.NewStyle().Background(cOrange).Foreground(lipgloss.Color("#000000")).Bold(true).Render(fmt.Sprintf(" PACKAGE %d: %s ", idx+1, gp.ScannableId)) + "\n"
		}

		// ── Order ─────────────────────────────────────────────────────────────────
		body += "\n" + secHdr("Order")
		body += row("Order ID", gp.ClientOrderId)
		if gp.DeliveryInstruct != "" {
			body += row("Instructions", gp.DeliveryInstruct)
		}
		if gp.DeliveryType != "" {
			body += row("Type", gp.DeliveryType)
		}
		// Notice: TimeStart/TimeEnd aren't in PackageDetails directly anymore since they were on enrichedPkg
		// We'll just show it if we have it on the group for now.
		if idx == 0 && p.TimeStart != 0 {
			body += row("Time Window", formatTimeWindow(p.TimeStart, p.TimeEnd))
		}

		// ── Physical ──────────────────────────────────────────────────────────────
		if gp.Weight != "" || gp.Dimensions != "" {
			body += "\n" + secHdr("Physical")
			if gp.Weight != "" {
				body += row("Weight", ui.FormatWeight(gp.Weight))
			}
			if gp.Dimensions != "" {
				body += row("Dimensions", ui.FormatDimensions(gp.Dimensions))
			}
		}

		// ── Images ────────────────────────────────────────────────────────────────
		if len(gp.Images) > 0 {
			body += "\n" + secHdr(fmt.Sprintf("Product Images (%d)", len(gp.Images)))
			for i, img := range gp.Images {
				if i >= 3 {
					body += dimV.Render(fmt.Sprintf("  … and %d more\n", len(gp.Images)-3))
					break
				}
				body += "  " + dimV.Render(fmt.Sprintf("[%d] ", i+1)) +
					lipgloss.NewStyle().Foreground(cCyan).Render(img) + "\n"
			}
		} else {
			body += "\n" + dimV.Render("  No product images\n")
		}
	}

	// ── Action hints ──────────────────────────────────────────────────────────
	body += "\n" + sep + "\n"
	body += m.hintBar([]string{"p", "Pickup", "S", "Simulate", "c", "Show Phone", "Esc", "Close"})

	panel := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(cOrange).
		Background(cDark).
		Padding(1, 2).
		Width(panelW).
		Render(body)

	// Center horizontally
	pad := (w - lipgloss.Width(panel)) / 2
	if pad < 0 {
		pad = 0
	}
	leftPad := strings.Repeat(" ", pad)
	var lines []string
	for _, line := range strings.Split(panel, "\n") {
		lines = append(lines, leftPad+line)
	}
	return strings.Join(lines, "\n")
}

// ─── Result Panel ─────────────────────────────────────────────────────────────

func (m dashModel) renderResultPanel(w int) string {
	panelW := w - 20
	if panelW > 100 {
		panelW = 100
	}
	if panelW < 50 {
		panelW = 50
	}
	innerW := panelW - 4

	borderCol := cCyan
	if m.resultIsError {
		borderCol = cRed
	}

	sep := lipgloss.NewStyle().Foreground(cBorder).Render(strings.Repeat("─", innerW))
	title := lipgloss.NewStyle().Bold(true).Foreground(cOrange).Render(m.resultTitle)
	body := title + "\n" + sep + "\n\n" + m.resultContent + "\n\n" + sep + "\n" +
		m.hintBar([]string{"Esc", "Close"})

	panel := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderCol).
		Background(cDark).
		Padding(1, 2).
		Width(panelW).
		Render(body)

	pad := (w - lipgloss.Width(panel)) / 2
	if pad < 0 {
		pad = 0
	}
	leftPad := strings.Repeat(" ", pad)
	var lines []string
	for _, line := range strings.Split(panel, "\n") {
		lines = append(lines, leftPad+line)
	}
	return strings.Join(lines, "\n")
}

// ─── Footer ───────────────────────────────────────────────────────────────────

func (m dashModel) renderFooter(w int) string {
	sep := lipgloss.NewStyle().Foreground(cBorder).Render(strings.Repeat("─", w))

	var hints string
	switch m.state {
	case vsSearch:
		hints = m.hintBar([]string{"Enter", "Apply", "Esc", "Cancel"})
	case vsDetail, vsResult:
		hints = m.hintBar([]string{"Esc", "Back"})
	default:
		hints = m.hintBar([]string{
			"Enter", "Details",
			"/", "Search",
			"1", "Pending",
			"2", "Picked Up",
			"3", "All",
			"p", "Pickup",
			"S", "Simulate",
			"r", "Refresh",
			"T", "Re-auth Token",
			"q", "Quit",
		})
	}

	bar := lipgloss.NewStyle().Background(cDark).Width(w).Padding(0, 1).Render(hints)
	return lipgloss.JoinVertical(lipgloss.Left, sep, bar)
}

// ─── Hint bar helper ──────────────────────────────────────────────────────────

func (m dashModel) hintBar(pairs []string) string {
	var parts []string
	for i := 0; i+1 < len(pairs); i += 2 {
		k := lipgloss.NewStyle().
			Background(cSurface).Foreground(cOrange).Bold(true).
			Padding(0, 1).Render(pairs[i])
		d := lipgloss.NewStyle().Foreground(cDim).Render(" " + pairs[i+1])
		parts = append(parts, k+d)
	}
	return strings.Join(parts, "   ")
}

// ─── Command ──────────────────────────────────────────────────────────────────

var dashboardCmd = &cobra.Command{
	Use:   "dashboard",
	Short: "Launch the interactive TUI dashboard",
	Run: func(cmd *cobra.Command, args []string) {
		p := tea.NewProgram(newDashModel(), tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			fmt.Printf("Error starting dashboard: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(dashboardCmd)
}
