package commands

import (
	"github.com/cortexproject/cortex-tools/pkg/client"
	"github.com/cortexproject/cortex-tools/pkg/config"
)

var (
	globalConfig          *config.Config
	globalContextOverride string
)

// SetConfig sets the global configuration and optional context override for all commands to use
func SetConfig(cfg *config.Config, contextName string) {
	globalConfig = cfg
	globalContextOverride = contextName
}

// GetConfig returns the global configuration
func GetConfig() *config.Config {
	return globalConfig
}

// ApplyConfigDefaults applies config file defaults to a client.Config
// Config file values are used as defaults if the field is empty
// Precedence: CLI flags/Env vars > --context flag > current-context > Defaults
func ApplyConfigDefaults(clientCfg *client.Config) error {
	if globalConfig == nil {
		return nil
	}

	// Get the context to use (from --context flag or current-context)
	var contextCfg *config.ContextConfig
	var err error

	if globalContextOverride != "" {
		// Use context from --context flag
		contextCfg, err = globalConfig.GetContext(globalContextOverride)
	} else {
		// Use current-context from config file
		contextCfg, err = globalConfig.GetCurrentContext()
	}

	if err != nil {
		// If there's no context or it's invalid, skip applying defaults
		return nil
	}

	// Apply config defaults only if the field is empty (not set by flag or env var)
	if clientCfg.Address == "" && contextCfg.Address != "" {
		clientCfg.Address = contextCfg.Address
	}

	if clientCfg.ID == "" && contextCfg.ID != "" {
		clientCfg.ID = contextCfg.ID
	}

	if clientCfg.User == "" && contextCfg.User != "" {
		clientCfg.User = contextCfg.User
	}

	if clientCfg.Key == "" && contextCfg.Key != "" {
		clientCfg.Key = contextCfg.Key
	}

	if clientCfg.AuthToken == "" && contextCfg.AuthToken != "" {
		clientCfg.AuthToken = contextCfg.AuthToken
	}

	if clientCfg.TLS.CAPath == "" && contextCfg.TLSCAPath != "" {
		clientCfg.TLS.CAPath = contextCfg.TLSCAPath
	}

	if clientCfg.TLS.CertPath == "" && contextCfg.TLSCertPath != "" {
		clientCfg.TLS.CertPath = contextCfg.TLSCertPath
	}

	if clientCfg.TLS.KeyPath == "" && contextCfg.TLSKeyPath != "" {
		clientCfg.TLS.KeyPath = contextCfg.TLSKeyPath
	}

	if clientCfg.RulerAPIPath == "" && contextCfg.RulerAPIPath != "" {
		clientCfg.RulerAPIPath = contextCfg.RulerAPIPath
	}

	// UseLegacyRoutes is a boolean, so we can't check for "empty"
	// We'll only apply the config value if it's true
	if !clientCfg.UseLegacyRoutes && contextCfg.UseLegacyRoutes {
		clientCfg.UseLegacyRoutes = contextCfg.UseLegacyRoutes
	}

	return nil
}
