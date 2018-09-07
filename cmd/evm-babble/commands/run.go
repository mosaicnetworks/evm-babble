package commands

import (
	"fmt"

	"github.com/mosaicnetworks/evm-babble/engine"
	"github.com/spf13/cobra"
)

//AddRunFlags adds flags to the Run command
func AddRunFlags(cmd *cobra.Command) {

	//Base
	cmd.Flags().String("datadir", config.BaseConfig.DataDir, "Top-level directory for configuration and data")
	cmd.Flags().String("log_level", config.BaseConfig.LogLevel, "debug, info, warn, error, fatal, panic")

	//Eth
	cmd.Flags().String("eth.genesis", config.Eth.Genesis, "Location of genesis file")
	cmd.Flags().String("eth.keystore", config.Eth.Keystore, "Location of Ethereum account keys")
	cmd.Flags().String("eth.pwd", config.Eth.PwdFile, "Password file to unlock accounts")
	cmd.Flags().String("eth.db", config.Eth.DbFile, "Eth database file")
	cmd.Flags().String("eth.api_addr", config.Eth.EthAPIAddr, "Address of HTTP API service")
	cmd.Flags().Int("eth.cache", config.Eth.Cache, "Megabytes of memory allocated to internal caching (min 16MB / database forced)")

	//Babble Socket
	cmd.Flags().String("babble.proxy_addr", config.Babble.ProxyAddr, "IP:PORT of Babble proxy")
	cmd.Flags().String("babble.client_addr", config.Babble.ClientAddr, "IP:PORT to bind client proxy")

	//Babble Inmem
	cmd.Flags().String("babble.dir", config.Babble.BabbleDir, "Directory contaning priv_key.pem and peers.json files")
	cmd.Flags().String("babble.node_addr", config.Babble.NodeAddr, "IP:PORT of Babble node")
	cmd.Flags().String("babble.api_addr", config.Babble.BabbleAPIAddr, "IP:PORT of Babble HTTP API service")
	cmd.Flags().Int("babble.heartbeat", config.Babble.Heartbeat, "Heartbeat time milliseconds (time between gossips)")
	cmd.Flags().Int("babble.tcp_timeout", config.Babble.TCPTimeout, "TCP timeout milliseconds")
	cmd.Flags().Int("babble.cache_size", config.Babble.CacheSize, "Number of items in LRU caches")
	cmd.Flags().Int("babble.sync_limit", config.Babble.SyncLimit, "Max number of Events per sync")
	cmd.Flags().Int("babble.max_pool", config.Babble.MaxPool, "Max number of pool connections")
	cmd.Flags().String("babble.store_type", config.Babble.StoreType, "badger,inmem")
	cmd.Flags().String("babble.store_path", config.Babble.StorePath, "File containing the store database")
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

	// engine, err := engine.NewBabbleSocketEngine(*config, logger)
	engine, err := engine.NewBabbleInmemEngine(*config, logger)
	if err != nil {
		return fmt.Errorf("Error building Engine: %s", err)
	}

	engine.Run()

	return nil
}
