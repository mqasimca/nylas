package ui

import (
	"encoding/json"
	"net/http"

	nylasadapter "github.com/mqasimca/nylas/internal/adapters/nylas"
	"github.com/mqasimca/nylas/internal/cli/common"
	"github.com/mqasimca/nylas/internal/domain"
)

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

	ctx, cancel := common.CreateContext()
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
