package nylas_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mqasimca/nylas/internal/adapters/nylas"
	"github.com/mqasimca/nylas/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Application Tests

func TestHTTPClient_ListApplications(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/applications", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		response := map[string]interface{}{
			"data": []map[string]interface{}{
				{
					"id":              "app-1",
					"application_id":  "app-id-1",
					"organization_id": "org-1",
					"region":          "us",
				},
				{
					"id":              "app-2",
					"application_id":  "app-id-2",
					"organization_id": "org-2",
					"region":          "eu",
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := nylas.NewHTTPClient()
	client.SetCredentials("client-id", "secret", "api-key")
	client.SetBaseURL(server.URL)

	ctx := context.Background()
	apps, err := client.ListApplications(ctx)

	require.NoError(t, err)
	assert.Len(t, apps, 2)
	assert.Equal(t, "app-1", apps[0].ID)
	assert.Equal(t, "us", apps[0].Region)
}

func TestHTTPClient_GetApplication(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/applications/app-123", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		response := map[string]interface{}{
			"data": map[string]interface{}{
				"id":              "app-123",
				"application_id":  "app-id-123",
				"organization_id": "org-456",
				"region":          "us",
				"environment":     "production",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := nylas.NewHTTPClient()
	client.SetCredentials("client-id", "secret", "api-key")
	client.SetBaseURL(server.URL)

	ctx := context.Background()
	app, err := client.GetApplication(ctx, "app-123")

	require.NoError(t, err)
	assert.Equal(t, "app-123", app.ID)
	assert.Equal(t, "app-id-123", app.ApplicationID)
	assert.Equal(t, "production", app.Environment)
}

func TestHTTPClient_CreateApplication(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/applications", r.URL.Path)
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		var body map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&body)
		assert.Equal(t, "New Application", body["name"])
		assert.Equal(t, "us", body["region"])

		response := map[string]interface{}{
			"data": map[string]interface{}{
				"id":     "app-new",
				"name":   "New Application",
				"region": "us",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(response) // Test helper, encode error not actionable
	}))
	defer server.Close()

	client := nylas.NewHTTPClient()
	client.SetCredentials("client-id", "secret", "api-key")
	client.SetBaseURL(server.URL)

	ctx := context.Background()
	req := &domain.CreateApplicationRequest{
		Name:   "New Application",
		Region: "us",
	}
	app, err := client.CreateApplication(ctx, req)

	require.NoError(t, err)
	assert.Equal(t, "app-new", app.ID)
	assert.Equal(t, "us", app.Region)
}

func TestHTTPClient_UpdateApplication(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/applications/app-456", r.URL.Path)
		assert.Equal(t, "PATCH", r.Method)

		var body map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&body)
		assert.Equal(t, "Updated Application", body["name"])

		response := map[string]interface{}{
			"data": map[string]interface{}{
				"id":   "app-456",
				"name": "Updated Application",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := nylas.NewHTTPClient()
	client.SetCredentials("client-id", "secret", "api-key")
	client.SetBaseURL(server.URL)

	ctx := context.Background()
	name := "Updated Application"
	req := &domain.UpdateApplicationRequest{
		Name: &name,
	}
	app, err := client.UpdateApplication(ctx, "app-456", req)

	require.NoError(t, err)
	assert.Equal(t, "app-456", app.ID)
}

func TestHTTPClient_DeleteApplication(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/applications/app-delete", r.URL.Path)
		assert.Equal(t, "DELETE", r.Method)

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := nylas.NewHTTPClient()
	client.SetCredentials("client-id", "secret", "api-key")
	client.SetBaseURL(server.URL)

	ctx := context.Background()
	err := client.DeleteApplication(ctx, "app-delete")

	require.NoError(t, err)
}

// Connector Tests

func TestHTTPClient_ListConnectors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/connectors", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		response := map[string]interface{}{
			"data": []map[string]interface{}{
				{
					"id":       "conn-1",
					"name":     "Google Connector",
					"provider": "google",
				},
				{
					"id":       "conn-2",
					"name":     "Microsoft Connector",
					"provider": "microsoft",
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := nylas.NewHTTPClient()
	client.SetCredentials("client-id", "secret", "api-key")
	client.SetBaseURL(server.URL)

	ctx := context.Background()
	connectors, err := client.ListConnectors(ctx)

	require.NoError(t, err)
	assert.Len(t, connectors, 2)
	assert.Equal(t, "conn-1", connectors[0].ID)
	assert.Equal(t, "google", connectors[0].Provider)
}

func TestHTTPClient_GetConnector(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/connectors/conn-123", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		response := map[string]interface{}{
			"data": map[string]interface{}{
				"id":       "conn-123",
				"name":     "IMAP Connector",
				"provider": "imap",
				"settings": map[string]interface{}{
					"imap_host": "imap.example.com",
					"imap_port": 993,
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := nylas.NewHTTPClient()
	client.SetCredentials("client-id", "secret", "api-key")
	client.SetBaseURL(server.URL)

	ctx := context.Background()
	connector, err := client.GetConnector(ctx, "conn-123")

	require.NoError(t, err)
	assert.Equal(t, "conn-123", connector.ID)
	assert.Equal(t, "imap", connector.Provider)
	assert.NotNil(t, connector.Settings)
}

func TestHTTPClient_CreateConnector(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/connectors", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		var body map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&body)
		assert.Equal(t, "New Connector", body["name"])
		assert.Equal(t, "google", body["provider"])

		response := map[string]interface{}{
			"data": map[string]interface{}{
				"id":       "conn-new",
				"name":     "New Connector",
				"provider": "google",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := nylas.NewHTTPClient()
	client.SetCredentials("client-id", "secret", "api-key")
	client.SetBaseURL(server.URL)

	ctx := context.Background()
	req := &domain.CreateConnectorRequest{
		Name:     "New Connector",
		Provider: "google",
	}
	connector, err := client.CreateConnector(ctx, req)

	require.NoError(t, err)
	assert.Equal(t, "conn-new", connector.ID)
	assert.Equal(t, "google", connector.Provider)
}

func TestHTTPClient_UpdateConnector(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/connectors/conn-789", r.URL.Path)
		assert.Equal(t, "PATCH", r.Method)

		var body map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&body)
		assert.Equal(t, "Updated Connector", body["name"])

		response := map[string]interface{}{
			"data": map[string]interface{}{
				"id":   "conn-789",
				"name": "Updated Connector",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := nylas.NewHTTPClient()
	client.SetCredentials("client-id", "secret", "api-key")
	client.SetBaseURL(server.URL)

	ctx := context.Background()
	name := "Updated Connector"
	req := &domain.UpdateConnectorRequest{
		Name: &name,
	}
	connector, err := client.UpdateConnector(ctx, "conn-789", req)

	require.NoError(t, err)
	assert.Equal(t, "conn-789", connector.ID)
}

func TestHTTPClient_DeleteConnector(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/connectors/conn-delete", r.URL.Path)
		assert.Equal(t, "DELETE", r.Method)

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := nylas.NewHTTPClient()
	client.SetCredentials("client-id", "secret", "api-key")
	client.SetBaseURL(server.URL)

	ctx := context.Background()
	err := client.DeleteConnector(ctx, "conn-delete")

	require.NoError(t, err)
}

// Connector Credential Tests

func TestHTTPClient_ListCredentials(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/connectors/conn-123/credentials", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		response := map[string]interface{}{
			"data": []map[string]interface{}{
				{
					"id":              "cred-1",
					"name":            "OAuth Credential",
					"credential_type": "oauth",
				},
				{
					"id":              "cred-2",
					"name":            "Service Account",
					"credential_type": "service_account",
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := nylas.NewHTTPClient()
	client.SetCredentials("client-id", "secret", "api-key")
	client.SetBaseURL(server.URL)

	ctx := context.Background()
	credentials, err := client.ListCredentials(ctx, "conn-123")

	require.NoError(t, err)
	assert.Len(t, credentials, 2)
	assert.Equal(t, "cred-1", credentials[0].ID)
	assert.Equal(t, "oauth", credentials[0].CredentialType)
}

func TestHTTPClient_GetCredential(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/credentials/cred-456", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		response := map[string]interface{}{
			"data": map[string]interface{}{
				"id":              "cred-456",
				"name":            "Test Credential",
				"credential_type": "oauth",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := nylas.NewHTTPClient()
	client.SetCredentials("client-id", "secret", "api-key")
	client.SetBaseURL(server.URL)

	ctx := context.Background()
	credential, err := client.GetCredential(ctx, "cred-456")

	require.NoError(t, err)
	assert.Equal(t, "cred-456", credential.ID)
	assert.Equal(t, "oauth", credential.CredentialType)
}

func TestHTTPClient_CreateCredential(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/connectors/conn-123/credentials", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		var body map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&body)
		assert.Equal(t, "New Credential", body["name"])
		assert.Equal(t, "oauth", body["credential_type"])

		response := map[string]interface{}{
			"data": map[string]interface{}{
				"id":              "cred-new",
				"name":            "New Credential",
				"credential_type": "oauth",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := nylas.NewHTTPClient()
	client.SetCredentials("client-id", "secret", "api-key")
	client.SetBaseURL(server.URL)

	ctx := context.Background()
	req := &domain.CreateCredentialRequest{
		Name:           "New Credential",
		CredentialType: "oauth",
	}
	credential, err := client.CreateCredential(ctx, "conn-123", req)

	require.NoError(t, err)
	assert.Equal(t, "cred-new", credential.ID)
	assert.Equal(t, "oauth", credential.CredentialType)
}

func TestHTTPClient_UpdateCredential(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/credentials/cred-789", r.URL.Path)
		assert.Equal(t, "PATCH", r.Method)

		var body map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&body)
		assert.Equal(t, "Updated Credential", body["name"])

		response := map[string]interface{}{
			"data": map[string]interface{}{
				"id":   "cred-789",
				"name": "Updated Credential",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := nylas.NewHTTPClient()
	client.SetCredentials("client-id", "secret", "api-key")
	client.SetBaseURL(server.URL)

	ctx := context.Background()
	name := "Updated Credential"
	req := &domain.UpdateCredentialRequest{
		Name: &name,
	}
	credential, err := client.UpdateCredential(ctx, "cred-789", req)

	require.NoError(t, err)
	assert.Equal(t, "cred-789", credential.ID)
}

func TestHTTPClient_DeleteCredential(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/credentials/cred-delete", r.URL.Path)
		assert.Equal(t, "DELETE", r.Method)

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := nylas.NewHTTPClient()
	client.SetCredentials("client-id", "secret", "api-key")
	client.SetBaseURL(server.URL)

	ctx := context.Background()
	err := client.DeleteCredential(ctx, "cred-delete")

	require.NoError(t, err)
}

// Grant Administration Tests

func TestHTTPClient_ListAllGrants(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/grants", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		response := map[string]interface{}{
			"data": []map[string]interface{}{
				{
					"id":           "grant-1",
					"provider":     "google",
					"email":        "user1@example.com",
					"grant_status": "valid",
				},
				{
					"id":           "grant-2",
					"provider":     "microsoft",
					"email":        "user2@example.com",
					"grant_status": "valid",
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := nylas.NewHTTPClient()
	client.SetCredentials("client-id", "secret", "api-key")
	client.SetBaseURL(server.URL)

	ctx := context.Background()
	grants, err := client.ListAllGrants(ctx, nil)

	require.NoError(t, err)
	assert.Len(t, grants, 2)
	assert.Equal(t, "grant-1", grants[0].ID)
	assert.Equal(t, "google", string(grants[0].Provider))
	assert.Equal(t, "valid", grants[0].GrantStatus)
}

func TestHTTPClient_ListAllGrants_WithParams(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/grants", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		// Check query parameters
		query := r.URL.Query()
		assert.Equal(t, "10", query.Get("limit"))
		assert.Equal(t, "conn-123", query.Get("connector_id"))

		response := map[string]interface{}{
			"data": []map[string]interface{}{
				{
					"id":           "grant-1",
					"provider":     "google",
					"email":        "user@example.com",
					"grant_status": "valid",
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := nylas.NewHTTPClient()
	client.SetCredentials("client-id", "secret", "api-key")
	client.SetBaseURL(server.URL)

	ctx := context.Background()
	params := &domain.GrantsQueryParams{
		Limit:       10,
		ConnectorID: "conn-123",
	}
	grants, err := client.ListAllGrants(ctx, params)

	require.NoError(t, err)
	assert.Len(t, grants, 1)
	assert.Equal(t, "google", string(grants[0].Provider))
}

func TestHTTPClient_GetGrantStats(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/grants", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		response := map[string]interface{}{
			"data": []map[string]interface{}{
				{
					"id":           "grant-1",
					"provider":     "google",
					"grant_status": "valid",
				},
				{
					"id":           "grant-2",
					"provider":     "microsoft",
					"grant_status": "valid",
				},
				{
					"id":           "grant-3",
					"provider":     "google",
					"grant_status": "invalid",
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := nylas.NewHTTPClient()
	client.SetCredentials("client-id", "secret", "api-key")
	client.SetBaseURL(server.URL)

	ctx := context.Background()
	stats, err := client.GetGrantStats(ctx)

	require.NoError(t, err)
	assert.Equal(t, 3, stats.Total)
	assert.Equal(t, 2, stats.ByProvider["google"])
	assert.Equal(t, 1, stats.ByProvider["microsoft"])
	assert.Equal(t, 2, stats.Valid)
	assert.Equal(t, 1, stats.Invalid)
}

// Mock Client Tests

func TestMockClient_AdminOperations(t *testing.T) {
	ctx := context.Background()
	mock := nylas.NewMockClient()

	// Application tests
	apps, err := mock.ListApplications(ctx)
	require.NoError(t, err)
	assert.NotEmpty(t, apps)

	app, err := mock.GetApplication(ctx, "app-123")
	require.NoError(t, err)
	assert.Equal(t, "app-123", app.ID)

	createAppReq := &domain.CreateApplicationRequest{Name: "Test App"}
	createdApp, err := mock.CreateApplication(ctx, createAppReq)
	require.NoError(t, err)
	assert.NotEmpty(t, createdApp.ID)

	appName := "Updated App"
	updateAppReq := &domain.UpdateApplicationRequest{Name: &appName}
	updatedApp, err := mock.UpdateApplication(ctx, "app-456", updateAppReq)
	require.NoError(t, err)
	assert.Equal(t, "app-456", updatedApp.ID)

	err = mock.DeleteApplication(ctx, "app-789")
	require.NoError(t, err)

	// Connector tests
	connectors, err := mock.ListConnectors(ctx)
	require.NoError(t, err)
	assert.NotEmpty(t, connectors)

	connector, err := mock.GetConnector(ctx, "conn-123")
	require.NoError(t, err)
	assert.Equal(t, "conn-123", connector.ID)

	createConnReq := &domain.CreateConnectorRequest{Name: "Test Connector", Provider: "google"}
	createdConn, err := mock.CreateConnector(ctx, createConnReq)
	require.NoError(t, err)
	assert.NotEmpty(t, createdConn.ID)

	connName := "Updated Connector"
	updateConnReq := &domain.UpdateConnectorRequest{Name: &connName}
	updatedConn, err := mock.UpdateConnector(ctx, "conn-456", updateConnReq)
	require.NoError(t, err)
	assert.Equal(t, "Updated Connector", updatedConn.Name)

	err = mock.DeleteConnector(ctx, "conn-789")
	require.NoError(t, err)

	// Credential tests
	credentials, err := mock.ListCredentials(ctx, "conn-123")
	require.NoError(t, err)
	assert.NotEmpty(t, credentials)

	credential, err := mock.GetCredential(ctx, "cred-456")
	require.NoError(t, err)
	assert.Equal(t, "cred-456", credential.ID)

	createCredReq := &domain.CreateCredentialRequest{Name: "Test Cred", CredentialType: "oauth"}
	createdCred, err := mock.CreateCredential(ctx, "conn-123", createCredReq)
	require.NoError(t, err)
	assert.NotEmpty(t, createdCred.ID)

	credName := "Updated Cred"
	updateCredReq := &domain.UpdateCredentialRequest{Name: &credName}
	updatedCred, err := mock.UpdateCredential(ctx, "cred-789", updateCredReq)
	require.NoError(t, err)
	assert.Equal(t, "Updated Cred", updatedCred.Name)

	err = mock.DeleteCredential(ctx, "cred-delete")
	require.NoError(t, err)

	// Grant tests
	grants, err := mock.ListAllGrants(ctx, nil)
	require.NoError(t, err)
	assert.NotEmpty(t, grants)

	stats, err := mock.GetGrantStats(ctx)
	require.NoError(t, err)
	assert.Greater(t, stats.Total, 0)
}

// Demo Client Tests

func TestDemoClient_AdminOperations(t *testing.T) {
	ctx := context.Background()
	demo := nylas.NewDemoClient()

	// Application tests
	apps, err := demo.ListApplications(ctx)
	require.NoError(t, err)
	assert.NotEmpty(t, apps)

	app, err := demo.GetApplication(ctx, "demo-app")
	require.NoError(t, err)
	assert.NotEmpty(t, app.ID)

	createAppReq := &domain.CreateApplicationRequest{Name: "Demo App"}
	createdApp, err := demo.CreateApplication(ctx, createAppReq)
	require.NoError(t, err)
	assert.NotEmpty(t, createdApp.ID)

	// Connector tests
	connectors, err := demo.ListConnectors(ctx)
	require.NoError(t, err)
	assert.NotEmpty(t, connectors)

	connector, err := demo.GetConnector(ctx, "demo-conn")
	require.NoError(t, err)
	assert.NotEmpty(t, connector.ID)

	// Credential tests
	credentials, err := demo.ListCredentials(ctx, "demo-conn")
	require.NoError(t, err)
	assert.NotEmpty(t, credentials)

	// Grant tests
	grants, err := demo.ListAllGrants(ctx, nil)
	require.NoError(t, err)
	assert.NotEmpty(t, grants)

	stats, err := demo.GetGrantStats(ctx)
	require.NoError(t, err)
	assert.Greater(t, stats.Total, 0)
	assert.NotEmpty(t, stats.ByProvider)
}

// Error Handling Tests

func TestHTTPClient_GetApplication_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{ // Test helper, encode error not actionable
			"error": map[string]string{
				"message": "Application not found",
			},
		})
	}))
	defer server.Close()

	client := nylas.NewHTTPClient()
	client.SetCredentials("client-id", "secret", "api-key")
	client.SetBaseURL(server.URL)

	ctx := context.Background()
	_, err := client.GetApplication(ctx, "nonexistent")

	require.Error(t, err)
}

func TestHTTPClient_GetConnector_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{ // Test helper, encode error not actionable
			"error": map[string]string{
				"message": "Connector not found",
			},
		})
	}))
	defer server.Close()

	client := nylas.NewHTTPClient()
	client.SetCredentials("client-id", "secret", "api-key")
	client.SetBaseURL(server.URL)

	ctx := context.Background()
	_, err := client.GetConnector(ctx, "nonexistent")

	require.Error(t, err)
}
