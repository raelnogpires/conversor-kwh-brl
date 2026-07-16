package history

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func testEntry(date time.Time) Entry {
	return Entry{
		Date:             date,
		Consumption1:     "105,5 kWh",
		Consumption2:     "67,2 kWh",
		TotalAmount:      "R$ 184,72",
		TotalConsumption: "172,7 kWh",
		Share1:           "61,09%",
		Share2:           "38,91%",
		Amount1:          "R$ 112,84",
		Amount2:          "R$ 71,88",
	}
}

func TestSaveCreatesParentsHeaderAndFirstEntry(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "history.csv")
	entry := testEntry(time.Date(2026, time.July, 15, 10, 30, 0, 0, time.UTC))

	if err := NewStore(path).Save(entry); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer file.Close()
	records, err := csv.NewReader(file).ReadAll()
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("record count = %d, want 2", len(records))
	}
	if !reflect.DeepEqual(records[0], header) {
		t.Errorf("header = %q, want %q", records[0], header)
	}
	if !reflect.DeepEqual(records[1], entry.record()) {
		t.Errorf("entry = %q, want %q", records[1], entry.record())
	}
}

func TestSaveAppendsEntries(t *testing.T) {
	path := filepath.Join(t.TempDir(), "history.csv")
	store := NewStore(path)
	first := testEntry(time.Date(2026, time.July, 14, 8, 0, 0, 0, time.UTC))
	second := testEntry(time.Date(2026, time.July, 15, 9, 0, 0, 0, time.FixedZone("BRT", -3*60*60)))
	second.Consumption1 = "90 kWh"

	if err := store.Save(first); err != nil {
		t.Fatalf("first Save() error = %v", err)
	}
	if err := store.Save(second); err != nil {
		t.Fatalf("second Save() error = %v", err)
	}

	entries, err := store.List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("entry count = %d, want 2", len(entries))
	}
	if !entries[0].Date.Equal(first.Date) || !entries[1].Date.Equal(second.Date) {
		t.Errorf("dates = %v and %v", entries[0].Date, entries[1].Date)
	}
	if entries[1].Consumption1 != second.Consumption1 {
		t.Errorf("second Consumption1 = %q, want %q", entries[1].Consumption1, second.Consumption1)
	}
}

func TestSaveAndListPreserveCommasAndAccents(t *testing.T) {
	store := NewStore(filepath.Join(t.TempDir(), "histórico.csv"))
	want := testEntry(time.Date(2026, time.July, 15, 12, 0, 0, 0, time.UTC))
	want.Consumption1 = "Morador José, 105,5 kWh"
	want.Share1 = "participação, 61,09%"

	if err := store.Save(want); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	entries, err := store.List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("entry count = %d, want 1", len(entries))
	}
	got := entries[0]
	if !got.Date.Equal(want.Date) {
		t.Errorf("Date = %v, want %v", got.Date, want.Date)
	}
	want.Date = got.Date
	if !reflect.DeepEqual(got, want) {
		t.Errorf("entry = %#v, want %#v", got, want)
	}
}

func TestListReturnsEmptyForMissingFile(t *testing.T) {
	entries, err := NewStore(filepath.Join(t.TempDir(), "missing.csv")).List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("entries = %#v, want empty", entries)
	}
}

func TestListRejectsInvalidHeader(t *testing.T) {
	path := filepath.Join(t.TempDir(), "history.csv")
	if err := os.WriteFile(path, []byte("Date,Consumption\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	_, err := NewStore(path).List()
	if err == nil || !strings.Contains(err.Error(), "invalid history header") {
		t.Fatalf("List() error = %v, want invalid header error", err)
	}
}

func TestListRejectsInvalidFields(t *testing.T) {
	tests := []struct {
		name string
		row  string
	}{
		{name: "field count", row: "2026-07-15T12:00:00Z,only-two-fields\n"},
		{name: "date", row: "15/07/2026,1,2,3,4,5,6,7,8\n"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "history.csv")
			var content strings.Builder
			writer := csv.NewWriter(&content)
			if err := writer.Write(header); err != nil {
				t.Fatalf("Write() error = %v", err)
			}
			writer.Flush()
			content.WriteString(test.row)
			if err := os.WriteFile(path, []byte(content.String()), 0o644); err != nil {
				t.Fatalf("WriteFile() error = %v", err)
			}

			if _, err := NewStore(path).List(); err == nil {
				t.Fatal("List() error = nil, want invalid fields error")
			}
		})
	}
}

func TestDeleteAtRemovesPositionalEntry(t *testing.T) {
	tests := []struct {
		name  string
		index int
		want  []string
	}{
		{name: "first", index: 0, want: []string{"registro 2", "registro 3"}},
		{name: "middle", index: 1, want: []string{"registro 1", "registro 3"}},
		{name: "last", index: 2, want: []string{"registro 1", "registro 2"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			store := NewStore(filepath.Join(t.TempDir(), "history.csv"))
			for index := 1; index <= 3; index++ {
				entry := testEntry(time.Date(2026, time.July, index, 12, 0, 0, 0, time.UTC))
				entry.Consumption1 = fmt.Sprintf("registro %d", index)
				if err := store.Save(entry); err != nil {
					t.Fatalf("Save() error = %v", err)
				}
			}

			if err := store.DeleteAt(test.index); err != nil {
				t.Fatalf("DeleteAt(%d) error = %v", test.index, err)
			}
			entries, err := store.List()
			if err != nil {
				t.Fatalf("List() error = %v", err)
			}
			got := make([]string, len(entries))
			for index, entry := range entries {
				got[index] = entry.Consumption1
			}
			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("remaining entries = %q, want %q", got, test.want)
			}
		})
	}
}

func TestDeleteAtOnlyEntryPreservesHeader(t *testing.T) {
	path := filepath.Join(t.TempDir(), "history.csv")
	store := NewStore(path)
	if err := store.Save(testEntry(time.Date(2026, time.July, 15, 12, 0, 0, 0, time.UTC))); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	if err := store.DeleteAt(0); err != nil {
		t.Fatalf("DeleteAt(0) error = %v", err)
	}
	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer file.Close()
	records, err := csv.NewReader(file).ReadAll()
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}
	if len(records) != 1 || !reflect.DeepEqual(records[0], header) {
		t.Fatalf("records after deletion = %q, want only header %q", records, header)
	}
}

func TestDeleteAtRejectsInvalidIndexWithoutChangingFile(t *testing.T) {
	for _, index := range []int{-1, 2, 3} {
		t.Run(fmt.Sprintf("index_%d", index), func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "history.csv")
			store := NewStore(path)
			for day := 1; day <= 2; day++ {
				if err := store.Save(testEntry(time.Date(2026, time.July, day, 12, 0, 0, 0, time.UTC))); err != nil {
					t.Fatalf("Save() error = %v", err)
				}
			}
			before, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("ReadFile() before deletion error = %v", err)
			}

			err = store.DeleteAt(index)
			if !errors.Is(err, ErrInvalidIndex) {
				t.Fatalf("DeleteAt(%d) error = %v, want ErrInvalidIndex", index, err)
			}
			after, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("ReadFile() after deletion error = %v", err)
			}
			if !bytes.Equal(after, before) {
				t.Fatal("invalid deletion changed the history file")
			}
		})
	}
}

func TestDeleteAtMissingFileReturnsInvalidIndexWithoutCreatingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.csv")
	err := NewStore(path).DeleteAt(0)
	if !errors.Is(err, ErrInvalidIndex) {
		t.Fatalf("DeleteAt(0) error = %v, want ErrInvalidIndex", err)
	}
	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("Stat() error = %v, want file to remain missing", err)
	}
}

func TestDeleteAtRemovesOneOfIdenticalDuplicates(t *testing.T) {
	store := NewStore(filepath.Join(t.TempDir(), "history.csv"))
	duplicate := testEntry(time.Date(2026, time.July, 15, 12, 0, 0, 0, time.UTC))
	last := testEntry(time.Date(2026, time.July, 16, 12, 0, 0, 0, time.UTC))
	last.Consumption1 = "registro distinto"
	for _, entry := range []Entry{duplicate, duplicate, last} {
		if err := store.Save(entry); err != nil {
			t.Fatalf("Save() error = %v", err)
		}
	}

	if err := store.DeleteAt(1); err != nil {
		t.Fatalf("DeleteAt(1) error = %v", err)
	}
	entries, err := store.List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("entry count = %d, want 2", len(entries))
	}
	if entries[0].Consumption1 != duplicate.Consumption1 || entries[1].Consumption1 != last.Consumption1 {
		t.Errorf("remaining entries = %#v, want one duplicate followed by distinct entry", entries)
	}
}
