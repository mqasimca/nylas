// Package gonylas provides a Go library for Nylas CLI plugins.
//
// This library allows plugins to interact with Nylas API using the same
// infrastructure as the core CLI, including:
//   - Configuration management
//   - API clients (Email, Calendar, Contacts)
//   - Authentication helpers
//
// # Quick Start
//
// Plugins receive configuration via environment variables set by the core CLI:
//
//	package main
//
//	import (
//	    "context"
//	    "fmt"
//	    "log"
//
//	    gonylas "github.com/mqasimca/nylas/pkg/go-nylas"
//	    "github.com/mqasimca/nylas/pkg/go-nylas/client"
//	    "github.com/mqasimca/nylas/pkg/go-nylas/config"
//	)
//
//	func main() {
//	    // Check CLI compatibility
//	    if err := config.CheckCompatibility(); err != nil {
//	        log.Fatal(err)
//	    }
//
//	    // Load config from environment (set by core CLI)
//	    cfg, err := config.LoadFromEnv()
//	    if err != nil {
//	        log.Fatal(err)
//	    }
//
//	    // Create email client
//	    emailClient, err := client.NewEmailClient(cfg)
//	    if err != nil {
//	        log.Fatal(err)
//	    }
//
//	    // List messages
//	    messages, err := emailClient.List(context.Background(), &client.ListOptions{Limit: 10})
//	    if err != nil {
//	        log.Fatal(err)
//	    }
//
//	    fmt.Printf("Found %d messages\n", len(messages))
//	}
//
// # Environment Variables
//
// The core CLI sets these environment variables when executing plugins:
//   - NYLAS_CLI_VERSION: Core CLI version
//   - NYLAS_API_KEY: API key for authentication
//   - NYLAS_GRANT_ID: Grant ID for the current user
//   - NYLAS_REGION: API region (us, eu, etc.)
//   - NYLAS_CLIENT_ID: OAuth client ID
//
// # Version Compatibility
//
// Plugins should check compatibility with the core CLI:
//
//	if err := config.CheckCompatibility(); err != nil {
//	    log.Fatal("incompatible CLI version:", err)
//	}
package gonylas

// Version is the go-nylas library version.
const Version = "1.0.0"
