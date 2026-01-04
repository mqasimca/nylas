package air

import (
	"net/http"
	"slices"

	"github.com/mqasimca/nylas/internal/domain"
)

// handleConfigStatus returns the current configuration status.
func (s *Server) handleConfigStatus(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}
	if s.handleDemoMode(w, ConfigStatusResponse{
		Configured:   true,
		Region:       "us",
		ClientID:     "demo-client-id",
		HasAPIKey:    true,
		GrantCount:   3,
		DefaultGrant: demoDefaultGrant(),
	}) {
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

// handleListGrants returns all authenticated accounts.
func (s *Server) handleListGrants(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}
	if s.handleDemoMode(w, GrantsResponse{
		Grants:       demoGrants(),
		DefaultGrant: demoDefaultGrant(),
	}) {
		return
	}

	grants, err := s.grantStore.ListGrants()
	if err != nil {
		writeJSON(w, http.StatusOK, GrantsResponse{Grants: []Grant{}})
		return
	}

	// Filter to only supported providers (Google, Microsoft)
	var grantList []Grant
	for _, g := range grants {
		if g.Provider.IsSupportedByAir() {
			grantList = append(grantList, grantFromDomain(g))
		}
	}

	defaultID, _ := s.grantStore.GetDefaultGrant()

	// If default grant is not a supported provider, pick the first supported account as default
	defaultIsSupported := false
	for _, g := range grantList {
		if g.ID == defaultID {
			defaultIsSupported = true
			break
		}
	}
	if !defaultIsSupported && len(grantList) > 0 {
		defaultID = grantList[0].ID
	}

	writeJSON(w, http.StatusOK, GrantsResponse{
		Grants:       grantList,
		DefaultGrant: defaultID,
	})
}

// handleSetDefaultGrant sets the default account.
func (s *Server) handleSetDefaultGrant(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}
	if s.handleDemoMode(w, SetDefaultGrantResponse{
		Success: true,
		Message: "Default account updated (demo mode)",
	}) {
		return
	}

	var req SetDefaultGrantRequest
	if !parseJSONBody(w, r, &req) {
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

// Demo mode helpers.
func demoGrants() []Grant {
	return []Grant{
		{ID: "demo-grant-001", Email: "alice@example.com", Provider: "google"},
		{ID: "demo-grant-002", Email: "bob@work.com", Provider: "microsoft"},
		{ID: "demo-grant-003", Email: "carol@company.org", Provider: "google"},
	}
}

func demoDefaultGrant() string {
	return "demo-grant-001"
}

// handleListFolders returns all folders for the current account.
func (s *Server) handleListFolders(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}
	if s.handleDemoMode(w, FoldersResponse{Folders: demoFolders()}) {
		return
	}

	// Check if configured
	if !s.requireConfig(w) {
		return
	}

	grantID, ok := s.requireDefaultGrant(w)
	if !ok {
		return
	}

	// Fetch folders from Nylas API
	ctx, cancel := s.withTimeout(r)
	defer cancel()

	folders, err := s.nylasClient.GetFolders(ctx, grantID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "Failed to fetch folders: " + err.Error(),
		})
		return
	}

	// Convert to response format
	resp := FoldersResponse{
		Folders: make([]FolderResponse, 0, len(folders)),
	}
	for _, f := range folders {
		resp.Folders = append(resp.Folders, FolderResponse{
			ID:           f.ID,
			Name:         f.Name,
			SystemFolder: f.SystemFolder,
			TotalCount:   f.TotalCount,
			UnreadCount:  f.UnreadCount,
		})
	}

	writeJSON(w, http.StatusOK, resp)
}
