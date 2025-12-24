package cli

import (
	"github.com/mqasimca/nylas/examples/minimal-feature/adapters"
	"github.com/mqasimca/nylas/examples/minimal-feature/ports"
)

// getWidgetService returns the widget service instance.
//
// In a real application, this would:
// - Read configuration (API URL, credentials)
// - Initialize the appropriate adapter (real vs mock)
// - Return the service interface
//
// For this example, we show the pattern.
func getWidgetService() ports.WidgetService {
	// In real app: read from config
	apiURL := "https://api.example.com"
	apiKey := "your-api-key"

	// Return adapter that implements the port
	return adapters.NewWidgetAdapter(apiURL, apiKey)
}

// For testing, you would use:
// service := &adapters.MockWidgetService{
//     ListFunc: func(ctx) ([]*Widget, error) {
//         return mockWidgets, nil
//     },
// }
