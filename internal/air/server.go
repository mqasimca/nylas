package air

import (
	"context"
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/mqasimca/nylas/internal/adapters/config"
	"github.com/mqasimca/nylas/internal/adapters/keyring"
	"github.com/mqasimca/nylas/internal/adapters/nylas"
	"github.com/mqasimca/nylas/internal/air/cache"
	authapp "github.com/mqasimca/nylas/internal/app/auth"
	"github.com/mqasimca/nylas/internal/domain"
	"github.com/mqasimca/nylas/internal/ports"
)

//go:embed static/css/*.css static/js/*.js static/icons/* static/sw.js static/favicon.svg
var staticFiles embed.FS

//go:embed templates/*.gohtml templates/partials/*.gohtml templates/pages/*.gohtml
var templateFiles embed.FS

// Server represents the Air web UI server.
type Server struct {
	addr        string
	demoMode    bool
	configSvc   *authapp.ConfigService
	configStore ports.ConfigStore
	secretStore ports.SecretStore
	grantStore  ports.GrantStore
	nylasClient ports.NylasClient
	templates   *template.Template

	// Cache components
	cacheManager  *cache.Manager
	cacheSettings *cache.Settings
	photoStore    *cache.PhotoStore              // Contact photo cache
	offlineQueues map[string]*cache.OfflineQueue // Per-email offline queues
	syncStopCh    chan struct{}                  // Channel to stop background sync
	syncWg        sync.WaitGroup                 // Wait group for sync goroutines
	isOnline      bool                           // Online status
	onlineMu      sync.RWMutex                   // Protects isOnline

	// Productivity features (Phase 6)
	splitInboxConfig *SplitInboxConfig        // Split inbox configuration
	splitInboxMu     sync.RWMutex             // Protects splitInboxConfig
	snoozedEmails    map[string]SnoozedEmail  // Snoozed emails by email ID
	snoozeMu         sync.RWMutex             // Protects snoozedEmails
	undoSendConfig   *UndoSendConfig          // Undo send configuration
	undoSendMu       sync.RWMutex             // Protects undoSendConfig
	pendingSends     map[string]PendingSend   // Pending sends in grace period
	pendingSendMu    sync.RWMutex             // Protects pendingSends
	emailTemplates   map[string]EmailTemplate // Email templates
	templatesMu      sync.RWMutex             // Protects emailTemplates
}

// NewServer creates a new Air server.
func NewServer(addr string) *Server {
	configStore := config.NewDefaultFileStore()
	secretStore, _ := keyring.NewSecretStore(config.DefaultConfigDir())
	grantStore := keyring.NewGrantStore(secretStore)
	configSvc := authapp.NewConfigService(configStore, secretStore)

	// Create Nylas client for API calls
	var nylasClient ports.NylasClient
	cfg, err := configStore.Load()
	if err == nil {
		apiKey, _ := secretStore.Get(ports.KeyAPIKey)
		clientID, _ := secretStore.Get(ports.KeyClientID)
		clientSecret, _ := secretStore.Get(ports.KeyClientSecret)

		if apiKey != "" {
			client := nylas.NewHTTPClient()
			client.SetRegion(cfg.Region)
			client.SetCredentials(clientID, clientSecret, apiKey)
			nylasClient = client
		}
	}

	// Load templates
	tmpl, err := loadTemplates()
	if err != nil {
		// Log error and fall back to nil
		fmt.Fprintf(os.Stderr, "Warning: Failed to load templates: %v\n", err)
		tmpl = nil
	}

	// Initialize cache
	cacheCfg := cache.DefaultConfig()
	cacheManager, _ := cache.NewManager(cacheCfg)
	cacheSettings, _ := cache.LoadSettings(cacheCfg.BasePath)

	// Initialize photo store with shared database
	var photoStore *cache.PhotoStore
	photoDB, err := cache.OpenSharedDB(cacheCfg.BasePath, "photos.db")
	if err == nil {
		photoStore, _ = cache.NewPhotoStore(photoDB, cacheCfg.BasePath, cache.DefaultPhotoTTL)
		// Prune expired photos on startup
		if photoStore != nil {
			go func() {
				if pruned, err := photoStore.Prune(); err == nil && pruned > 0 {
					fmt.Fprintf(os.Stderr, "Pruned %d expired photos from cache\n", pruned)
				}
			}()
		}
	}

	return &Server{
		addr:          addr,
		demoMode:      false,
		configSvc:     configSvc,
		configStore:   configStore,
		secretStore:   secretStore,
		grantStore:    grantStore,
		nylasClient:   nylasClient,
		templates:     tmpl,
		cacheManager:  cacheManager,
		cacheSettings: cacheSettings,
		photoStore:    photoStore,
		offlineQueues: make(map[string]*cache.OfflineQueue),
		syncStopCh:    make(chan struct{}),
		isOnline:      true,
	}
}

// NewDemoServer creates an Air server in demo mode with sample data.
func NewDemoServer(addr string) *Server {
	tmpl, err := loadTemplates()
	if err != nil {
		tmpl = nil
	}

	return &Server{
		addr:      addr,
		demoMode:  true,
		templates: tmpl,
	}
}

// Start starts the HTTP server.
func (s *Server) Start() error {
	mux := http.NewServeMux()

	// API routes - Config & Grants
	mux.HandleFunc("/api/config", s.handleConfigStatus)
	mux.HandleFunc("/api/grants", s.handleListGrants)
	mux.HandleFunc("/api/grants/default", s.handleSetDefaultGrant)

	// API routes - Email (Phase 3)
	mux.HandleFunc("/api/folders", s.handleListFolders)
	mux.HandleFunc("/api/emails", s.handleListEmails)
	mux.HandleFunc("/api/emails/", s.handleEmailByID) // Handles /api/emails/:id and actions

	// API routes - Compose & Send (Phase 3)
	mux.HandleFunc("/api/drafts", s.handleDrafts)     // POST to create, GET to list
	mux.HandleFunc("/api/drafts/", s.handleDraftByID) // GET, PUT, DELETE, POST .../send
	mux.HandleFunc("/api/send", s.handleSendMessage)  // POST to send directly

	// API routes - Calendar (Phase 4)
	mux.HandleFunc("/api/calendars", s.handleListCalendars)    // GET calendars
	mux.HandleFunc("/api/events", s.handleEventsRoute)         // GET list, POST create
	mux.HandleFunc("/api/events/conflicts", s.handleConflicts) // GET conflicts for time range
	mux.HandleFunc("/api/events/", s.handleEventByID)          // GET, PUT, DELETE by ID
	mux.HandleFunc("/api/availability", s.handleAvailability)  // GET/POST find available times
	mux.HandleFunc("/api/freebusy", s.handleFreeBusy)          // GET/POST free/busy info

	// API routes - Contacts (Phase 5)
	mux.HandleFunc("/api/contacts", s.handleContactsRoute)        // GET list, POST create
	mux.HandleFunc("/api/contacts/search", s.handleContactSearch) // GET search contacts
	mux.HandleFunc("/api/contacts/", s.handleContactByID)         // GET, PUT, DELETE by ID
	mux.HandleFunc("/api/contact-groups", s.handleContactGroups)  // GET groups

	// API routes - Productivity (Phase 6)
	mux.HandleFunc("/api/inbox/split", s.handleSplitInbox)           // GET/PUT split inbox config
	mux.HandleFunc("/api/inbox/categorize", s.handleCategorizeEmail) // POST categorize email
	mux.HandleFunc("/api/inbox/vip", s.handleVIPSenders)             // GET/POST/DELETE VIP senders
	mux.HandleFunc("/api/snooze", s.handleSnooze)                    // GET/POST/DELETE snooze
	mux.HandleFunc("/api/scheduled", s.handleScheduledSend)          // GET/POST/DELETE scheduled send
	mux.HandleFunc("/api/undo-send", s.handleUndoSend)               // GET/PUT/POST undo send
	mux.HandleFunc("/api/pending-sends", s.handlePendingSends)       // GET pending sends
	mux.HandleFunc("/api/templates", s.handleTemplates)              // GET/POST email templates
	mux.HandleFunc("/api/templates/", s.handleTemplateByID)          // GET/PUT/DELETE/expand template

	// API routes - Cache (Phase 8)
	mux.HandleFunc("/api/cache/status", s.handleCacheStatus)     // GET cache status
	mux.HandleFunc("/api/cache/sync", s.handleCacheSync)         // POST trigger sync
	mux.HandleFunc("/api/cache/clear", s.handleCacheClear)       // POST clear cache
	mux.HandleFunc("/api/cache/search", s.handleCacheSearch)     // GET search cached data
	mux.HandleFunc("/api/cache/settings", s.handleCacheSettings) // GET/PUT cache settings

	// API routes - AI (Claude Code integration)
	mux.HandleFunc("/api/ai/summarize", s.handleAISummarize)              // POST summarize email
	mux.HandleFunc("/api/ai/smart-replies", s.handleAISmartReplies)       // POST smart reply suggestions
	mux.HandleFunc("/api/ai/enhanced-summary", s.handleAIEnhancedSummary) // POST enhanced summary with action items
	mux.HandleFunc("/api/ai/auto-label", s.handleAIAutoLabel)             // POST auto-label email
	mux.HandleFunc("/api/ai/thread-summary", s.handleAIThreadSummary)     // POST summarize email thread
	mux.HandleFunc("/api/ai/complete", s.handleAIComplete)                // POST smart compose autocomplete
	mux.HandleFunc("/api/ai/nl-search", s.handleNLSearch)                 // POST natural language search

	// API routes - Bundles (smart email categorization)
	mux.HandleFunc("/api/bundles", s.handleGetBundles)                  // GET list bundles
	mux.HandleFunc("/api/bundles/categorize", s.handleBundleCategorize) // POST categorize into bundle
	mux.HandleFunc("/api/bundles/emails", s.handleGetBundleEmails)      // GET emails in bundle

	// API routes - Notetaker (meeting recordings)
	mux.HandleFunc("/api/notetakers", s.handleNotetakersRoute)         // GET list, POST create
	mux.HandleFunc("/api/notetakers/", s.handleNotetakerByID)          // GET, DELETE by ID
	mux.HandleFunc("/api/notetakers/media", s.handleGetNotetakerMedia) // GET media for notetaker

	// API routes - Screener (sender approval)
	mux.HandleFunc("/api/screener", s.handleGetScreenedSenders)  // GET pending senders
	mux.HandleFunc("/api/screener/add", s.handleAddToScreener)   // POST add to screener
	mux.HandleFunc("/api/screener/allow", s.handleScreenerAllow) // POST allow sender
	mux.HandleFunc("/api/screener/block", s.handleScreenerBlock) // POST block sender

	// API routes - AI Configuration
	mux.HandleFunc("/api/ai/config", s.handleAIConfigRoute)     // GET/PUT AI config
	mux.HandleFunc("/api/ai/test", s.handleTestAIConnection)    // POST test connection
	mux.HandleFunc("/api/ai/usage", s.handleGetAIUsage)         // GET usage stats
	mux.HandleFunc("/api/ai/providers", s.handleGetAIProviders) // GET available providers

	// API routes - Read Receipts
	mux.HandleFunc("/api/receipts", s.handleGetReadReceipts)              // GET receipts
	mux.HandleFunc("/api/receipts/settings", s.handleReadReceiptSettings) // GET/PUT settings
	mux.HandleFunc("/api/track/open", s.handleTrackOpen)                  // GET tracking pixel

	// API routes - Reply Later
	mux.HandleFunc("/api/reply-later", s.handleReplyLaterRoute)             // GET list, POST add
	mux.HandleFunc("/api/reply-later/update", s.handleUpdateReplyLater)     // PUT update
	mux.HandleFunc("/api/reply-later/remove", s.handleRemoveFromReplyLater) // DELETE remove

	// API routes - Focus Mode
	mux.HandleFunc("/api/focus", s.handleFocusModeRoute)             // GET state, POST start, DELETE stop
	mux.HandleFunc("/api/focus/break", s.handleStartBreak)           // POST start break
	mux.HandleFunc("/api/focus/settings", s.handleFocusModeSettings) // GET/PUT settings

	// API routes - Analytics
	mux.HandleFunc("/api/analytics/dashboard", s.handleGetAnalyticsDashboard)    // GET dashboard
	mux.HandleFunc("/api/analytics/trends", s.handleGetAnalyticsTrends)          // GET trends
	mux.HandleFunc("/api/analytics/focus-time", s.handleGetFocusTimeSuggestions) // GET suggestions
	mux.HandleFunc("/api/analytics/productivity", s.handleGetProductivityStats)  // GET productivity

	// Static files (CSS, JS, icons)
	staticFS, _ := fs.Sub(staticFiles, "static")
	fileServer := http.FileServer(http.FS(staticFS))

	// Wrap JS files with no-cache headers to prevent stale caching
	noCacheJS := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
		fileServer.ServeHTTP(w, r)
	})

	// Serve static files for specific paths
	mux.Handle("/css/", fileServer)
	mux.Handle("/js/", noCacheJS)
	mux.Handle("/icons/", fileServer)

	// Service worker (must be served from root for proper scope)
	mux.HandleFunc("/sw.js", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/javascript")
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		data, err := staticFiles.ReadFile("static/sw.js")
		if err != nil {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write(data)
	})

	// Favicon
	mux.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/svg+xml")
		data, err := staticFiles.ReadFile("static/favicon.svg")
		if err != nil {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write(data)
	})
	mux.HandleFunc("/favicon.svg", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/svg+xml")
		data, err := staticFiles.ReadFile("static/favicon.svg")
		if err != nil {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write(data)
	})

	// Template-rendered index page
	mux.HandleFunc("/", s.handleIndex)

	// Start background sync if not in demo mode and cache is enabled
	if !s.demoMode && s.cacheManager != nil && s.cacheSettings != nil && s.cacheSettings.IsCacheEnabled() {
		s.startBackgroundSync()
	}

	// Apply middleware chain for performance and security
	// Order matters: CORS → Security → Compression → Cache → Monitoring → MethodOverride → Handler
	handler := CORSMiddleware(
		SecurityHeadersMiddleware(
			CompressionMiddleware(
				CacheMiddleware(
					PerformanceMonitoringMiddleware(
						MethodOverrideMiddleware(mux))))))

	server := &http.Server{
		Addr:              s.addr,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
		MaxHeaderBytes:    1 << 20, // 1 MB
	}

	return server.ListenAndServe()
}

// Stop gracefully stops the server and background processes.
func (s *Server) Stop() error {
	// Signal background sync to stop
	close(s.syncStopCh)

	// Wait for sync goroutines to finish
	s.syncWg.Wait()

	// Close cache manager
	if s.cacheManager != nil {
		return s.cacheManager.Close()
	}

	return nil
}

// IsOnline returns whether the server has network connectivity.
func (s *Server) IsOnline() bool {
	s.onlineMu.RLock()
	defer s.onlineMu.RUnlock()
	return s.isOnline
}

// SetOnline updates the online status.
func (s *Server) SetOnline(online bool) {
	s.onlineMu.Lock()
	s.isOnline = online
	s.onlineMu.Unlock()

	// If coming back online, process offline queue
	if online && s.cacheManager != nil {
		s.processOfflineQueues()
	}
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

// processOfflineQueues processes all pending offline actions.
func (s *Server) processOfflineQueues() {
	for email, queue := range s.offlineQueues {
		s.processOfflineQueue(email, queue)
	}
}

// processOfflineQueue processes a single account's offline queue.
func (s *Server) processOfflineQueue(email string, queue *cache.OfflineQueue) {
	if s.nylasClient == nil || !s.IsOnline() {
		return
	}

	// Get the grant ID for this email
	var grantID string
	grants, err := s.grantStore.ListGrants()
	if err != nil {
		return
	}
	for _, g := range grants {
		if g.Email == email {
			grantID = g.ID
			break
		}
	}
	if grantID == "" {
		return
	}

	ctx := context.Background()

	for {
		action, err := queue.Dequeue()
		if err != nil || action == nil {
			break
		}

		// Process the action
		err = s.processOfflineAction(ctx, grantID, action)
		if err != nil {
			// Mark as failed and re-queue if retries left
			if action.Attempts < 3 {
				_ = queue.MarkFailed(action.ID, err)
			}
		}
	}
}

// processOfflineAction processes a single offline action.
func (s *Server) processOfflineAction(ctx context.Context, grantID string, action *cache.QueuedAction) error {
	switch action.Type {
	case cache.ActionMarkRead, cache.ActionMarkUnread:
		var payload cache.MarkReadPayload
		if err := action.GetActionData(&payload); err != nil {
			return err
		}
		_, err := s.nylasClient.UpdateMessage(ctx, grantID, payload.EmailID, &domain.UpdateMessageRequest{
			Unread: &payload.Unread,
		})
		return err

	case cache.ActionStar, cache.ActionUnstar:
		var payload cache.StarPayload
		if err := action.GetActionData(&payload); err != nil {
			return err
		}
		_, err := s.nylasClient.UpdateMessage(ctx, grantID, payload.EmailID, &domain.UpdateMessageRequest{
			Starred: &payload.Starred,
		})
		return err

	case cache.ActionDelete:
		return s.nylasClient.DeleteMessage(ctx, grantID, action.ResourceID)

	case cache.ActionMove:
		var payload cache.MovePayload
		if err := action.GetActionData(&payload); err != nil {
			return err
		}
		_, err := s.nylasClient.UpdateMessage(ctx, grantID, payload.EmailID, &domain.UpdateMessageRequest{
			Folders: []string{payload.FolderID},
		})
		return err

	default:
		return fmt.Errorf("unknown action type: %s", action.Type)
	}
}

// domainMessageToCached converts a domain message to a cached email.
func domainMessageToCached(msg *domain.Message) *cache.CachedEmail {
	var fromName, fromEmail string
	if len(msg.From) > 0 {
		fromName = msg.From[0].Name
		fromEmail = msg.From[0].Email
	}

	var folderID string
	if len(msg.Folders) > 0 {
		folderID = msg.Folders[0]
	}

	return &cache.CachedEmail{
		ID:             msg.ID,
		ThreadID:       msg.ThreadID,
		FolderID:       folderID,
		Subject:        msg.Subject,
		Snippet:        msg.Snippet,
		FromName:       fromName,
		FromEmail:      fromEmail,
		To:             participantsToStrings(msg.To),
		CC:             participantsToStrings(msg.Cc),
		BCC:            participantsToStrings(msg.Bcc),
		Date:           msg.Date,
		Unread:         msg.Unread,
		Starred:        msg.Starred,
		HasAttachments: len(msg.Attachments) > 0,
		BodyHTML:       msg.Body,
		BodyText:       msg.Body, // Simplified
		CachedAt:       time.Now(),
	}
}

// domainEventToCached converts a domain event to a cached event.
func domainEventToCached(evt *domain.Event, calendarID string) *cache.CachedEvent {
	return &cache.CachedEvent{
		ID:           evt.ID,
		CalendarID:   calendarID,
		Title:        evt.Title,
		Description:  evt.Description,
		Location:     evt.Location,
		StartTime:    time.Unix(evt.When.StartTime, 0),
		EndTime:      time.Unix(evt.When.EndTime, 0),
		AllDay:       evt.When.Object == "date" || evt.When.Object == "datespan",
		Status:       evt.Status,
		Busy:         evt.Busy,
		Participants: eventParticipantsToStrings(evt.Participants),
		CachedAt:     time.Now(),
	}
}

// domainContactToCached converts a domain contact to a cached contact.
func domainContactToCached(c *domain.Contact) *cache.CachedContact {
	var email, phone, company, jobTitle string
	if len(c.Emails) > 0 {
		email = c.Emails[0].Email
	}
	if len(c.PhoneNumbers) > 0 {
		phone = c.PhoneNumbers[0].Number
	}
	if len(c.CompanyName) > 0 {
		company = c.CompanyName
	}
	if len(c.JobTitle) > 0 {
		jobTitle = c.JobTitle
	}

	return &cache.CachedContact{
		ID:          c.ID,
		Email:       email,
		GivenName:   c.GivenName,
		Surname:     c.Surname,
		DisplayName: c.GivenName + " " + c.Surname,
		Phone:       phone,
		Company:     company,
		JobTitle:    jobTitle,
		Notes:       c.Notes,
		CachedAt:    time.Now(),
	}
}

// participantsToStrings converts email participants to a slice of strings.
func participantsToStrings(participants []domain.EmailParticipant) []string {
	result := make([]string, 0, len(participants))
	for _, p := range participants {
		if p.Name != "" {
			result = append(result, p.Name+" <"+p.Email+">")
		} else {
			result = append(result, p.Email)
		}
	}
	return result
}

// eventParticipantsToStrings converts event participants to a slice of strings.
func eventParticipantsToStrings(participants []domain.Participant) []string {
	result := make([]string, 0, len(participants))
	for _, p := range participants {
		if p.Name != "" {
			result = append(result, p.Name+" <"+p.Email+">")
		} else {
			result = append(result, p.Email)
		}
	}
	return result
}

// handleIndex renders the main page.
func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	// Only handle root path
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	// Fall back to static file if templates not loaded
	if s.templates == nil {
		http.Error(w, "Templates not loaded", http.StatusInternalServerError)
		return
	}

	// Build page data - use real data when available, fall back to mock
	data := s.buildPageData()

	// Render template
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.templates.ExecuteTemplate(w, "base", data); err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
	}
}

// buildPageData gathers all data needed for the page.
func (s *Server) buildPageData() PageData {
	// Start with mock data as base
	data := buildMockPageData()

	// Demo mode: return mock data
	if s.demoMode {
		return data
	}

	// Non-demo mode: clear mock data so JavaScript loads real data
	// This prevents the "flash" of mock data before real data loads
	data.Emails = nil
	data.SelectedEmail = nil
	data.Events = nil
	data.Calendars = nil
	data.Contacts = nil

	// Get real config status
	status, err := s.configSvc.GetStatus()
	if err == nil && status.IsConfigured {
		data.Configured = true
		data.ClientID = status.ClientID
		data.Region = status.Region
		data.HasAPIKey = status.HasAPIKey
	}

	// Get real grants (filter to supported providers: Google, Microsoft)
	grants, err := s.grantStore.ListGrants()
	if err == nil && len(grants) > 0 {
		// Filter to supported providers only
		var supportedGrants []domain.GrantInfo
		for _, g := range grants {
			if g.Provider.IsSupportedByAir() {
				supportedGrants = append(supportedGrants, g)
			}
		}

		if len(supportedGrants) > 0 {
			// Get default grant ID
			defaultID, _ := s.grantStore.GetDefaultGrant()

			// Check if default is a supported provider, otherwise use first supported account
			defaultIsSupported := false
			for _, g := range supportedGrants {
				if g.ID == defaultID {
					defaultIsSupported = true
					break
				}
			}
			if !defaultIsSupported {
				defaultID = supportedGrants[0].ID
			}

			// Find default grant info
			for _, g := range supportedGrants {
				if g.ID == defaultID {
					data.UserEmail = g.Email
					data.UserName = extractName(g.Email)
					data.UserAvatar = initials(g.Email)
					data.DefaultGrantID = g.ID
					data.Provider = string(g.Provider)
					break
				}
			}

			// Build grants list for UI (supported providers only)
			data.Grants = make([]GrantInfo, 0, len(supportedGrants))
			for _, g := range supportedGrants {
				data.Grants = append(data.Grants, GrantInfo{
					ID:        g.ID,
					Email:     g.Email,
					Provider:  string(g.Provider),
					IsDefault: g.ID == defaultID,
				})
			}
			data.AccountsCount = len(supportedGrants)
		}
	}

	return data
}

// extractName extracts a display name from an email address.
func extractName(email string) string {
	// Simple extraction: use the part before @ and capitalize
	for i, c := range email {
		if c == '@' {
			name := email[:i]
			// Capitalize first letter
			if len(name) > 0 {
				return string(name[0]-32) + name[1:]
			}
			return name
		}
	}
	return email
}

// initials returns the initials from an email address.
func initials(email string) string {
	// Get first letter of email
	if len(email) == 0 {
		return "?"
	}
	// Uppercase first letter
	c := email[0]
	if c >= 'a' && c <= 'z' {
		c -= 32
	}
	return string(c)
}

// loadTemplates parses all template files.
func loadTemplates() (*template.Template, error) {
	return template.New("").Funcs(templateFuncs).ParseFS(
		templateFiles,
		"templates/*.gohtml",
		"templates/partials/*.gohtml",
		"templates/pages/*.gohtml",
	)
}

// Template functions.
var templateFuncs = template.FuncMap{
	"safeHTML": func(s string) template.HTML {
		//nolint:gosec // G203: We control the input, this is for rendering pre-defined HTML
		return template.HTML(s)
	},
}
