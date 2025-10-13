package Operations

import (
	"MIA_2S2025_P1_202105668/Logica/Disk"
	"MIA_2S2025_P1_202105668/Logica/System"
	"MIA_2S2025_P1_202105668/Logica/Users"
	"MIA_2S2025_P1_202105668/Models"
	"errors"
	"strings"
)

func Chown(params map[string]string) error {
	path := params["path"]
	newOwner := params["usuario"]
	_, isRecursive := params["r"]

	session := Users.GetCurrentSession()
	mountInfo, _ := Disk.GetMountInfoByID(session.MountID)
	partitionInfo, superBloque, _ := Users.GetPartitionAndSuperBlock(mountInfo)

	userManager := Users.NewUserManager(mountInfo.DiskPath, partitionInfo, superBloque)
	records, _ := userManager.ReadUsersFile()
	newOwnerUser := userManager.FindUserByName(records, newOwner)
	if newOwnerUser == nil {
		return errors.New("ERROR: El usuario no existe")
	}

	manager := &System.EXT2Manager{}
	manager.SetPartitionInfo(partitionInfo)
	manager.SetSuperBlock(superBloque)
	manager.SetDiskPath(mountInfo.DiskPath)
	fileManager := System.NewEXT2FileManager(manager)

	path = normalizePath(path)
	inodeNum, err := findFileInode(fileManager, path)
	if err != nil {
		return errors.New("ERROR: La ruta no existe")
	}

	inodo, _ := readInode(fileManager, inodeNum)

	if session.UserID != 1 && int32(session.UserID) != inodo.I_uid {
		return nil
	}

	if isRecursive && inodo.I_type == Models.INODO_DIRECTORIO {
		chownRecursive(fileManager, path, inodeNum, int32(newOwnerUser.ID), session.UserID)
	} else {
		changeOwner(fileManager, inodeNum, inodo, int32(newOwnerUser.ID))
	}

	return nil
}

func changeOwner(fileManager *System.EXT2FileManager, inodeNum int32, inodo *Models.Inodo, newOwnerID int32) {
	inodo.I_uid = newOwnerID
	inodo.I_ctime = float64(Models.GetCurrentUnixTime())
	writeInode(fileManager, inodeNum, inodo)
}

func chownRecursive(fileManager *System.EXT2FileManager, currentPath string, currentInodeNum int32, newOwnerID int32, sessionUserID int) {
	currentInodo, _ := readInode(fileManager, currentInodeNum)
	canChange := sessionUserID == 1 || int32(sessionUserID) == currentInodo.I_uid

	if canChange {
		changeOwner(fileManager, currentInodeNum, currentInodo, newOwnerID)
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

				canChangeEntry := sessionUserID == 1 || int32(sessionUserID) == entryInodo.I_uid

				if canChangeEntry {
					if entryInodo.I_type == Models.INODO_DIRECTORIO {
						chownRecursive(fileManager, entryPath, entry.B_inodo, newOwnerID, sessionUserID)
					} else {
						changeOwner(fileManager, entry.B_inodo, entryInodo, newOwnerID)
					}
				}
			}
		}
	}
}
