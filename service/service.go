package service

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
	"github.com/babbleio/evm-babble/common"
	"github.com/babbleio/evm-babble/state"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/gorilla/mux"
)

var defaultGas = big.NewInt(90000)

type Service struct {
	sync.Mutex
	state    *state.State
	submitCh chan []byte
	dataDir  string
	apiAddr  string
	keyStore *keystore.KeyStore
	pwdFile  string
	logger   *logrus.Logger
}

func NewService(dataDir, apiAddr, pwdFile string,
	state *state.State,
	submitCh chan []byte,
	logger *logrus.Logger) *Service {
	return &Service{
		dataDir:  dataDir,
		apiAddr:  apiAddr,
		pwdFile:  pwdFile,
		state:    state,
		submitCh: submitCh,
		logger:   logger}
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

	if len(m.keyStore.Accounts()) == 0 {
		return nil
	}

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

	if _, err := os.Stat(genesisFile); os.IsNotExist(err) {
		return nil
	}

	contents, err := ioutil.ReadFile(genesisFile)
	if err != nil {
		return err
	}

	var genesis struct {
		Alloc common.AccountMap
	}

	if err := json.Unmarshal(contents, &genesis); err != nil {
		return err
	}

	if err := m.state.CreateAccounts(genesis.Alloc); err != nil {
		return err
	}
	return nil
}

func (m *Service) serveAPI() {
	r := mux.NewRouter()
	r.HandleFunc("/account/{address}", m.makeHandler(accountHandler)).Methods("GET")
	r.HandleFunc("/accounts", m.makeHandler(accountsHandler)).Methods("GET")
	r.HandleFunc("/call", m.makeHandler(callHandler)).Methods("POST")
	r.HandleFunc("/tx", m.makeHandler(transactionHandler)).Methods("POST")
	r.HandleFunc("/rawtx", m.makeHandler(rawTransactionHandler)).Methods("POST")
	r.HandleFunc("/tx/{tx_hash}", m.makeHandler(transactionReceiptHandler)).Methods("GET")
	http.Handle("/", &CORSServer{r})
	http.ListenAndServe(m.apiAddr, nil)
}

type CORSServer struct {
	r *mux.Router
}

func (s *CORSServer) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if origin := req.Header.Get("Origin"); origin != "" {
		rw.Header().Set("Access-Control-Allow-Origin", origin)
		rw.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		rw.Header().Set("Access-Control-Allow-Headers",
			"Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	}
	// Stop here if its Preflighted OPTIONS request
	if req.Method == "OPTIONS" {
		return
	}
	// Lets Gorilla work
	s.r.ServeHTTP(rw, req)
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
