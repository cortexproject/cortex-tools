package main

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/cortexproject/cortex-tools/pkg/commands"
	"github.com/cortexproject/cortex-tools/pkg/config"
	"github.com/cortexproject/cortex-tools/pkg/version"
)

var (
	ruleCommand           commands.RuleCommand
	alertCommand          commands.AlertCommand
	alertmanagerCommand   commands.AlertmanagerCommand
	logConfig             commands.LoggerConfig
	pushGateway           commands.PushGatewayConfig
	loadgenCommand        commands.LoadgenCommand
	remoteReadCommand     commands.RemoteReadCommand
	aclCommand            commands.AccessControlCommand
	analyseCommand        commands.AnalyseCommand
	bucketValidateCommand commands.BucketValidationCommand
	configCommand         commands.ConfigCommand

	configPath  string
	contextName string
)

func loadConfig(_ *kingpin.ParseContext) error {
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Warnf("Failed to load config file: %v", err)
		return nil // Don't fail if config file has errors
	}
	if cfg != nil {
		log.Debugf("Loaded config from %s", configPath)
	}

	// If --context flag is specified, validate it exists
	if contextName != "" {
		if cfg == nil {
			return fmt.Errorf("--context flag requires a config file, but no config file found at %s", configPath)
		}
		// Validate context exists
		_, err := cfg.GetContext(contextName)
		if err != nil {
			return fmt.Errorf("context %q not found in config file: %w", contextName, err)
		}
		log.Debugf("Using context %q from --context flag", contextName)
	}

	// Set the global config and context override for commands to use
	commands.SetConfig(cfg, contextName)
	return nil
}

func main() {
	app := kingpin.New("cortextool", "A command-line tool to manage cortex.")
	app.Flag("config", "Path to cortextool config file.").
		Default(config.DefaultConfigPath()).
		StringVar(&configPath)
	app.Flag("context", "Name of the context to use from config file.").
		StringVar(&contextName)

	// Load config after flags are parsed but before commands run
	app.PreAction(loadConfig)

	logConfig.Register(app)
	alertCommand.Register(app)
	alertmanagerCommand.Register(app)
	ruleCommand.Register(app)
	pushGateway.Register(app)
	loadgenCommand.Register(app)
	remoteReadCommand.Register(app)
	aclCommand.Register(app)
	analyseCommand.Register(app)
	bucketValidateCommand.Register(app)
	configCommand.Register(app, &configPath)

	app.Command("version", "Get the version of the cortextool CLI").Action(func(_ *kingpin.ParseContext) error {
		fmt.Print(version.Template)
		version.CheckLatest()

		return nil
	})

	kingpin.MustParse(app.Parse(os.Args[1:]))

	pushGateway.Stop()
}
