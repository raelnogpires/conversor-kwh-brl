// Package history persiste o histórico de rateios em arquivos CSV.
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

// O cabeçalho fixa o esquema do CSV: cada posição corresponde ao campo de
// Entry na mesma posição produzida por record. A ordem e os nomes também
// funcionam como uma versão simples do formato persistido.
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

// ErrInvalidIndex indica que não existe registro de histórico no índice informado.
var ErrInvalidIndex = errors.New("history index out of range")

// Entry representa um resultado de rateio persistido. Exceto pela data, os
// valores permanecem como texto para conservar exatamente a apresentação
// produzida pela interface, incluindo casas decimais, moeda e percentuais.
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

// Store persiste registros de rateio no caminho de um arquivo CSV. Ele foi
// projetado para o fluxo de uma única janela e não implementa locking entre
// processos; chamadas concorrentes podem sobrescrever alterações umas das outras.
type Store struct {
	path string
}

// NewStore cria um Store que usa path como arquivo CSV.
func NewStore(path string) *Store {
	return &Store{path: path}
}

// Save acrescenta um registro ao histórico. Ele cria os diretórios pais e o
// arquivo quando necessário; em um arquivo vazio, grava o cabeçalho antes do
// primeiro registro para estabelecer o esquema esperado por List e DeleteAt.
// Um arquivo não vazio é considerado preexistente e não é validado antes do
// append; eventual corrupção será detectada posteriormente durante a leitura.
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
	// O fechamento é explícito para que erros tardios do sistema de arquivos
	// sejam devolvidos ao chamador, em vez de se perderem em um defer.
	if err := file.Close(); err != nil {
		return fmt.Errorf("close history file: %w", err)
	}
	return nil
}

// List lê todos os registros na ordem do CSV. A ausência do arquivo representa
// um histórico ainda não iniciado e, portanto, resulta em uma lista vazia.
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
	// A comparação exata impede interpretar arquivos com colunas ausentes,
	// reordenadas ou pertencentes a outro formato como um histórico válido.
	if !sameFields(gotHeader, header) {
		return nil, fmt.Errorf("invalid history header: got %q", gotHeader)
	}

	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("read history entries: %w", err)
	}

	entries := make([]Entry, 0, len(records))
	for index, record := range records {
		// Cada linha é validada quanto à quantidade de campos e à data; o número
		// informado no erro inclui o cabeçalho, que ocupa a primeira linha.
		entry, err := entryFromRecord(record)
		if err != nil {
			return nil, fmt.Errorf("invalid history entry on row %d: %w", index+2, err)
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

// DeleteAt remove o registro no índice informado e preserva, na mesma ordem,
// o cabeçalho e todos os demais registros CSV. O índice é apenas uma posição
// do snapshot lido pela GUI, não uma identidade estável: o chamador deve evitar
// alterações concorrentes no arquivo entre listar e excluir.
func (s *Store) DeleteAt(index int) error {
	if index < 0 {
		return fmt.Errorf("%w: %d", ErrInvalidIndex, index)
	}

	file, err := os.Open(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("%w: %d", ErrInvalidIndex, index)
	}
	if err != nil {
		return fmt.Errorf("open history file: %w", err)
	}

	info, err := file.Stat()
	if err != nil {
		_ = file.Close()
		return fmt.Errorf("inspect history file: %w", err)
	}

	reader := csv.NewReader(file)
	gotHeader, err := reader.Read()
	if errors.Is(err, io.EOF) {
		_ = file.Close()
		return errors.New("invalid history header: file is empty")
	}
	if err != nil {
		_ = file.Close()
		return fmt.Errorf("read history header: %w", err)
	}
	// A exclusão exige o mesmo esquema aceito por List para não reescrever como
	// válido um arquivo incompatível ou corrompido.
	if !sameFields(gotHeader, header) {
		_ = file.Close()
		return fmt.Errorf("invalid history header: got %q", gotHeader)
	}

	records, err := reader.ReadAll()
	if err != nil {
		_ = file.Close()
		return fmt.Errorf("read history entries: %w", err)
	}
	// O arquivo de origem é fechado explicitamente antes da substituição. Além
	// de reportar falhas, isso evita manter aberto o arquivo que será renomeado.
	if err := file.Close(); err != nil {
		return fmt.Errorf("close history file: %w", err)
	}
	if index >= len(records) {
		return fmt.Errorf("%w: %d", ErrInvalidIndex, index)
	}
	// Todos os registros são validados antes da regravação, não apenas aquele
	// que será removido, para que a operação não perpetue linhas inválidas.
	for row, record := range records {
		if _, err := entryFromRecord(record); err != nil {
			return fmt.Errorf("invalid history entry on row %d: %w", row+2, err)
		}
	}

	records = append(records[:index], records[index+1:]...)
	directory := filepath.Dir(s.path)
	// O arquivo temporário fica no mesmo diretório do histórico para que o
	// Rename ocorra no mesmo sistema de arquivos e substitua o conteúdo de uma
	// só vez, sem expor aos leitores um CSV parcialmente escrito.
	temporary, err := os.CreateTemp(directory, ".history-*.tmp")
	if err != nil {
		return fmt.Errorf("create temporary history file: %w", err)
	}
	temporaryPath := temporary.Name()
	// Se qualquer etapa falhar, remove o temporário; após o Rename, o caminho
	// antigo já não existe e a remoção torna-se inofensiva.
	defer os.Remove(temporaryPath)

	// CreateTemp usa permissões próprias e restritivas. Copiar as permissões do
	// arquivo original evita alterá-las como efeito colateral da exclusão.
	if err := temporary.Chmod(info.Mode().Perm()); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("set temporary history permissions: %w", err)
	}
	writer := csv.NewWriter(temporary)
	if err := writer.Write(gotHeader); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("write history header: %w", err)
	}
	writer.WriteAll(records)
	writer.Flush()
	if err := writer.Error(); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("write history entries: %w", err)
	}
	// Sync solicita que os bytes do novo CSV cheguem ao armazenamento antes de
	// ele substituir o arquivo vigente, reduzindo o risco de perda após sucesso.
	if err := temporary.Sync(); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("sync temporary history file: %w", err)
	}
	// Fechar antes do Rename confirma eventuais erros finais de escrita e torna
	// a troca compatível também com sistemas que não renomeiam arquivos abertos.
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("close temporary history file: %w", err)
	}
	// Rename publica o arquivo completo no caminho definitivo. Mantê-lo no mesmo
	// diretório evita uma cópia entre sistemas de arquivos; as garantias exatas
	// de atomicidade continuam dependendo do sistema operacional e do sistema de
	// arquivos.
	if err := os.Rename(temporaryPath, s.path); err != nil {
		return fmt.Errorf("replace history file: %w", err)
	}
	return nil
}

func (e Entry) record() []string {
	// A data usa RFC3339 para ter uma representação textual estável e reversível;
	// os demais campos já chegam formatados e são persistidos sem conversão.
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
	// Validar a aridade antes de acessar posições distingue uma linha malformada
	// de um valor textual vazio, que é preservado como parte do registro.
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
