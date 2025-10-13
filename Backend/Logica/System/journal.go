package System

import (
	"MIA_2S2025_P1_202105668/Models"
	"encoding/binary"
	"os"
)

// JournalManager gestiona las operaciones del journal de EXT3
type JournalManager struct {
	diskPath      string
	partitionInfo *Models.Partition
	journal       *Models.Journal
	journalPos    int64
}

// NewJournalManager crea un nuevo gestor de journal
func NewJournalManager(diskPath string, partitionInfo *Models.Partition) *JournalManager {
	journalPos := partitionInfo.PartStart + int64(Models.SUPERBLOQUE_SIZE)

	return &JournalManager{
		diskPath:      diskPath,
		partitionInfo: partitionInfo,
		journalPos:    journalPos,
	}
}

// InitializeJournal inicializa un nuevo journal en el disco
func (jm *JournalManager) InitializeJournal() error {
	journal := Models.NewJournal()
	jm.journal = &journal

	return jm.WriteJournal()
}

// LoadJournal carga el journal desde el disco
func (jm *JournalManager) LoadJournal() error {
	file, err := os.Open(jm.diskPath)
	if err != nil {
		return err
	}
	defer file.Close()

	var journal Models.Journal
	_, err = file.Seek(jm.journalPos, 0)
	if err != nil {
		return err
	}
	err = binary.Read(file, binary.LittleEndian, &journal)
	if err != nil {
		return err
	}

	jm.journal = &journal
	return nil
}

// WriteJournal escribe el journal al disco
func (jm *JournalManager) WriteJournal() error {
	file, err := os.OpenFile(jm.diskPath, os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Seek(jm.journalPos, 0)
	if err != nil {
		return err
	}
	err = binary.Write(file, binary.LittleEndian, jm.journal)
	if err != nil {
		return err
	}

	// Forzar escritura al disco
	err = file.Sync()
	if err != nil {
		return err
	}

	return nil
}

// LogOperation registra una operacion en el journal
func (jm *JournalManager) LogOperation(operation string, path string, content string) error {
	if jm.journal == nil {
		if err := jm.LoadJournal(); err != nil {
			return err
		}
	}

	// Crear nueva entrada de information
	info := Models.NewInformation(operation, path, content)

	// Agregar al journal
	index := jm.journal.J_count

	// Verificar que no se exceda el limite de entradas
	if index >= Models.JOURNAL_MAX_ENTRIES {
		// Si esta lleno, se puede implementar rotacion o error
		return nil
	}

	jm.journal.J_content[index] = info
	jm.journal.J_count++

	// Escribir journal actualizado al disco
	return jm.WriteJournal()
}

// GetJournalEntries retorna todas las entradas del journal
func (jm *JournalManager) GetJournalEntries() ([]Models.Information, error) {
	if jm.journal == nil {
		if err := jm.LoadJournal(); err != nil {
			return nil, err
		}
	}

	// Validar que J_count este dentro de los limites validos
	if jm.journal.J_count < 0 {
		jm.journal.J_count = 0
	}
	if jm.journal.J_count > Models.JOURNAL_MAX_ENTRIES {
		jm.journal.J_count = Models.JOURNAL_MAX_ENTRIES
	}

	// Filtrar solo las entradas validas (con fecha != 0)
	var validEntries []Models.Information
	for i := int32(0); i < jm.journal.J_count; i++ {
		entry := jm.journal.J_content[i]
		// Una entrada es valida si tiene fecha diferente de 0
		if entry.I_date != 0 {
			validEntries = append(validEntries, entry)
		}
	}

	return validEntries, nil
}

// ClearJournal limpia todas las entradas del journal
func (jm *JournalManager) ClearJournal() error {
	journal := Models.NewJournal()
	jm.journal = &journal

	return jm.WriteJournal()
}

// GetJournalCount retorna el numero de entradas en el journal
func (jm *JournalManager) GetJournalCount() int32 {
	if jm.journal == nil {
		jm.LoadJournal()
	}

	if jm.journal != nil {
		return jm.journal.J_count
	}

	return 0
}
