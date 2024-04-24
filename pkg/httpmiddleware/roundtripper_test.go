package httpmiddleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestTenantIDRoundTripper_RoundTrip tests the RoundTrip method of TenantIDRoundTripper.
// It creates a new TenantIDRoundTripper with a TenantName and the default http transport as Next.
// It then sends a request and checks if the X-Scope-OrgID header of the request is correctly set to the TenantName.
// It also checks if the status code of the response is OK (200).
func TestTenantIDRoundTripper_RoundTrip(t *testing.T) {
	// Create a new TenantIDRoundTripper with a TenantName and the default http transport as Next.
	roundTripper := &TenantIDRoundTripper{
		TenantName: "test-tenant",
		Next:       http.DefaultTransport,
	}

	// Create a new request.
	req := httptest.NewRequest("GET", "https://example.com", nil)

	// Send the request using the RoundTrip method of the TenantIDRoundTripper.
	resp, err := roundTripper.RoundTrip(req)

	// Check if there was an error.
	require.NoError(t, err)

	// Check if the status code of the response is OK (200).
	require.Equal(t, http.StatusOK, resp.StatusCode)

	// Check if the X-Scope-OrgID header of the request is correctly set to the TenantName.
	require.Equal(t, "test-tenant", req.Header.Get("X-Scope-OrgID"))
}
