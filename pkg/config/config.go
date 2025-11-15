package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the cortextool configuration file structure
type Config struct {
	CurrentContext string         `yaml:"current-context"`
	Contexts       []NamedContext `yaml:"contexts"`
	Clusters       []NamedCluster `yaml:"clusters"`
	Users          []NamedUser    `yaml:"users"`
}

// NamedContext associates a name with a context
type NamedContext struct {
	Name    string  `yaml:"name"`
	Context Context `yaml:"context"`
}

// Context references a cluster and user
type Context struct {
	Cluster string `yaml:"cluster"`
	User    string `yaml:"user"`
}

// NamedCluster associates a name with a cluster
type NamedCluster struct {
	Name    string  `yaml:"name"`
	Cluster Cluster `yaml:"cluster"`
}

// Cluster contains cluster connection information
type Cluster struct {
	Address         string `yaml:"address"`
	TLSCAPath       string `yaml:"tls-ca-path,omitempty"`
	TLSCertPath     string `yaml:"tls-cert-path,omitempty"`
	TLSKeyPath      string `yaml:"tls-key-path,omitempty"`
	UseLegacyRoutes bool   `yaml:"use-legacy-routes,omitempty"`
	RulerAPIPath    string `yaml:"ruler-api-path,omitempty"`
}

// NamedUser associates a name with user credentials
type NamedUser struct {
	Name string `yaml:"name"`
	User User   `yaml:"user"`
}

// User contains authentication information
type User struct {
	ID        string `yaml:"id"`
	User      string `yaml:"user,omitempty"`
	Key       string `yaml:"key,omitempty"`
	AuthToken string `yaml:"auth-token,omitempty"`
}

// ContextConfig represents the merged configuration from a specific context
type ContextConfig struct {
	Address         string
	TLSCAPath       string
	TLSCertPath     string
	TLSKeyPath      string
	UseLegacyRoutes bool
	RulerAPIPath    string
	ID              string
	User            string
	Key             string
	AuthToken       string
}

// DefaultConfigPath returns the default config file path
func DefaultConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".cortextool", "config")
}

// LoadConfig loads the configuration from the specified path
func LoadConfig(path string) (*Config, error) {
	if path == "" {
		path = DefaultConfigPath()
	}

	// If file doesn't exist, return empty config (not an error)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

// SaveConfig saves the configuration to the specified path
func SaveConfig(config *Config, path string) error {
	if path == "" {
		path = DefaultConfigPath()
	}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GetCurrentContext returns the merged configuration for the current context
func (c *Config) GetCurrentContext() (*ContextConfig, error) {
	if c == nil {
		return nil, nil
	}

	if c.CurrentContext == "" {
		return nil, fmt.Errorf("no current context set")
	}

	// Find the context
	var ctx *Context
	for _, nc := range c.Contexts {
		if nc.Name == c.CurrentContext {
			ctx = &nc.Context
			break
		}
	}
	if ctx == nil {
		return nil, fmt.Errorf("context %q not found", c.CurrentContext)
	}

	// Find the cluster
	var cluster *Cluster
	for _, nc := range c.Clusters {
		if nc.Name == ctx.Cluster {
			cluster = &nc.Cluster
			break
		}
	}
	if cluster == nil {
		return nil, fmt.Errorf("cluster %q not found", ctx.Cluster)
	}

	// Find the user
	var user *User
	for _, nu := range c.Users {
		if nu.Name == ctx.User {
			user = &nu.User
			break
		}
	}
	if user == nil {
		return nil, fmt.Errorf("user %q not found", ctx.User)
	}

	// Merge into ContextConfig
	return &ContextConfig{
		Address:         cluster.Address,
		TLSCAPath:       cluster.TLSCAPath,
		TLSCertPath:     cluster.TLSCertPath,
		TLSKeyPath:      cluster.TLSKeyPath,
		UseLegacyRoutes: cluster.UseLegacyRoutes,
		RulerAPIPath:    cluster.RulerAPIPath,
		ID:              user.ID,
		User:            user.User,
		Key:             user.Key,
		AuthToken:       user.AuthToken,
	}, nil
}

// GetContext returns a specific context by name
func (c *Config) GetContext(name string) (*ContextConfig, error) {
	if c == nil {
		return nil, nil
	}

	// Find the context
	var ctx *Context
	for _, nc := range c.Contexts {
		if nc.Name == name {
			ctx = &nc.Context
			break
		}
	}
	if ctx == nil {
		return nil, fmt.Errorf("context %q not found", name)
	}

	// Find the cluster
	var cluster *Cluster
	for _, nc := range c.Clusters {
		if nc.Name == ctx.Cluster {
			cluster = &nc.Cluster
			break
		}
	}
	if cluster == nil {
		return nil, fmt.Errorf("cluster %q not found", ctx.Cluster)
	}

	// Find the user
	var user *User
	for _, nu := range c.Users {
		if nu.Name == ctx.User {
			user = &nu.User
			break
		}
	}
	if user == nil {
		return nil, fmt.Errorf("user %q not found", ctx.User)
	}

	// Merge into ContextConfig
	return &ContextConfig{
		Address:         cluster.Address,
		TLSCAPath:       cluster.TLSCAPath,
		TLSCertPath:     cluster.TLSCertPath,
		TLSKeyPath:      cluster.TLSKeyPath,
		UseLegacyRoutes: cluster.UseLegacyRoutes,
		RulerAPIPath:    cluster.RulerAPIPath,
		ID:              user.ID,
		User:            user.User,
		Key:             user.Key,
		AuthToken:       user.AuthToken,
	}, nil
}

// SetContext creates or updates a context
func (c *Config) SetContext(name, cluster, user string) {
	// Check if context exists
	for i, nc := range c.Contexts {
		if nc.Name == name {
			c.Contexts[i].Context.Cluster = cluster
			c.Contexts[i].Context.User = user
			return
		}
	}

	// Add new context
	c.Contexts = append(c.Contexts, NamedContext{
		Name: name,
		Context: Context{
			Cluster: cluster,
			User:    user,
		},
	})
}

// SetCluster creates or updates a cluster
func (c *Config) SetCluster(name string, cluster Cluster) {
	// Check if cluster exists
	for i, nc := range c.Clusters {
		if nc.Name == name {
			c.Clusters[i].Cluster = cluster
			return
		}
	}

	// Add new cluster
	c.Clusters = append(c.Clusters, NamedCluster{
		Name:    name,
		Cluster: cluster,
	})
}

// SetUser creates or updates a user
func (c *Config) SetUser(name string, user User) {
	// Check if user exists
	for i, nu := range c.Users {
		if nu.Name == name {
			c.Users[i].User = user
			return
		}
	}

	// Add new user
	c.Users = append(c.Users, NamedUser{
		Name: name,
		User: user,
	})
}

// DeleteContext removes a context
func (c *Config) DeleteContext(name string) bool {
	for i, nc := range c.Contexts {
		if nc.Name == name {
			c.Contexts = append(c.Contexts[:i], c.Contexts[i+1:]...)
			return true
		}
	}
	return false
}

// DeleteCluster removes a cluster
func (c *Config) DeleteCluster(name string) bool {
	for i, nc := range c.Clusters {
		if nc.Name == name {
			c.Clusters = append(c.Clusters[:i], c.Clusters[i+1:]...)
			return true
		}
	}
	return false
}

// DeleteUser removes a user
func (c *Config) DeleteUser(name string) bool {
	for i, nu := range c.Users {
		if nu.Name == name {
			c.Users = append(c.Users[:i], c.Users[i+1:]...)
			return true
		}
	}
	return false
}
