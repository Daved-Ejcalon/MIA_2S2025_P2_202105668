package Disk

import (
	"MIA_2S2025_P1_202105668/Logica/System"
	"fmt"
)

// Mkfs formatea una partición montada con sistema de archivos EXT2 o EXT3
func Mkfs(mountID string, formatType string, fs string) error {
	// Validar que el ID de montaje esté presente
	if mountID == "" {
		return fmt.Errorf("parametro -id requerido")
	}

	// Buscar la partición montada por ID
	mountInfo, err := findMountedPartitionByID(mountID)
	if err != nil {
		return fmt.Errorf("particion no montada")
	}

	// Crear estructura de montaje para el sistema de archivos
	systemMountInfo := &System.MountInfo{
		DiskPath:      mountInfo.DiskPath,
		PartitionName: mountInfo.PartitionName,
		MountID:       mountInfo.MountID,
		DiskLetter:    mountInfo.DiskLetter,
		PartNumber:    mountInfo.PartNumber,
	}

	// Seleccionar el tipo de sistema de archivos
	if fs == "3fs" {
		// Formatear con EXT3
		ext3Manager := System.NewEXT3Manager(systemMountInfo)
		if ext3Manager == nil {
			return fmt.Errorf("error inicializando EXT3")
		}

		err = ext3Manager.FormatPartition()
		if err != nil {
			return err
		}

		// Registrar la operación de formato en el journal
		err = ext3Manager.LogOperation("mkfs", "format", "EXT3")
		if err != nil {
			return fmt.Errorf("error registrando operación de formato en journal: %v", err)
		}
	} else {
		// Formatear con EXT2 (default)
		ext2Manager := System.NewEXT2Manager(systemMountInfo)
		if ext2Manager == nil {
			return fmt.Errorf("error inicializando EXT2")
		}

		err = ext2Manager.FormatPartition()
		if err != nil {
			return err
		}
	}

	return nil
}

// findMountedPartitionByID busca una partición montada por su ID único
func findMountedPartitionByID(mountID string) (*MountInfo, error) {
	mountedPartitions := GetMountedPartitions()

	// Buscar en lista de particiones montadas
	for _, mount := range mountedPartitions {
		if mount.MountID == mountID {
			return &mount, nil
		}
	}

	return nil, fmt.Errorf("partición no encontrada")
}
