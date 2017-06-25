package evmbabble

import (
	"encoding/json"
	"io/ioutil"
	"math/big"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/Sirupsen/logrus"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/gorilla/mux"
)

var defaultGas = big.NewInt(90000)

type Service struct {
	sync.Mutex
	proxy    *Proxy
	dataDir  string
	apiAddr  string
	keyStore *keystore.KeyStore
	pwdFile  string
	logger   *logrus.Logger
}

func NewService(dataDir, apiAddr, pwdFile string, logger *logrus.Logger) *Service {
	return &Service{
		dataDir: dataDir,
		apiAddr: apiAddr,
		pwdFile: pwdFile,
		logger:  logger}
}

func (m *Service) Init(proxy *Proxy) error {
	m.proxy = proxy
	return nil
}

func (m *Service) Run() {
	m.checkErr(m.makeKeyStore())

	m.checkErr(m.unlockAccounts())

	m.checkErr(m.createGenesisAccounts())

	m.logger.Info("serving api...")
	m.serveAPI()
}

func (m *Service) makeKeyStore() error {

	scryptN := keystore.StandardScryptN
	scryptP := keystore.StandardScryptP

	keydir := filepath.Join(m.dataDir, "keystore")
	if err := os.MkdirAll(keydir, 0700); err != nil {
		return err
	}

	m.keyStore = keystore.NewKeyStore(keydir, scryptN, scryptP)

	return nil
}

func (m *Service) unlockAccounts() error {
	pwd, err := m.readPwd()
	if err != nil {
		m.logger.WithError(err).Error("Reading PwdFile")
		return err
	}

	for _, ac := range m.keyStore.Accounts() {
		if err := m.keyStore.Unlock(ac, string(pwd)); err != nil {
			return err
		}
		m.logger.WithField("address", ac.Address.Hex()).Debug("Unlocked account")
	}
	return nil
}

func (m *Service) createGenesisAccounts() error {
	genesisFile := filepath.Join(m.dataDir, "genesis.json")

	contents, err := ioutil.ReadFile(genesisFile)
	if err != nil {
		return err
	}

	var genesis struct {
		Alloc AccountMap
	}

	if err := json.Unmarshal(contents, &genesis); err != nil {
		return err
	}
	state, err := m.getState()
	if err != nil {
		return err
	}
	if err := state.CreateAccounts(genesis.Alloc); err != nil {
		return err
	}
	return nil
}

func (m *Service) getState() (*State, error) {
	return m.proxy.GetState(), nil
}

func (m *Service) serveAPI() {
	router := mux.NewRouter()
	router.HandleFunc("/accounts", m.makeHandler(accountsHandler)).Methods("GET")
	router.HandleFunc("/tx", m.makeHandler(transactionHandler)).Methods("POST")
	router.HandleFunc("/tx/{tx_hash}", m.makeHandler(transactionReceiptHandler)).Methods("GET")
	http.ListenAndServe(m.apiAddr, router)
}

func (m *Service) makeHandler(fn func(http.ResponseWriter, *http.Request, *Service)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m.Lock()
		fn(w, r, m)
		m.Unlock()
	}
}

func (m *Service) checkErr(err error) {
	if err != nil {
		m.logger.WithError(err).Error("ERROR")
		os.Exit(1)
	}
}

func (m *Service) readPwd() (pwd string, err error) {
	text, err := ioutil.ReadFile(m.pwdFile)
	if err != nil {
		return "", err
	}
	lines := strings.Split(string(text), "\n")
	// Sanitise DOS line endings.
	for i := range lines {
		lines[i] = strings.TrimRight(lines[i], "\r")
	}
	return lines[0], nil
}
