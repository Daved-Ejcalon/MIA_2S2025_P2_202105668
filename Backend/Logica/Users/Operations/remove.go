package Operations

import (
	"MIA_2S2025_P1_202105668/Logica/Disk"
	"MIA_2S2025_P1_202105668/Logica/System"
	"MIA_2S2025_P1_202105668/Logica/Users"
	"MIA_2S2025_P1_202105668/Models"
	"errors"
	"strings"
)

func Remove(params map[string]string) error {
	path := params["path"]

	session := Users.GetCurrentSession()
	mountInfo, _ := Disk.GetMountInfoByID(session.MountID)
	partitionInfo, superBloque, _ := Users.GetPartitionAndSuperBlock(mountInfo)

	manager := &System.EXT2Manager{}
	manager.SetPartitionInfo(partitionInfo)
	manager.SetSuperBlock(superBloque)
	manager.SetDiskPath(mountInfo.DiskPath)

	fileManager := System.NewEXT2FileManager(manager)
	path = normalizePath(path)

	inodeNum, err := findFileInode(fileManager, path)
	if err != nil {
		return errors.New("ERROR: El archivo o carpeta no existe o no tiene permisos de escritura")
	}

	inodo, _ := readInode(fileManager, inodeNum)

	hasPermission := System.ValidateFileWritePermission(
		inodo.I_uid,
		inodo.I_gid,
		inodo.I_perm,
		session.UserID,
		session.GroupID,
	)

	if !hasPermission {
		return errors.New("ERROR: El archivo o carpeta no existe o no tiene permisos de escritura")
	}

	if inodo.I_type == Models.INODO_ARCHIVO {
		removeFile(fileManager, path, inodeNum, inodo)
		return nil
	} else if inodo.I_type == Models.INODO_DIRECTORIO {
		canDelete, _ := canDeleteDirectory(fileManager, path, session.UserID, session.GroupID)

		if !canDelete {
			return errors.New("ERROR: El archivo o carpeta no existe o no tiene permisos de escritura")
		}

		removeDirectory(fileManager, path, inodeNum, session.UserID, session.GroupID)
		return nil
	}

	return nil
}

func canDeleteDirectory(fileManager *System.EXT2FileManager, dirPath string, userID, groupID int) (bool, []string) {
	var failedItems []string

	dirInodeNum, _ := findFileInode(fileManager, dirPath)
	dirInodo, _ := readInode(fileManager, dirInodeNum)

	for i := 0; i < 12; i++ {
		if dirInodo.I_block[i] == Models.FREE_BLOCK {
			break
		}

		dirBlock, _ := readDirectoryBlock(fileManager, dirInodo.I_block[i])

		for _, entry := range dirBlock.B_content {
			if entry.B_inodo == Models.FREE_INODE {
				continue
			}

			entryName := strings.TrimRight(string(entry.B_name[:]), "\x00")

			if entryName == "." || entryName == ".." || entryName == "" {
				continue
			}

			entryInodo, _ := readInode(fileManager, entry.B_inodo)

			hasPermission := System.ValidateFileWritePermission(
				entryInodo.I_uid,
				entryInodo.I_gid,
				entryInodo.I_perm,
				userID,
				groupID,
			)

			if !hasPermission {
				failedItems = append(failedItems, dirPath+"/"+entryName)
				continue
			}

			if entryInodo.I_type == Models.INODO_DIRECTORIO {
				subDirPath := dirPath + "/" + entryName
				canDelete, subFailedItems := canDeleteDirectory(fileManager, subDirPath, userID, groupID)
				if !canDelete {
					failedItems = append(failedItems, subFailedItems...)
				}
			}
		}
	}

	return len(failedItems) == 0, failedItems
}

func removeDirectory(fileManager *System.EXT2FileManager, dirPath string, dirInodeNum int32, userID, groupID int) {
	dirInodo, _ := readInode(fileManager, dirInodeNum)

	for i := 0; i < 12; i++ {
		if dirInodo.I_block[i] == Models.FREE_BLOCK {
			break
		}

		dirBlock, _ := readDirectoryBlock(fileManager, dirInodo.I_block[i])

		for _, entry := range dirBlock.B_content {
			if entry.B_inodo == Models.FREE_INODE {
				continue
			}

			entryName := strings.TrimRight(string(entry.B_name[:]), "\x00")

			if entryName == "." || entryName == ".." || entryName == "" {
				continue
			}

			entryInodo, _ := readInode(fileManager, entry.B_inodo)

			if entryInodo.I_type == Models.INODO_DIRECTORIO {
				subDirPath := dirPath + "/" + entryName
				removeDirectory(fileManager, subDirPath, entry.B_inodo, userID, groupID)
			} else {
				filePath := dirPath + "/" + entryName
				removeFile(fileManager, filePath, entry.B_inodo, entryInodo)
			}
		}
	}

	for i := 0; i < 12; i++ {
		if dirInodo.I_block[i] != Models.FREE_BLOCK {
			updateBlockBitmap(fileManager, dirInodo.I_block[i], false)
		}
	}

	updateInodeBitmap(fileManager, dirInodeNum, false)
	removeEntryFromParent(fileManager, dirPath, dirInodeNum)
}

func removeFile(fileManager *System.EXT2FileManager, filePath string, inodeNum int32, inodo *Models.Inodo) {
	for i := 0; i < 12; i++ {
		if inodo.I_block[i] != Models.FREE_BLOCK {
			updateBlockBitmap(fileManager, inodo.I_block[i], false)
		}
	}

	updateInodeBitmap(fileManager, inodeNum, false)
	removeEntryFromParent(fileManager, filePath, inodeNum)
}

func removeEntryFromParent(fileManager *System.EXT2FileManager, itemPath string, inodeNum int32) {
	parentPath, itemName := splitPath(itemPath)

	parentInodeNum, _ := findFileInode(fileManager, parentPath)
	parentInodo, _ := readInode(fileManager, parentInodeNum)

	for i := 0; i < 12; i++ {
		if parentInodo.I_block[i] == Models.FREE_BLOCK {
			break
		}

		dirBlock, _ := readDirectoryBlock(fileManager, parentInodo.I_block[i])

		for j := 0; j < len(dirBlock.B_content); j++ {
			if dirBlock.B_content[j].B_inodo == inodeNum {
				entryName := strings.TrimRight(string(dirBlock.B_content[j].B_name[:]), "\x00")
				if entryName == itemName {
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
