package commands

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
	"gopkg.in/yaml.v3"

	"github.com/cortexproject/cortex-tools/pkg/config"
)

// ConfigCommand handles config file management
type ConfigCommand struct {
	configPathPtr *string // pointer to the global config path

	// set-context flags
	contextName string
	clusterName string
	userName    string

	// set-cluster flags
	address         string
	tlsCAPath       string
	tlsCertPath     string
	tlsKeyPath      string
	useLegacyRoutes bool
	rulerAPIPath    string

	// set-credentials flags
	id        string
	user      string
	key       string
	authToken string

	// use-context flags
	useContextName string
}

// Register config command and its subcommands
func (c *ConfigCommand) Register(app *kingpin.Application, configPath *string) {
	c.configPathPtr = configPath
	configCmd := app.Command("config", "Manage cortextool configuration.")

	// view command
	configCmd.Command("view", "Display merged configuration.").
		Action(c.viewConfig)

	// get-contexts command
	configCmd.Command("get-contexts", "List all contexts.").
		Action(c.getContexts)

	// current-context command
	configCmd.Command("current-context", "Display the current context.").
		Action(c.currentContext)

	// use-context command
	useContextCmd := configCmd.Command("use-context", "Set the current context.").
		Action(c.useContext)
	useContextCmd.Arg("name", "Context name to use.").
		Required().
		StringVar(&c.useContextName)

	// set-context command
	setContextCmd := configCmd.Command("set-context", "Create or modify a context.").
		Action(c.setContext)
	setContextCmd.Arg("name", "Context name.").
		Required().
		StringVar(&c.contextName)
	setContextCmd.Flag("cluster", "Cluster name for the context.").
		StringVar(&c.clusterName)
	setContextCmd.Flag("user", "User name for the context.").
		StringVar(&c.userName)

	// set-cluster command
	setClusterCmd := configCmd.Command("set-cluster", "Create or modify a cluster.").
		Action(c.setCluster)
	setClusterCmd.Arg("name", "Cluster name.").
		Required().
		StringVar(&c.clusterName)
	setClusterCmd.Flag("address", "Cortex cluster address.").
		StringVar(&c.address)
	setClusterCmd.Flag("tls-ca-path", "TLS CA certificate path.").
		StringVar(&c.tlsCAPath)
	setClusterCmd.Flag("tls-cert-path", "TLS client certificate path.").
		StringVar(&c.tlsCertPath)
	setClusterCmd.Flag("tls-key-path", "TLS client key path.").
		StringVar(&c.tlsKeyPath)
	setClusterCmd.Flag("use-legacy-routes", "Use legacy API routes.").
		BoolVar(&c.useLegacyRoutes)
	setClusterCmd.Flag("ruler-api-path", "Custom ruler API path.").
		StringVar(&c.rulerAPIPath)

	// set-credentials command
	setCredsCmd := configCmd.Command("set-credentials", "Create or modify user credentials.").
		Action(c.setCredentials)
	setCredsCmd.Arg("name", "User name.").
		Required().
		StringVar(&c.userName)
	setCredsCmd.Flag("id", "Tenant ID.").
		StringVar(&c.id)
	setCredsCmd.Flag("user", "Basic auth username.").
		StringVar(&c.user)
	setCredsCmd.Flag("key", "Basic auth password/key.").
		StringVar(&c.key)
	setCredsCmd.Flag("auth-token", "Bearer token for JWT auth.").
		StringVar(&c.authToken)

	// delete-context command
	deleteContextCmd := configCmd.Command("delete-context", "Delete a context.").
		Action(c.deleteContext)
	deleteContextCmd.Arg("name", "Context name to delete.").
		Required().
		StringVar(&c.contextName)
}

func (c *ConfigCommand) loadOrCreateConfig() (*config.Config, error) {
	cfg, err := config.LoadConfig(*c.configPathPtr)
	if err != nil {
		return nil, err
	}
	if cfg == nil {
		cfg = &config.Config{}
	}
	return cfg, nil
}

func (c *ConfigCommand) viewConfig(_ *kingpin.ParseContext) error {
	cfg, err := config.LoadConfig(*c.configPathPtr)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	if cfg == nil {
		fmt.Println("No configuration file found.")
		return nil
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	fmt.Print(string(data))
	return nil
}

func (c *ConfigCommand) getContexts(_ *kingpin.ParseContext) error {
	cfg, err := config.LoadConfig(*c.configPathPtr)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	if cfg == nil {
		fmt.Println("No configuration file found.")
		return nil
	}

	if len(cfg.Contexts) == 0 {
		fmt.Println("No contexts found.")
		return nil
	}

	fmt.Printf("CURRENT   NAME\n")
	for _, ctx := range cfg.Contexts {
		current := ""
		if ctx.Name == cfg.CurrentContext {
			current = "*"
		}
		fmt.Printf("%-9s %s\n", current, ctx.Name)
	}
	return nil
}

func (c *ConfigCommand) currentContext(_ *kingpin.ParseContext) error {
	cfg, err := config.LoadConfig(*c.configPathPtr)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	if cfg == nil {
		fmt.Println("No configuration file found.")
		return nil
	}

	if cfg.CurrentContext == "" {
		fmt.Println("No current context set.")
		return nil
	}

	fmt.Println(cfg.CurrentContext)
	return nil
}

func (c *ConfigCommand) useContext(_ *kingpin.ParseContext) error {
	cfg, err := c.loadOrCreateConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Check if context exists
	found := false
	for _, ctx := range cfg.Contexts {
		if ctx.Name == c.useContextName {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("context %q not found", c.useContextName)
	}

	cfg.CurrentContext = c.useContextName

	if err := config.SaveConfig(cfg, *c.configPathPtr); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	log.Infof("Switched to context %q", c.useContextName)
	return nil
}

func (c *ConfigCommand) setContext(_ *kingpin.ParseContext) error {
	cfg, err := c.loadOrCreateConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// If cluster or user not specified, check if context already exists and keep existing values
	existing := false
	for _, ctx := range cfg.Contexts {
		if ctx.Name == c.contextName {
			existing = true
			if c.clusterName == "" {
				c.clusterName = ctx.Context.Cluster
			}
			if c.userName == "" {
				c.userName = ctx.Context.User
			}
			break
		}
	}

	if !existing {
		if c.clusterName == "" || c.userName == "" {
			return fmt.Errorf("--cluster and --user are required when creating a new context")
		}
	}

	cfg.SetContext(c.contextName, c.clusterName, c.userName)

	if err := config.SaveConfig(cfg, *c.configPathPtr); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	log.Infof("Context %q set", c.contextName)
	return nil
}

func (c *ConfigCommand) setCluster(_ *kingpin.ParseContext) error {
	cfg, err := c.loadOrCreateConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Check if cluster exists and merge with existing values
	cluster := config.Cluster{
		Address:         c.address,
		TLSCAPath:       c.tlsCAPath,
		TLSCertPath:     c.tlsCertPath,
		TLSKeyPath:      c.tlsKeyPath,
		UseLegacyRoutes: c.useLegacyRoutes,
		RulerAPIPath:    c.rulerAPIPath,
	}

	// If cluster exists, merge with existing values (don't overwrite with empty strings)
	for _, existing := range cfg.Clusters {
		if existing.Name == c.clusterName {
			if cluster.Address == "" {
				cluster.Address = existing.Cluster.Address
			}
			if cluster.TLSCAPath == "" {
				cluster.TLSCAPath = existing.Cluster.TLSCAPath
			}
			if cluster.TLSCertPath == "" {
				cluster.TLSCertPath = existing.Cluster.TLSCertPath
			}
			if cluster.TLSKeyPath == "" {
				cluster.TLSKeyPath = existing.Cluster.TLSKeyPath
			}
			if cluster.RulerAPIPath == "" {
				cluster.RulerAPIPath = existing.Cluster.RulerAPIPath
			}
			// For boolean, we keep the new value (could be explicitly set to false)
			break
		}
	}

	cfg.SetCluster(c.clusterName, cluster)

	if err := config.SaveConfig(cfg, *c.configPathPtr); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	log.Infof("Cluster %q set", c.clusterName)
	return nil
}

func (c *ConfigCommand) setCredentials(_ *kingpin.ParseContext) error {
	cfg, err := c.loadOrCreateConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	user := config.User{
		ID:        c.id,
		User:      c.user,
		Key:       c.key,
		AuthToken: c.authToken,
	}

	// If user exists, merge with existing values (don't overwrite with empty strings)
	for _, existing := range cfg.Users {
		if existing.Name == c.userName {
			if user.ID == "" {
				user.ID = existing.User.ID
			}
			if user.User == "" {
				user.User = existing.User.User
			}
			if user.Key == "" {
				user.Key = existing.User.Key
			}
			if user.AuthToken == "" {
				user.AuthToken = existing.User.AuthToken
			}
			break
		}
	}

	cfg.SetUser(c.userName, user)

	if err := config.SaveConfig(cfg, *c.configPathPtr); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	log.Infof("Credentials for user %q set", c.userName)
	return nil
}

func (c *ConfigCommand) deleteContext(_ *kingpin.ParseContext) error {
	cfg, err := c.loadOrCreateConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if !cfg.DeleteContext(c.contextName) {
		return fmt.Errorf("context %q not found", c.contextName)
	}

	// If we deleted the current context, clear it
	if cfg.CurrentContext == c.contextName {
		cfg.CurrentContext = ""
	}

	if err := config.SaveConfig(cfg, *c.configPathPtr); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	log.Infof("Context %q deleted", c.contextName)
	return nil
}
