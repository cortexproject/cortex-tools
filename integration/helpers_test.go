//go:build integration || integration_utf8

package integration

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cortexproject/cortex-tools/pkg/client"
)

func cortexAddress() string {
	if addr := os.Getenv("CORTEX_ADDRESS"); addr != "" {
		return addr
	}
	return "http://localhost:9009"
}

func newClient(t *testing.T) *client.CortexClient {
	t.Helper()
	c, err := client.New(client.Config{
		Address: cortexAddress(),
		ID:      "fake",
	})
	require.NoError(t, err)
	return c
}
