package evmbabble

import (
	"time"

	bproxy "github.com/babbleio/babble/proxy/babble"
	"github.com/Sirupsen/logrus"
)

type Config struct {
	proxyAddr  string //bind address of this app proxy
	babbleAddr string //address of babble node
	apiAddr    string //address of HTTP API service
	ethDir     string //directory containing eth config
	pwdFile    string //file containing password to unlock ethereum accounts
	timeout    time.Duration
}

func NewConfig(proxyAddr, babbleAddr, apiAddr, ethDir, pwdFile string, timeout time.Duration) Config {
	return Config{
		proxyAddr:  proxyAddr,
		babbleAddr: babbleAddr,
		apiAddr:    apiAddr,
		ethDir:     ethDir,
		pwdFile:    pwdFile,
		timeout:    timeout,
	}
}

type Proxy struct {
	service     *Service
	state       *State
	babbleProxy *bproxy.SocketBabbleProxy
	logger      *logrus.Logger
}

func NewProxy(config Config, logger *logrus.Logger) (*Proxy, error) {
	service := NewService(config.ethDir, config.apiAddr, config.pwdFile, logger)
	state, err := NewState(logger)
	if err != nil {
		return nil, err
	}

	babbleProxy, err := bproxy.NewSocketBabbleProxy(config.babbleAddr,
		config.proxyAddr,
		config.timeout)
	if err != nil {
		return nil, err
	}

	return &Proxy{
		service:     service,
		state:       state,
		babbleProxy: babbleProxy,
		logger:      logger,
	}, nil
}

func (p *Proxy) Run() error {
	if err := p.state.Init(p); err != nil {
		return err
	}

	if err := p.service.Init(p); err != nil {
		return err
	}

	go p.service.Run()

	p.Serve()

	return nil
}

func (p *Proxy) Serve() {
	for {
		select {
		case tx := <-p.babbleProxy.CommitCh():
			if err := p.state.AppendTx(tx); err != nil {
				p.logger.WithError(err).Error("AppendTx")
				break
			}
			if err := p.state.Commit(); err != nil {
				p.logger.WithError(err).Error("Commit")
				break
			}
		}
	}
}

func (p *Proxy) SubmitTransaction(tx []byte) error {
	return p.babbleProxy.SubmitTx(tx)
}

func (p *Proxy) GetState() *State {
	return p.state
}
