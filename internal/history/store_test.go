package history

import (
	"encoding/csv"
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
