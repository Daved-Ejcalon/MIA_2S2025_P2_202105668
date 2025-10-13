package Operations

import (
	"MIA_2S2025_P1_202105668/Logica/Disk"
	"MIA_2S2025_P1_202105668/Logica/System"
	"MIA_2S2025_P1_202105668/Logica/Users"
	"MIA_2S2025_P1_202105668/Models"
	"errors"
	"strings"
)

func Rename(params map[string]string) error {
	path := params["path"]
	newName := params["name"]

	if len(newName) > 11 {
		newName = newName[:11]
	}

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

	hasWritePermission := System.ValidateFileWritePermission(
		inodo.I_uid,
		inodo.I_gid,
		inodo.I_perm,
		session.UserID,
		session.GroupID,
	)

	if !hasWritePermission {
		return errors.New("ERROR: El archivo o carpeta no existe o no tiene permisos de escritura")
	}

	parentPath, currentName := splitPath(path)
	parentInodeNum, _ := findFileInode(fileManager, parentPath)
	parentInodo, _ := readInode(fileManager, parentInodeNum)

	exists, _ := checkNameExistsInDirectory(fileManager, parentInodo, newName)

	if exists {
		return errors.New("ERROR: Ya existe un archivo con el mismo nombre")
	}

	renameEntryInDirectory(fileManager, parentInodo, inodeNum, currentName, newName)
	inodo.I_mtime = float64(Models.GetCurrentUnixTime())
	writeInode(fileManager, inodeNum, inodo)

	return nil
}

func renameEntryInDirectory(fileManager *System.EXT2FileManager, dirInodo *Models.Inodo, inodeNum int32, oldName, newName string) error {
	for i := 0; i < 12; i++ {
		if dirInodo.I_block[i] == Models.FREE_BLOCK {
			break
		}

		dirBlock, _ := readDirectoryBlock(fileManager, dirInodo.I_block[i])

		for j := 0; j < len(dirBlock.B_content); j++ {
			if dirBlock.B_content[j].B_inodo == inodeNum {
				entryName := strings.TrimRight(string(dirBlock.B_content[j].B_name[:]), "\x00")
				if entryName == oldName {
					for k := range dirBlock.B_content[j].B_name {
						dirBlock.B_content[j].B_name[k] = 0
					}
					copy(dirBlock.B_content[j].B_name[:], newName)
					writeDirectoryBlock(fileManager, dirInodo.I_block[i], dirBlock)
					return nil
				}
			}
		}
	}

	return nil
}
