package ui

import (
	"embed"
	"html/template"
	"io/fs"
	"net/http"
	"time"

	"github.com/mqasimca/nylas/internal/adapters/config"
	"github.com/mqasimca/nylas/internal/adapters/keyring"
	authapp "github.com/mqasimca/nylas/internal/app/auth"
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
