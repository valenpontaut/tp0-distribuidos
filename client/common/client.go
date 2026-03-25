package common

import (
	"fmt"
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
	MaxAmount     int
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
	for i := 1; i <= c.config.LoopAmount; i++ {
		conn, err := net.Dial("tcp", c.config.ServerAddress)
		if err == nil {
			c.conn = conn
			return nil
		}
		log.Warningf(
			"action: connect | result: fail | client_id: %v | attempt: %v/%v | error: %v",
			c.config.ID, i, c.config.LoopAmount, err,
		)
		time.Sleep(c.config.LoopPeriod)
	}
	log.Criticalf("action: connect | result: fail | client_id: %v | error: max retries exceeded", c.config.ID)
	return fmt.Errorf("could not connect to server after %d attempts", c.config.LoopAmount)
}

// StartClientLoop reads bets from the agency CSV in batches and sends each batch to the server.
func (c *Client) StartClientLoop() {
	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGTERM)

	reader, err := NewAgencyReader(c.config.ID)
	if err != nil {
		log.Criticalf("action: open_agency_file | result: fail | client_id: %v | error: %v", c.config.ID, err)
		return
	}
	defer reader.Close()

	if err := c.createClientSocket(); err != nil {
		return
	}
	defer c.conn.Close()

	for {
		select {
		case <-sigterm:
			log.Infof("action: shutdown | result: success | client_id: %v", c.config.ID)
			return
		default:
		}

		batch, err := reader.NextBatch(c.config.MaxAmount)
		if err != nil {
			log.Errorf("action: read_batch | result: fail | client_id: %v | error: %v", c.config.ID, err)
			return
		}
		if len(batch) == 0 {
			break
		}

		if err := SendBatch(c.conn, c.config.ID, batch); err != nil {
			log.Errorf("action: send_batch | result: fail | client_id: %v | error: %v", c.config.ID, err)
			return
		}

		confirmation, err := RecvConfirmation(c.conn)
		if err != nil {
			log.Errorf("action: recibir_confirmacion | result: fail | client_id: %v | error: %v", c.config.ID, err)
			return
		}
		if confirmation != "OK" {
			log.Errorf("action: recibir_confirmacion | result: fail | client_id: %v | msg: %v", c.config.ID, confirmation)
			return
		}

		log.Infof("action: apuesta_enviada | result: success | cantidad: %v", len(batch))
	}

	if err := SendEOF(c.conn); err != nil {
		log.Errorf("action: send_eof | result: fail | client_id: %v | error: %v", c.config.ID, err)
		return
	}
	log.Infof("action: loop_finished | result: success | client_id: %v", c.config.ID)
}
