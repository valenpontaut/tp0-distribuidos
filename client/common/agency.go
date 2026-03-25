package common

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
)

type AgencyReader struct {
	reader *csv.Reader
	file   *os.File
}

// opens the CSV file for the given agency ID
func NewAgencyReader(agencyID string) (*AgencyReader, error) {
	path := fmt.Sprintf("/data/agency-%s.csv", agencyID)
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("could not open agency file: %w", err)
	}
	return &AgencyReader{
		reader: csv.NewReader(f),
		file:   f,
	}, nil
}

// reads up to maxAmount bets from the CSV
func (r *AgencyReader) NextBatch(maxAmount int) ([]BetInfo, error) {
	var bets []BetInfo
	for len(bets) < maxAmount {
		record, err := r.reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error reading CSV: %w", err)
		}
		if len(record) != 5 {
			log.Warningf("action: read_csv | result: skip | reason: invalid field count | row: %v", record)
			continue
		}
		bet, err := NewBetInfo(record[0], record[1], record[2], record[3], record[4])
		if err != nil {
			log.Warningf("action: read_csv | result: skip | reason: %v", err)
			continue
		}
		bets = append(bets, bet)
	}
	return bets, nil
}

// Close closes the underlying CSV file.
func (r *AgencyReader) Close() {
	r.file.Close()
}