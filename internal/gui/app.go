// Package gui implements the Fyne desktop interface for Rateio Luz.
package gui

import (
	"errors"
	"fmt"
	"image/color"
	"os"
	"path/filepath"
	"strings"
	"time"

	"rateio-luz/assets"
	"rateio-luz/internal/calculator"
	"rateio-luz/internal/history"
	"rateio-luz/internal/presentation"
	"rateio-luz/internal/validation"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
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
	tabs   *container.AppTabs

	rateioTab  *container.TabItem
	historyTab *container.TabItem
	store      historyStore
	now        func() time.Time

	consumption1Entry *widget.Entry
	consumption2Entry *widget.Entry
	totalAmountEntry  *widget.Entry
	calculateButton   *widget.Button
	clearButton       *widget.Button
	footerLabel       *widget.Label

	errorBox   *fyne.Container
	errorLabel *widget.Label
	resultCard *widget.Card
	saveButton *widget.Button
	saveStatus *widget.Label
	snapshot   *history.Entry

	totalConsumptionValue *widget.Label
	share1Value           *widget.Label
	share2Value           *widget.Label
	amount1Value          *widget.Label
	amount2Value          *widget.Label
	reconciliationValue   *widget.Label

	refreshHistoryButton *widget.Button
	historyStatusCard    *widget.Card
	historyStatusLabel   *widget.Label
	historyList          *fyne.Container
	historyScroll        *container.Scroll
	historyEntries       []history.Entry
}

type historyStore interface {
	Save(history.Entry) error
	List() ([]history.Entry, error)
}

// NewWindow builds the main Rateio Luz window without starting the event loop.
func NewWindow(application fyne.App) fyne.Window {
	return newScreen(application).window
}

func newScreen(application fyne.App) *screen {
	path := defaultHistoryPath()
	// Store.Save reports any remaining filesystem error to the user. Creating
	// the application directory here also supports stores that expect it to
	// exist already.
	_ = os.MkdirAll(filepath.Dir(path), 0o700)
	return newScreenWithStore(application, history.NewStore(path))
}

func newScreenWithStore(application fyne.App, store historyStore) *screen {
	s := &screen{
		window:            application.NewWindow(AppName),
		store:             store,
		now:               time.Now,
		consumption1Entry: newEntry("Ex.: 105,5", theme.AccountIcon()),
		consumption2Entry: newEntry("Ex.: 67,2", theme.AccountIcon()),
		totalAmountEntry:  newEntry("Ex.: 184,72", theme.DocumentIcon()),
	}

	s.errorLabel = widget.NewLabel("")
	s.errorLabel.Wrapping = fyne.TextWrapWord
	s.errorLabel.Importance = widget.DangerImportance
	s.errorBox = newSurface(
		errorWash,
		colorWithAlpha(theme.ErrorColor(), 0x58),
		10,
		container.NewBorder(nil, nil, widget.NewIcon(theme.ErrorIcon()), nil, s.errorLabel),
	)
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
	s.scroll = container.NewVScroll(container.NewPadded(newConstrainedContainer(760, content)))
	s.rateioTab = container.NewTabItemWithIcon("Rateio", theme.HomeIcon(), s.scroll)
	s.historyTab = container.NewTabItemWithIcon("Histórico", theme.HistoryIcon(), s.buildHistoryTab())
	s.tabs = container.NewAppTabs(s.rateioTab, s.historyTab)
	s.tabs.OnSelected = func(selected *container.TabItem) {
		if selected == s.historyTab {
			s.loadHistory()
		}
	}

	s.window.SetContent(container.NewBorder(s.buildHeader(), s.buildFooter(), nil, nil, s.tabs))
	s.window.Resize(fyne.NewSize(780, 760))
	s.window.CenterOnScreen()
	return s
}

func defaultHistoryPath() string {
	configDir, err := os.UserConfigDir()
	if err != nil || strings.TrimSpace(configDir) == "" {
		homeDir, homeErr := os.UserHomeDir()
		if homeErr == nil && strings.TrimSpace(homeDir) != "" {
			configDir = filepath.Join(homeDir, ".config")
		} else {
			configDir = os.TempDir()
		}
	}
	return filepath.Join(configDir, "rateio-luz", "historico.csv")
}

func newEntry(placeholder string, icon fyne.Resource) *widget.Entry {
	entry := widget.NewEntry()
	entry.SetPlaceHolder(placeholder)
	entry.SetIcon(icon)
	return entry
}

func colorWithAlpha(value color.Color, alpha uint8) color.NRGBA {
	converted := color.NRGBAModel.Convert(value).(color.NRGBA)
	converted.A = alpha
	return converted
}

func newSurface(fill, stroke color.Color, radius float32, content fyne.CanvasObject) *fyne.Container {
	background := canvas.NewRectangle(fill)
	background.CornerRadius = radius
	background.StrokeColor = stroke
	background.StrokeWidth = 1
	return container.NewStack(background, container.NewPadded(content))
}

func sectionLabel(text string) *widget.Label {
	label := widget.NewLabelWithStyle(text, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	label.SizeName = theme.SizeNameCaptionText
	label.Importance = widget.LowImportance
	return label
}

func fieldGroup(marker, title, unit string, entry *widget.Entry) fyne.CanvasObject {
	markerBackground := canvas.NewRectangle(softBlue)
	markerBackground.CornerRadius = 9
	markerBackground.StrokeColor = colorWithAlpha(accent, 0x70)
	markerBackground.StrokeWidth = 1
	markerBackground.SetMinSize(fyne.NewSize(38, 30))
	markerText := canvas.NewText(marker, MarineBlue)
	markerText.TextSize = 12
	markerText.TextStyle = fyne.TextStyle{Bold: true}
	markerText.Alignment = fyne.TextAlignCenter
	markerBadge := container.NewStack(markerBackground, container.NewCenter(markerText))

	titleLabel := widget.NewLabelWithStyle(title, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	unitLabel := widget.NewLabel("· " + unit)
	unitLabel.SizeName = theme.SizeNameCaptionText
	unitLabel.Importance = widget.LowImportance
	header := container.NewBorder(nil, nil, markerBadge, nil, container.NewHBox(titleLabel, unitLabel))

	return newSurface(surface, line, 12, container.NewVBox(header, entry))
}

func resultTile(title string, amount, share *widget.Label) fyne.CanvasObject {
	amount.Alignment = fyne.TextAlignLeading
	amount.Selectable = true
	share.Selectable = true
	caption := widget.NewLabel("Valor a pagar")
	caption.SizeName = theme.SizeNameCaptionText
	caption.Importance = widget.LowImportance
	shareLine := container.NewBorder(nil, nil, widget.NewLabel("Proporção"), nil, share)
	return newSurface(
		surface,
		line,
		12,
		container.NewVBox(sectionLabel(title), caption, amount, widget.NewSeparator(), shareLine),
	)
}

type maxWidthLayout struct {
	maxWidth float32
}

func (l *maxWidthLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	if len(objects) == 0 {
		return
	}
	width := size.Width
	if width > l.maxWidth {
		width = l.maxWidth
	}
	objects[0].Resize(fyne.NewSize(width, size.Height))
	objects[0].Move(fyne.NewPos((size.Width-width)/2, 0))
}

func (l *maxWidthLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	if len(objects) == 0 {
		return fyne.Size{}
	}
	minimum := objects[0].MinSize()
	if minimum.Width > l.maxWidth {
		minimum.Width = l.maxWidth
	}
	return minimum
}

func newConstrainedContainer(maxWidth float32, content fyne.CanvasObject) *fyne.Container {
	return container.New(&maxWidthLayout{maxWidth: maxWidth}, content)
}

func (s *screen) buildHeader() fyne.CanvasObject {
	background := canvas.NewHorizontalGradient(deepNavy, MarineBlue)
	background.SetMinSize(fyne.NewSize(0, 104))

	logoBackground := canvas.NewRectangle(colorWithAlpha(OffWhite, 0x18))
	logoBackground.CornerRadius = 16
	logoBackground.StrokeColor = colorWithAlpha(OffWhite, 0x38)
	logoBackground.StrokeWidth = 1
	logoBackground.SetMinSize(fyne.NewSize(68, 68))
	logo := canvas.NewImageFromResource(assets.Icon)
	logo.FillMode = canvas.ImageFillContain
	logo.SetMinSize(fyne.NewSize(52, 52))
	logoTile := container.NewStack(logoBackground, container.NewPadded(logo))

	title := canvas.NewText(AppName, OffWhite)
	title.TextSize = 29
	title.TextStyle = fyne.TextStyle{Bold: true}
	subtitle := canvas.NewText("CONSUMO JUSTO  •  CONTA DIVIDIDA", colorWithAlpha(OffWhite, 0xC8))
	subtitle.TextSize = 12
	subtitle.TextStyle = fyne.TextStyle{Bold: true}

	headerContent := container.NewPadded(container.NewPadded(container.NewHBox(
		logoTile,
		container.NewVBox(title, subtitle),
	)))
	accentRule := canvas.NewHorizontalGradient(accent, colorWithAlpha(accent, 0x00))
	accentRule.SetMinSize(fyne.NewSize(0, 3))
	return container.NewBorder(nil, accentRule, nil, nil, container.NewStack(background, headerContent))
}

func (s *screen) buildFooter() fyne.CanvasObject {
	s.footerLabel = widget.NewLabel("Noguires-Pires\nAll rights reserved.")
	s.footerLabel.Alignment = fyne.TextAlignCenter
	s.footerLabel.SizeName = theme.SizeNameCaptionText
	s.footerLabel.Importance = widget.LowImportance
	separator := canvas.NewRectangle(line)
	separator.SetMinSize(fyne.NewSize(0, 1))
	return container.NewBorder(separator, nil, nil, nil, container.NewPadded(s.footerLabel))
}

func (s *screen) buildInputCard() fyne.CanvasObject {
	residents := container.NewGridWithColumns(2,
		fieldGroup("01", "Consumo do morador 1", "em kWh", s.consumption1Entry),
		fieldGroup("02", "Consumo do morador 2", "em kWh", s.consumption2Entry),
	)
	bill := fieldGroup("R$", "Valor total da conta", "em reais", s.totalAmountEntry)

	help := widget.NewLabel("Você pode usar vírgula ou ponto nos decimais — sem separador de milhares.")
	help.Wrapping = fyne.TextWrapWord
	help.SizeName = theme.SizeNameCaptionText
	help.Importance = widget.LowImportance
	helpBox := container.NewBorder(nil, nil, widget.NewIcon(theme.InfoIcon()), nil, help)

	actions := container.NewGridWithColumns(2,
		s.clearButton,
		s.calculateButton,
	)
	return widget.NewCard(
		"Monte o rateio",
		"Informe os dois consumos e o total da conta.",
		container.NewVBox(residents, bill, helpBox, s.errorBox, widget.NewSeparator(), actions),
	)
}

func (s *screen) buildResultCard() {
	s.totalConsumptionValue = resultValueLabel()
	s.share1Value = resultValueLabel()
	s.share2Value = resultValueLabel()
	s.amount1Value = resultValueLabel()
	s.amount2Value = resultValueLabel()
	s.amount1Value.SizeName = theme.SizeNameHeadingText
	s.amount2Value.SizeName = theme.SizeNameHeadingText
	s.reconciliationValue = widget.NewLabel("")
	s.reconciliationValue.Wrapping = fyne.TextWrapWord
	s.reconciliationValue.Importance = widget.SuccessImportance
	s.saveButton = widget.NewButtonWithIcon("Salvar no histórico", theme.DocumentSaveIcon(), s.saveSnapshot)
	s.saveButton.Importance = widget.HighImportance
	s.saveButton.Disable()
	s.saveStatus = widget.NewLabel("")
	s.saveStatus.Wrapping = fyne.TextWrapWord
	s.saveStatus.Alignment = fyne.TextAlignCenter
	s.saveStatus.Hide()

	totalLine := newSurface(
		softBlue,
		colorWithAlpha(accent, 0x70),
		12,
		container.NewBorder(nil, nil,
			sectionLabel("CONSUMO COMBINADO"),
			nil,
			s.totalConsumptionValue,
		),
	)
	payouts := container.NewGridWithColumns(2,
		resultTile("MORADOR 1", s.amount1Value, s.share1Value),
		resultTile("MORADOR 2", s.amount2Value, s.share2Value),
	)
	confirmation := newSurface(
		successWash,
		colorWithAlpha(theme.SuccessColor(), 0x50),
		10,
		container.NewBorder(nil, nil, widget.NewIcon(theme.ConfirmIcon()), nil, s.reconciliationValue),
	)
	s.resultCard = widget.NewCard(
		"Divisão calculada",
		"Cada valor acompanha a proporção de consumo.",
		container.NewVBox(
			totalLine,
			payouts,
			confirmation,
			widget.NewSeparator(),
			s.saveButton,
			s.saveStatus,
		),
	)
	s.resultCard.Hide()
}

func (s *screen) buildHistoryTab() fyne.CanvasObject {
	title := widget.NewLabelWithStyle("Rateios salvos", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	title.SizeName = theme.SizeNameSubHeadingText
	help := widget.NewLabel("Consulte os cálculos que você escolheu guardar neste dispositivo.")
	help.Wrapping = fyne.TextWrapWord
	help.Importance = widget.LowImportance
	s.refreshHistoryButton = widget.NewButtonWithIcon("Atualizar", theme.ViewRefreshIcon(), s.loadHistory)

	toolbar := container.NewBorder(
		nil,
		nil,
		nil,
		container.NewCenter(s.refreshHistoryButton),
		container.NewVBox(title, help),
	)

	s.historyStatusLabel = widget.NewLabel("O histórico será carregado ao abrir esta aba.")
	s.historyStatusLabel.Wrapping = fyne.TextWrapWord
	s.historyStatusCard = widget.NewCard(
		"Seu histórico",
		"",
		container.NewBorder(nil, nil, widget.NewIcon(theme.InfoIcon()), nil, s.historyStatusLabel),
	)
	s.historyList = container.NewVBox()
	s.historyScroll = container.NewVScroll(container.NewPadded(newConstrainedContainer(760, container.NewVBox(
		s.historyStatusCard,
		s.historyList,
	))))

	return container.NewBorder(container.NewPadded(newConstrainedContainer(760, toolbar)), nil, nil, nil, s.historyScroll)
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
	s.snapshot = &history.Entry{
		Date:             s.now(),
		Consumption1:     presentation.KWh(input.Consumption1),
		Consumption2:     presentation.KWh(input.Consumption2),
		TotalAmount:      presentation.BRL(input.TotalAmountCents),
		TotalConsumption: s.totalConsumptionValue.Text,
		Share1:           s.share1Value.Text,
		Share2:           s.share2Value.Text,
		Amount1:          s.amount1Value.Text,
		Amount2:          s.amount2Value.Text,
	}
	s.saveStatus.Hide()
	s.saveButton.Enable()
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
	s.snapshot = nil
	s.saveButton.Disable()
	s.saveStatus.Hide()
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
	s.snapshot = nil
	s.saveButton.Disable()
	s.saveStatus.Hide()
	s.errorLabel.SetText(message)
	s.errorBox.Show()
	if entry != nil {
		entry.SetValidationError(errors.New(message))
		s.window.Canvas().Focus(entry)
	}
}

func (s *screen) saveSnapshot() {
	if s.snapshot == nil {
		return
	}

	if err := s.store.Save(*s.snapshot); err != nil {
		s.saveStatus.SetText("Não foi possível salvar este rateio. Verifique o local do histórico e tente novamente.")
		s.saveStatus.Importance = widget.DangerImportance
		s.saveStatus.Refresh()
		s.saveStatus.Show()
		s.saveButton.Enable()
		return
	}

	s.saveStatus.SetText("Rateio salvo. Ele já está disponível na aba Histórico.")
	s.saveStatus.Importance = widget.SuccessImportance
	s.saveStatus.Refresh()
	s.saveStatus.Show()
	s.saveButton.Disable()
}

func (s *screen) loadHistory() {
	entries, err := s.store.List()
	s.historyList.RemoveAll()
	s.historyScroll.ScrollToTop()
	if err != nil {
		s.historyEntries = nil
		s.showHistoryState(
			"Não foi possível abrir o histórico",
			"O arquivo não pôde ser lido. Verifique-o e use Atualizar para tentar novamente.",
		)
		return
	}

	s.historyEntries = entries
	if len(entries) == 0 {
		s.showHistoryState(
			"Nenhum rateio salvo ainda",
			"Faça um cálculo na aba Rateio e escolha “Salvar no histórico”.",
		)
		return
	}

	s.historyStatusCard.Hide()
	for _, entry := range entries {
		s.historyList.Add(historyCard(entry))
	}
	s.historyList.Refresh()
	s.historyScroll.Refresh()
	s.historyScroll.ScrollToTop()
}

func (s *screen) showHistoryState(title, message string) {
	s.historyStatusCard.SetTitle(title)
	s.historyStatusLabel.SetText(message)
	s.historyStatusCard.Show()
	s.historyStatusCard.Refresh()
}

func historyCard(entry history.Entry) fyne.CanvasObject {
	date := "Data não informada"
	if !entry.Date.IsZero() {
		date = entry.Date.Local().Format("02/01/2006 às 15:04")
	}

	totalAmount := historyValueLabel(entry.TotalAmount)
	totalAmount.SizeName = theme.SizeNameSubHeadingText
	totalAmount.Selectable = true
	totalLine := newSurface(
		softBlue,
		colorWithAlpha(accent, 0x60),
		10,
		container.NewBorder(nil, nil, sectionLabel("TOTAL DA CONTA"), nil, totalAmount),
	)
	totalConsumption := container.NewBorder(
		nil,
		nil,
		widget.NewLabel("Consumo combinado"),
		nil,
		historyValueLabel(entry.TotalConsumption),
	)
	residents := container.NewGridWithColumns(2,
		historyResidentTile("MORADOR 1", entry.Consumption1, entry.Share1, entry.Amount1),
		historyResidentTile("MORADOR 2", entry.Consumption2, entry.Share2, entry.Amount2),
	)
	return widget.NewCard(date, "Rateio proporcional salvo", container.NewVBox(totalLine, totalConsumption, residents))
}

func historyValueLabel(value string) *widget.Label {
	label := widget.NewLabelWithStyle(value, fyne.TextAlignTrailing, fyne.TextStyle{Bold: true})
	label.Wrapping = fyne.TextWrapWord
	return label
}

func historyResidentTile(title, consumption, share, amount string) fyne.CanvasObject {
	amountLabel := historyValueLabel(amount)
	amountLabel.Alignment = fyne.TextAlignLeading
	amountLabel.SizeName = theme.SizeNameSubHeadingText
	amountLabel.Selectable = true
	consumptionLine := container.NewBorder(
		nil,
		nil,
		widget.NewLabel("Consumo"),
		nil,
		historyValueLabel(consumption),
	)
	shareLine := container.NewBorder(
		nil,
		nil,
		widget.NewLabel("Proporção"),
		nil,
		historyValueLabel(share),
	)
	return newSurface(
		surface,
		line,
		10,
		container.NewVBox(sectionLabel(title), amountLabel, widget.NewSeparator(), consumptionLine, shareLine),
	)
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
