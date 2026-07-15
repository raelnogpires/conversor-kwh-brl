package gui

import (
	"strings"
	"testing"

	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
)

func newTestScreen(t *testing.T) *screen {
	t.Helper()
	application := test.NewApp()
	application.Settings().SetTheme(NewTheme())
	s := newScreen(application)
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
	if s.resultCard.Visible() || s.errorBox.Visible() {
		t.Fatal("feedback should be hidden before the first calculation")
	}
	if s.consumption1Entry.PlaceHolder == "" || s.totalAmountEntry.PlaceHolder == "" {
		t.Fatal("input examples should be visible as placeholders")
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
