package Operations

import (
	"MIA_2S2025_P1_202105668/Logica/Disk"
	"MIA_2S2025_P1_202105668/Logica/System"
	"MIA_2S2025_P1_202105668/Logica/Users"
	"MIA_2S2025_P1_202105668/Models"
	"errors"
	"strconv"
	"strings"
)

func Chmod(params map[string]string) error {
	path := params["path"]
	ugo := params["ugo"]
	_, isRecursive := params["r"]

	session := Users.GetCurrentSession()
	if session.UserID != 1 {
		return errors.New("ERROR: Solo el usuario root puede ejecutar este comando")
	}

	u, err := strconv.Atoi(string(ugo[0]))
	if err != nil || u < 0 || u > 7 {
		return errors.New("ERROR: Los permisos deben estar en el rango 0-7")
	}

	g, err := strconv.Atoi(string(ugo[1]))
	if err != nil || g < 0 || g > 7 {
		return errors.New("ERROR: Los permisos deben estar en el rango 0-7")
	}

	o, err := strconv.Atoi(string(ugo[2]))
	if err != nil || o < 0 || o > 7 {
		return errors.New("ERROR: Los permisos deben estar en el rango 0-7")
	}

	mountInfo, _ := Disk.GetMountInfoByID(session.MountID)
	partitionInfo, superBloque, _ := Users.GetPartitionAndSuperBlock(mountInfo)

	manager := &System.EXT2Manager{}
	manager.SetPartitionInfo(partitionInfo)
	manager.SetSuperBlock(superBloque)
	manager.SetDiskPath(mountInfo.DiskPath)
	fileManager := System.NewEXT2FileManager(manager)

	path = normalizePath(path)
	inodeNum, _ := findFileInode(fileManager, path)
	inodo, _ := readInode(fileManager, inodeNum)

	if !isRecursive && int32(session.UserID) != inodo.I_uid {
		return errors.New("ERROR: El archivo no pertenece al usuario actual")
	}

	perms := int32(u*100 + g*10 + o)

	if isRecursive && inodo.I_type == Models.INODO_DIRECTORIO {
		chmodRecursive(fileManager, path, inodeNum, perms, session.UserID)
	} else {
		changePermissions(fileManager, inodeNum, inodo, perms)
	}

	return nil
}

func changePermissions(fileManager *System.EXT2FileManager, inodeNum int32, inodo *Models.Inodo, perms int32) {
	inodo.I_perm = Models.SetPermissions(perms)
	inodo.I_ctime = float64(Models.GetCurrentUnixTime())
	writeInode(fileManager, inodeNum, inodo)
}

func chmodRecursive(fileManager *System.EXT2FileManager, currentPath string, currentInodeNum int32, perms int32, sessionUserID int) {
	currentInodo, _ := readInode(fileManager, currentInodeNum)

	if int32(sessionUserID) == currentInodo.I_uid || sessionUserID == 1 {
		changePermissions(fileManager, currentInodeNum, currentInodo, perms)
	}

	if currentInodo.I_type == Models.INODO_DIRECTORIO {
		for i := 0; i < 12; i++ {
			if currentInodo.I_block[i] == Models.FREE_BLOCK {
				break
			}

			dirBlock, err := readDirectoryBlock(fileManager, currentInodo.I_block[i])
			if err != nil {
				continue
			}

			for _, entry := range dirBlock.B_content {
				if entry.B_inodo == Models.FREE_INODE {
					continue
				}

				entryName := strings.TrimRight(string(entry.B_name[:]), "\x00")

				if entryName == "." || entryName == ".." || entryName == "" {
					continue
				}

				entryInodo, _ := readInode(fileManager, entry.B_inodo)

				var entryPath string
				if currentPath == "/" {
					entryPath = "/" + entryName
				} else {
					entryPath = currentPath + "/" + entryName
				}

				if int32(sessionUserID) == entryInodo.I_uid || sessionUserID == 1 {
					if entryInodo.I_type == Models.INODO_DIRECTORIO {
						chmodRecursive(fileManager, entryPath, entry.B_inodo, perms, sessionUserID)
					} else {
						changePermissions(fileManager, entry.B_inodo, entryInodo, perms)
					}
				}
			}
		}
	}
}
