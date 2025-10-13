package System

import (
	"MIA_2S2025_P1_202105668/Models"
	"encoding/binary"
	"os"
)

// EXT3Manager extiende EXT2Manager con soporte de journaling
type EXT3Manager struct {
	*EXT2Manager
	journalManager *JournalManager
}

// NewEXT3Manager crea un nuevo manager para EXT3
func NewEXT3Manager(mountInfo *MountInfo) *EXT3Manager {
	ext2Manager := NewEXT2Manager(mountInfo)
	if ext2Manager == nil {
		return nil
	}

	manager := &EXT3Manager{
		EXT2Manager: ext2Manager,
	}

	return manager
}

// FormatPartition inicializa una particion con sistema de archivos EXT3
func (e *EXT3Manager) FormatPartition() error {
	// Cargar metadatos de la particion desde el MBR
	err := e.LoadPartitionInfo()
	if err != nil {
		return err
	}

	// Calcular distribucion del espacio EXT3 (similar a EXT2 + Journal)
	err = e.calculateEXT3Layout()
	if err != nil {
		return err
	}

	// Escribir SuperBloque con metadatos del sistema (tipo 3 para EXT3)
	err = e.writeSuperBloqueEXT3()
	if err != nil {
		return err
	}

	// Inicializar journal
	err = e.initializeJournal()
	if err != nil {
		return err
	}

	// Inicializar bitmaps de inodos y bloques
	err = e.initializeBitmaps()
	if err != nil {
		return err
	}

	// Crear directorio raiz con entradas . y ..
	err = e.createRootDirectory()
	if err != nil {
		return err
	}

	// Crear archivo users.txt con usuario root
	err = e.createUsersFile()
	if err != nil {
		return err
	}

	return nil
}

// calculateEXT3Layout calcula la distribucion del espacio para EXT3 con Journaling
func (e *EXT3Manager) calculateEXT3Layout() error {
	partitionSize := e.partitionInfo.PartSize

	// Obtener tamanos de estructuras
	superBlockSize := int64(Models.GetSuperBloqueSize())
	inodoSize := int64(Models.GetInodoSize())
	blockSize := int64(Models.GetBloqueSize())

	// Formula EXT3: tamaño_particion = sizeof(superblock) + n*sizeof(Journaling) + n + 3*n + n*sizeof(inodos) + 3*n*sizeof(block)
	// Donde: n = numero de inodos
	// sizeof(Journaling) = constante 50 (segun especificacion)
	// n = bitmap de inodos
	// 3*n = bitmap de bloques (3 tipos: carpetas, archivos, contenido)
	// n*sizeof(inodos) = tabla de inodos
	// 3*n*sizeof(block) = area de bloques

	// Despejando n:
	// n = (tamaño_particion - sizeof(superblock)) / (sizeof(Journaling) + 1 + 3 + sizeof(inodos) + 3*sizeof(block))

	journalingConstant := int64(50) // Constante especificada
	numerator := float64(partitionSize - superBlockSize)
	denominator := float64(journalingConstant + 1 + 3 + inodoSize + 3*blockSize)
	n := numerator / denominator

	// Calcular cantidad de inodos usando floor (minimo 1)
	inodesCount := int32(n)
	if inodesCount < 1 {
		return nil
	}

	// Relacion 3:1 bloques por inodo (3 tipos de bloques)
	blocksCount := 3 * inodesCount

	// Crear superbloque con tipo 3 (EXT3)
	e.superBloque = &Models.SuperBloque{}
	*e.superBloque = Models.NewSuperBloque(inodesCount, blocksCount)
	e.superBloque.S_filesystem_type = 3

	return nil
}

// writeSuperBloqueEXT3 escribe el superbloque con tipo EXT3
func (e *EXT3Manager) writeSuperBloqueEXT3() error {
	if e.superBloque == nil {
		return nil
	}

	// Cambiar el tipo de sistema de archivos a EXT3
	e.superBloque.S_filesystem_type = 3

	file, err := os.OpenFile(e.diskPath, os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Escribir superbloque en la posicion de inicio de la particion
	file.Seek(e.partitionInfo.PartStart, 0)
	err = binary.Write(file, binary.LittleEndian, e.superBloque)
	if err != nil {
		return err
	}

	return nil
}

// initializeJournal inicializa el journal de EXT3
func (e *EXT3Manager) initializeJournal() error {
	// Crear journal manager
	e.journalManager = NewJournalManager(e.diskPath, e.partitionInfo)

	// Inicializar journal en el disco
	return e.journalManager.InitializeJournal()
}

// LogOperation registra una operacion en el journal
func (e *EXT3Manager) LogOperation(operation string, path string, content string) error {
	if e.journalManager == nil {
		e.journalManager = NewJournalManager(e.diskPath, e.partitionInfo)
	}

	return e.journalManager.LogOperation(operation, path, content)
}

// GetJournalEntries retorna todas las entradas del journal
func (e *EXT3Manager) GetJournalEntries() ([]Models.Information, error) {
	if e.journalManager == nil {
		e.journalManager = NewJournalManager(e.diskPath, e.partitionInfo)
	}

	return e.journalManager.GetJournalEntries()
}
