package Disk

import (
	"MIA_2S2025_P1_202105668/Models"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
)

// DiskInfo contiene información sobre un disco
type DiskInfo struct {
	Name              string          `json:"name"`
	Path              string          `json:"path"`
	Size              int64           `json:"size"`
	Fit               string          `json:"fit"`
	MountedPartitions int             `json:"mountedPartitions"`
	Partitions        []PartitionInfo `json:"partitions"`
}

// PartitionInfo contiene información sobre una partición
type PartitionInfo struct {
	Name      string `json:"name"`
	ID        string `json:"id"`
	Type      string `json:"type"`
	Size      int64  `json:"size"`
	Fit       string `json:"fit"`
	IsMounted bool   `json:"isMounted"`
}

// GetAllDisksInfo obtiene información de todos los discos creados
func GetAllDisksInfo() ([]DiskInfo, error) {
	var disksInfo []DiskInfo
	processedPaths := make(map[string]bool)

	// Obtener particiones montadas
	mountedPartitions := GetMountedPartitions()

	// Recorrer las particiones montadas para obtener información de los discos
	for _, mount := range mountedPartitions {
		// Evitar procesar el mismo disco múltiples veces
		if processedPaths[mount.DiskPath] {
			continue
		}
		processedPaths[mount.DiskPath] = true

		// Leer información del disco
		diskInfo, err := readDiskInfo(mount.DiskPath)
		if err != nil {
			continue // Ignorar discos con errores
		}

		// Obtener todas las particiones del disco (montadas y no montadas)
		allPartitions := getAllPartitionsInfo(mount.DiskPath, mountedPartitions)

		// Contar particiones montadas
		mountedCount := 0
		for _, p := range allPartitions {
			if p.IsMounted {
				mountedCount++
			}
		}

		diskInfo.MountedPartitions = mountedCount
		diskInfo.Partitions = allPartitions
		disksInfo = append(disksInfo, diskInfo)
	}

	return disksInfo, nil
}

// readDiskInfo lee información básica del MBR del disco
func readDiskInfo(diskPath string) (DiskInfo, error) {
	file, err := os.Open(diskPath)
	if err != nil {
		return DiskInfo{}, err
	}
	defer file.Close()

	var mbr Models.MBR
	err = binary.Read(file, binary.LittleEndian, &mbr)
	if err != nil {
		return DiskInfo{}, err
	}

	// Obtener información del archivo
	fileInfo, err := file.Stat()
	if err != nil {
		return DiskInfo{}, err
	}

	// Convertir fit a string legible
	fitStr := "WF" // Default
	switch mbr.DiskFit {
	case 'F':
		fitStr = "FF (First Fit)"
	case 'B':
		fitStr = "BF (Best Fit)"
	case 'W':
		fitStr = "WF (Worst Fit)"
	}

	return DiskInfo{
		Name:              filepath.Base(diskPath),
		Path:              diskPath,
		Size:              fileInfo.Size(),
		Fit:               fitStr,
		MountedPartitions: 0, // Se actualizará después
		Partitions:        []PartitionInfo{},
	}, nil
}

// getPartitionSize obtiene el tamaño de una partición específica
func getPartitionSize(diskPath string, partitionName string) int64 {
	file, err := os.Open(diskPath)
	if err != nil {
		return 0
	}
	defer file.Close()

	var mbr Models.MBR
	err = binary.Read(file, binary.LittleEndian, &mbr)
	if err != nil {
		return 0
	}

	// Buscar la partición por nombre
	for _, partition := range mbr.Partitions {
		if partition.PartStatus != 0 && partition.GetName() == partitionName {
			return partition.PartSize
		}
	}

	return 0
}

// GetDiskInfoByPath obtiene información de un disco específico por ruta
func GetDiskInfoByPath(diskPath string) (DiskInfo, error) {
	diskInfo, err := readDiskInfo(diskPath)
	if err != nil {
		return DiskInfo{}, fmt.Errorf("error leyendo disco: %v", err)
	}

	// Contar particiones montadas
	mountedPartitions := GetMountedPartitions()
	mountedCount := 0
	var partitions []PartitionInfo

	for _, mount := range mountedPartitions {
		if mount.DiskPath == diskPath {
			mountedCount++
			partitions = append(partitions, PartitionInfo{
				Name: mount.PartitionName,
				ID:   mount.MountID,
				Type: "Primary/Logical",
				Size: getPartitionSize(diskPath, mount.PartitionName),
			})
		}
	}

	diskInfo.MountedPartitions = mountedCount
	diskInfo.Partitions = partitions

	return diskInfo, nil
}

// getAllPartitionsInfo obtiene información de todas las particiones de un disco (montadas y no montadas)
func getAllPartitionsInfo(diskPath string, mountedPartitions []MountInfo) []PartitionInfo {
	file, err := os.Open(diskPath)
	if err != nil {
		return []PartitionInfo{}
	}
	defer file.Close()

	var mbr Models.MBR
	err = binary.Read(file, binary.LittleEndian, &mbr)
	if err != nil {
		return []PartitionInfo{}
	}

	var partitions []PartitionInfo

	// Crear un mapa de particiones montadas para búsqueda rápida
	mountedMap := make(map[string]MountInfo)
	for _, mount := range mountedPartitions {
		if mount.DiskPath == diskPath {
			mountedMap[mount.PartitionName] = mount
		}
	}

	// Recorrer las particiones del MBR
	for _, partition := range mbr.Partitions {
		if partition.PartStatus == 0 {
			continue // Partición no activa
		}

		partName := partition.GetName()
		if partName == "" {
			continue
		}

		// Determinar tipo
		partType := "Primary"
		if partition.PartType == 'E' {
			partType = "Extended"
		} else if partition.PartType == 'L' {
			partType = "Logical"
		}

		// Determinar fit
		fitStr := "WF"
		switch partition.PartFit {
		case 'F':
			fitStr = "FF (First Fit)"
		case 'B':
			fitStr = "BF (Best Fit)"
		case 'W':
			fitStr = "WF (Worst Fit)"
		}

		// Verificar si está montada
		mountInfo, isMounted := mountedMap[partName]
		mountID := ""
		if isMounted {
			mountID = mountInfo.MountID
		}

		partitions = append(partitions, PartitionInfo{
			Name:      partName,
			ID:        mountID,
			Type:      partType,
			Size:      partition.PartSize,
			Fit:       fitStr,
			IsMounted: isMounted,
		})

		// Si es extendida, procesar particiones lógicas (EBR)
		if partition.PartType == 'E' {
			logicalPartitions := getLogicalPartitionsInfo(file, partition.PartStart, diskPath, mountedMap)
			partitions = append(partitions, logicalPartitions...)
		}
	}

	return partitions
}

// getLogicalPartitionsInfo obtiene información de particiones lógicas desde EBRs
func getLogicalPartitionsInfo(file *os.File, extendedStart int64, diskPath string, mountedMap map[string]MountInfo) []PartitionInfo {
	var logicalPartitions []PartitionInfo
	currentEBRPos := extendedStart

	for currentEBRPos != -1 {
		file.Seek(currentEBRPos, 0)

		var ebr Models.EBR
		err := binary.Read(file, binary.LittleEndian, &ebr)
		if err != nil || ebr.PartMount == 0 {
			break
		}

		partName := ebr.GetLogicalPartitionName()
		if partName != "" {
			// Determinar fit
			fitStr := "WF"
			switch ebr.PartFit {
			case 'F':
				fitStr = "FF (First Fit)"
			case 'B':
				fitStr = "BF (Best Fit)"
			case 'W':
				fitStr = "WF (Worst Fit)"
			}

			// Verificar si está montada
			mountInfo, isMounted := mountedMap[partName]
			mountID := ""
			if isMounted {
				mountID = mountInfo.MountID
			}

			logicalPartitions = append(logicalPartitions, PartitionInfo{
				Name:      partName,
				ID:        mountID,
				Type:      "Logical",
				Size:      ebr.PartS,
				Fit:       fitStr,
				IsMounted: isMounted,
			})
		}

		// Siguiente EBR
		if ebr.PartNext == -1 {
			break
		}
		currentEBRPos = ebr.PartNext
	}

	return logicalPartitions
}
