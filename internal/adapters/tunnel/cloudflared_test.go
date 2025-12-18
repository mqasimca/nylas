package tunnel

import (
	"context"
	"testing"
	"time"

	"github.com/mqasimca/nylas/internal/ports"
	"github.com/stretchr/testify/assert"
)

func TestNewCloudflaredTunnel(t *testing.T) {
	tunnel := NewCloudflaredTunnel("http://localhost:3000")
	assert.NotNil(t, tunnel)
	assert.Equal(t, "http://localhost:3000", tunnel.localURL)
	assert.Equal(t, ports.TunnelStatusDisconnected, tunnel.status)
}

func TestCloudflaredTunnel_Status(t *testing.T) {
	tunnel := NewCloudflaredTunnel("http://localhost:3000")

	// Initial status should be disconnected
	assert.Equal(t, ports.TunnelStatusDisconnected, tunnel.Status())
	assert.Equal(t, "", tunnel.StatusMessage())
}

func TestCloudflaredTunnel_GetPublicURL(t *testing.T) {
	tunnel := NewCloudflaredTunnel("http://localhost:3000")

	// Initially no public URL
	assert.Equal(t, "", tunnel.GetPublicURL())
}

func TestCloudflaredTunnel_Stop(t *testing.T) {
	tunnel := NewCloudflaredTunnel("http://localhost:3000")

	// Stop should not error when not started
	err := tunnel.Stop()
	assert.NoError(t, err)
	assert.Equal(t, ports.TunnelStatusDisconnected, tunnel.Status())
}

func TestIsCloudflaredInstalled(t *testing.T) {
	// This test just checks the function runs without panic
	// The result depends on whether cloudflared is installed
	_ = IsCloudflaredInstalled()
}

// TestCloudflaredTunnel_StartNotInstalled tests behavior when cloudflared is not installed
func TestCloudflaredTunnel_StartNotInstalled(t *testing.T) {
	if IsCloudflaredInstalled() {
		t.Skip("cloudflared is installed, skipping not-installed test")
	}

	tunnel := NewCloudflaredTunnel("http://localhost:3000")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := tunnel.Start(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cloudflared not found")
}

// TestCloudflaredTunnel_Integration is a full integration test
// It requires cloudflared to be installed and is skipped in CI
func TestCloudflaredTunnel_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	if !IsCloudflaredInstalled() {
		t.Skip("cloudflared not installed, skipping integration test")
	}

	tunnel := NewCloudflaredTunnel("http://localhost:3000")
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Start tunnel
	url, err := tunnel.Start(ctx)
	if err != nil {
		t.Fatalf("failed to start tunnel: %v", err)
	}

	assert.NotEmpty(t, url)
	assert.Contains(t, url, "trycloudflare.com")
	assert.Equal(t, ports.TunnelStatusConnected, tunnel.Status())
	assert.Equal(t, url, tunnel.GetPublicURL())

	// Stop tunnel
	err = tunnel.Stop()
	assert.NoError(t, err)
	assert.Equal(t, ports.TunnelStatusDisconnected, tunnel.Status())
}

func TestCloudflaredTunnel_StatusTransitions(t *testing.T) {
	tunnel := NewCloudflaredTunnel("http://localhost:3000")

	// Verify initial state
	assert.Equal(t, ports.TunnelStatusDisconnected, tunnel.Status())

	// Simulate status transitions manually
	tunnel.mu.Lock()
	tunnel.status = ports.TunnelStatusStarting
	tunnel.statusMessage = "Starting..."
	tunnel.mu.Unlock()

	assert.Equal(t, ports.TunnelStatusStarting, tunnel.Status())
	assert.Equal(t, "Starting...", tunnel.StatusMessage())

	tunnel.mu.Lock()
	tunnel.status = ports.TunnelStatusConnected
	tunnel.publicURL = "https://test.trycloudflare.com"
	tunnel.mu.Unlock()

	assert.Equal(t, ports.TunnelStatusConnected, tunnel.Status())
	assert.Equal(t, "https://test.trycloudflare.com", tunnel.GetPublicURL())
}

func TestCloudflaredTunnel_ConcurrentAccess(t *testing.T) {
	tunnel := NewCloudflaredTunnel("http://localhost:3000")

	// Test concurrent reads don't cause races
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func() {
			_ = tunnel.Status()
			_ = tunnel.StatusMessage()
			_ = tunnel.GetPublicURL()
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}
