package commands

import (
	"path/filepath"

	"github.com/mosaicnetworks/evm-babble/engine"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	config = engine.DefaultConfig()
	logger = logrus.New()
)

func logLevel(l string) logrus.Level {
	switch l {
	case "debug":
		return logrus.DebugLevel
	case "info":
		return logrus.InfoLevel
	case "warn":
		return logrus.WarnLevel
	case "error":
		return logrus.ErrorLevel
	case "fatal":
		return logrus.FatalLevel
	case "panic":
		return logrus.PanicLevel
	default:
		return logrus.DebugLevel
	}
}

// ParseConfig retrieves the default environment configuration,
// sets up the Tendermint root and ensures that the root exists
func ParseConfig() (*engine.Config, error) {
	conf := engine.DefaultConfig()
	err := viper.Unmarshal(conf)
	if err != nil {
		return nil, err
	}
	return conf, err
}

//RootCmd is the root command for evm-babble
var RootCmd = &cobra.Command{
	Use:   "evm-babble",
	Short: "LightWeight EVM app for Babble",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) (err error) {
		if cmd.Name() == VersionCmd.Name() {
			return nil
		}

		if err := bindFlagsLoadViper(cmd, args); err != nil {
			return err
		}

		config, err = ParseConfig()
		if err != nil {
			return err
		}

		logger = logrus.New()
		logger.Level = logLevel(config.BaseConfig.LogLevel)

		logger.WithFields(logrus.Fields{
			"Base":   config.BaseConfig,
			"Eth":    config.Eth,
			"Babble": config.Babble}).Debug("Config")

		return nil
	},
}

// Bind all flags and read the config into viper
func bindFlagsLoadViper(cmd *cobra.Command, args []string) error {
	// cmd.Flags() includes flags from this command and all persistent flags from the parent
	if err := viper.BindPFlags(cmd.Flags()); err != nil {
		return err
	}

	viper.SetConfigName("config")                                           // name of config file (without extension)
	viper.AddConfigPath(config.BaseConfig.DataDir)                          // search root directory
	viper.AddConfigPath(filepath.Join(config.BaseConfig.DataDir, "config")) // search root directory /config

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		// stderr, so if we redirect output to json file, this doesn't appear
		// fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	} else if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
		// ignore not found error, return other errors
		return err
	}
	return nil
}
