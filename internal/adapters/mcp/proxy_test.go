package mcp

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mqasimca/nylas/internal/domain"
)

func TestNewProxy(t *testing.T) {
	t.Parallel()

	proxy := NewProxy("test-api-key")

	if proxy == nil {
		t.Fatal("NewProxy returned nil")
	}
	if proxy.apiKey != "test-api-key" {
		t.Errorf("expected apiKey 'test-api-key', got '%s'", proxy.apiKey)
	}
	if proxy.endpoint != NylasMCPEndpoint {
		t.Errorf("expected endpoint '%s', got '%s'", NylasMCPEndpoint, proxy.endpoint)
	}
	if proxy.httpClient == nil {
		t.Error("httpClient is nil")
	}
}

func TestProxy_SetDefaultGrant(t *testing.T) {
	t.Parallel()

	proxy := NewProxy("test-api-key")

	// Initially empty
	if proxy.defaultGrant != "" {
		t.Errorf("expected empty defaultGrant, got '%s'", proxy.defaultGrant)
	}

	// Set grant
	proxy.SetDefaultGrant("grant-123")
	if proxy.defaultGrant != "grant-123" {
		t.Errorf("expected defaultGrant 'grant-123', got '%s'", proxy.defaultGrant)
	}
}

func TestProxy_createErrorResponse(t *testing.T) {
	t.Parallel()

	proxy := NewProxy("test-key")

	tests := []struct {
		name   string
		req    *rpcRequest
		err    error
		wantID any
	}{
		{
			name:   "with numeric id",
			req:    &rpcRequest{JSONRPC: "2.0", ID: float64(1), Method: "test"},
			err:    http.ErrNotSupported,
			wantID: float64(1),
		},
		{
			name:   "with string id",
			req:    &rpcRequest{JSONRPC: "2.0", ID: "abc", Method: "test"},
			err:    http.ErrNotSupported,
			wantID: "abc",
		},
		{
			name:   "without id",
			req:    nil,
			err:    http.ErrNotSupported,
			wantID: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := proxy.createErrorResponse(tt.req, tt.err)

			var resp struct {
				JSONRPC string `json:"jsonrpc"`
				ID      any    `json:"id"`
				Error   struct {
					Code    int    `json:"code"`
					Message string `json:"message"`
				} `json:"error"`
			}

			if err := json.Unmarshal(result, &resp); err != nil {
				t.Fatalf("failed to unmarshal response: %v", err)
			}

			if resp.JSONRPC != "2.0" {
				t.Errorf("expected jsonrpc '2.0', got '%s'", resp.JSONRPC)
			}
			if resp.ID != tt.wantID {
				t.Errorf("expected id %v, got %v", tt.wantID, resp.ID)
			}
			if resp.Error.Code != -32603 {
				t.Errorf("expected error code -32603, got %d", resp.Error.Code)
			}
		})
	}
}

func TestProxy_forward(t *testing.T) {
	t.Parallel()

	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify headers
		if r.Header.Get("Authorization") != "Bearer test-api-key" {
			t.Errorf("expected Authorization header 'Bearer test-api-key', got '%s'", r.Header.Get("Authorization"))
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected Content-Type 'application/json', got '%s'", r.Header.Get("Content-Type"))
		}

		// Return a response
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Mcp-Session-Id", "test-session-123")
		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":{"test":"ok"}}`))
	}))
	defer server.Close()

	proxy := NewProxy("test-api-key")
	proxy.endpoint = server.URL

	request := []byte(`{"jsonrpc":"2.0","id":1,"method":"test"}`)
	response, err := proxy.forward(t.Context(), request, nil)

	if err != nil {
		t.Fatalf("forward failed: %v", err)
	}

	// Verify response
	var resp map[string]any
	if err := json.Unmarshal(response, &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp["jsonrpc"] != "2.0" {
		t.Errorf("expected jsonrpc '2.0', got '%v'", resp["jsonrpc"])
	}

	// Verify session ID was stored
	if proxy.sessionID != "test-session-123" {
		t.Errorf("expected sessionID 'test-session-123', got '%s'", proxy.sessionID)
	}
}

func TestProxy_forward_WithDefaultGrant(t *testing.T) {
	t.Parallel()

	// Create a mock server that verifies the grant_id is injected into tool calls
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Read the request body to verify grant_id was injected
		body, _ := io.ReadAll(r.Body)

		var req struct {
			Method string `json:"method"`
			Params struct {
				Arguments map[string]any `json:"arguments"`
			} `json:"params"`
		}
		if err := json.Unmarshal(body, &req); err == nil {
			if req.Method == "tools/call" {
				if grantID, ok := req.Params.Arguments["grant_id"].(string); !ok || grantID != "test-grant-456" {
					t.Errorf("expected grant_id 'test-grant-456' in arguments, got '%v'", req.Params.Arguments["grant_id"])
				}
			}
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":{"grant":"ok"}}`))
	}))
	defer server.Close()

	proxy := NewProxy("test-api-key")
	proxy.endpoint = server.URL
	proxy.SetDefaultGrant("test-grant-456")

	// Test with a tools/call request
	request := []byte(`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"list_messages","arguments":{}}}`)
	response, err := proxy.forward(t.Context(), request, nil)

	if err != nil {
		t.Fatalf("forward failed: %v", err)
	}

	if response == nil {
		t.Fatal("expected response, got nil")
	}
}

func TestProxy_injectDefaultGrant(t *testing.T) {
	t.Parallel()

	proxy := NewProxy("test-api-key")
	proxy.SetDefaultGrant("my-grant-id")

	tests := []struct {
		name       string
		input      string
		wantGrant  bool
		grantValue string
	}{
		{
			name:       "injects grant_id into tools/call for list_messages",
			input:      `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"list_messages","arguments":{}}}`,
			wantGrant:  true,
			grantValue: "my-grant-id",
		},
		{
			name:       "does not override existing grant_id",
			input:      `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"list_messages","arguments":{"grant_id":"existing"}}}`,
			wantGrant:  true,
			grantValue: "existing",
		},
		{
			name:       "does not override existing identifier",
			input:      `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"list_messages","arguments":{"identifier":"user@example.com"}}}`,
			wantGrant:  false,
			grantValue: "",
		},
		{
			name:      "ignores non-tools/call methods",
			input:     `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`,
			wantGrant: false,
		},
		{
			name:      "does not inject grant_id for epoch_to_datetime utility tool",
			input:     `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"epoch_to_datetime","arguments":{"batch":[{"epoch_time":1735063516,"timezone":"America/Los_Angeles"}]}}}`,
			wantGrant: false,
		},
		{
			name:      "does not inject grant_id for current_time utility tool",
			input:     `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"current_time","arguments":{"timezone":"America/Los_Angeles"}}}`,
			wantGrant: false,
		},
		{
			name:      "does not inject grant_id for datetime_to_epoch utility tool",
			input:     `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"datetime_to_epoch","arguments":{"batch":[{"date":"2024-12-24","time":"10:00:00","timezone":"America/Los_Angeles"}]}}}`,
			wantGrant: false,
		},
		{
			name:       "injects grant_id for create_event",
			input:      `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"create_event","arguments":{}}}`,
			wantGrant:  true,
			grantValue: "my-grant-id",
		},
		{
			name:       "injects grant_id for list_calendars",
			input:      `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"list_calendars","arguments":{}}}`,
			wantGrant:  true,
			grantValue: "my-grant-id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := proxy.injectDefaultGrant([]byte(tt.input), nil)

			var parsed struct {
				Params struct {
					Arguments map[string]any `json:"arguments"`
				} `json:"params"`
			}
			if err := json.Unmarshal(result, &parsed); err != nil {
				t.Fatalf("failed to parse result: %v", err)
			}

			grantID, hasGrant := parsed.Params.Arguments["grant_id"].(string)
			if tt.wantGrant {
				if !hasGrant {
					t.Error("expected grant_id in arguments")
				} else if grantID != tt.grantValue {
					t.Errorf("expected grant_id '%s', got '%s'", tt.grantValue, grantID)
				}
			}
		})
	}
}

func TestProxy_forward_SSE(t *testing.T) {
	t.Parallel()

	// Create a mock server that returns SSE
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte("data: {\"jsonrpc\":\"2.0\",\"id\":1,\"result\":{\"sse\":true}}\n\n"))
	}))
	defer server.Close()

	proxy := NewProxy("test-api-key")
	proxy.endpoint = server.URL

	request := []byte(`{"jsonrpc":"2.0","id":1,"method":"test"}`)
	response, err := proxy.forward(t.Context(), request, nil)

	if err != nil {
		t.Fatalf("forward failed: %v", err)
	}

	// Verify SSE response was parsed
	var resp map[string]any
	if err := json.Unmarshal(response, &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	result, ok := resp["result"].(map[string]any)
	if !ok {
		t.Fatalf("expected result to be a map, got %T", resp["result"])
	}
	if result["sse"] != true {
		t.Errorf("expected sse true, got %v", result["sse"])
	}
}

func TestProxy_forward_Error(t *testing.T) {
	t.Parallel()

	// Create a mock server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Bad request"))
	}))
	defer server.Close()

	proxy := NewProxy("test-api-key")
	proxy.endpoint = server.URL

	request := []byte(`{"jsonrpc":"2.0","id":1,"method":"test"}`)
	_, err := proxy.forward(t.Context(), request, nil)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestProxy_readSSE(t *testing.T) {
	t.Parallel()

	proxy := NewProxy("test-key")

	tests := []struct {
		name     string
		input    string
		wantLen  int
		wantJSON bool
	}{
		{
			name:     "single message",
			input:    "data: {\"id\":1}\n\n",
			wantLen:  1,
			wantJSON: true,
		},
		{
			name:     "multiple messages",
			input:    "data: {\"id\":1}\n\ndata: {\"id\":2}\n\n",
			wantLen:  2,
			wantJSON: true,
		},
		{
			name:     "empty",
			input:    "",
			wantLen:  0,
			wantJSON: false,
		},
		{
			name:     "with comments",
			input:    ": comment\ndata: {\"id\":1}\n\n",
			wantLen:  1,
			wantJSON: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := proxy.readSSE(strings.NewReader(tt.input))
			if err != nil {
				t.Fatalf("readSSE failed: %v", err)
			}

			if tt.wantLen == 0 {
				if result != nil {
					t.Errorf("expected nil result, got %s", string(result))
				}
				return
			}

			if tt.wantJSON {
				if !json.Valid(result) {
					t.Errorf("expected valid JSON, got %s", string(result))
				}
			}
		})
	}
}

// mockGrantStore implements ports.GrantStore for testing.
type mockGrantStore struct {
	grants       []domain.GrantInfo
	defaultGrant string
}

func (m *mockGrantStore) SaveGrant(info domain.GrantInfo) error {
	m.grants = append(m.grants, info)
	return nil
}

func (m *mockGrantStore) GetGrant(grantID string) (*domain.GrantInfo, error) {
	for _, g := range m.grants {
		if g.ID == grantID {
			return &g, nil
		}
	}
	return nil, domain.ErrGrantNotFound
}

func (m *mockGrantStore) GetGrantByEmail(email string) (*domain.GrantInfo, error) {
	for _, g := range m.grants {
		if g.Email == email {
			return &g, nil
		}
	}
	return nil, domain.ErrGrantNotFound
}

func (m *mockGrantStore) ListGrants() ([]domain.GrantInfo, error) {
	return m.grants, nil
}

func (m *mockGrantStore) DeleteGrant(grantID string) error {
	return nil
}

func (m *mockGrantStore) SetDefaultGrant(grantID string) error {
	m.defaultGrant = grantID
	return nil
}

func (m *mockGrantStore) GetDefaultGrant() (string, error) {
	if m.defaultGrant == "" {
		return "", domain.ErrNoDefaultGrant
	}
	return m.defaultGrant, nil
}

func (m *mockGrantStore) ClearGrants() error {
	m.grants = nil
	m.defaultGrant = ""
	return nil
}

func TestProxy_handleLocalToolCall(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		request      string
		grantStore   *mockGrantStore
		defaultGrant string
		wantHandled  bool
		wantGrantID  string
		wantEmail    string
		wantError    bool
	}{
		{
			name:    "returns default grant when no email provided",
			request: `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"get_grant","arguments":{}}}`,
			grantStore: &mockGrantStore{
				grants: []domain.GrantInfo{
					{ID: "grant-123", Email: "user@example.com", Provider: "google"},
				},
			},
			defaultGrant: "grant-123",
			wantHandled:  true,
			wantGrantID:  "grant-123",
			wantEmail:    "user@example.com",
		},
		{
			name:    "returns first grant when no default set",
			request: `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"get_grant","arguments":{}}}`,
			grantStore: &mockGrantStore{
				grants: []domain.GrantInfo{
					{ID: "first-grant", Email: "first@example.com", Provider: "google"},
					{ID: "second-grant", Email: "second@example.com", Provider: "microsoft"},
				},
			},
			defaultGrant: "",
			wantHandled:  true,
			wantGrantID:  "first-grant",
			wantEmail:    "first@example.com",
		},
		{
			name:    "passes through when email is provided",
			request: `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"get_grant","arguments":{"email":"other@example.com"}}}`,
			grantStore: &mockGrantStore{
				grants: []domain.GrantInfo{
					{ID: "grant-123", Email: "user@example.com", Provider: "google"},
				},
			},
			defaultGrant: "grant-123",
			wantHandled:  false,
		},
		{
			name:         "passes through for non-get_grant tools",
			request:      `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"list_messages","arguments":{}}}`,
			grantStore:   &mockGrantStore{},
			defaultGrant: "",
			wantHandled:  false,
		},
		{
			name:         "passes through for non-tools/call methods",
			request:      `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`,
			grantStore:   &mockGrantStore{},
			defaultGrant: "",
			wantHandled:  false,
		},
		{
			name:         "returns error when no grants exist",
			request:      `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"get_grant","arguments":{}}}`,
			grantStore:   &mockGrantStore{},
			defaultGrant: "",
			wantHandled:  true,
			wantError:    true,
		},
		{
			name:         "passes through when no grant store configured",
			request:      `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"get_grant","arguments":{}}}`,
			grantStore:   nil,
			defaultGrant: "",
			wantHandled:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proxy := NewProxy("test-api-key")
			if tt.grantStore != nil {
				proxy.SetGrantStore(tt.grantStore)
			}
			if tt.defaultGrant != "" {
				proxy.SetDefaultGrant(tt.defaultGrant)
			}

			// Parse the request to pass as *rpcRequest
			var req rpcRequest
			_ = json.Unmarshal([]byte(tt.request), &req)
			response, handled := proxy.handleLocalToolCall(&req)

			if handled != tt.wantHandled {
				t.Errorf("expected handled=%v, got %v", tt.wantHandled, handled)
			}

			if !tt.wantHandled {
				return
			}

			// Parse the response
			var resp struct {
				JSONRPC string `json:"jsonrpc"`
				ID      any    `json:"id"`
				Result  struct {
					Content []struct {
						Type string `json:"type"`
						Text string `json:"text"`
					} `json:"content"`
					IsError bool `json:"isError"`
				} `json:"result"`
			}
			if err := json.Unmarshal(response, &resp); err != nil {
				t.Fatalf("failed to parse response: %v", err)
			}

			if resp.JSONRPC != "2.0" {
				t.Errorf("expected jsonrpc '2.0', got '%s'", resp.JSONRPC)
			}

			if tt.wantError {
				if !resp.Result.IsError {
					t.Error("expected isError=true")
				}
				return
			}

			if len(resp.Result.Content) == 0 {
				t.Fatal("expected content in response")
			}

			// Parse the text content as JSON
			var grantResult struct {
				GrantID  string `json:"grant_id"`
				Email    string `json:"email"`
				Provider string `json:"provider"`
			}
			if err := json.Unmarshal([]byte(resp.Result.Content[0].Text), &grantResult); err != nil {
				t.Fatalf("failed to parse grant result: %v", err)
			}

			if grantResult.GrantID != tt.wantGrantID {
				t.Errorf("expected grant_id '%s', got '%s'", tt.wantGrantID, grantResult.GrantID)
			}
			if grantResult.Email != tt.wantEmail {
				t.Errorf("expected email '%s', got '%s'", tt.wantEmail, grantResult.Email)
			}
		})
	}
}

func TestProxy_SetGrantStore(t *testing.T) {
	t.Parallel()

	proxy := NewProxy("test-api-key")

	// Initially nil
	if proxy.grantStore != nil {
		t.Error("expected grantStore to be nil initially")
	}

	// Set grant store
	store := &mockGrantStore{}
	proxy.SetGrantStore(store)

	if proxy.grantStore == nil {
		t.Error("expected grantStore to be set")
	}
}

func TestProxy_modifyInitializeResponse(t *testing.T) {
	t.Parallel()

	proxy := NewProxy("test-api-key")

	tests := []struct {
		name                 string
		response             string
		wantTimezoneGuidance bool
	}{
		{
			name: "adds timezone guidance to initialize response",
			response: `{
				"jsonrpc": "2.0",
				"id": 1,
				"result": {
					"serverInfo": {"name": "nylas"},
					"instructions": "Nylas MCP server instructions."
				}
			}`,
			wantTimezoneGuidance: true,
		},
		{
			name: "handles empty instructions",
			response: `{
				"jsonrpc": "2.0",
				"id": 1,
				"result": {
					"serverInfo": {"name": "nylas"}
				}
			}`,
			wantTimezoneGuidance: true,
		},
		{
			name:                 "handles invalid JSON",
			response:             `not valid json`,
			wantTimezoneGuidance: false,
		},
		{
			name: "handles missing result",
			response: `{
				"jsonrpc": "2.0",
				"id": 1,
				"error": {"code": -1, "message": "error"}
			}`,
			wantTimezoneGuidance: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := proxy.modifyInitializeResponse([]byte(tt.response))

			hasGuidance := strings.Contains(string(result), "Timezone Consistency")
			if hasGuidance != tt.wantTimezoneGuidance {
				t.Errorf("modifyInitializeResponse() timezone guidance = %v, want %v", hasGuidance, tt.wantTimezoneGuidance)
			}

			if tt.wantTimezoneGuidance {
				// Verify key guidance points are present
				if !strings.Contains(string(result), "epoch_to_datetime") {
					t.Error("expected guidance to mention epoch_to_datetime tool")
				}
				// Should contain the detected timezone
				if !strings.Contains(string(result), "user's local timezone is") {
					t.Error("expected guidance to include detected timezone")
				}
			}
		})
	}
}

func TestProxy_modifyToolsListResponse(t *testing.T) {
	t.Parallel()

	proxy := NewProxy("test-api-key")

	tests := []struct {
		name              string
		response          string
		wantEmailOptional bool
		wantDescModified  bool
	}{
		{
			name: "modifies get_grant to make email optional",
			response: `{
				"jsonrpc": "2.0",
				"id": 1,
				"result": {
					"tools": [
						{
							"name": "get_grant",
							"description": "Look up grant by email address.",
							"inputSchema": {
								"type": "object",
								"properties": {
									"email": {"type": "string", "description": "the email address"}
								},
								"required": ["email"]
							}
						},
						{
							"name": "list_messages",
							"description": "List messages",
							"inputSchema": {
								"type": "object",
								"properties": {},
								"required": ["grant_id"]
							}
						}
					]
				}
			}`,
			wantEmailOptional: true,
			wantDescModified:  true,
		},
		{
			name: "handles empty required array",
			response: `{
				"jsonrpc": "2.0",
				"id": 1,
				"result": {
					"tools": [
						{
							"name": "get_grant",
							"description": "Look up grant",
							"inputSchema": {
								"type": "object",
								"properties": {"email": {"type": "string"}},
								"required": []
							}
						}
					]
				}
			}`,
			wantEmailOptional: true,
			wantDescModified:  true,
		},
		{
			name: "preserves other tools unchanged",
			response: `{
				"jsonrpc": "2.0",
				"id": 1,
				"result": {
					"tools": [
						{
							"name": "list_messages",
							"description": "List messages",
							"inputSchema": {
								"type": "object",
								"required": ["grant_id"]
							}
						}
					]
				}
			}`,
			wantEmailOptional: false,
			wantDescModified:  false,
		},
		{
			name:              "handles invalid JSON",
			response:          `not json`,
			wantEmailOptional: false,
			wantDescModified:  false,
		},
		{
			name:              "handles missing result",
			response:          `{"jsonrpc":"2.0","id":1}`,
			wantEmailOptional: false,
			wantDescModified:  false,
		},
		{
			name:              "handles missing tools",
			response:          `{"jsonrpc":"2.0","id":1,"result":{}}`,
			wantEmailOptional: false,
			wantDescModified:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := proxy.modifyToolsListResponse([]byte(tt.response))

			// Parse result
			var resp map[string]any
			if err := json.Unmarshal(result, &resp); err != nil {
				if tt.wantEmailOptional || tt.wantDescModified {
					t.Fatalf("failed to parse result: %v", err)
				}
				return // Expected for invalid JSON test
			}

			// Find get_grant tool if it exists
			resultObj, ok := resp["result"].(map[string]any)
			if !ok {
				return
			}

			tools, ok := resultObj["tools"].([]any)
			if !ok {
				return
			}

			for _, tool := range tools {
				toolMap, ok := tool.(map[string]any)
				if !ok {
					continue
				}

				name, _ := toolMap["name"].(string)
				if name != "get_grant" {
					continue
				}

				// Check if email is optional (not in required)
				inputSchema, ok := toolMap["inputSchema"].(map[string]any)
				if ok {
					required, _ := inputSchema["required"].([]any)
					emailRequired := false
					for _, r := range required {
						if r == "email" {
							emailRequired = true
							break
						}
					}
					if tt.wantEmailOptional && emailRequired {
						t.Error("expected email to be optional, but it's still required")
					}
				}

				// Check if description was modified
				desc, _ := toolMap["description"].(string)
				hasModifiedDesc := strings.Contains(desc, "default authenticated grant")
				if tt.wantDescModified && !hasModifiedDesc {
					t.Error("expected description to be modified")
				}
			}
		})
	}
}

func TestProxy_forward_ModifiesToolsList(t *testing.T) {
	t.Parallel()

	// Create a mock server that returns a tools/list response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"jsonrpc": "2.0",
			"id": 1,
			"result": {
				"tools": [
					{
						"name": "get_grant",
						"description": "Look up grant by email.",
						"inputSchema": {
							"type": "object",
							"properties": {"email": {"type": "string"}},
							"required": ["email"]
						}
					}
				]
			}
		}`))
	}))
	defer server.Close()

	proxy := NewProxy("test-api-key")
	proxy.endpoint = server.URL

	// Send a tools/list request
	request := []byte(`{"jsonrpc":"2.0","id":1,"method":"tools/list"}`)
	var req rpcRequest
	_ = json.Unmarshal(request, &req)
	response, err := proxy.forward(t.Context(), request, &req)
	if err != nil {
		t.Fatalf("forward failed: %v", err)
	}

	// Parse response and verify get_grant was modified
	var resp map[string]any
	if err := json.Unmarshal(response, &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	result := resp["result"].(map[string]any)
	tools := result["tools"].([]any)
	getGrantTool := tools[0].(map[string]any)

	// Verify email is no longer required
	inputSchema := getGrantTool["inputSchema"].(map[string]any)
	required, _ := inputSchema["required"].([]any)
	for _, r := range required {
		if r == "email" {
			t.Error("expected email to be removed from required, but it's still there")
		}
	}

	// Verify description was modified
	desc := getGrantTool["description"].(string)
	if !strings.Contains(desc, "default authenticated grant") {
		t.Error("expected description to be modified")
	}
}
