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

//go:embed static/*
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
