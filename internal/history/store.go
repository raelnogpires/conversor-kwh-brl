// Package history persists rateio history in CSV files.
package history

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

var header = []string{
	"Data",
	"Consumo 1",
	"Consumo 2",
	"Valor total",
	"Consumo total",
	"Percentual 1",
	"Percentual 2",
	"Valor 1",
	"Valor 2",
}

// Entry is one persisted rateio result.
type Entry struct {
	Date             time.Time
	Consumption1     string
	Consumption2     string
	TotalAmount      string
	TotalConsumption string
	Share1           string
	Share2           string
	Amount1          string
	Amount2          string
}

// Store persists rateio entries at a CSV file path.
type Store struct {
	path string
}

// NewStore creates a Store that uses path as its CSV file.
func NewStore(path string) *Store {
	return &Store{path: path}
}

// Save appends an entry, creating the file, its parent directories, and the
// CSV header when necessary.
func (s *Store) Save(entry Entry) error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return fmt.Errorf("create history directory: %w", err)
	}

	file, err := os.OpenFile(s.path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return fmt.Errorf("open history file: %w", err)
	}

	info, err := file.Stat()
	if err != nil {
		_ = file.Close()
		return fmt.Errorf("inspect history file: %w", err)
	}

	writer := csv.NewWriter(file)
	if info.Size() == 0 {
		if err := writer.Write(header); err != nil {
			_ = file.Close()
			return fmt.Errorf("write history header: %w", err)
		}
	}
	if err := writer.Write(entry.record()); err != nil {
		_ = file.Close()
		return fmt.Errorf("write history entry: %w", err)
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		_ = file.Close()
		return fmt.Errorf("flush history file: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close history file: %w", err)
	}
	return nil
}

// List reads all entries. A missing history file is treated as an empty list.
func (s *Store) List() ([]Entry, error) {
	file, err := os.Open(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return []Entry{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("open history file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	gotHeader, err := reader.Read()
	if errors.Is(err, io.EOF) {
		return nil, errors.New("invalid history header: file is empty")
	}
	if err != nil {
		return nil, fmt.Errorf("read history header: %w", err)
	}
	if !sameFields(gotHeader, header) {
		return nil, fmt.Errorf("invalid history header: got %q", gotHeader)
	}

	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("read history entries: %w", err)
	}

	entries := make([]Entry, 0, len(records))
	for index, record := range records {
		entry, err := entryFromRecord(record)
		if err != nil {
			return nil, fmt.Errorf("invalid history entry on row %d: %w", index+2, err)
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

func (e Entry) record() []string {
	return []string{
		e.Date.Format(time.RFC3339),
		e.Consumption1,
		e.Consumption2,
		e.TotalAmount,
		e.TotalConsumption,
		e.Share1,
		e.Share2,
		e.Amount1,
		e.Amount2,
	}
}

func entryFromRecord(record []string) (Entry, error) {
	if len(record) != len(header) {
		return Entry{}, fmt.Errorf("got %d fields, want %d", len(record), len(header))
	}
	date, err := time.Parse(time.RFC3339, record[0])
	if err != nil {
		return Entry{}, fmt.Errorf("invalid date %q: %w", record[0], err)
	}
	return Entry{
		Date:             date,
		Consumption1:     record[1],
		Consumption2:     record[2],
		TotalAmount:      record[3],
		TotalConsumption: record[4],
		Share1:           record[5],
		Share2:           record[6],
		Amount1:          record[7],
		Amount2:          record[8],
	}, nil
}

func sameFields(first, second []string) bool {
	if len(first) != len(second) {
		return false
	}
	for index := range first {
		if first[index] != second[index] {
			return false
		}
	}
	return true
}
