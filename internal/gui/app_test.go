package gui

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"rateio-luz/internal/history"

	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
)

func newTestScreen(t *testing.T) *screen {
	t.Helper()
	store := history.NewStore(filepath.Join(t.TempDir(), "historico.csv"))
	return newTestScreenWithStore(t, store)
}

func newTestScreenWithStore(t *testing.T, store historyStore) *screen {
	t.Helper()
	application := test.NewApp()
	application.Settings().SetTheme(NewTheme())
	s := newScreenWithStore(application, store)
	s.window.Show()
	t.Cleanup(func() {
		s.window.Close()
		application.Quit()
	})
	return s
}

func TestScreenIdentityAndInitialState(t *testing.T) {
	s := newTestScreen(t)

	if s.window.Title() != AppName {
		t.Errorf("window title = %q, want %q", s.window.Title(), AppName)
	}
	if s.calculateButton.Text != "Calcular rateio" || s.calculateButton.Importance != widget.HighImportance {
		t.Errorf("main action is not prominent: %+v", s.calculateButton)
	}
	if s.clearButton.Text != "Limpar" {
		t.Errorf("clear button = %q", s.clearButton.Text)
	}
	if len(s.tabs.Items) != 2 || s.tabs.Items[0].Text != "Rateio" || s.tabs.Items[1].Text != "Histórico" {
		t.Fatalf("tabs = %+v, want Rateio and Histórico", s.tabs.Items)
	}
	if s.tabs.Selected() != s.rateioTab {
		t.Fatal("Rateio should be the initially selected tab")
	}
	if s.resultCard.Visible() || s.errorBox.Visible() {
		t.Fatal("feedback should be hidden before the first calculation")
	}
	if !s.saveButton.Disabled() || s.snapshot != nil {
		t.Fatal("saving should be unavailable before a valid calculation")
	}
	if s.consumption1Entry.PlaceHolder == "" || s.totalAmountEntry.PlaceHolder == "" {
		t.Fatal("input examples should be visible as placeholders")
	}
	if s.footerLabel == nil {
		t.Fatal("footer label was not created")
	}
	if s.footerLabel.Text != "Noguires-Pires\nAll rights reserved." {
		t.Fatalf("footer = %q, want requested rights notice", s.footerLabel.Text)
	}
}

func TestCalculateDisplaysOrganizedResultAndReconciliation(t *testing.T) {
	s := newTestScreen(t)
	s.consumption1Entry.SetText("105,5")
	s.consumption2Entry.SetText("67,2")
	s.totalAmountEntry.SetText("184,72")

	test.Tap(s.calculateButton)

	if !s.resultCard.Visible() || s.errorBox.Visible() {
		t.Fatal("successful calculation should show only the result card")
	}
	wants := map[string]string{
		"total consumption": s.totalConsumptionValue.Text,
		"share 1":           s.share1Value.Text,
		"share 2":           s.share2Value.Text,
		"amount 1":          s.amount1Value.Text,
		"amount 2":          s.amount2Value.Text,
	}
	expected := map[string]string{
		"total consumption": "172,7 kWh",
		"share 1":           "61,09%",
		"share 2":           "38,91%",
		"amount 1":          "R$ 112,84",
		"amount 2":          "R$ 71,88",
	}
	for field, got := range wants {
		if got != expected[field] {
			t.Errorf("%s = %q, want %q", field, got, expected[field])
		}
	}
	if got := s.reconciliationValue.Text; got != "Conferência: R$ 112,84 + R$ 71,88 = R$ 184,72. Total conferido." {
		t.Errorf("reconciliation = %q", got)
	}
	if s.saveButton.Disabled() {
		t.Fatal("a valid calculation should enable saving")
	}
	if s.snapshot == nil {
		t.Fatal("a valid calculation should prepare a history snapshot")
	}
	if s.snapshot.Consumption1 != "105,5 kWh" || s.snapshot.Consumption2 != "67,2 kWh" || s.snapshot.TotalAmount != "R$ 184,72" {
		t.Errorf("snapshot inputs = %+v", s.snapshot)
	}
}

func TestSaveAndLoadHistoryUsingTemporaryPath(t *testing.T) {
	store := history.NewStore(filepath.Join(t.TempDir(), "historico.csv"))
	s := newTestScreenWithStore(t, store)
	fixedDate := time.Date(2026, time.July, 15, 14, 30, 0, 0, time.Local)
	s.now = func() time.Time { return fixedDate }
	s.consumption1Entry.SetText("10")
	s.consumption2Entry.SetText("30")
	s.totalAmountEntry.SetText("80")

	test.Tap(s.calculateButton)
	if s.saveButton.Disabled() {
		t.Fatal("save button should be enabled after calculation")
	}
	test.Tap(s.saveButton)

	if !s.saveButton.Disabled() {
		t.Fatal("save button should be disabled after the snapshot is saved")
	}
	if !s.saveStatus.Visible() || !strings.Contains(s.saveStatus.Text, "Rateio salvo") {
		t.Fatalf("save confirmation = %q", s.saveStatus.Text)
	}
	entries, err := store.List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("List() returned %d entries, want 1", len(entries))
	}
	want := history.Entry{
		Date:             fixedDate,
		Consumption1:     "10 kWh",
		Consumption2:     "30 kWh",
		TotalAmount:      "R$ 80,00",
		TotalConsumption: "40 kWh",
		Share1:           "25,00%",
		Share2:           "75,00%",
		Amount1:          "R$ 20,00",
		Amount2:          "R$ 60,00",
	}
	if got := entries[0]; !got.Date.Equal(want.Date) || got.Consumption1 != want.Consumption1 ||
		got.Consumption2 != want.Consumption2 || got.TotalAmount != want.TotalAmount ||
		got.TotalConsumption != want.TotalConsumption || got.Share1 != want.Share1 ||
		got.Share2 != want.Share2 || got.Amount1 != want.Amount1 || got.Amount2 != want.Amount2 {
		t.Errorf("saved entry = %+v, want %+v", got, want)
	}

	s.tabs.Select(s.historyTab)
	if len(s.historyEntries) != 1 || len(s.historyList.Objects) != 1 {
		t.Fatalf("history tab has %d entries and %d cards, want 1", len(s.historyEntries), len(s.historyList.Objects))
	}
	if s.historyStatusCard.Visible() {
		t.Fatal("non-empty history should hide the empty state")
	}

	second := want
	second.Date = fixedDate.Add(time.Hour)
	second.TotalAmount = "R$ 90,00"
	if err := store.Save(second); err != nil {
		t.Fatalf("saving second entry: %v", err)
	}
	test.Tap(s.refreshHistoryButton)
	if len(s.historyEntries) != 2 || s.historyEntries[1].TotalAmount != "R$ 90,00" {
		t.Fatalf("updated history = %+v", s.historyEntries)
	}
}

func TestHistoryTabShowsEmptyState(t *testing.T) {
	s := newTestScreen(t)

	s.tabs.Select(s.historyTab)

	if !s.historyStatusCard.Visible() {
		t.Fatal("empty history should show its state card")
	}
	if s.historyStatusCard.Title != "Nenhum rateio salvo ainda" {
		t.Fatalf("empty state title = %q", s.historyStatusCard.Title)
	}
	if len(s.historyEntries) != 0 || len(s.historyList.Objects) != 0 {
		t.Fatal("empty history should not render entry cards")
	}
}

func TestValidationMessagesAreFriendlyAndDoNotShowStaleResult(t *testing.T) {
	tests := []struct {
		name        string
		first       string
		second      string
		bill        string
		wantMessage string
	}{
		{"empty field", "", "10", "50", "Preencha o consumo do morador 1."},
		{"non numeric", "abc", "10", "50", "Digite um número válido para o consumo do morador 1."},
		{"negative consumption", "-1", "10", "50", "O consumo do morador 1 não pode ser negativo."},
		{"zero total", "0", "0", "50", "Informe um consumo maior que zero para pelo menos um morador."},
		{"empty bill", "10", "20", "", "Preencha o valor total da conta."},
		{"invalid bill", "10", "20", "R$ 50", "Digite um número válido para o valor total da conta."},
		{"negative bill", "10", "20", "-50", "O valor total da conta não pode ser negativo."},
		{"bill precision", "10", "20", "50,001", "No valor total da conta, use no máximo 2 casas decimais."},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			s := newTestScreen(t)
			s.consumption1Entry.SetText(testCase.first)
			s.consumption2Entry.SetText(testCase.second)
			s.totalAmountEntry.SetText(testCase.bill)

			test.Tap(s.calculateButton)

			if !s.errorBox.Visible() || s.resultCard.Visible() {
				t.Fatal("invalid input should show only validation feedback")
			}
			if s.errorLabel.Text != testCase.wantMessage {
				t.Errorf("message = %q, want %q", s.errorLabel.Text, testCase.wantMessage)
			}
		})
	}
}

func TestUserCanEditCalculateAgainAndClear(t *testing.T) {
	s := newTestScreen(t)
	s.consumption1Entry.SetText("100")
	s.consumption2Entry.SetText("100")
	s.totalAmountEntry.SetText("200")
	test.Tap(s.calculateButton)
	if s.amount1Value.Text != "R$ 100,00" || s.amount2Value.Text != "R$ 100,00" {
		t.Fatalf("first calculation = %q and %q", s.amount1Value.Text, s.amount2Value.Text)
	}

	s.consumption1Entry.SetText("1")
	if s.resultCard.Visible() {
		t.Fatal("editing an input should hide the stale result")
	}
	s.consumption2Entry.SetText("3")
	s.totalAmountEntry.SetText("0,05")
	test.Tap(s.calculateButton)
	if s.amount1Value.Text != "R$ 0,01" || s.amount2Value.Text != "R$ 0,04" {
		t.Fatalf("second calculation = %q and %q", s.amount1Value.Text, s.amount2Value.Text)
	}

	test.Tap(s.clearButton)
	if strings.Join([]string{s.consumption1Entry.Text, s.consumption2Entry.Text, s.totalAmountEntry.Text}, "") != "" {
		t.Fatal("clear should empty every input")
	}
	if s.resultCard.Visible() || s.errorBox.Visible() {
		t.Fatal("clear should hide all feedback")
	}
	if focused := s.window.Canvas().Focused(); focused != s.consumption1Entry {
		t.Errorf("focus after clear = %T, want first consumption entry", focused)
	}
}

func TestEnterMovesThroughFieldsAndRunsMainAction(t *testing.T) {
	s := newTestScreen(t)
	s.consumption1Entry.SetText("10")
	s.consumption2Entry.SetText("30")
	s.totalAmountEntry.SetText("80")

	s.consumption1Entry.OnSubmitted(s.consumption1Entry.Text)
	if focused := s.window.Canvas().Focused(); focused != s.consumption2Entry {
		t.Fatalf("first Enter focused %T", focused)
	}
	s.consumption2Entry.OnSubmitted(s.consumption2Entry.Text)
	if focused := s.window.Canvas().Focused(); focused != s.totalAmountEntry {
		t.Fatalf("second Enter focused %T", focused)
	}
	s.totalAmountEntry.OnSubmitted(s.totalAmountEntry.Text)
	if !s.resultCard.Visible() {
		t.Fatal("Enter on the bill field should calculate")
	}
}
