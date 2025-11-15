package config

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestLoadConfig(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config")

	testConfig := &Config{
		CurrentContext: "test-context",
		Contexts: []NamedContext{
			{
				Name: "test-context",
				Context: Context{
					Cluster: "test-cluster",
					User:    "test-user",
				},
			},
		},
		Clusters: []NamedCluster{
			{
				Name: "test-cluster",
				Cluster: Cluster{
					Address:         "https://cortex.example.com",
					TLSCAPath:       "/path/to/ca.crt",
					TLSCertPath:     "/path/to/client.crt",
					TLSKeyPath:      "/path/to/client.key",
					UseLegacyRoutes: true,
					RulerAPIPath:    "/custom/ruler/path",
				},
			},
		},
		Users: []NamedUser{
			{
				Name: "test-user",
				User: User{
					ID:        "test-tenant",
					AuthToken: "test-token",
				},
			},
		},
	}

	// Write test config
	data, err := yaml.Marshal(testConfig)
	if err != nil {
		t.Fatalf("failed to marshal test config: %v", err)
	}
	if err := os.WriteFile(configPath, data, 0600); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	// Test loading
	loaded, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if loaded == nil {
		t.Fatal("LoadConfig returned nil config")
	}

	if loaded.CurrentContext != "test-context" {
		t.Errorf("expected current-context 'test-context', got %q", loaded.CurrentContext)
	}

	if len(loaded.Contexts) != 1 {
		t.Errorf("expected 1 context, got %d", len(loaded.Contexts))
	}

	if len(loaded.Clusters) != 1 {
		t.Errorf("expected 1 cluster, got %d", len(loaded.Clusters))
	}

	if len(loaded.Users) != 1 {
		t.Errorf("expected 1 user, got %d", len(loaded.Users))
	}
}

func TestLoadConfigNonExistent(t *testing.T) {
	// Loading a non-existent config should return nil, nil (not an error)
	cfg, err := LoadConfig("/path/that/does/not/exist")
	if err != nil {
		t.Errorf("expected no error for non-existent config, got: %v", err)
	}
	if cfg != nil {
		t.Errorf("expected nil config for non-existent file, got: %v", cfg)
	}
}

func TestGetCurrentContext(t *testing.T) {
	cfg := &Config{
		CurrentContext: "test-context",
		Contexts: []NamedContext{
			{
				Name: "test-context",
				Context: Context{
					Cluster: "test-cluster",
					User:    "test-user",
				},
			},
		},
		Clusters: []NamedCluster{
			{
				Name: "test-cluster",
				Cluster: Cluster{
					Address:         "https://cortex.example.com",
					TLSCAPath:       "/path/to/ca.crt",
					UseLegacyRoutes: true,
				},
			},
		},
		Users: []NamedUser{
			{
				Name: "test-user",
				User: User{
					ID:        "test-tenant",
					User:      "basic-user",
					Key:       "basic-pass",
					AuthToken: "test-token",
				},
			},
		},
	}

	ctx, err := cfg.GetCurrentContext()
	if err != nil {
		t.Fatalf("GetCurrentContext failed: %v", err)
	}

	if ctx.Address != "https://cortex.example.com" {
		t.Errorf("expected address 'https://cortex.example.com', got %q", ctx.Address)
	}

	if ctx.TLSCAPath != "/path/to/ca.crt" {
		t.Errorf("expected tls-ca-path '/path/to/ca.crt', got %q", ctx.TLSCAPath)
	}

	if ctx.UseLegacyRoutes != true {
		t.Errorf("expected use-legacy-routes true, got %v", ctx.UseLegacyRoutes)
	}

	if ctx.ID != "test-tenant" {
		t.Errorf("expected id 'test-tenant', got %q", ctx.ID)
	}

	if ctx.User != "basic-user" {
		t.Errorf("expected user 'basic-user', got %q", ctx.User)
	}

	if ctx.Key != "basic-pass" {
		t.Errorf("expected key 'basic-pass', got %q", ctx.Key)
	}

	if ctx.AuthToken != "test-token" {
		t.Errorf("expected auth-token 'test-token', got %q", ctx.AuthToken)
	}
}

func TestGetCurrentContextErrors(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "no current context set",
			config: &Config{
				CurrentContext: "",
			},
			wantErr: true,
		},
		{
			name: "context not found",
			config: &Config{
				CurrentContext: "missing-context",
				Contexts:       []NamedContext{},
			},
			wantErr: true,
		},
		{
			name: "cluster not found",
			config: &Config{
				CurrentContext: "test-context",
				Contexts: []NamedContext{
					{
						Name: "test-context",
						Context: Context{
							Cluster: "missing-cluster",
							User:    "test-user",
						},
					},
				},
				Clusters: []NamedCluster{},
			},
			wantErr: true,
		},
		{
			name: "user not found",
			config: &Config{
				CurrentContext: "test-context",
				Contexts: []NamedContext{
					{
						Name: "test-context",
						Context: Context{
							Cluster: "test-cluster",
							User:    "missing-user",
						},
					},
				},
				Clusters: []NamedCluster{
					{Name: "test-cluster", Cluster: Cluster{Address: "http://test"}},
				},
				Users: []NamedUser{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.config.GetCurrentContext()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetCurrentContext() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSetContext(t *testing.T) {
	cfg := &Config{}

	// Add a new context
	cfg.SetContext("new-context", "cluster1", "user1")

	if len(cfg.Contexts) != 1 {
		t.Fatalf("expected 1 context, got %d", len(cfg.Contexts))
	}

	if cfg.Contexts[0].Name != "new-context" {
		t.Errorf("expected name 'new-context', got %q", cfg.Contexts[0].Name)
	}

	if cfg.Contexts[0].Context.Cluster != "cluster1" {
		t.Errorf("expected cluster 'cluster1', got %q", cfg.Contexts[0].Context.Cluster)
	}

	// Update existing context
	cfg.SetContext("new-context", "cluster2", "user2")

	if len(cfg.Contexts) != 1 {
		t.Fatalf("expected 1 context after update, got %d", len(cfg.Contexts))
	}

	if cfg.Contexts[0].Context.Cluster != "cluster2" {
		t.Errorf("expected updated cluster 'cluster2', got %q", cfg.Contexts[0].Context.Cluster)
	}
}

func TestSetCluster(t *testing.T) {
	cfg := &Config{}

	cluster := Cluster{
		Address:     "https://cortex.example.com",
		TLSCAPath:   "/path/to/ca.crt",
		TLSCertPath: "/path/to/client.crt",
		TLSKeyPath:  "/path/to/client.key",
	}

	// Add a new cluster
	cfg.SetCluster("test-cluster", cluster)

	if len(cfg.Clusters) != 1 {
		t.Fatalf("expected 1 cluster, got %d", len(cfg.Clusters))
	}

	if cfg.Clusters[0].Name != "test-cluster" {
		t.Errorf("expected name 'test-cluster', got %q", cfg.Clusters[0].Name)
	}

	if cfg.Clusters[0].Cluster.Address != cluster.Address {
		t.Errorf("expected address %q, got %q", cluster.Address, cfg.Clusters[0].Cluster.Address)
	}

	// Update existing cluster
	updatedCluster := Cluster{
		Address: "https://cortex2.example.com",
	}
	cfg.SetCluster("test-cluster", updatedCluster)

	if len(cfg.Clusters) != 1 {
		t.Fatalf("expected 1 cluster after update, got %d", len(cfg.Clusters))
	}

	if cfg.Clusters[0].Cluster.Address != updatedCluster.Address {
		t.Errorf("expected updated address %q, got %q", updatedCluster.Address, cfg.Clusters[0].Cluster.Address)
	}
}

func TestSetUser(t *testing.T) {
	cfg := &Config{}

	user := User{
		ID:        "tenant-123",
		AuthToken: "token-abc",
	}

	// Add a new user
	cfg.SetUser("test-user", user)

	if len(cfg.Users) != 1 {
		t.Fatalf("expected 1 user, got %d", len(cfg.Users))
	}

	if cfg.Users[0].Name != "test-user" {
		t.Errorf("expected name 'test-user', got %q", cfg.Users[0].Name)
	}

	if cfg.Users[0].User.ID != user.ID {
		t.Errorf("expected id %q, got %q", user.ID, cfg.Users[0].User.ID)
	}

	// Update existing user
	updatedUser := User{
		ID:        "tenant-456",
		AuthToken: "token-xyz",
	}
	cfg.SetUser("test-user", updatedUser)

	if len(cfg.Users) != 1 {
		t.Fatalf("expected 1 user after update, got %d", len(cfg.Users))
	}

	if cfg.Users[0].User.ID != updatedUser.ID {
		t.Errorf("expected updated id %q, got %q", updatedUser.ID, cfg.Users[0].User.ID)
	}
}

func TestDeleteContext(t *testing.T) {
	cfg := &Config{
		Contexts: []NamedContext{
			{Name: "context1", Context: Context{Cluster: "c1", User: "u1"}},
			{Name: "context2", Context: Context{Cluster: "c2", User: "u2"}},
		},
	}

	// Delete existing context
	deleted := cfg.DeleteContext("context1")
	if !deleted {
		t.Error("expected DeleteContext to return true")
	}

	if len(cfg.Contexts) != 1 {
		t.Fatalf("expected 1 context after delete, got %d", len(cfg.Contexts))
	}

	if cfg.Contexts[0].Name != "context2" {
		t.Errorf("expected remaining context 'context2', got %q", cfg.Contexts[0].Name)
	}

	// Delete non-existent context
	deleted = cfg.DeleteContext("non-existent")
	if deleted {
		t.Error("expected DeleteContext to return false for non-existent context")
	}
}

func TestSaveConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test-config")

	cfg := &Config{
		CurrentContext: "test",
		Contexts: []NamedContext{
			{Name: "test", Context: Context{Cluster: "c", User: "u"}},
		},
		Clusters: []NamedCluster{
			{Name: "c", Cluster: Cluster{Address: "http://test"}},
		},
		Users: []NamedUser{
			{Name: "u", User: User{ID: "tenant"}},
		},
	}

	// Save config
	err := SaveConfig(cfg, configPath)
	if err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("config file was not created")
	}

	// Load and verify
	loaded, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("failed to load saved config: %v", err)
	}

	if loaded.CurrentContext != cfg.CurrentContext {
		t.Errorf("expected current-context %q, got %q", cfg.CurrentContext, loaded.CurrentContext)
	}
}

func TestDefaultConfigPath(t *testing.T) {
	path := DefaultConfigPath()
	if path == "" {
		t.Error("DefaultConfigPath returned empty string")
	}

	// Should end with .cortextool/config
	if !filepath.IsAbs(path) {
		t.Error("DefaultConfigPath should return absolute path")
	}
}
