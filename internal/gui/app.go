// Package gui implements the Fyne desktop interface for Rateio Luz.
package gui

import (
	"errors"
	"fmt"
	"strings"

	"rateio-luz/assets"
	"rateio-luz/internal/calculator"
	"rateio-luz/internal/presentation"
	"rateio-luz/internal/validation"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

const (
	// AppName is the human-readable desktop application name.
	AppName = "Rateio Luz"
	// AppID is the stable identifier used by Fyne and desktop platforms.
	AppID = "com.raelpires.rateioluz"
)

type screen struct {
	window fyne.Window
	scroll *container.Scroll

	consumption1Entry *widget.Entry
	consumption2Entry *widget.Entry
	totalAmountEntry  *widget.Entry
	calculateButton   *widget.Button
	clearButton       *widget.Button

	errorBox   *fyne.Container
	errorLabel *widget.Label
	resultCard *widget.Card

	totalConsumptionValue *widget.Label
	share1Value           *widget.Label
	share2Value           *widget.Label
	amount1Value          *widget.Label
	amount2Value          *widget.Label
	reconciliationValue   *widget.Label
}

// NewWindow builds the main Rateio Luz window without starting the event loop.
func NewWindow(application fyne.App) fyne.Window {
	return newScreen(application).window
}

func newScreen(application fyne.App) *screen {
	s := &screen{
		window:            application.NewWindow(AppName),
		consumption1Entry: newEntry("Ex.: 105,5", theme.AccountIcon()),
		consumption2Entry: newEntry("Ex.: 67,2", theme.AccountIcon()),
		totalAmountEntry:  newEntry("Ex.: 184,72", theme.DocumentIcon()),
	}

	s.errorLabel = widget.NewLabel("")
	s.errorLabel.Wrapping = fyne.TextWrapWord
	s.errorBox = container.NewBorder(nil, nil, widget.NewIcon(theme.ErrorIcon()), nil, s.errorLabel)
	s.errorBox.Hide()

	s.buildResultCard()
	s.calculateButton = widget.NewButtonWithIcon("Calcular rateio", theme.ConfirmIcon(), s.calculate)
	s.calculateButton.Importance = widget.HighImportance
	s.clearButton = widget.NewButtonWithIcon("Limpar", theme.ContentClearIcon(), s.clear)
	s.clearButton.Importance = widget.LowImportance
	s.bindInputEvents()

	content := container.NewVBox(
		s.buildInputCard(),
		s.resultCard,
	)
	s.scroll = container.NewVScroll(container.NewPadded(content))
	s.window.SetContent(container.NewBorder(s.buildHeader(), nil, nil, nil, s.scroll))
	s.window.Resize(fyne.NewSize(720, 700))
	s.window.CenterOnScreen()
	return s
}

func newEntry(placeholder string, icon fyne.Resource) *widget.Entry {
	entry := widget.NewEntry()
	entry.SetPlaceHolder(placeholder)
	entry.SetIcon(icon)
	return entry
}

func (s *screen) buildHeader() fyne.CanvasObject {
	background := canvas.NewRectangle(MarineBlue)
	logo := canvas.NewImageFromResource(assets.Icon)
	logo.FillMode = canvas.ImageFillContain
	logo.SetMinSize(fyne.NewSize(58, 58))

	title := canvas.NewText(AppName, OffWhite)
	title.TextSize = 26
	title.TextStyle = fyne.TextStyle{Bold: true}
	subtitle := canvas.NewText("Divida a conta de energia de forma simples e justa", OffWhite)
	subtitle.TextSize = 14

	headerContent := container.NewPadded(container.NewHBox(
		logo,
		container.NewVBox(title, subtitle),
	))
	return container.NewStack(background, headerContent)
}

func (s *screen) buildInputCard() fyne.CanvasObject {
	form := widget.NewForm(
		widget.NewFormItem("Consumo do morador 1 (kWh)", s.consumption1Entry),
		widget.NewFormItem("Consumo do morador 2 (kWh)", s.consumption2Entry),
		widget.NewFormItem("Valor total da conta (R$)", s.totalAmountEntry),
	)
	help := widget.NewLabel("Use vírgula ou ponto nos decimais. Exemplo de conta: 184,72.")
	help.Wrapping = fyne.TextWrapWord

	actions := container.NewHBox(
		layout.NewSpacer(),
		s.clearButton,
		s.calculateButton,
	)
	return widget.NewCard(
		"Dados para o rateio",
		"Informe o consumo individual e o valor total da conta.",
		container.NewVBox(form, help, s.errorBox, actions),
	)
}

func (s *screen) buildResultCard() {
	s.totalConsumptionValue = resultValueLabel()
	s.share1Value = resultValueLabel()
	s.share2Value = resultValueLabel()
	s.amount1Value = resultValueLabel()
	s.amount2Value = resultValueLabel()
	s.reconciliationValue = widget.NewLabel("")
	s.reconciliationValue.Wrapping = fyne.TextWrapWord

	details := container.New(layout.NewFormLayout(),
		widget.NewLabel("Consumo total"), s.totalConsumptionValue,
		widget.NewLabel("Percentual do morador 1"), s.share1Value,
		widget.NewLabel("Percentual do morador 2"), s.share2Value,
		widget.NewLabel("Morador 1 paga"), s.amount1Value,
		widget.NewLabel("Morador 2 paga"), s.amount2Value,
	)
	confirmation := container.NewBorder(
		nil,
		nil,
		widget.NewIcon(theme.ConfirmIcon()),
		nil,
		s.reconciliationValue,
	)
	s.resultCard = widget.NewCard(
		"Resultado do rateio",
		"Valores proporcionais ao consumo informado.",
		container.NewVBox(details, widget.NewSeparator(), confirmation),
	)
	s.resultCard.Hide()
}

func resultValueLabel() *widget.Label {
	return widget.NewLabelWithStyle("", fyne.TextAlignTrailing, fyne.TextStyle{Bold: true})
}

func (s *screen) bindInputEvents() {
	entries := []*widget.Entry{s.consumption1Entry, s.consumption2Entry, s.totalAmountEntry}
	for _, current := range entries {
		entry := current
		entry.OnChanged = func(string) {
			entry.SetValidationError(nil)
			s.hideFeedback()
		}
	}

	s.consumption1Entry.OnSubmitted = func(string) {
		s.window.Canvas().Focus(s.consumption2Entry)
	}
	s.consumption2Entry.OnSubmitted = func(string) {
		s.window.Canvas().Focus(s.totalAmountEntry)
	}
	s.totalAmountEntry.OnSubmitted = func(string) {
		s.calculate()
	}
}

func (s *screen) calculate() {
	s.clearFieldErrors()
	input, err := validation.ParseAndValidate(
		s.consumption1Entry.Text,
		s.consumption2Entry.Text,
		s.totalAmountEntry.Text,
	)
	if err != nil {
		s.showValidationError(err)
		return
	}

	result, err := calculator.Calculate(
		input.Consumption1,
		input.Consumption2,
		input.TotalAmountCents,
	)
	if err != nil {
		s.showError("Não foi possível concluir o cálculo. Revise os valores informados.", nil)
		return
	}
	if result.Amount1Cents+result.Amount2Cents != input.TotalAmountCents {
		s.showError("Não foi possível conferir o total da conta.", nil)
		return
	}

	s.totalConsumptionValue.SetText(presentation.KWh(result.TotalConsumption))
	s.share1Value.SetText(presentation.Percentage(result.Share1) + "%")
	s.share2Value.SetText(presentation.Percentage(result.Share2) + "%")
	s.amount1Value.SetText(presentation.BRL(result.Amount1Cents))
	s.amount2Value.SetText(presentation.BRL(result.Amount2Cents))
	s.reconciliationValue.SetText(fmt.Sprintf(
		"Conferência: %s + %s = %s. Total conferido.",
		presentation.BRL(result.Amount1Cents),
		presentation.BRL(result.Amount2Cents),
		presentation.BRL(input.TotalAmountCents),
	))
	s.errorBox.Hide()
	s.resultCard.Show()
	s.scroll.Refresh()
	s.scroll.ScrollToBottom()
}

func (s *screen) clear() {
	s.consumption1Entry.SetText("")
	s.consumption2Entry.SetText("")
	s.totalAmountEntry.SetText("")
	s.clearFieldErrors()
	s.hideFeedback()
	s.scroll.ScrollToTop()
	s.window.Canvas().Focus(s.consumption1Entry)
}

func (s *screen) hideFeedback() {
	s.errorBox.Hide()
	s.resultCard.Hide()
}

func (s *screen) clearFieldErrors() {
	s.consumption1Entry.SetValidationError(nil)
	s.consumption2Entry.SetValidationError(nil)
	s.totalAmountEntry.SetValidationError(nil)
}

func (s *screen) showValidationError(err error) {
	message := friendlyValidationMessage(err)
	s.showError(message, s.entryForValidationError(err))
}

func (s *screen) showError(message string, entry *widget.Entry) {
	s.resultCard.Hide()
	s.errorLabel.SetText(message)
	s.errorBox.Show()
	if entry != nil {
		entry.SetValidationError(errors.New(message))
		s.window.Canvas().Focus(entry)
	}
}

func (s *screen) entryForValidationError(err error) *widget.Entry {
	message := err.Error()
	switch {
	case strings.HasPrefix(message, "consumo do consumidor 1:"),
		message == "o consumo total deve ser maior que zero":
		return s.consumption1Entry
	case strings.HasPrefix(message, "consumo do consumidor 2:"):
		return s.consumption2Entry
	case strings.HasPrefix(message, "valor total da conta:"):
		return s.totalAmountEntry
	default:
		return nil
	}
}

func friendlyValidationMessage(err error) string {
	message := err.Error()
	if message == "o consumo total deve ser maior que zero" {
		return "Informe um consumo maior que zero para pelo menos um morador."
	}

	field, detail, found := strings.Cut(message, ": ")
	if !found {
		return "Revise os dados informados e tente novamente."
	}
	field = strings.Replace(field, "consumidor", "morador", 1)
	switch detail {
	case "entrada vazia":
		return "Preencha o " + field + "."
	case "não pode ser negativo":
		return "O " + field + " não pode ser negativo."
	case "número inválido":
		return "Digite um número válido para o " + field + "."
	case "use apenas um separador decimal, sem separador de milhares":
		return "Use apenas um separador decimal no " + field + ", sem separador de milhares."
	case "use no máximo 2 casas decimais":
		return "No valor total da conta, use no máximo 2 casas decimais."
	case "valor fora do intervalo permitido":
		return "O valor total da conta é maior que o limite permitido."
	default:
		return "Revise o " + field + ": " + detail + "."
	}
}
