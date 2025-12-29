package air

import (
	"context"
	"time"

	"github.com/mqasimca/nylas/internal/air/cache"
)

// startBackgroundSync starts background sync goroutines for all accounts.
func (s *Server) startBackgroundSync() {
	// Get all grants
	grants, err := s.grantStore.ListGrants()
	if err != nil || len(grants) == 0 {
		return
	}

	// Start sync for each supported account
	for _, grant := range grants {
		if !grant.Provider.IsSupportedByAir() {
			continue
		}

		s.syncWg.Add(1)
		go s.syncAccountLoop(grant.Email, grant.ID)
	}
}

// syncAccountLoop runs the sync loop for a single account.
func (s *Server) syncAccountLoop(email, grantID string) {
	defer s.syncWg.Done()

	interval := s.cacheSettings.GetSyncInterval()
	if interval < time.Minute {
		interval = 5 * time.Minute
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Initial sync
	s.syncAccount(email, grantID)

	for {
		select {
		case <-s.syncStopCh:
			return
		case <-ticker.C:
			s.syncAccount(email, grantID)
		}
	}
}

// syncAccount syncs a single account's data from the API.
func (s *Server) syncAccount(email, grantID string) {
	if s.nylasClient == nil || !s.IsOnline() {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Sync emails
	s.syncEmails(ctx, email, grantID)

	// Sync folders
	s.syncFolders(ctx, email, grantID)

	// Sync events
	s.syncEvents(ctx, email, grantID)

	// Sync contacts
	s.syncContacts(ctx, email, grantID)
}

// syncEmails syncs emails from the API to cache.
func (s *Server) syncEmails(ctx context.Context, email, grantID string) {
	store, err := s.getEmailStore(email)
	if err != nil {
		return
	}

	syncStore, err := s.getSyncStore(email)
	if err != nil {
		return
	}

	// Get last sync state
	state, _ := syncStore.Get("emails")
	if state == nil {
		state = &cache.SyncState{Resource: "emails"}
	}

	// Fetch emails from API
	messages, err := s.nylasClient.GetMessages(ctx, grantID, 100)
	if err != nil {
		s.SetOnline(false)
		return
	}
	s.SetOnline(true)

	// Cache emails
	for i := range messages {
		cached := domainMessageToCached(&messages[i])
		_ = store.Put(cached)
	}

	// Update sync state
	state.LastSync = time.Now()
	_ = syncStore.Set(state)
}

// syncFolders syncs folders from the API to cache.
func (s *Server) syncFolders(ctx context.Context, email, grantID string) {
	store, err := s.getFolderStore(email)
	if err != nil {
		return
	}

	// Fetch folders from API
	folders, err := s.nylasClient.GetFolders(ctx, grantID)
	if err != nil {
		return
	}

	// Cache folders
	for i := range folders {
		f := &folders[i]
		cached := &cache.CachedFolder{
			ID:          f.ID,
			Name:        f.Name,
			Type:        f.SystemFolder,
			TotalCount:  f.TotalCount,
			UnreadCount: f.UnreadCount,
			CachedAt:    time.Now(),
		}
		_ = store.Put(cached)
	}
}

// syncEvents syncs calendar events from the API to cache.
func (s *Server) syncEvents(ctx context.Context, email, grantID string) {
	store, err := s.getEventStore(email)
	if err != nil {
		return
	}

	// Fetch calendars first
	calendars, err := s.nylasClient.GetCalendars(ctx, grantID)
	if err != nil {
		return
	}

	// Fetch events for each calendar
	for i := range calendars {
		cal := &calendars[i]
		events, err := s.nylasClient.GetEvents(ctx, grantID, cal.ID, nil)
		if err != nil {
			continue
		}

		for j := range events {
			cached := domainEventToCached(&events[j], cal.ID)
			_ = store.Put(cached)
		}
	}
}

// syncContacts syncs contacts from the API to cache.
func (s *Server) syncContacts(ctx context.Context, email, grantID string) {
	store, err := s.getContactStore(email)
	if err != nil {
		return
	}

	// Fetch contacts from API
	contacts, err := s.nylasClient.GetContacts(ctx, grantID, nil)
	if err != nil {
		return
	}

	// Cache contacts
	for i := range contacts {
		cached := domainContactToCached(&contacts[i])
		_ = store.Put(cached)
	}
}
