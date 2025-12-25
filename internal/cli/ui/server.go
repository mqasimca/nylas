package ui

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"html/template"
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
		configSvc:   configSvc,
		configStore: configStore,
		secretStore: secretStore,
		grantStore:  grantStore,
		templates:   tmpl,
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

	var req SetupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
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

	var req SetDefaultGrantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
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
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, ExecResponse{
			Error: "Invalid request body",
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
