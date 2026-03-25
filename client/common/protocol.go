package common

import (
	"encoding/binary"
	"fmt"
	"net"
	"strings"
)

const headerSize = 4

// sends a length-prefixed message over conn: 4-byte big-endian payload size followed by the bet fields joined by '|'
// Uses sendAll to avoid short-writes
func SendBet(conn net.Conn, clientID string, bet BetInfo) error {
	payload := strings.Join([]string{
		clientID,
		bet.nombre,
		bet.apellido,
		bet.dni,
		bet.nacimiento,
		bet.numero,
	}, "|")

	data := []byte(payload)
	header := make([]byte, headerSize)
	binary.BigEndian.PutUint32(header, uint32(len(data)))

	if err := sendAll(conn, header); err != nil {
		return fmt.Errorf("failed to send length header: %w", err)
	}
	if err := sendAll(conn, data); err != nil {
		return fmt.Errorf("failed to send payload: %w", err)
	}
	return nil
}

// reads a length-prefixed response from the server
// Uses recvAll to avoid short-reads
func RecvConfirmation(conn net.Conn) (string, error) {
	header := make([]byte, headerSize)
	if err := recvAll(conn, header); err != nil {
		return "", fmt.Errorf("failed to read length header: %w", err)
	}

	length := binary.BigEndian.Uint32(header)
	buf := make([]byte, length)
	if err := recvAll(conn, buf); err != nil {
		return "", fmt.Errorf("failed to read confirmation: %w", err)
	}
	return string(buf), nil
}

// sendAll loops over conn.Write until all bytes are sent, avoiding short-writes.
func sendAll(conn net.Conn, data []byte) error {
	sent := 0
	for sent < len(data) {
		n, err := conn.Write(data[sent:])
		if err != nil {
			return err
		}
		sent += n
	}
	return nil
}

// recvAll loops over conn.Read until buf is completely filled, avoiding short-reads.
func recvAll(conn net.Conn, buf []byte) error {
	received := 0
	for received < len(buf) {
		n, err := conn.Read(buf[received:])
		if err != nil {
			return err
		}
		received += n
	}
	return nil
}
