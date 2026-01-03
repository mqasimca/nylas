package air

import (
	"fmt"
	"net/http"

	"github.com/mqasimca/nylas/internal/air/cache"
)

// requireDefaultGrant gets the default grant ID, writing an error response if not available.
// Returns the grant ID and true if successful, or empty string and false if error written.
// Callers should return immediately when ok is false.
func (s *Server) requireDefaultGrant(w http.ResponseWriter) (grantID string, ok bool) {
	grantID, err := s.grantStore.GetDefaultGrant()
	if err != nil || grantID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "No default account. Please select an account first.",
		})
		return "", false
	}
	return grantID, true
}

// getEmailStore returns the email store for the given email account.
func (s *Server) getEmailStore(email string) (*cache.EmailStore, error) {
	if s.cacheManager == nil {
		return nil, fmt.Errorf("cache not initialized")
	}
	db, err := s.cacheManager.GetDB(email)
	if err != nil {
		return nil, err
	}
	return cache.NewEmailStore(db), nil
}

// getEventStore returns the event store for the given email account.
func (s *Server) getEventStore(email string) (*cache.EventStore, error) {
	if s.cacheManager == nil {
		return nil, fmt.Errorf("cache not initialized")
	}
	db, err := s.cacheManager.GetDB(email)
	if err != nil {
		return nil, err
	}
	return cache.NewEventStore(db), nil
}

// getContactStore returns the contact store for the given email account.
func (s *Server) getContactStore(email string) (*cache.ContactStore, error) {
	if s.cacheManager == nil {
		return nil, fmt.Errorf("cache not initialized")
	}
	db, err := s.cacheManager.GetDB(email)
	if err != nil {
		return nil, err
	}
	return cache.NewContactStore(db), nil
}

// getFolderStore returns the folder store for the given email account.
func (s *Server) getFolderStore(email string) (*cache.FolderStore, error) {
	if s.cacheManager == nil {
		return nil, fmt.Errorf("cache not initialized")
	}
	db, err := s.cacheManager.GetDB(email)
	if err != nil {
		return nil, err
	}
	return cache.NewFolderStore(db), nil
}

// getSyncStore returns the sync store for the given email account.
func (s *Server) getSyncStore(email string) (*cache.SyncStore, error) {
	if s.cacheManager == nil {
		return nil, fmt.Errorf("cache not initialized")
	}
	db, err := s.cacheManager.GetDB(email)
	if err != nil {
		return nil, err
	}
	return cache.NewSyncStore(db), nil
}
