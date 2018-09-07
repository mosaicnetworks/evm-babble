package commands

import (
	"fmt"

	"github.com/mosaicnetworks/evm-babble/engine"
	"github.com/spf13/cobra"
)

func AddRunFlags(cmd *cobra.Command) {

	//Base Flags
	cmd.Flags().String("datadir", config.BaseConfig.DataDir, "Top-level directory for configuration and data")
	cmd.Flags().String("api_addr", config.BaseConfig.APIAddr, "IP:PORT to bind API server")
	cmd.Flags().String("pwd", config.BaseConfig.PwdFile, "Password file to unlock accounts")
	cmd.Flags().String("db", config.BaseConfig.DbFile, "Eth database file")
	cmd.Flags().Int("cache", config.BaseConfig.Cache, "Megabytes of memory allocated to internal caching (min 16MB / database forced)")

	//Babble Flags
	cmd.Flags().String("babble_addr", config.Babble.BabbleAddr, "IP:PORT of Babble node")
	cmd.Flags().String("proxy_addr", config.Babble.ProxyAddr, "IP:PORT to bind proxy server")
}

// NewRunCmd returns the command that allows the CLI to start a node.
func NewRunCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run the evm-babble node",
		RunE:  run,
	}

	AddRunFlags(cmd)
	return cmd
}

func run(cmd *cobra.Command, args []string) error {

	logger.WithField("config", config).Debug("Config")

	engine, err := engine.NewBabbleSocketEngine(*config, logger)
	if err != nil {
		return fmt.Errorf("Error building Engine: %s", err)
	}

	engine.Run()

	return nil
}
