// Package mcp provides an MCP proxy that forwards requests to the Nylas MCP server.
package mcp

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/mqasimca/nylas/internal/domain"
	"github.com/mqasimca/nylas/internal/ports"
)

const (
	// NylasMCPEndpointUS is the US regional MCP endpoint.
	NylasMCPEndpointUS = "https://mcp.us.nylas.com"
	// NylasMCPEndpointEU is the EU regional MCP endpoint.
	NylasMCPEndpointEU = "https://mcp.eu.nylas.com"

	// DefaultTimeout for HTTP requests.
	DefaultTimeout = 90 * time.Second
)

// GetMCPEndpoint returns the appropriate MCP endpoint for the given region.
func GetMCPEndpoint(region string) string {
	switch strings.ToLower(region) {
	case "eu":
		return NylasMCPEndpointEU
	default:
		return NylasMCPEndpointUS
	}
}

// rpcRequest represents a JSON-RPC request structure.
// Defined once to avoid duplicate parsing.
type rpcRequest struct {
	JSONRPC string `json:"jsonrpc"`
	ID      any    `json:"id"`
	Method  string `json:"method"`
	Params  struct {
		Name      string         `json:"name"`
		Arguments map[string]any `json:"arguments"`
	} `json:"params"`
}

// Proxy forwards MCP requests from STDIO to the Nylas MCP server.
type Proxy struct {
	endpoint     string
	apiKey       string
	authHeader   string // Cached "Bearer <apiKey>" value
	defaultGrant string
	grantStore   ports.GrantStore
	httpClient   *http.Client
	sessionID    string
	mu           sync.RWMutex
}

// NewProxy creates a new MCP proxy with the given API key and region.
func NewProxy(apiKey, region string) *Proxy {
	return &Proxy{
		endpoint:   GetMCPEndpoint(region),
		apiKey:     apiKey,
		authHeader: "Bearer " + apiKey, // Cache auth header
		httpClient: &http.Client{
			Timeout: DefaultTimeout,
		},
	}
}

// SetDefaultGrant sets the default grant ID to use for requests.
// This helps the MCP server know which account to use by default.
func (p *Proxy) SetDefaultGrant(grantID string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.defaultGrant = grantID
}

// SetGrantStore sets the grant store for local grant operations.
// This allows the proxy to respond to grant queries locally without
// requiring the AI to provide an email address.
func (p *Proxy) SetGrantStore(store ports.GrantStore) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.grantStore = store
}

// Run starts the proxy, reading from stdin and writing to stdout.
func (p *Proxy) Run(ctx context.Context) error {
	reader := bufio.NewReader(os.Stdin)
	writer := bufio.NewWriter(os.Stdout)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Read a line (JSON-RPC message)
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return fmt.Errorf("reading stdin: %w", err)
		}

		// Skip empty lines
		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			continue
		}

		// Parse JSON once for all operations
		var req rpcRequest
		if err := json.Unmarshal(line, &req); err != nil {
			// Not valid JSON - forward as-is, let server handle error
			response, fwdErr := p.forward(ctx, line, nil)
			if fwdErr != nil {
				errorResp := p.createErrorResponse(nil, fwdErr)
				_, _ = writer.Write(append(errorResp, '\n'))
				_ = writer.Flush()
				continue
			}
			if len(response) > 0 {
				_, _ = writer.Write(append(response, '\n'))
				_ = writer.Flush()
			}
			continue
		}

		// Try to handle locally first (for get_grant without email)
		if localResponse, handled := p.handleLocalToolCall(&req); handled {
			if len(localResponse) > 0 {
				if _, err := writer.Write(append(localResponse, '\n')); err != nil {
					return fmt.Errorf("writing local response: %w", err)
				}
				_ = writer.Flush()
			}
			continue
		}

		// Forward to Nylas MCP server
		response, err := p.forward(ctx, line, &req)
		if err != nil {
			// Write error response
			errorResp := p.createErrorResponse(&req, err)
			if _, writeErr := writer.Write(append(errorResp, '\n')); writeErr != nil {
				return fmt.Errorf("writing error response: %w", writeErr)
			}
			_ = writer.Flush()
			continue
		}

		// Write response
		if len(response) > 0 {
			if _, err := writer.Write(append(response, '\n')); err != nil {
				return fmt.Errorf("writing response: %w", err)
			}
			_ = writer.Flush()
		}
	}
}

// forward sends a request to the Nylas MCP server and returns the response.
// The parsed rpcRequest is optional - if nil, request is forwarded as-is.
func (p *Proxy) forward(ctx context.Context, request []byte, parsed *rpcRequest) ([]byte, error) {
	// Check request types that need response modification
	isToolsList := parsed != nil && parsed.Method == "tools/list"
	isInitialize := parsed != nil && parsed.Method == "initialize"

	// Inject default grant into tool calls if not specified
	request = p.injectDefaultGrant(request, parsed)

	req, err := http.NewRequestWithContext(ctx, "POST", p.endpoint, bytes.NewReader(request))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	// Set required headers (use cached auth header)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	req.Header.Set("Authorization", p.authHeader)

	// Include session ID and default grant if we have them (read lock)
	p.mu.RLock()
	if p.sessionID != "" {
		req.Header.Set("Mcp-Session-Id", p.sessionID)
	}
	if p.defaultGrant != "" {
		req.Header.Set("X-Nylas-Grant-Id", p.defaultGrant)
	}
	p.mu.RUnlock()

	// Send request
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sending request: %w", err)
	}
	defer resp.Body.Close()

	// Store session ID if provided
	if sessionID := resp.Header.Get("Mcp-Session-Id"); sessionID != "" {
		p.mu.Lock()
		p.sessionID = sessionID
		p.mu.Unlock()
	}

	// Handle response based on content type
	contentType := resp.Header.Get("Content-Type")

	// Handle 202 Accepted (no body)
	if resp.StatusCode == http.StatusAccepted {
		return nil, nil
	}

	// Handle errors
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("server returned %d: %s", resp.StatusCode, string(body))
	}

	// Handle SSE stream
	if strings.HasPrefix(contentType, "text/event-stream") {
		body, err := p.readSSE(resp.Body)
		if err != nil {
			return nil, err
		}
		// Modify responses as needed
		if isToolsList {
			body = p.modifyToolsListResponse(body)
		}
		if isInitialize {
			body = p.modifyInitializeResponse(body)
		}
		return body, nil
	}

	// Handle JSON response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	// Modify responses as needed
	if isToolsList {
		body = p.modifyToolsListResponse(body)
	}
	if isInitialize {
		body = p.modifyInitializeResponse(body)
	}

	return body, nil
}

// readSSE reads Server-Sent Events and extracts JSON-RPC messages.
func (p *Proxy) readSSE(reader io.Reader) ([]byte, error) {
	scanner := bufio.NewScanner(reader)
	var responses []json.RawMessage

	for scanner.Scan() {
		line := scanner.Text()

		// SSE data lines start with "data: "
		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")
			if data != "" {
				responses = append(responses, json.RawMessage(data))
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading SSE: %w", err)
	}

	// Return single response or batch
	if len(responses) == 0 {
		return nil, nil
	}
	if len(responses) == 1 {
		return responses[0], nil
	}

	// Batch multiple responses
	batch, err := json.Marshal(responses)
	if err != nil {
		return nil, fmt.Errorf("marshaling batch: %w", err)
	}
	return batch, nil
}

// toolsRequiringGrant lists tools that accept a grant_id parameter at the root level.
// Utility tools like time converters should NOT have grant_id injected.
// Excluded tools (don't accept grant_id at root level):
//   - confirm_send_message, confirm_send_draft: only validate message content
//   - availability: grant_id goes inside participants array, not at root
var toolsRequiringGrant = map[string]bool{
	"get_grant":        true,
	"list_calendars":   true,
	"list_events":      true,
	"create_event":     true,
	"update_event":     true,
	"list_messages":    true,
	"list_threads":     true,
	"get_folder_by_id": true,
	"create_draft":     true,
	"update_draft":     true,
	"send_draft":       true,
	"send_message":     true,
}

// injectDefaultGrant injects the default grant_id into tool call requests if not already specified.
// Uses the pre-parsed request if available to avoid re-parsing.
func (p *Proxy) injectDefaultGrant(request []byte, parsed *rpcRequest) []byte {
	p.mu.RLock()
	defaultGrant := p.defaultGrant
	p.mu.RUnlock()

	if defaultGrant == "" {
		return request
	}

	// Use parsed request if available, otherwise parse
	var req *rpcRequest
	if parsed != nil {
		req = parsed
	} else {
		var r rpcRequest
		if err := json.Unmarshal(request, &r); err != nil {
			return request // Not valid JSON, pass through
		}
		req = &r
	}

	// Only process tools/call requests
	if req.Method != "tools/call" {
		return request
	}

	// Only inject grant_id for tools that need it
	// Utility tools like epoch_to_datetime, current_time, datetime_to_epoch don't accept grant_id
	if !toolsRequiringGrant[req.Params.Name] {
		return request
	}

	// Check if grant_id or identifier is already specified
	if req.Params.Arguments == nil {
		req.Params.Arguments = make(map[string]any)
	}

	// Don't override if already set
	if _, hasGrantID := req.Params.Arguments["grant_id"]; hasGrantID {
		return request
	}
	if _, hasIdentifier := req.Params.Arguments["identifier"]; hasIdentifier {
		return request
	}

	// Inject the default grant_id
	req.Params.Arguments["grant_id"] = defaultGrant

	// Re-marshal the request
	modified, err := json.Marshal(req)
	if err != nil {
		return request // Marshal failed, use original
	}

	return modified
}

// handleLocalToolCall checks if a tool call can be handled locally.
// Returns the response and true if handled locally, nil and false otherwise.
// Uses the pre-parsed request to avoid re-parsing.
func (p *Proxy) handleLocalToolCall(req *rpcRequest) ([]byte, bool) {
	p.mu.RLock()
	grantStore := p.grantStore
	defaultGrant := p.defaultGrant
	p.mu.RUnlock()

	// Need grant store for local handling
	if grantStore == nil {
		return nil, false
	}

	// Only handle tools/call for get_grant
	if req.Method != "tools/call" || req.Params.Name != "get_grant" {
		return nil, false
	}

	// Check if email is provided - if so, let cloud handle it
	if req.Params.Arguments != nil {
		if email, ok := req.Params.Arguments["email"].(string); ok && email != "" {
			return nil, false
		}
	}

	// No email provided - return the default grant from local storage
	var grantInfo *domain.GrantInfo
	var err error

	if defaultGrant != "" {
		grantInfo, err = grantStore.GetGrant(defaultGrant)
	}

	// If no default grant or not found, try to get the first available grant
	if grantInfo == nil || err != nil {
		grants, listErr := grantStore.ListGrants()
		if listErr == nil && len(grants) > 0 {
			grantInfo = &grants[0]
		}
	}

	if grantInfo == nil {
		// Return error response
		return p.createToolErrorResponse(req.ID, "No authenticated grants found. Please run 'nylas auth login' first."), true
	}

	// Build successful response
	return p.createToolSuccessResponse(req.ID, map[string]any{
		"grant_id": grantInfo.ID,
		"email":    grantInfo.Email,
		"provider": string(grantInfo.Provider),
	}), true
}

// createToolSuccessResponse creates a successful MCP tool call response.
func (p *Proxy) createToolSuccessResponse(id any, result map[string]any) []byte {
	// Format result as text content (MCP tool response format)
	resultJSON, _ := json.Marshal(result)

	response := map[string]any{
		"jsonrpc": "2.0",
		"id":      id,
		"result": map[string]any{
			"content": []map[string]any{
				{
					"type": "text",
					"text": string(resultJSON),
				},
			},
		},
	}

	resp, _ := json.Marshal(response)
	return resp
}

// createToolErrorResponse creates an error response for a tool call.
func (p *Proxy) createToolErrorResponse(id any, message string) []byte {
	response := map[string]any{
		"jsonrpc": "2.0",
		"id":      id,
		"result": map[string]any{
			"content": []map[string]any{
				{
					"type": "text",
					"text": message,
				},
			},
			"isError": true,
		},
	}

	resp, _ := json.Marshal(response)
	return resp
}

// createErrorResponse creates a JSON-RPC error response.
// Uses the pre-parsed request if available to get the ID.
func (p *Proxy) createErrorResponse(req *rpcRequest, err error) []byte {
	var id any
	if req != nil {
		id = req.ID
	}

	errorResp := map[string]any{
		"jsonrpc": "2.0",
		"id":      id,
		"error": map[string]any{
			"code":    -32603,
			"message": err.Error(),
		},
	}

	result, _ := json.Marshal(errorResp)
	return result
}

// modifyToolsListResponse modifies the tools/list response to make get_grant email optional.
// This allows AI assistants to call get_grant without providing an email,
// which triggers the local grant lookup in handleLocalToolCall.
func (p *Proxy) modifyToolsListResponse(response []byte) []byte {
	// Parse the JSON-RPC response
	var rpcResp map[string]any
	if err := json.Unmarshal(response, &rpcResp); err != nil {
		return response
	}

	// Navigate to result.tools
	result, ok := rpcResp["result"].(map[string]any)
	if !ok {
		return response
	}

	tools, ok := result["tools"].([]any)
	if !ok {
		return response
	}

	// Find and modify the get_grant tool
	for _, tool := range tools {
		toolMap, ok := tool.(map[string]any)
		if !ok {
			continue
		}

		name, ok := toolMap["name"].(string)
		if !ok || name != "get_grant" {
			continue
		}

		// Found get_grant - modify its inputSchema to make email optional
		inputSchema, ok := toolMap["inputSchema"].(map[string]any)
		if !ok {
			continue
		}

		// Remove "email" from required array
		required, ok := inputSchema["required"].([]any)
		if ok {
			newRequired := make([]any, 0, len(required))
			for _, r := range required {
				if r != "email" {
					newRequired = append(newRequired, r)
				}
			}
			inputSchema["required"] = newRequired
		}

		// Update the description to indicate email is optional
		if desc, ok := toolMap["description"].(string); ok {
			toolMap["description"] = desc + " If email is not provided, returns the default authenticated grant."
		}

		break
	}

	// Re-marshal the modified response
	modified, err := json.Marshal(rpcResp)
	if err != nil {
		return response
	}

	return modified
}

// modifyInitializeResponse enhances the initialize response with timezone guidance.
// This ensures AI assistants display all timestamps consistently in the user's timezone.
func (p *Proxy) modifyInitializeResponse(response []byte) []byte {
	// Parse the JSON-RPC response
	var rpcResp map[string]any
	if err := json.Unmarshal(response, &rpcResp); err != nil {
		return response
	}

	// Navigate to result
	result, ok := rpcResp["result"].(map[string]any)
	if !ok {
		return response
	}

	// Get existing instructions
	instructions, _ := result["instructions"].(string)

	// Detect system timezone
	localZone, _ := time.Now().Zone()
	tzName := time.Local.String()
	if tzName == "Local" {
		tzName = localZone // Fallback to abbreviation if no IANA name
	}

	// Append timezone guidance with detected timezone
	timezoneGuidance := fmt.Sprintf(`

IMPORTANT - Timezone Consistency:
The user's local timezone is: %s (%s)
When displaying ANY timestamps to users (from emails, events, availability, etc.):
1. Always use epoch_to_datetime tool with timezone "%s" to convert Unix timestamps
2. Display ALL times in %s, never in UTC or the event's original timezone
3. Format times clearly (e.g., "2:00 PM %s")`, tzName, localZone, tzName, localZone, localZone)

	result["instructions"] = instructions + timezoneGuidance
	rpcResp["result"] = result

	// Re-marshal the modified response
	modified, err := json.Marshal(rpcResp)
	if err != nil {
		return response
	}

	return modified
}
