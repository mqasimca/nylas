package air

import (
	"fmt"

	"github.com/mqasimca/nylas/internal/air/cache"
)

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
