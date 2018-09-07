package engine

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
)

var (
	defaultLogLevel = "debug"
	defaultAPIAddr  = "127.0.0.1:8080"
	defaultCache    = 128
	defaultEthDir   = fmt.Sprintf("%s/eth", defaultDataDir())
	defaultPwdFile  = fmt.Sprintf("%s/pwd.txt", defaultEthDir)
	defaultDbFile   = fmt.Sprintf("%s/chaindata", defaultEthDir)

	defaultBabbleAddr    = "127.0.0.1:1339"
	defaultBabbleAPIAddr = "127.0.0.1:80"
	defaultHeartbeat     = 500
	defaultTCPTimeout    = 1000
	defaultCacheSize     = 50000
	defaultSyncLimit     = 1000
	defaultMaxPool       = 2
	defaultStoreType     = "badger"
	defaultBabbleDir     = fmt.Sprintf("%s/babble", defaultDataDir())
	defaultPeersFile     = fmt.Sprintf("%s/peers.json", defaultBabbleDir)
	defaultStorePath     = fmt.Sprintf("%s/badger_db", defaultBabbleDir)
)

//Config contains de configuration for an EVM-Babble node
type Config struct {
	//Top level options use an anonymous struct
	BaseConfig `mapstructure:",squash"`
	//Options for Babble
	Babble *BabbleConfig `mapstructure:"babble"`
}

//DefaultConfig returns the default configuration for an EVM-Babble node
func DefaultConfig() *Config {
	return &Config{
		BaseConfig: DefaultBaseConfig(),
		Babble:     DefaultBabbleConfig(),
	}
}

/*******************************************************************************
BASE CONFIG
*******************************************************************************/

//BaseConfig contains the top level configuration for an EVM-Babble node
type BaseConfig struct {

	//Top-level directory of evm-babble data
	DataDir string `mapstructure:"datadir"`

	//Directory containing eth config
	EthDir string `mapstructure:"eth_dir"`

	//File containing passwords to unlock ethereum accounts
	PwdFile string `mapstructure:"pwd"`

	//File containing the levelDB database
	DbFile string `mapstructure:"db"`

	//Address of HTTP API Service
	APIAddr string `mapstructure:"api_addr"`

	//Megabytes of memory allocated to internal caching (min 16MB / database forced)
	Cache int `mapstructure:"cache"`

	//Debug, info, warn, error, fatal, panic
	LogLevel string `mapstructure:"log_level"`
}

//DefaultBaseConfig returns the default top-level configuration for EVM-Babble
func DefaultBaseConfig() BaseConfig {
	return BaseConfig{
		DataDir:  defaultDataDir(),
		EthDir:   defaultEthDir,
		PwdFile:  defaultPwdFile,
		DbFile:   defaultDbFile,
		APIAddr:  defaultAPIAddr,
		Cache:    defaultCache,
		LogLevel: defaultLogLevel,
	}
}

/*******************************************************************************
BABBLE CONFIG           XXX this should probably be in Babble itself XXX
*******************************************************************************/

//BabbleConfig contains the configuration of a Babble node
type BabbleConfig struct {
	BabbleDir  string `mapstructure:"babble_dir"`
	ProxyAddr  string `mapstructure:"proxy_addr"`
	BabbleAddr string `mapstructure:"babble_addr"`
	APIAddr    string `mapstructure:"api_addr"`
	PeersFile  string `mapstructure:"peers_file"`
	Heartbeat  int    `mapstructure:"heartbeat"`
	TCPTimeout int    `mapstructure:"tcp_timeout"`
	CacheSize  int    `mapstructure:"cache_size"`
	SyncLimit  int    `mapstructure:"sync_limit"`
	MaxPool    int    `mapstructure:"max_pool"`
	StoreType  string `mapstructure:"store_type"`
	StorePath  string `mapstructure:"store_path"`
}

//DefaultBabbleConfig returns the default configuration for a Babble node
func DefaultBabbleConfig() *BabbleConfig {
	return &BabbleConfig{
		BabbleDir:  defaultBabbleDir,
		BabbleAddr: defaultBabbleAddr,
		PeersFile:  defaultPeersFile,
		Heartbeat:  defaultHeartbeat,
		TCPTimeout: defaultTCPTimeout,
		CacheSize:  defaultCacheSize,
		SyncLimit:  defaultSyncLimit,
		MaxPool:    defaultMaxPool,
		StoreType:  defaultStoreType,
		StorePath:  defaultStorePath,
	}
}

/*******************************************************************************
FILE HELPERS
*******************************************************************************/

func defaultDataDir() string {
	// Try to place the data folder in the user's home dir
	home := homeDir()
	if home != "" {
		if runtime.GOOS == "darwin" {
			return filepath.Join(home, "Library", "BABBLE")
		} else if runtime.GOOS == "windows" {
			return filepath.Join(home, "AppData", "Roaming", "EVMBABBE")
		} else {
			return filepath.Join(home, ".evm-babble")
		}
	}
	// As we cannot guess a stable location, return empty and handle later
	return ""
}

func homeDir() string {
	if home := os.Getenv("HOME"); home != "" {
		return home
	}
	if usr, err := user.Current(); err == nil {
		return usr.HomeDir
	}
	return ""
}
