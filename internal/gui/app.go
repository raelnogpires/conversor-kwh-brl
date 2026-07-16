// Package gui implementa a interface desktop do Rateio Luz com o toolkit Fyne.
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
	// AppName é o nome legível exibido pela aplicação desktop.
	AppName = "Rateio Luz"
	// AppID é o identificador estável usado pelo Fyne e pelas plataformas desktop.
	AppID = "com.raelpires.rateioluz"
)

// screen concentra os widgets e o estado mutável de uma janela. Manter essas
// referências juntas permite que callbacks atualizem a mesma tela sem misturar
// regras de validação, cálculo ou persistência à construção visual.
type screen struct {
	// Estrutura principal de navegação e rolagem da tela de rateio.
	window fyne.Window
	scroll *container.Scroll
	tabs   *container.AppTabs

	// Abas e dependências injetáveis. now é uma função para que a criação do
	// registro histórico possa usar um relógio controlado em testes.
	rateioTab  *container.TabItem
	historyTab *container.TabItem
	store      historyStore
	now        func() time.Time

	// Entradas e ações do formulário que inicia o cálculo.
	consumption1Entry *widget.Entry
	consumption2Entry *widget.Entry
	totalAmountEntry  *widget.Entry
	calculateButton   *widget.Button
	clearButton       *widget.Button
	footerLabel       *widget.Label

	// Feedback da última tentativa e snapshot do último cálculo válido. O
	// snapshot só existe enquanto o resultado visível ainda pode ser salvo.
	errorBox   *fyne.Container
	errorLabel *widget.Label
	resultCard *widget.Card
	saveButton *widget.Button
	saveStatus *widget.Label
	snapshot   *history.Entry

	// Labels atualizados em conjunto quando o cálculo é concluído com sucesso.
	totalConsumptionValue *widget.Label
	share1Value           *widget.Label
	share2Value           *widget.Label
	amount1Value          *widget.Label
	amount2Value          *widget.Label
	reconciliationValue   *widget.Label

	// Estado da aba Histórico. Entradas e botões são guardados na mesma ordem
	// exibida; cada callback de exclusão captura o índice correspondente no store.
	refreshHistoryButton *widget.Button
	historyStatusCard    *widget.Card
	historyStatusLabel   *widget.Label
	historyList          *fyne.Container
	historyScroll        *container.Scroll
	historyEntries       []history.Entry
	historyDeleteButtons []*widget.Button
}

// historyStore é a fronteira injetável entre a GUI e a persistência. A tela
// depende apenas das operações que usa, o que evita acoplá-la ao arquivo CSV e
// permite substituir o armazenamento em testes.
type historyStore interface {
	Save(history.Entry) error
	List() ([]history.Entry, error)
	DeleteAt(index int) error
}

// NewWindow constrói a janela principal do Rateio Luz sem iniciar o loop de eventos.
func NewWindow(application fyne.App) fyne.Window {
	return newScreen(application).window
}

// newScreen usa o armazenamento real da aplicação. A criação antecipada da
// pasta é uma tentativa de preparação; erros definitivos continuam sendo
// tratados por Save e apresentados ao usuário no momento apropriado.
func newScreen(application fyne.App) *screen {
	path := defaultHistoryPath()
	// Store.Save informa ao usuário qualquer erro restante do sistema de
	// arquivos. Criar o diretório aqui também atende stores que esperam que ele
	// já exista.
	_ = os.MkdirAll(filepath.Dir(path), 0o700)
	return newScreenWithStore(application, history.NewStore(path))
}

// newScreenWithStore monta a tela com uma implementação de histórico recebida
// de fora. Primeiro cria o estado compartilhado, depois os componentes que
// registram callbacks e, por fim, liga tudo à janela.
func newScreenWithStore(application fyne.App, store historyStore) *screen {
	s := &screen{
		window:            application.NewWindow(AppName),
		store:             store,
		now:               time.Now,
		consumption1Entry: newEntry("Ex.: 105,5", theme.AccountIcon()),
		consumption2Entry: newEntry("Ex.: 67,2", theme.AccountIcon()),
		totalAmountEntry:  newEntry("Ex.: 184,72", theme.DocumentIcon()),
	}

	// O aviso de erro é construído uma vez e apenas alternado entre oculto e
	// visível, evitando recriar o layout a cada validação.
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

	// O conteúdo do rateio tem largura máxima para continuar legível em janelas
	// largas, mas permanece dentro de uma rolagem em alturas menores.
	content := container.NewVBox(
		s.buildInputCard(),
		s.resultCard,
	)
	s.scroll = container.NewVScroll(container.NewPadded(newConstrainedContainer(760, content)))
	s.rateioTab = container.NewTabItemWithIcon("Rateio", theme.HomeIcon(), s.scroll)
	s.historyTab = container.NewTabItemWithIcon("Histórico", theme.HistoryIcon(), s.buildHistoryTab())
	s.tabs = container.NewAppTabs(s.rateioTab, s.historyTab)
	// O histórico é carregado sob demanda: abrir a aplicação não faz I/O até o
	// usuário visitar a aba, e cada nova seleção reflete mudanças no arquivo.
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

// defaultHistoryPath escolhe um local persistente por usuário e aplica
// fallbacks seguros quando o sistema não fornece um diretório de configuração.
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

// newEntry padroniza as pistas visuais das entradas numéricas.
func newEntry(placeholder string, icon fyne.Resource) *widget.Entry {
	entry := widget.NewEntry()
	entry.SetPlaceHolder(placeholder)
	entry.SetIcon(icon)
	return entry
}

// colorWithAlpha preserva a cor recebida e ajusta somente sua transparência.
func colorWithAlpha(value color.Color, alpha uint8) color.NRGBA {
	converted := color.NRGBAModel.Convert(value).(color.NRGBA)
	converted.A = alpha
	return converted
}

// newSurface cria o padrão visual reutilizado por avisos e blocos de resultado:
// fundo arredondado, borda e espaçamento interno ao redor do conteúdo.
func newSurface(fill, stroke color.Color, radius float32, content fyne.CanvasObject) *fyne.Container {
	background := canvas.NewRectangle(fill)
	background.CornerRadius = radius
	background.StrokeColor = stroke
	background.StrokeWidth = 1
	return container.NewStack(background, container.NewPadded(content))
}

// sectionLabel produz títulos secundários consistentes dentro dos cartões.
func sectionLabel(text string) *widget.Label {
	label := widget.NewLabelWithStyle(text, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	label.SizeName = theme.SizeNameCaptionText
	label.Importance = widget.LowImportance
	return label
}

// fieldGroup reúne marcador, descrição, unidade e entrada em uma única unidade
// visual para o usuário relacionar claramente cada valor ao seu significado.
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

// resultTile apresenta, lado a lado com o outro morador, valor e proporção de
// um participante. Os valores são selecionáveis para facilitar sua cópia.
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

// maxWidthLayout limita apenas a largura do primeiro filho e o centraliza. Esse
// layout mantém formulários confortáveis sem impedir que ocupem telas estreitas.
type maxWidthLayout struct {
	maxWidth float32
}

// Layout implementa fyne.Layout para posicionar o conteúdo dentro do limite.
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

// MinSize respeita tanto o mínimo pedido pelo filho quanto a largura máxima.
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

// newConstrainedContainer aplica maxWidthLayout a um único conteúdo.
func newConstrainedContainer(maxWidth float32, content fyne.CanvasObject) *fyne.Container {
	return container.New(&maxWidthLayout{maxWidth: maxWidth}, content)
}

// buildHeader monta a identidade visual fixa no topo da janela.
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

// buildFooter cria o rodapé compartilhado pelas duas abas.
func (s *screen) buildFooter() fyne.CanvasObject {
	s.footerLabel = widget.NewLabel("Noguires-Pires\nAll rights reserved.")
	s.footerLabel.Alignment = fyne.TextAlignCenter
	s.footerLabel.SizeName = theme.SizeNameCaptionText
	s.footerLabel.Importance = widget.LowImportance
	separator := canvas.NewRectangle(line)
	separator.SetMinSize(fyne.NewSize(0, 1))
	return container.NewBorder(separator, nil, nil, nil, container.NewPadded(s.footerLabel))
}

// buildInputCard organiza os campos, a orientação de formato e as ações do
// formulário. Os widgets já pertencem a screen para serem usados nos callbacks.
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

// buildResultCard prepara a área inicialmente oculta que receberá o resultado
// e o comando de persistência depois de um cálculo válido.
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

// buildHistoryTab monta a barra de atualização, o estado vazio/erro e a lista
// rolável. Os dados só serão preenchidos por loadHistory.
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

// resultValueLabel padroniza números de resultado alinhados à direita.
func resultValueLabel() *widget.Label {
	return widget.NewLabelWithStyle("", fyne.TextAlignTrailing, fyne.TextStyle{Bold: true})
}

// bindInputEvents define o comportamento reativo do formulário: editar remove
// feedback obsoleto, Enter avança o foco e o último campo dispara o cálculo.
func (s *screen) bindInputEvents() {
	entries := []*widget.Entry{s.consumption1Entry, s.consumption2Entry, s.totalAmountEntry}
	for _, current := range entries {
		// A variável local garante que cada closure limpe a validação da entrada
		// que originou seu próprio evento.
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

// calculate coordena o caso de uso da tela. A GUI coleta textos, delega parsing
// e validação, chama a regra de rateio, formata a saída e só então altera o
// estado visual; ela não reimplementa regras dos pacotes de domínio.
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
	// Esta conferência defensiva impede exibir ou salvar um rateio cujo
	// arredondamento não reconcilie exatamente com o valor da conta.
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
	// O histórico recebe os mesmos textos apresentados ao usuário. Assim, uma
	// futura leitura reproduz a visualização sem recalcular valores antigos.
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

// clear restaura o formulário ao estado inicial e devolve o foco ao primeiro campo.
func (s *screen) clear() {
	s.consumption1Entry.SetText("")
	s.consumption2Entry.SetText("")
	s.totalAmountEntry.SetText("")
	s.clearFieldErrors()
	s.hideFeedback()
	s.scroll.ScrollToTop()
	s.window.Canvas().Focus(s.consumption1Entry)
}

// hideFeedback invalida resultado e snapshot sempre que as entradas mudam, para
// impedir que o usuário salve um cálculo que não corresponde mais ao formulário.
func (s *screen) hideFeedback() {
	s.errorBox.Hide()
	s.resultCard.Hide()
	s.snapshot = nil
	s.saveButton.Disable()
	s.saveStatus.Hide()
}

// clearFieldErrors remove as marcações de validação dos três campos.
func (s *screen) clearFieldErrors() {
	s.consumption1Entry.SetValidationError(nil)
	s.consumption2Entry.SetValidationError(nil)
	s.totalAmountEntry.SetValidationError(nil)
}

// showValidationError traduz o erro de domínio e associa o aviso ao campo adequado.
func (s *screen) showValidationError(err error) {
	message := friendlyValidationMessage(err)
	s.showError(message, s.entryForValidationError(err))
}

// showError centraliza o estado visual de falha: oculta qualquer resultado
// anterior, exibe a mensagem e, quando possível, foca a entrada responsável.
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

// saveSnapshot persiste apenas o último cálculo ainda válido. Em caso de erro,
// mantém o botão habilitado para nova tentativa; após sucesso, desabilita-o para
// evitar salvar o mesmo snapshot duas vezes pelo mesmo resultado.
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

// loadHistory relê todo o store e reconstrói a lista visual. Antes disso, limpa
// entradas e botões antigos para que os índices de exclusão continuem alinhados
// à ordem devolvida pela persistência. O store atual faz I/O síncrono; para um
// histórico pequeno isso simplifica o fluxo, mas uma implementação maior deve
// mover leitura e regravação para fora do callback da interface.
func (s *screen) loadHistory() {
	entries, err := s.store.List()
	s.historyList.RemoveAll()
	s.historyDeleteButtons = nil
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
	for index, entry := range entries {
		s.historyList.Add(s.historyCard(entry, index))
	}
	s.historyList.Refresh()
	s.historyScroll.Refresh()
	s.historyScroll.ScrollToTop()
}

// deleteHistoryEntry delega a exclusão ao store pelo índice exibido. Após o
// sucesso, recarrega a aba inteira para refletir a nova ordem dos registros.
func (s *screen) deleteHistoryEntry(index int) {
	if err := s.store.DeleteAt(index); err != nil {
		s.showHistoryState(
			"Não foi possível excluir o rateio",
			"O registro não pôde ser excluído. Use Atualizar para tentar novamente.",
		)
		return
	}
	s.loadHistory()
}

// showHistoryState reutiliza um único cartão para estados vazio e de erro.
func (s *screen) showHistoryState(title, message string) {
	s.historyStatusCard.SetTitle(title)
	s.historyStatusLabel.SetText(message)
	s.historyStatusCard.Show()
	s.historyStatusCard.Refresh()
}

// historyCard transforma uma entrada persistida em um cartão e captura seu
// índice no callback de exclusão correspondente.
func (s *screen) historyCard(entry history.Entry, index int) fyne.CanvasObject {
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
	deleteButton := widget.NewButtonWithIcon("Excluir", theme.DeleteIcon(), func() {
		s.deleteHistoryEntry(index)
	})
	deleteButton.Importance = widget.DangerImportance
	s.historyDeleteButtons = append(s.historyDeleteButtons, deleteButton)
	actions := container.NewBorder(nil, nil, nil, deleteButton, widget.NewSeparator())
	return widget.NewCard(date, "Rateio proporcional salvo", container.NewVBox(totalLine, totalConsumption, residents, actions))
}

// historyValueLabel padroniza valores armazenados, inclusive textos mais longos.
func historyValueLabel(value string) *widget.Label {
	label := widget.NewLabelWithStyle(value, fyne.TextAlignTrailing, fyne.TextStyle{Bold: true})
	label.Wrapping = fyne.TextWrapWord
	return label
}

// historyResidentTile agrupa consumo, proporção e valor salvo de um morador.
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

// entryForValidationError converte o prefixo estável dos erros de validação na
// entrada que deve receber foco. Erros gerais de consumo apontam ao primeiro campo.
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

// friendlyValidationMessage adapta mensagens técnicas da validação para a
// linguagem da interface, sem perder o detalhe necessário para correção.
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
