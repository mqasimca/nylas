package ui

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"html/template"
	"io"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"slices"
	"strings"
	"time"

	"github.com/mqasimca/nylas/internal/adapters/config"
	"github.com/mqasimca/nylas/internal/adapters/keyring"
	nylasadapter "github.com/mqasimca/nylas/internal/adapters/nylas"
	authapp "github.com/mqasimca/nylas/internal/app/auth"
	"github.com/mqasimca/nylas/internal/domain"
	"github.com/mqasimca/nylas/internal/ports"
)

//go:embed static/* static/css/* static/js/*
var staticFiles embed.FS

// Server represents the web UI server.
type Server struct {
	addr        string
	demoMode    bool
	configSvc   *authapp.ConfigService
	configStore ports.ConfigStore
	secretStore ports.SecretStore
	grantStore  ports.GrantStore
	templates   *template.Template
}

// NewServer creates a new UI server.
func NewServer(addr string) *Server {
	configStore := config.NewDefaultFileStore()
	secretStore, _ := keyring.NewSecretStore(config.DefaultConfigDir())
	grantStore := keyring.NewGrantStore(secretStore)
	configSvc := authapp.NewConfigService(configStore, secretStore)

	// Load templates
	tmpl, err := loadTemplates()
	if err != nil {
		// Fall back to nil; will serve static files only
		tmpl = nil
	}

	return &Server{
		addr:        addr,
		demoMode:    false,
		configSvc:   configSvc,
		configStore: configStore,
		secretStore: secretStore,
		grantStore:  grantStore,
		templates:   tmpl,
	}
}

// NewDemoServer creates a UI server in demo mode with sample data.
func NewDemoServer(addr string) *Server {
	// Load templates
	tmpl, err := loadTemplates()
	if err != nil {
		tmpl = nil
	}

	return &Server{
		addr:      addr,
		demoMode:  true,
		templates: tmpl,
		// Other fields are nil - demo mode doesn't use real stores
	}
}

// demoGrants returns sample grants for demo mode.
func demoGrants() []Grant {
	return []Grant{
		{ID: "demo-grant-001", Email: "alice@example.com", Provider: "google"},
		{ID: "demo-grant-002", Email: "bob@work.com", Provider: "microsoft"},
		{ID: "demo-grant-003", Email: "carol@company.org", Provider: "google"},
	}
}

// demoDefaultGrant returns the default grant ID for demo mode.
func demoDefaultGrant() string {
	return "demo-grant-001"
}

// getDemoCommandOutput returns sample output for demo mode commands.
func getDemoCommandOutput(command string) string {
	cmd := strings.TrimSpace(command)
	args := strings.Fields(cmd)
	if len(args) == 0 {
		return "Demo mode - no command specified"
	}

	baseCmd := args[0]
	if len(args) >= 2 {
		baseCmd = args[0] + " " + args[1]
	}

	switch baseCmd {
	case "email list":
		return `Demo Mode - Sample Emails

  ★ ●  alice@example.com       Weekly Team Sync - Agenda        2 min ago
    ●  bob@work.com            Project Update: Q4 Goals         15 min ago
  ★    calendar@google.com     Reminder: Design Review          1 hour ago
    ●  notifications@github    [nylas/cli] New PR opened        2 hours ago
       support@nylas.com       Welcome to Nylas!                1 day ago

Showing 5 of 127 messages`

	case "email threads":
		return `Demo Mode - Sample Threads

  ★ ●  Team Weekly Standup     5 messages    alice, bob, carol    2 min ago
    ●  Project Planning Q1     12 messages   team@company.org     1 hour ago
  ★    Design Review           3 messages    design@example.com   3 hours ago
       Onboarding Docs         2 messages    hr@company.org       1 day ago

Showing 4 threads`

	case "calendar list":
		return `Demo Mode - Sample Calendars

  ID                     NAME                 PRIMARY
  cal-primary-001        Work Calendar        ✓
  cal-personal-002       Personal
  cal-team-003           Team Events

3 calendars found`

	case "calendar events":
		return `Demo Mode - Sample Events

  TODAY
  09:00 - 10:00   Team Standup                  Conference Room A
  14:00 - 15:00   Design Review                 Zoom Meeting

  TOMORROW
  10:00 - 11:00   1:1 with Manager              Office
  15:00 - 16:00   Sprint Planning               Conference Room B

4 upcoming events`

	case "auth status":
		return `Demo Mode - Authentication Status

  Status:     Configured ✓
  Region:     US
  Client ID:  demo-client-id
  API Key:    ********configured

  Default Account: alice@example.com (Google)`

	case "auth list":
		return `Demo Mode - Connected Accounts

  ✓  alice@example.com    Google      demo-grant-001 (default)
     bob@work.com         Microsoft   demo-grant-002
     carol@company.org    Google      demo-grant-003

3 accounts connected`

	case "contacts list", "contacts list --id":
		return `Demo Mode - Sample Contacts

Found 5 contact(s):

ID                                     NAME                   EMAIL                      PHONE
demo-contact-001-alice-johnson-12345   Alice Johnson          alice@example.com          +1-555-0101
demo-contact-002-bob-smith-67890       Bob Smith              bob@work.com               +1-555-0102
demo-contact-003-carol-williams-11111  Carol Williams         carol@company.org          +1-555-0103
demo-contact-004-david-brown-22222     David Brown            david@startup.io           +1-555-0104
demo-contact-005-eve-davis-33333       Eve Davis              eve@consulting.com         +1-555-0105

Showing 5 of 127 contacts`

	case "contacts groups":
		return `Demo Mode - Contact Groups

  ID                     NAME                 MEMBERS
  grp-001                Work                 23
  grp-002                Personal             15
  grp-003                VIP Clients          8
  grp-004                Newsletter           156

4 groups found`

	case "inbound list":
		return `Demo Mode - Inbound Inboxes

  ID                     ADDRESS                           STATUS
  inbox-001              support@yourapp.nylas.email       Active
  inbox-002              leads@yourapp.nylas.email         Active
  inbox-003              tickets@yourapp.nylas.email       Active

3 inbound inboxes`

	case "inbound messages":
		return `Demo Mode - Inbound Messages

  ★ ●  customer@email.com    Need help with billing       5 min ago
    ●  lead@company.com      Interested in your product   1 hour ago
       partner@business.com  Partnership inquiry          3 hours ago

Showing 3 of 42 messages`

	case "scheduler configurations":
		return `Demo Mode - Scheduler Configurations

  ID                     NAME                 DURATION    AVAILABILITY
  cfg-001                30-min Meeting       30 min      Mon-Fri 9-5
  cfg-002                1-hour Consultation  60 min      Mon-Wed 10-4
  cfg-003                Quick Call           15 min      Daily 8-8

3 configurations`

	case "scheduler bookings":
		return `Demo Mode - Bookings

  UPCOMING
  Dec 27  10:00 AM   30-min Meeting       john@client.com
  Dec 28  02:00 PM   1-hour Consultation  jane@partner.org
  Dec 30  09:00 AM   Quick Call           mike@prospect.io

3 upcoming bookings`

	case "scheduler sessions":
		return `Demo Mode - Scheduling Sessions

  ID                     CONFIGURATION        STATUS      EXPIRES
  sess-001               30-min Meeting       Active      Dec 31, 2024
  sess-002               1-hour Consultation  Active      Jan 15, 2025

2 active sessions`

	case "scheduler pages":
		return `Demo Mode - Scheduling Pages

  ID                     SLUG                 CONFIGURATION        STATUS
  page-001               meet-with-alice      30-min Meeting       Published
  page-002               consultation         1-hour Consultation  Published
  page-003               quick-chat           Quick Call           Draft

3 scheduling pages`

	case "timezone list":
		return `Demo Mode - Time Zones

  REGION          ZONE                    OFFSET    CURRENT TIME
  America         America/New_York        -05:00    10:30 AM EST
  America         America/Los_Angeles     -08:00    7:30 AM PST
  Europe          Europe/London           +00:00    3:30 PM GMT
  Europe          Europe/Paris            +01:00    4:30 PM CET
  Asia            Asia/Tokyo              +09:00    12:30 AM JST

Showing 5 of 594 time zones`

	case "timezone info":
		return `Demo Mode - Time Zone Info

  Zone:      America/New_York
  Offset:    -05:00 (EST)
  DST:       Observed
  Current:   Thu Dec 25, 2024 10:30:00 AM EST

  Next DST Transition:
  Mar 9, 2025 02:00 AM → 03:00 AM (EDT, -04:00)`

	case "timezone convert":
		return `Demo Mode - Time Conversion

  FROM:   Dec 25, 2024 10:30 AM America/New_York (EST)
  TO:     Dec 26, 2024 12:30 AM Asia/Tokyo (JST)

  Time difference: +14:00`

	case "timezone find-meeting":
		return `Demo Mode - Meeting Time Finder

  Zones: America/New_York, Europe/London, Asia/Tokyo

  Best meeting times (next 7 days):
  ┌─────────────────────────────────────────────────────────┐
  │ Dec 26  9:00 AM EST │ 2:00 PM GMT │ 11:00 PM JST        │
  │ Dec 27  9:00 AM EST │ 2:00 PM GMT │ 11:00 PM JST        │
  │ Dec 30  9:00 AM EST │ 2:00 PM GMT │ 11:00 PM JST        │
  └─────────────────────────────────────────────────────────┘

3 available time slots found`

	case "timezone dst":
		return `Demo Mode - DST Transitions

  Zone: America/New_York
  Year: 2025

  Spring Forward:  Mar 9, 2025 02:00 AM → 03:00 AM (EST → EDT)
  Fall Back:       Nov 2, 2025 02:00 AM → 01:00 AM (EDT → EST)

  Current offset: -05:00 (EST)
  Summer offset:  -04:00 (EDT)`

	case "webhook list":
		return `Demo Mode - Webhooks

  ID                     CALLBACK URL                           TRIGGERS        STATUS
  wh-001                 https://example.com/webhook/events     message.*       Active
  wh-002                 https://api.company.io/nylas            calendar.*      Active
  wh-003                 https://hooks.app.com/contacts          contact.*       Paused

3 webhooks configured`

	case "webhook triggers":
		return `Demo Mode - Available Webhook Triggers

  CATEGORY       TRIGGER                    DESCRIPTION
  message        message.created            New message received
  message        message.updated            Message modified
  message        message.opened             Message opened (tracking)
  calendar       calendar.created           New calendar added
  calendar       event.created              New event created
  calendar       event.updated              Event modified
  calendar       event.deleted              Event deleted
  contact        contact.created            New contact added
  contact        contact.updated            Contact modified
  contact        contact.deleted            Contact deleted
  grant          grant.created              New grant connected
  grant          grant.expired              Grant expired
  grant          grant.deleted              Grant removed

14 trigger types available`

	case "webhook test":
		return `Demo Mode - Webhook Test

  Webhook:    wh-001
  URL:        https://example.com/webhook/events

  Test payload sent successfully!

  Response:
  Status:     200 OK
  Latency:    142ms
  Body:       {"received": true}`

	case "webhook server":
		return `Demo Mode - Webhook Server

  Starting local webhook receiver...

  Local URL:      http://localhost:9000/webhook
  Tunnel URL:     https://abc123.ngrok.io/webhook

  Ready to receive webhooks!

  Press Ctrl+C to stop the server.`

	case "otp get":
		return `Demo Mode - OTP Code

  Account:    alice@example.com
  Service:    GitHub
  Code:       847293
  Expires:    28 seconds

  Code copied to clipboard!`

	case "otp watch":
		return `Demo Mode - Watching for OTP codes...

  Monitoring: alice@example.com
  Filter:     All services

  Waiting for new OTP codes...
  (Demo mode - would show real-time codes)

  Press Ctrl+C to stop.`

	case "otp list":
		return `Demo Mode - Configured OTP Accounts

  EMAIL                      DEFAULT    LAST OTP
  alice@example.com          ✓          2 min ago
  bob@work.com                          1 hour ago

2 accounts configured`

	case "otp messages":
		return `Demo Mode - Recent OTP Messages

  TIME          FROM                    SERVICE         CODE
  2 min ago     noreply@github.com      GitHub          847293
  15 min ago    verify@google.com       Google          531842
  1 hour ago    security@amazon.com     Amazon          729461
  2 hours ago   auth@microsoft.com      Microsoft       184629

4 recent OTP messages`

	case "admin applications":
		return `Demo Mode - Applications

  ID                     NAME                     CREATED
  app-001                Production App           Jan 15, 2024
  app-002                Development App          Mar 22, 2024

2 applications`

	case "admin connectors":
		return `Demo Mode - Connectors

  ID                     PROVIDER       NAME                 STATUS
  conn-001               google         Google Workspace     Active
  conn-002               microsoft      Microsoft 365        Active
  conn-003               imap           Custom IMAP          Inactive

3 connectors configured`

	case "admin credentials":
		return `Demo Mode - Credentials

  ID                     NAME                     TYPE           CREATED
  cred-001               Google OAuth             oauth2         Jan 15, 2024
  cred-002               MS Graph API             oauth2         Jan 20, 2024
  cred-003               IMAP Server              password       Feb 10, 2024

3 credentials stored`

	case "admin grants":
		return `Demo Mode - Grants

  ID                     EMAIL                      PROVIDER       STATUS
  grant-001              alice@example.com          Google         Active
  grant-002              bob@work.com               Microsoft      Active
  grant-003              carol@company.org          Google         Expired

3 grants`

	case "notetaker list":
		return `Demo Mode - Notetakers

  ID                     MEETING                          STATUS        CREATED
  nt-001                 Team Standup                     Completed     Dec 24
  nt-002                 Client Presentation              Recording     Now
  nt-003                 Sprint Planning                  Scheduled     Dec 27

3 notetakers`

	case "notetaker create":
		return `Demo Mode - Create Notetaker

  Created notetaker successfully!

  ID:           nt-004
  Meeting:      https://zoom.us/j/123456789
  Status:       Scheduled
  Join Time:    Immediately when meeting starts

  The notetaker bot will join and record the meeting.`

	case "notetaker media":
		return `Demo Mode - Notetaker Media

  Notetaker:    nt-001
  Meeting:      Team Standup
  Duration:     32 minutes

  Available Media:
  ✓  Video Recording    128 MB    video.mp4
  ✓  Audio Recording    24 MB     audio.mp3
  ✓  Transcript         42 KB     transcript.txt
  ✓  Summary            8 KB      summary.md

  Use --download to save files locally.`

	case "version":
		return `nylas version dev (demo mode)
Built: 2024-01-01T00:00:00Z
Go: go1.24`

	default:
		return "Demo Mode - Command: " + cmd + "\n\n(This is sample output. Connect your account with 'nylas auth login' to see real data.)"
	}
}

// Start starts the HTTP server.
func (s *Server) Start() error {
	mux := http.NewServeMux()

	// API routes
	mux.HandleFunc("/api/config/status", s.handleConfigStatus)
	mux.HandleFunc("/api/config/setup", s.handleConfigSetup)
	mux.HandleFunc("/api/grants", s.handleListGrants)
	mux.HandleFunc("/api/grants/default", s.handleSetDefaultGrant)
	mux.HandleFunc("/api/exec", s.handleExecCommand)

	// Static files (CSS, JS)
	staticFS, _ := fs.Sub(staticFiles, "static")
	fileServer := http.FileServer(http.FS(staticFS))

	// Serve static files for specific paths
	mux.Handle("/css/", fileServer)
	mux.Handle("/js/", fileServer)
	mux.Handle("/app.js", fileServer)

	// Template-rendered index page
	mux.HandleFunc("/", s.handleIndex)

	server := &http.Server{
		Addr:              s.addr,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	return server.ListenAndServe()
}

// handleIndex renders the main page with server-side data.
func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	// Only handle root path
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	// Fall back to static file if templates not loaded
	if s.templates == nil {
		staticFS, _ := fs.Sub(staticFiles, "static")
		http.FileServer(http.FS(staticFS)).ServeHTTP(w, r)
		return
	}

	// Build page data
	data := s.buildPageData()

	// Render template
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.templates.ExecuteTemplate(w, "base", data); err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
	}
}

// buildPageData gathers all data needed for the page.
func (s *Server) buildPageData() PageData {
	data := PageData{
		Commands: GetDefaultCommands(),
		DemoMode: s.demoMode,
	}

	// Demo mode: return sample data
	if s.demoMode {
		data.Configured = true
		data.ClientID = "demo-client-id"
		data.Region = "us"
		data.HasAPIKey = true
		data.DefaultGrant = demoDefaultGrant()
		data.Grants = demoGrants()
		data.DefaultGrantEmail = "alice@example.com"
		return data
	}

	// Get config status
	status, err := s.configSvc.GetStatus()
	if err == nil && status.IsConfigured {
		data.Configured = true
		data.ClientID = status.ClientID
		data.Region = status.Region
		data.HasAPIKey = status.HasAPIKey
		data.DefaultGrant = status.DefaultGrant
	}

	// Get grants and default grant from store
	grants, err := s.grantStore.ListGrants()
	if err == nil {
		// Get default grant ID from store (more reliable than config)
		defaultID, _ := s.grantStore.GetDefaultGrant()
		if defaultID != "" {
			data.DefaultGrant = defaultID
		}

		for _, g := range grants {
			data.Grants = append(data.Grants, grantFromDomain(g))

			// Set default grant email
			if g.ID == data.DefaultGrant {
				data.DefaultGrantEmail = g.Email
			}
		}
	}

	return data
}

// ConfigStatusResponse represents the config status API response.
type ConfigStatusResponse struct {
	Configured   bool   `json:"configured"`
	Region       string `json:"region"`
	ClientID     string `json:"client_id,omitempty"`
	HasAPIKey    bool   `json:"has_api_key"`
	GrantCount   int    `json:"grant_count"`
	DefaultGrant string `json:"default_grant,omitempty"`
}

func (s *Server) handleConfigStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Demo mode: return sample configured status
	if s.demoMode {
		writeJSON(w, http.StatusOK, ConfigStatusResponse{
			Configured:   true,
			Region:       "us",
			ClientID:     "demo-client-id",
			HasAPIKey:    true,
			GrantCount:   3,
			DefaultGrant: demoDefaultGrant(),
		})
		return
	}

	status, err := s.configSvc.GetStatus()
	if err != nil {
		writeJSON(w, http.StatusOK, ConfigStatusResponse{Configured: false})
		return
	}

	resp := ConfigStatusResponse{
		Configured:   status.IsConfigured,
		Region:       status.Region,
		ClientID:     status.ClientID,
		HasAPIKey:    status.HasAPIKey,
		GrantCount:   status.GrantCount,
		DefaultGrant: status.DefaultGrant,
	}

	writeJSON(w, http.StatusOK, resp)
}

// SetupRequest represents the setup API request.
type SetupRequest struct {
	APIKey string `json:"api_key"`
	Region string `json:"region"`
}

// SetupResponse represents the setup API response.
type SetupResponse struct {
	Success      bool          `json:"success"`
	Message      string        `json:"message"`
	ClientID     string        `json:"client_id,omitempty"`
	Applications []Application `json:"applications,omitempty"`
	Grants       []Grant       `json:"grants,omitempty"`
	Error        string        `json:"error,omitempty"`
}

// Application represents a Nylas application.
type Application struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Environment string `json:"environment"`
}

// Grant represents an authenticated account.
type Grant struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Provider string `json:"provider"`
}

// grantFromDomain converts a domain.GrantInfo to a Grant for API responses.
func grantFromDomain(g domain.GrantInfo) Grant {
	return Grant{
		ID:       g.ID,
		Email:    g.Email,
		Provider: string(g.Provider),
	}
}

func (s *Server) handleConfigSetup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Demo mode: simulate successful setup
	if s.demoMode {
		writeJSON(w, http.StatusOK, SetupResponse{
			Success:  true,
			Message:  "Demo mode - configuration simulated",
			ClientID: "demo-client-id",
			Applications: []Application{
				{ID: "demo-app", Name: "Demo Application", Environment: "production"},
			},
			Grants: demoGrants(),
		})
		return
	}

	var req SetupRequest
	if err := json.NewDecoder(limitedBody(w, r)).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, SetupResponse{
			Success: false,
			Error:   "Invalid request body",
		})
		return
	}

	if req.APIKey == "" {
		writeJSON(w, http.StatusBadRequest, SetupResponse{
			Success: false,
			Error:   "API key is required",
		})
		return
	}

	if req.Region == "" {
		req.Region = "us"
	}

	// Create Nylas client to detect applications
	client := nylasadapter.NewHTTPClient()
	client.SetRegion(req.Region)
	client.SetCredentials("", "", req.APIKey)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// List applications to get Client ID
	apps, err := client.ListApplications(ctx)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, SetupResponse{
			Success: false,
			Error:   "Invalid API key or could not connect to Nylas: " + err.Error(),
		})
		return
	}

	if len(apps) == 0 {
		writeJSON(w, http.StatusBadRequest, SetupResponse{
			Success: false,
			Error:   "No applications found for this API key",
		})
		return
	}

	// Use the first application's Client ID
	app := apps[0]
	clientID := app.ApplicationID
	if clientID == "" {
		clientID = app.ID
	}

	// Save configuration
	if err := s.configSvc.SetupConfig(req.Region, clientID, "", req.APIKey); err != nil {
		writeJSON(w, http.StatusInternalServerError, SetupResponse{
			Success: false,
			Error:   "Failed to save configuration: " + err.Error(),
		})
		return
	}

	// Update client with credentials for grant lookup
	client.SetCredentials(clientID, "", req.APIKey)

	// Fetch existing grants
	grants, _ := client.ListGrants(ctx)

	// Save grants locally
	var grantList []Grant
	for i, grant := range grants {
		if !grant.IsValid() {
			continue
		}

		grantInfo := domain.GrantInfo{
			ID:       grant.ID,
			Email:    grant.Email,
			Provider: grant.Provider,
		}

		_ = s.grantStore.SaveGrant(grantInfo)

		// Set first grant as default
		if i == 0 {
			_ = s.grantStore.SetDefaultGrant(grant.ID)
		}

		grantList = append(grantList, grantFromDomain(grantInfo))
	}

	// Build response
	var appList []Application
	for _, a := range apps {
		id := a.ApplicationID
		if id == "" {
			id = a.ID
		}
		appList = append(appList, Application{
			ID:          id,
			Name:        id, // Use ID as name since Application struct doesn't have Name field
			Environment: a.Environment,
		})
	}

	writeJSON(w, http.StatusOK, SetupResponse{
		Success:      true,
		Message:      "Configuration saved successfully",
		ClientID:     clientID,
		Applications: appList,
		Grants:       grantList,
	})
}

// GrantsResponse represents the grants list API response.
type GrantsResponse struct {
	Grants       []Grant `json:"grants"`
	DefaultGrant string  `json:"default_grant"`
}

func (s *Server) handleListGrants(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Demo mode: return sample grants
	if s.demoMode {
		writeJSON(w, http.StatusOK, GrantsResponse{
			Grants:       demoGrants(),
			DefaultGrant: demoDefaultGrant(),
		})
		return
	}

	grants, err := s.grantStore.ListGrants()
	if err != nil {
		writeJSON(w, http.StatusOK, GrantsResponse{Grants: []Grant{}})
		return
	}

	var grantList []Grant
	for _, g := range grants {
		grantList = append(grantList, grantFromDomain(g))
	}

	defaultID, _ := s.grantStore.GetDefaultGrant()

	writeJSON(w, http.StatusOK, GrantsResponse{
		Grants:       grantList,
		DefaultGrant: defaultID,
	})
}

// SetDefaultGrantRequest represents the request to set default grant.
type SetDefaultGrantRequest struct {
	GrantID string `json:"grant_id"`
}

// SetDefaultGrantResponse represents the response for setting default grant.
type SetDefaultGrantResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

func (s *Server) handleSetDefaultGrant(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Demo mode: simulate success
	if s.demoMode {
		writeJSON(w, http.StatusOK, SetDefaultGrantResponse{
			Success: true,
			Message: "Default account updated (demo mode)",
		})
		return
	}

	var req SetDefaultGrantRequest
	if err := json.NewDecoder(limitedBody(w, r)).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, SetDefaultGrantResponse{
			Success: false,
			Error:   "Invalid request body",
		})
		return
	}

	if req.GrantID == "" {
		writeJSON(w, http.StatusBadRequest, SetDefaultGrantResponse{
			Success: false,
			Error:   "Grant ID is required",
		})
		return
	}

	// Verify grant exists
	grants, err := s.grantStore.ListGrants()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, SetDefaultGrantResponse{
			Success: false,
			Error:   "Failed to list grants",
		})
		return
	}

	// Use slices.ContainsFunc (Go 1.21+) for cleaner lookup
	found := slices.ContainsFunc(grants, func(g domain.GrantInfo) bool {
		return g.ID == req.GrantID
	})

	if !found {
		writeJSON(w, http.StatusNotFound, SetDefaultGrantResponse{
			Success: false,
			Error:   "Grant not found",
		})
		return
	}

	if err := s.grantStore.SetDefaultGrant(req.GrantID); err != nil {
		writeJSON(w, http.StatusInternalServerError, SetDefaultGrantResponse{
			Success: false,
			Error:   "Failed to set default grant: " + err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, SetDefaultGrantResponse{
		Success: true,
		Message: "Default account updated",
	})
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

// maxRequestBodySize is the maximum allowed request body size (1MB).
// This prevents memory exhaustion attacks via large payloads.
const maxRequestBodySize = 1 << 20 // 1MB

// limitedBody wraps a request body with a size limit.
// Returns an error response if the body exceeds the limit.
func limitedBody(w http.ResponseWriter, r *http.Request) io.ReadCloser {
	return http.MaxBytesReader(w, r.Body, maxRequestBodySize)
}

// ExecRequest represents a command execution request.
type ExecRequest struct {
	Command string `json:"command"`
}

// ExecResponse represents a command execution response.
type ExecResponse struct {
	Output string `json:"output,omitempty"`
	Error  string `json:"error,omitempty"`
}

// dangerousChars are shell metacharacters that could indicate injection attempts.
// While exec.CommandContext doesn't use a shell, we reject these for defense in depth.
var dangerousChars = []string{";", "|", "&", "`", "$", "(", ")", "<", ">", "\\", "\n", "\x00"}

// containsDangerousChars checks if a command contains shell metacharacters.
func containsDangerousChars(cmd string) bool {
	for _, char := range dangerousChars {
		if strings.Contains(cmd, char) {
			return true
		}
	}
	return false
}

// Allowed commands for security (whitelist approach).
var allowedCommands = map[string]bool{
	// Auth commands
	"auth login":     true,
	"auth logout":    true,
	"auth status":    true,
	"auth whoami":    true,
	"auth list":      true,
	"auth show":      true,
	"auth switch":    true,
	"auth add":       true,
	"auth remove":    true,
	"auth revoke":    true,
	"auth config":    true,
	"auth providers": true,
	"auth detect":    true,
	"auth scopes":    true,
	"auth token":     true,
	"auth migrate":   true,
	// Email commands
	"email list":          true,
	"email read":          true,
	"email send":          true,
	"email search":        true,
	"email delete":        true,
	"email mark":          true,
	"email drafts":        true,
	"email folders":       true,
	"email threads":       true,
	"email scheduled":     true,
	"email attachments":   true,
	"email metadata":      true,
	"email tracking-info": true,
	"email ai":            true,
	"email smart-compose": true,
	// Email folder subcommands
	"email folders list":   true,
	"email folders show":   true,
	"email folders create": true,
	"email folders rename": true,
	"email folders delete": true,
	// Email drafts subcommands
	"email drafts list":   true,
	"email drafts show":   true,
	"email drafts create": true,
	"email drafts delete": true,
	"email drafts send":   true,
	// Email threads subcommands
	"email threads list":   true,
	"email threads show":   true,
	"email threads search": true,
	"email threads delete": true,
	"email threads mark":   true,
	// Email scheduled subcommands
	"email scheduled list":   true,
	"email scheduled show":   true,
	"email scheduled cancel": true,
	// Email attachments subcommands
	"email attachments list":     true,
	"email attachments show":     true,
	"email attachments download": true,
	// Calendar commands
	"calendar list":         true,
	"calendar show":         true,
	"calendar create":       true,
	"calendar update":       true,
	"calendar delete":       true,
	"calendar events":       true,
	"calendar availability": true,
	"calendar find-time":    true,
	"calendar recurring":    true,
	"calendar schedule":     true,
	"calendar virtual":      true,
	"calendar ai":           true,
	// Calendar events subcommands
	"calendar events list":   true,
	"calendar events show":   true,
	"calendar events create": true,
	"calendar events update": true,
	"calendar events delete": true,
	"calendar events rsvp":   true,
	// Calendar availability subcommands
	"calendar availability check": true,
	"calendar availability find":  true,
	// Contacts commands
	"contacts list":   true,
	"contacts show":   true,
	"contacts create": true,
	"contacts update": true,
	"contacts delete": true,
	"contacts groups": true,
	"contacts search": true,
	"contacts photo":  true,
	"contacts sync":   true,
	// Contacts groups subcommands
	"contacts groups list":   true,
	"contacts groups show":   true,
	"contacts groups create": true,
	"contacts groups delete": true,
	// Inbound commands
	"inbound list":     true,
	"inbound show":     true,
	"inbound create":   true,
	"inbound delete":   true,
	"inbound messages": true,
	"inbound monitor":  true,
	// Scheduler commands
	"scheduler configurations": true,
	"scheduler sessions":       true,
	"scheduler bookings":       true,
	"scheduler pages":          true,
	// Scheduler configurations subcommands
	"scheduler configurations list":   true,
	"scheduler configurations show":   true,
	"scheduler configurations create": true,
	"scheduler configurations update": true,
	"scheduler configurations delete": true,
	// Scheduler sessions subcommands
	"scheduler sessions list":   true,
	"scheduler sessions show":   true,
	"scheduler sessions create": true,
	"scheduler sessions delete": true,
	// Scheduler bookings subcommands
	"scheduler bookings list":    true,
	"scheduler bookings show":    true,
	"scheduler bookings create":  true,
	"scheduler bookings confirm": true,
	"scheduler bookings cancel":  true,
	"scheduler bookings delete":  true,
	// Scheduler pages subcommands
	"scheduler pages list":   true,
	"scheduler pages show":   true,
	"scheduler pages create": true,
	"scheduler pages update": true,
	"scheduler pages delete": true,
	// Timezone commands (offline utilities)
	"timezone list":         true,
	"timezone info":         true,
	"timezone convert":      true,
	"timezone find-meeting": true,
	"timezone dst":          true,
	// Webhook commands
	"webhook list":     true,
	"webhook show":     true,
	"webhook create":   true,
	"webhook update":   true,
	"webhook delete":   true,
	"webhook triggers": true,
	"webhook test":     true,
	"webhook server":   true,
	// OTP commands
	"otp get":      true,
	"otp watch":    true,
	"otp list":     true,
	"otp messages": true,
	// Admin commands
	"admin applications": true,
	"admin connectors":   true,
	"admin credentials":  true,
	"admin grants":       true,
	// Admin applications subcommands
	"admin applications list": true,
	"admin applications show": true,
	// Admin connectors subcommands
	"admin connectors list":   true,
	"admin connectors show":   true,
	"admin connectors create": true,
	"admin connectors update": true,
	"admin connectors delete": true,
	// Admin credentials subcommands
	"admin credentials list":   true,
	"admin credentials show":   true,
	"admin credentials create": true,
	"admin credentials delete": true,
	// Admin grants subcommands
	"admin grants list":   true,
	"admin grants show":   true,
	"admin grants delete": true,
	// Notetaker commands
	"notetaker list":   true,
	"notetaker show":   true,
	"notetaker create": true,
	"notetaker delete": true,
	"notetaker media":  true,
	// Other
	"version": true,
}

func (s *Server) handleExecCommand(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ExecRequest
	if err := json.NewDecoder(limitedBody(w, r)).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, ExecResponse{
			Error: "Invalid request body",
		})
		return
	}

	// Demo mode: return sample output
	if s.demoMode {
		writeJSON(w, http.StatusOK, ExecResponse{
			Output: getDemoCommandOutput(req.Command),
		})
		return
	}

	// Validate command is allowed (check base command)
	cmd := strings.TrimSpace(req.Command)

	// Check for empty command
	if cmd == "" {
		writeJSON(w, http.StatusForbidden, ExecResponse{
			Error: "Command not allowed: empty command",
		})
		return
	}

	// Check for shell metacharacters (defense in depth)
	if containsDangerousChars(cmd) {
		writeJSON(w, http.StatusForbidden, ExecResponse{
			Error: "Command not allowed: contains dangerous characters",
		})
		return
	}

	args := strings.Fields(cmd)

	// Extract base command - try 3 words, then 2, then 1
	// e.g., "calendar events list --days 7" -> "calendar events list"
	baseCmd := ""
	allowed := false

	// Try 3-word command first (e.g., "calendar events list")
	if len(args) >= 3 {
		baseCmd = args[0] + " " + args[1] + " " + args[2]
		allowed = allowedCommands[baseCmd]
	}

	// Try 2-word command (e.g., "email list")
	if !allowed && len(args) >= 2 {
		baseCmd = args[0] + " " + args[1]
		allowed = allowedCommands[baseCmd]
	}

	// Try 1-word command (e.g., "version")
	if !allowed && len(args) >= 1 {
		baseCmd = args[0]
		allowed = allowedCommands[baseCmd]
	}

	if !allowed {
		writeJSON(w, http.StatusForbidden, ExecResponse{
			Error: "Command not allowed: " + cmd,
		})
		return
	}

	// Execute the nylas command using the same binary that started this server
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Use the current executable path instead of relying on PATH
	execPath, err := os.Executable()
	if err != nil {
		execPath = "nylas" // Fallback to PATH lookup
	}

	// #nosec G204 -- execPath is the current binary from os.Executable(), user input validated in args (not in execPath)
	execCmd := exec.CommandContext(ctx, execPath, args...)
	var stdout, stderr bytes.Buffer
	execCmd.Stdout = &stdout
	execCmd.Stderr = &stderr

	err = execCmd.Run()

	output := stdout.String()
	if output == "" {
		output = stderr.String()
	}

	if err != nil {
		// Command failed but may still have useful output
		if output != "" {
			writeJSON(w, http.StatusOK, ExecResponse{
				Output: output,
			})
		} else {
			writeJSON(w, http.StatusOK, ExecResponse{
				Error: "Command failed: " + err.Error(),
			})
		}
		return
	}

	writeJSON(w, http.StatusOK, ExecResponse{
		Output: output,
	})
}
