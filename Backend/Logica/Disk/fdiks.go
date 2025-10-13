package Disk

import (
	"MIA_2S2025_P1_202105668/Logica/Partition"
	"MIA_2S2025_P1_202105668/Models"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"strings"
)

// Fdisk crea, elimina o modifica particiones en un disco virtual existente (primarias, extendidas o lógicas)
func Fdisk(size int64, unit string, fit string, path string, ptype string, name string, deleteMode string, add int64) error {
	unit = strings.ToUpper(unit)
	fit = strings.ToUpper(fit)
	ptype = strings.ToUpper(ptype)
	deleteMode = strings.ToUpper(deleteMode)

	// Si deleteMode está especificado, ejecutar lógica de eliminación
	if deleteMode != "" {
		return deleteFdisk(path, name, deleteMode)
	}

	// Si add está especificado, ejecutar lógica de redimensionamiento
	if add != 0 {
		return resizeFdisk(path, name, add, unit)
	}

	// Validaciones básicas de entrada para creación
	if size <= 0 {
		return fmt.Errorf("tamaño inválido")
	}
	if name == "" {
		return errors.New("nombre requerido")
	}
	if strings.TrimSpace(name) == "" {
		return errors.New("nombre inválido")
	}

	// Conversión de unidades a bytes (K por defecto)
	if unit == "" {
		unit = "K"
	}
	switch unit {
	case "B":
	case "K":
		size *= 1024
	case "M":
		size *= 1024 * 1024
	default:
		return fmt.Errorf("unidad inválida")
	}

	if fit == "" {
		fit = "WF"
	}
	switch fit {
	case "BF", "FF", "WF":
	default:
		return fmt.Errorf("fit inválido")
	}

	if ptype == "" {
		ptype = "P"
	}
	switch ptype {
	case "P", "E", "L":
	default:
		return fmt.Errorf("tipo inválido")
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("archivo no existe")
	}

	file, err := os.OpenFile(path, os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("error abriendo disco")
	}
	defer file.Close()

	var mbr Models.MBR
	file.Seek(0, 0)
	err = binary.Read(file, binary.LittleEndian, &mbr)
	if err != nil {
		return fmt.Errorf("error leyendo MBR")
	}

	// Verificar nombres duplicados
	for _, partition := range mbr.Partitions {
		if partition.PartStatus != 0 && partition.GetName() == name {
			return fmt.Errorf("nombre duplicado")
		}
	}

	// Contar particiones existentes y encontrar partición extendida
	var primaryCount, extendedCount int
	var extended *Models.Partition
	for i := range mbr.Partitions {
		p := &mbr.Partitions[i]
		if p.PartStatus != 0 {
			switch p.PartType {
			case 'P':
				primaryCount++
			case 'E':
				extendedCount++
				extended = p
			}
		}
	}

	// Validar límites de particiones según tipo (máximo 4 particiones en MBR)
	if ptype == "P" && primaryCount >= 4 {
		return fmt.Errorf("demasiadas particiones primarias")
	}

	if ptype == "P" && primaryCount >= 3 && extendedCount == 1 {
		return fmt.Errorf("límite de particiones alcanzado")
	}

	if ptype == "E" && extendedCount >= 1 {
		return fmt.Errorf("partición extendida existente")
	}

	if ptype == "L" && extendedCount == 0 {
		return fmt.Errorf("partición extendida requerida")
	}

	// Buscar slot disponible en la tabla de particiones
	slotIndex := -1
	for i, partition := range mbr.Partitions {
		if partition.PartStatus == 0 {
			slotIndex = i
			break
		}
	}
	if slotIndex == -1 {
		return fmt.Errorf("sin espacios disponibles")
	}

	// Manejar particiones lógicas usando EBR Manager
	if ptype == "L" {
		ebrMgr := Partition.NewEBRManager(path, extended)

		// Inicializar primer EBR si es necesario
		if err := ebrMgr.CreateFirstEBR(); err != nil {
			return fmt.Errorf("error inicializando EBR: %v", err)
		}

		if err := ebrMgr.AddLogicalPartition(name, size, fitToByte(fit)); err != nil {
			return fmt.Errorf("error creando partición lógica: %v", err)
		}

		return nil
	}

	// Calcular posición de inicio para particiones primarias/extendidas
	startPosition := int64(Models.GetMBRSize())
	
	// Encontrar la posición donde termina la última partición
	for _, partition := range mbr.Partitions {
		if partition.PartStatus != 0 {
			endPosition := partition.PartStart + partition.PartSize
			if endPosition > startPosition {
				startPosition = endPosition
			}
		}
	}

	// Verificar que hay espacio suficiente
	if startPosition+size > mbr.MbrSize {
		return fmt.Errorf("espacio insuficiente en el disco")
	}

	// Crear y escribir nueva partición en el MBR
	newPartition := Models.Partition{
		PartStatus:      1,
		PartType:        ptype[0],
		PartFit:         fitToByte(fit),
		PartStart:       startPosition,
		PartSize:        size,
		PartCorrelative: -1,
	}
	newPartition.SetPartitionName(name)

	mbr.Partitions[slotIndex] = newPartition

	file.Seek(0, 0)
	if err := binary.Write(file, binary.LittleEndian, &mbr); err != nil {
		return fmt.Errorf("error actualizando MBR")
	}

	return nil
}

// fitToByte convierte el algoritmo de ajuste a byte para almacenamiento
func fitToByte(fit string) byte {
	s := strings.ToUpper(fit)
	if s == "BF" || s == "B" {
		return Models.FIT_BEST
	}
	if s == "FF" || s == "F" {
		return Models.FIT_FIRST
	}
	return Models.FIT_WORST
}

// deleteFdisk elimina una partición del disco (Fast o Full)
func deleteFdisk(path string, name string, deleteMode string) error {
	// Validar modo de eliminación
	if deleteMode != "FAST" && deleteMode != "FULL" {
		return fmt.Errorf("modo de eliminación inválido: use FAST o FULL")
	}

	file, err := os.OpenFile(path, os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Leer MBR
	var mbr Models.MBR
	file.Seek(0, 0)
	err = binary.Read(file, binary.LittleEndian, &mbr)
	if err != nil {
		return err
	}

	// Buscar la partición por nombre
	partitionIndex := -1
	var targetPartition *Models.Partition
	for i := range mbr.Partitions {
		p := &mbr.Partitions[i]
		if p.PartStatus != 0 && p.GetName() == name {
			partitionIndex = i
			targetPartition = p
			break
		}
	}

	if partitionIndex == -1 {
		return fmt.Errorf("partición '%s' no existe", name)
	}

	// Si es partición extendida, eliminar particiones lógicas en cascada
	if targetPartition.PartType == 'E' {
		deleteLogicalPartitions(file, targetPartition, deleteMode)
	}

	// Modo FULL: rellenar con \0
	if deleteMode == "FULL" {
		zeros := make([]byte, targetPartition.PartSize)
		file.Seek(targetPartition.PartStart, 0)
		file.Write(zeros)
	}

	// Marcar partición como vacía en la tabla
	mbr.Partitions[partitionIndex] = Models.Partition{
		PartStatus:      0,
		PartType:        0,
		PartFit:         0,
		PartStart:       0,
		PartSize:        0,
		PartCorrelative: -1,
	}

	// Escribir MBR actualizado
	file.Seek(0, 0)
	binary.Write(file, binary.LittleEndian, &mbr)

	// Mensaje de éxito
	fmt.Printf("Partición '%s' eliminada exitosamente\n", name)

	return nil
}

// deleteLogicalPartitions elimina todas las particiones lógicas dentro de una extendida
func deleteLogicalPartitions(file *os.File, extended *Models.Partition, deleteMode string) {
	if deleteMode == "FULL" {
		zeros := make([]byte, extended.PartSize)
		file.Seek(extended.PartStart, 0)
		file.Write(zeros)
	}
}

// resizeFdisk redimensiona una partición existente (agregar o quitar espacio)
func resizeFdisk(path string, name string, add int64, unit string) error {
	unit = strings.ToUpper(unit)

	// Convertir add a bytes según la unidad
	if unit == "" {
		unit = "K"
	}
	switch unit {
	case "B":
		// add ya está en bytes
	case "K":
		add *= 1024
	case "M":
		add *= 1024 * 1024
	}

	file, err := os.OpenFile(path, os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	var mbr Models.MBR
	file.Seek(0, 0)
	err = binary.Read(file, binary.LittleEndian, &mbr)
	if err != nil {
		return err
	}

	// Buscar la partición por nombre
	partitionIndex := -1
	var targetPartition *Models.Partition
	for i := range mbr.Partitions {
		p := &mbr.Partitions[i]
		if p.PartStatus != 0 && p.GetName() == name {
			partitionIndex = i
			targetPartition = p
			break
		}
	}

	if partitionIndex == -1 {
		return errors.New("")
	}

	// Calcular nuevo tamaño
	newSize := targetPartition.PartSize + add

	// Validar que no quede espacio negativo al quitar
	if newSize <= 0 {
		return errors.New("")
	}

	// Si se está agregando espacio, validar que hay espacio libre después
	if add > 0 {
		endPosition := targetPartition.PartStart + targetPartition.PartSize
		nextStart := mbr.MbrSize // Por defecto, el final del disco

		// Buscar la siguiente partición
		for i := range mbr.Partitions {
			p := &mbr.Partitions[i]
			if p.PartStatus != 0 && p.PartStart > targetPartition.PartStart && p.PartStart < nextStart {
				nextStart = p.PartStart
			}
		}

		availableSpace := nextStart - endPosition
		if add > availableSpace {
			return errors.New("")
		}
	}

	// Actualizar el tamaño de la partición
	mbr.Partitions[partitionIndex].PartSize = newSize

	// Escribir MBR actualizado
	file.Seek(0, 0)
	binary.Write(file, binary.LittleEndian, &mbr)

	return nil
}
