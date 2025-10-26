package System

import (
	"MIA_2S2025_P1_202105668/Models"
	"encoding/binary"
	"fmt"
	"os"
	"strings"
)

type RecoveryManager struct {
	diskPath      string
	partitionInfo *Models.Partition
	superBloque   *Models.SuperBloque
	ext3Manager   *EXT3Manager
}

func NewRecoveryManager(diskPath string, partitionInfo *Models.Partition) *RecoveryManager {
	return &RecoveryManager{
		diskPath:      diskPath,
		partitionInfo: partitionInfo,
	}
}

func (rm *RecoveryManager) RecoverFileSystem() error {
	file, err := os.Open(rm.diskPath)
	if err != nil {
		return err
	}
	defer file.Close()

	var sb Models.SuperBloque
	file.Seek(rm.partitionInfo.PartStart, 0)
	err = binary.Read(file, binary.LittleEndian, &sb)
	if err != nil {
		return err
	}

	rm.superBloque = &sb

	if rm.superBloque.S_filesystem_type != 3 {
		return fmt.Errorf("ERROR: La partición no tiene sistema de archivos EXT3")
	}

	journalManager := NewJournalManager(rm.diskPath, rm.partitionInfo, rm.superBloque)
	entries, err := journalManager.GetJournalEntries()
	if err != nil {
		return err
	}

	if len(entries) == 0 {
		fmt.Println("No hay entradas en el journal para recuperar")
		return nil
	}

	lastFormatIndex := -1
	for i := len(entries) - 1; i >= 0; i-- {
		operation := strings.TrimRight(string(entries[i].I_operation[:]), "\x00")
		if operation == "format" || operation == "mkfs" {
			lastFormatIndex = i
			break
		}
	}

	if lastFormatIndex == -1 {
		fmt.Println("No se encontró operación de formato en el journal")
		return nil
	}

	fmt.Printf("Recuperando sistema de archivos al estado anterior al formato (entrada %d)...\n", lastFormatIndex)

	// Crear MountInfo temporal para inicializar el gestor EXT3
	tempMountInfo := &MountInfo{
		DiskPath:      rm.diskPath,
		PartitionName: string(rm.partitionInfo.PartName[:]),
		MountID:       "recovery",
		DiskLetter:    'X',
		PartNumber:    0,
	}

	// Crear EXT3Manager correctamente usando NewEXT3Manager
	rm.ext3Manager = NewEXT3Manager(tempMountInfo)
	if rm.ext3Manager == nil {
		return fmt.Errorf("error inicializando EXT3Manager para recuperación")
	}

	// Asignar el superbloque y partition info que ya tenemos
	rm.ext3Manager.EXT2Manager.superBloque = &sb
	rm.ext3Manager.EXT2Manager.partitionInfo = rm.partitionInfo
	rm.ext3Manager.journalManager = journalManager

	entriesToRecover := entries[:lastFormatIndex]

	// Formatear la partición para limpiarla antes de recuperar
	err = rm.ext3Manager.FormatPartition()
	if err != nil {
		return fmt.Errorf("error formateando partición durante recuperación: %v", err)
	}

	for i, entry := range entriesToRecover {
		operation := strings.TrimRight(string(entry.I_operation[:]), "\x00")
		path := strings.TrimRight(string(entry.I_path[:]), "\x00")
		content := strings.TrimRight(string(entry.I_content[:]), "\x00")

		fmt.Printf("[%d/%d] Recuperando operación: %s en %s\n", i+1, len(entriesToRecover), operation, path)

		err := rm.replayOperation(operation, path, content)
		if err != nil {
			fmt.Printf("  ADVERTENCIA: No se pudo recuperar operación %s: %v\n", operation, err)
		}
	}

	fmt.Println("Recuperación del sistema de archivos completada")
	return nil
}

func (rm *RecoveryManager) replayOperation(operation, path, content string) error {
	fileManager := NewEXT2FileManager(rm.ext3Manager.EXT2Manager)

	switch operation {
	case "mkdir":
		return rm.recoverMkdir(fileManager, path)
	case "mkfile":
		return rm.recoverMkfile(fileManager, path, content)
	case "remove":
		return rm.recoverRemove(fileManager, path)
	case "edit":
		return rm.recoverEdit(fileManager, path, content)
	default:
		return nil
	}
}

func (rm *RecoveryManager) recoverMkdir(fileManager *EXT2FileManager, path string) error {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 0 {
		return nil
	}

	currentPath := ""
	for _, part := range parts {
		if part == "" {
			continue
		}

		if currentPath == "" {
			currentPath = "/" + part
		} else {
			currentPath = currentPath + "/" + part
		}

		_, err := fileManager.findFileInode(currentPath)
		if err != nil {
			parentPath := "/"
			if idx := strings.LastIndex(currentPath, "/"); idx > 0 {
				parentPath = currentPath[:idx]
			}

			parentInodeNum, _ := fileManager.findFileInode(parentPath)
			freeInode, _ := fileManager.findFreeInode()
			freeBlock, _ := fileManager.findFreeBlock()

			newDirInodo := Models.Inodo{
				I_uid:   1,
				I_gid:   1,
				I_s:     64,
				I_atime: float64(Models.GetCurrentUnixTime()),
				I_ctime: float64(Models.GetCurrentUnixTime()),
				I_mtime: float64(Models.GetCurrentUnixTime()),
				I_type:  Models.INODO_DIRECTORIO,
				I_perm:  Models.SetPermissions(664),
			}

			for i := range newDirInodo.I_block {
				newDirInodo.I_block[i] = Models.FREE_BLOCK
			}
			newDirInodo.I_block[0] = freeBlock

			dirBlock := Models.BloqueCarpeta{}
			for i := range dirBlock.B_content {
				dirBlock.B_content[i].B_inodo = Models.FREE_INODE
			}
			dirBlock.B_content[0].B_inodo = freeInode
			copy(dirBlock.B_content[0].B_name[:], ".")
			dirBlock.B_content[1].B_inodo = parentInodeNum
			copy(dirBlock.B_content[1].B_name[:], "..")

			fileManager.writeInode(freeInode, &newDirInodo)
			fileManager.writeDirectoryBlock(freeBlock, &dirBlock)
			fileManager.markInodeAsUsed(freeInode)
			fileManager.markBlockAsUsed(freeBlock)
			fileManager.addEntryToDirectory(parentInodeNum, part, freeInode)
		}
	}

	return nil
}

func (rm *RecoveryManager) recoverMkfile(fileManager *EXT2FileManager, path, content string) error {
	idx := strings.LastIndex(path, "/")
	var parentPath string

	if idx <= 0 {
		parentPath = "/"
	} else {
		parentPath = path[:idx]
	}

	rm.recoverMkdir(fileManager, parentPath)
	fileManager.WriteFileContent(path, content, 1, 1, 664)

	return nil
}

func (rm *RecoveryManager) recoverRemove(fileManager *EXT2FileManager, path string) error {
	return nil
}

func (rm *RecoveryManager) recoverEdit(fileManager *EXT2FileManager, path, content string) error {
	_, err := fileManager.findFileInode(path)
	if err != nil {
		rm.recoverMkfile(fileManager, path, content)
		return nil
	}

	inodeNum, _ := fileManager.findFileInode(path)
	fileManager.overwriteFileContent(inodeNum, content)

	return nil
}
