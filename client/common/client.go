package common

import (
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("log")

// ClientConfig Configuration used by the client
type ClientConfig struct {
	ID            string
	ServerAddress string
	LoopAmount    int
	LoopPeriod    time.Duration
}

// Client Entity that encapsulates how
type Client struct {
	config ClientConfig
	conn   net.Conn
}

// NewClient Initializes a new client receiving the configuration
// as a parameter
func NewClient(config ClientConfig) *Client {
	client := &Client{
		config: config,
	}
	return client
}

// CreateClientSocket Initializes client socket. In case of
// failure, error is printed in stdout/stderr and exit 1
// is returned
func (c *Client) createClientSocket() error {
	conn, err := net.Dial("tcp", c.config.ServerAddress)
	if err != nil {
		log.Criticalf(
			"action: connect | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
		return err
	}
	c.conn = conn
	return nil
}

// StartClientLoop Send messages to the client until some time threshold is met
func (c *Client) StartClientLoop() {
	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGTERM)

	bet, err := NewBetFromEnv()
	if err != nil {
		log.Criticalf("action: apuesta_leida | result: fail | client_id: %v | error: %v", c.config.ID, err)
		return
	}

	select {
	case <-sigterm:
		log.Infof("action: shutdown | result: success | client_id: %v", c.config.ID)
		return
	default:
	}

	if err := c.createClientSocket(); err != nil {
		return
	}
	defer c.conn.Close()

	if err := SendBet(c.conn, c.config.ID, bet); err != nil {
		log.Errorf("action: apuesta_enviada | result: fail | client_id: %v | error: %v", c.config.ID, err)
		return
	}

	if _, err := RecvConfirmation(c.conn); err != nil {
		log.Errorf("action: recibir_confirmacion | result: fail | client_id: %v | error: %v", c.config.ID, err)
		return
	}

	log.Infof("action: apuesta_enviada | result: success | dni: %v | numero: %v", bet.dni, bet.numero)
}
