package httpmiddleware

import "net/http"

// TenantIDRoundTripper is a custom implementation of http.RoundTripper.
// It adds a tenant name to the request header before passing it to the next RoundTripper.
type TenantIDRoundTripper struct {
	// TenantName is the name of the tenant to be added to the request header.
	TenantName string
	// Next is the next RoundTripper in the chain.
	Next http.RoundTripper
}

// RoundTrip adds the tenant name to the request header and then passes the request to the next RoundTripper.
// If TenantName is not set, it simply passes the request to the next RoundTripper.
// It returns the response from the next RoundTripper and any error encountered.
func (r *TenantIDRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if r.TenantName != "" {
		req.Header.Set("X-Scope-OrgID", r.TenantName)
	}
	return r.Next.RoundTrip(req)
}
