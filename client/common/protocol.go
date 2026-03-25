package common

import (
	"encoding/binary"
	"fmt"
	"net"
	"strings"
)

const headerSize = 4
const maxChunkSize = 8 * 1024

// sends a batch of bets as a single length-prefixed message over conn: 4-byte big-endian payload size followed by the bet fields joined by '|'
// Sent in chunks of at most 8KB to avoid large writes
// Uses sendAll to avoid short-writes
func SendBatch(conn net.Conn, clientID string, bets []BetInfo) error {
	payload := encodeBets(clientID, bets)

	header := make([]byte, headerSize)
	binary.BigEndian.PutUint32(header, uint32(len(payload)))

	if err := sendAll(conn, header); err != nil {
		return fmt.Errorf("failed to send length header: %w", err)
	}
	if err := sendAll(conn, payload); err != nil {
		return fmt.Errorf("failed to send payload: %w", err)
	}
	return nil
}

// serializes all bets into a flat '|' separated payload
// The server reads clientID once (first field) then groups remaining fields by 5.
func encodeBets(clientID string, bets []BetInfo) []byte {
	fields := []string{clientID}
	for _, bet := range bets {
		fields = append(fields, bet.nombre, bet.apellido, bet.dni, bet.nacimiento, bet.numero)
	}
	return []byte(strings.Join(fields, "|"))
}

// SendEOF signals the server that the client has finished sending all batches.
func SendEOF(conn net.Conn) error {
	data := []byte("EOF")
	header := make([]byte, headerSize)
	binary.BigEndian.PutUint32(header, uint32(len(data)))
	if err := sendAll(conn, header); err != nil {
		return err
	}
	return sendAll(conn, data)
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

// sendAll loops over conn.Write in chunks of at most maxChunkSize, avoiding short-writes.
func sendAll(conn net.Conn, data []byte) error {
	sent := 0
	for sent < len(data) {
		end := sent + maxChunkSize
		if end > len(data) {
			end = len(data)
		}
		n, err := conn.Write(data[sent:end])
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
