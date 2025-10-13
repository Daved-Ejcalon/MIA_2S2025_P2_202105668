package Operations

import (
	"MIA_2S2025_P1_202105668/Logica/Disk"
	"MIA_2S2025_P1_202105668/Logica/System"
	"MIA_2S2025_P1_202105668/Logica/Users"
	"MIA_2S2025_P1_202105668/Models"
	"errors"
	"strings"
)

func Move(params map[string]string) error {
	sourcePath := params["path"]
	destPath := params["destino"]

	session := Users.GetCurrentSession()
	mountInfo, _ := Disk.GetMountInfoByID(session.MountID)
	partitionInfo, superBloque, _ := Users.GetPartitionAndSuperBlock(mountInfo)

	manager := &System.EXT2Manager{}
	manager.SetPartitionInfo(partitionInfo)
	manager.SetSuperBlock(superBloque)
	manager.SetDiskPath(mountInfo.DiskPath)

	fileManager := System.NewEXT2FileManager(manager)
	sourcePath = normalizePath(sourcePath)
	destPath = normalizePath(destPath)

	sourceInodeNum, err := findFileInode(fileManager, sourcePath)
	if err != nil {
		return errors.New("ERROR: No existe la ruta")
	}

	sourceInodo, _ := readInode(fileManager, sourceInodeNum)

	hasWritePermission := System.ValidateFileWritePermission(
		sourceInodo.I_uid,
		sourceInodo.I_gid,
		sourceInodo.I_perm,
		session.UserID,
		session.GroupID,
	)

	if !hasWritePermission {
		return errors.New("ERROR: No tiene permisos de escritura")
	}

	destInodeNum, err := findFileInode(fileManager, destPath)
	if err != nil {
		return errors.New("ERROR: La carpeta destino no existe")
	}

	destInodo, _ := readInode(fileManager, destInodeNum)

	if destInodo.I_type != Models.INODO_DIRECTORIO {
		return errors.New("ERROR: La carpeta destino no existe")
	}

	hasWritePermissionDest := System.ValidateFileWritePermission(
		destInodo.I_uid,
		destInodo.I_gid,
		destInodo.I_perm,
		session.UserID,
		session.GroupID,
	)

	if !hasWritePermissionDest {
		return errors.New("ERROR: No tiene permisos de escritura sobre la carpeta destino")
	}

	sourceParentPath, sourceName := splitPath(sourcePath)
	sourceParentInodeNum, _ := findFileInode(fileManager, sourceParentPath)

	exists, _ := checkNameExistsInDirectory(fileManager, destInodo, sourceName)
	if exists {
		return nil
	}

	removeEntryFromParentDirectory(fileManager, sourceParentInodeNum, sourceInodeNum, sourceName)
	addEntryToDestinationDirectory(fileManager, destInodeNum, sourceName, sourceInodeNum)

	if sourceInodo.I_type == Models.INODO_DIRECTORIO {
		updateParentReference(fileManager, sourceInodo, destInodeNum)
	}

	sourceInodo.I_mtime = float64(Models.GetCurrentUnixTime())
	writeInode(fileManager, sourceInodeNum, sourceInodo)

	return nil
}

func removeEntryFromParentDirectory(fileManager *System.EXT2FileManager, parentInodeNum int32, targetInodeNum int32, targetName string) {
	parentInodo, _ := readInode(fileManager, parentInodeNum)

	for i := 0; i < 12; i++ {
		if parentInodo.I_block[i] == Models.FREE_BLOCK {
			break
		}

		dirBlock, _ := readDirectoryBlock(fileManager, parentInodo.I_block[i])

		for j := 0; j < len(dirBlock.B_content); j++ {
			if dirBlock.B_content[j].B_inodo == targetInodeNum {
				entryName := strings.TrimRight(string(dirBlock.B_content[j].B_name[:]), "\x00")
				if entryName == targetName {
					dirBlock.B_content[j].B_inodo = int32(Models.FREE_INODE)
					for k := range dirBlock.B_content[j].B_name {
						dirBlock.B_content[j].B_name[k] = 0
					}

					writeDirectoryBlock(fileManager, parentInodo.I_block[i], dirBlock)
					return
				}
			}
		}
	}
}

func addEntryToDestinationDirectory(fileManager *System.EXT2FileManager, destInodeNum int32, fileName string, fileInodeNum int32) {
	destInodo, _ := readInode(fileManager, destInodeNum)

	for i := 0; i < 12; i++ {
		if destInodo.I_block[i] == Models.FREE_BLOCK {
			break
		}

		dirBlock, _ := readDirectoryBlock(fileManager, destInodo.I_block[i])

		for j := 0; j < len(dirBlock.B_content); j++ {
			if dirBlock.B_content[j].B_inodo == Models.FREE_INODE {
				dirBlock.B_content[j].B_inodo = fileInodeNum
				copy(dirBlock.B_content[j].B_name[:], fileName)
				writeDirectoryBlock(fileManager, destInodo.I_block[i], dirBlock)
				return
			}
		}
	}

	for i := 0; i < 12; i++ {
		if destInodo.I_block[i] == Models.FREE_BLOCK {
			newBlockNum, _ := findFreeBlock(fileManager)
			destInodo.I_block[i] = newBlockNum

			newDirBlock := &Models.BloqueCarpeta{}
			for j := 0; j < len(newDirBlock.B_content); j++ {
				newDirBlock.B_content[j].B_inodo = Models.FREE_INODE
			}

			newDirBlock.B_content[0].B_inodo = fileInodeNum
			copy(newDirBlock.B_content[0].B_name[:], fileName)

			writeDirectoryBlock(fileManager, newBlockNum, newDirBlock)
			updateBlockBitmap(fileManager, newBlockNum, true)
			writeInode(fileManager, destInodeNum, destInodo)
			return
		}
	}
}

func updateParentReference(fileManager *System.EXT2FileManager, dirInodo *Models.Inodo, newParentInodeNum int32) {
	if dirInodo.I_block[0] == Models.FREE_BLOCK {
		return
	}

	dirBlock, _ := readDirectoryBlock(fileManager, dirInodo.I_block[0])

	for i := 0; i < len(dirBlock.B_content); i++ {
		entryName := strings.TrimRight(string(dirBlock.B_content[i].B_name[:]), "\x00")
		if entryName == ".." {
			dirBlock.B_content[i].B_inodo = newParentInodeNum
			writeDirectoryBlock(fileManager, dirInodo.I_block[0], dirBlock)
			return
		}
	}
}
