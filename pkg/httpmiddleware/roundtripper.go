package httpmiddleware

import "net/http"

type TenantIDRoundTripper struct {
	TenantName string
	Next       http.RoundTripper
}

func (r *TenantIDRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if r.TenantName != "" {
		req.Header.Set("X-Scope-OrgID", r.TenantName)
	}
	return r.Next.RoundTrip(req)
}
