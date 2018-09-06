package engine

import "time"

type Config struct {
	proxyAddr    string //bind address of this app proxy
	babbleAddr   string //address of babble node
	apiAddr      string //address of HTTP API service
	ethDir       string //directory containing eth config
	pwdFile      string //file containing password to unlock ethereum accounts
	databaseFile string //file containing LevelDB database
	cache        int    //Megabytes of memory allocated to internal caching (min 16MB / database forced)
	timeout      time.Duration

	//babble_inmem
	privKey    string //private key in PEM format (actual string, not file name)
	peers      string //peers.json (actual json string, not file name)
	heartbeat  int
	tcpTimeout int
	cacheSize  int
	syncLimit  int
	storeType  string
	storePath  string
}

func NewConfig(proxyAddr,
	babbleAddr,
	apiAddr,
	ethDir,
	pwdFile,
	dbFile string,
	cache int,
	timeout time.Duration) Config {

	return Config{
		proxyAddr:    proxyAddr,
		babbleAddr:   babbleAddr,
		apiAddr:      apiAddr,
		ethDir:       ethDir,
		pwdFile:      pwdFile,
		databaseFile: dbFile,
		cache:        cache,
		timeout:      timeout,
	}
}

func NewConfig2(proxyAddr,
	babbleAddr,
	apiAddr,
	ethDir,
	pwdFile,
	dbFile string,
	cache int,
	timeout time.Duration,
	privKey string,
	peers string,
	heartbeat int,
	tcpTimeout int,
	cacheSize int,
	syncLimit int,
	storeType string,
	storePath string) Config {

	return Config{
		proxyAddr:    proxyAddr,
		babbleAddr:   babbleAddr,
		apiAddr:      apiAddr,
		ethDir:       ethDir,
		pwdFile:      pwdFile,
		databaseFile: dbFile,
		cache:        cache,
		timeout:      timeout,
		privKey:      privKey,
		peers:        peers,
		heartbeat:    heartbeat,
		tcpTimeout:   tcpTimeout,
		cacheSize:    cacheSize,
		syncLimit:    syncLimit,
		storeType:    storeType,
		storePath:    storePath,
	}
}
