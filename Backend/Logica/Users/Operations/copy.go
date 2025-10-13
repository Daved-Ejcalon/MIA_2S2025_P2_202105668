package Operations

import (
	"MIA_2S2025_P1_202105668/Logica/Disk"
	"MIA_2S2025_P1_202105668/Logica/System"
	"MIA_2S2025_P1_202105668/Logica/Users"
	"MIA_2S2025_P1_202105668/Models"
	"encoding/binary"
	"errors"
	"os"
	"strings"
)

func Copy(params map[string]string) error {
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

	destInodeNum, err := findFileInode(fileManager, destPath)
	if err != nil {
		return errors.New("ERROR: No existe la carpeta destino")
	}

	destInodo, _ := readInode(fileManager, destInodeNum)

	if destInodo.I_type != Models.INODO_DIRECTORIO {
		return errors.New("ERROR: No existe la carpeta destino")
	}

	hasWritePermission := System.ValidateFileWritePermission(
		destInodo.I_uid,
		destInodo.I_gid,
		destInodo.I_perm,
		session.UserID,
		session.GroupID,
	)

	if !hasWritePermission {
		return errors.New("ERROR: No tiene permisos de escritura sobre la carpeta destino")
	}

	_, sourceName := splitPath(sourcePath)
	exists, _ := checkNameExistsInDirectory(fileManager, destInodo, sourceName)

	if exists {
		return nil
	}

	if sourceInodo.I_type == Models.INODO_ARCHIVO {
		copyFile(fileManager, sourceInodeNum, sourceInodo, destInodeNum, sourceName, session.UserID, session.GroupID)
	} else {
		copyDirectory(fileManager, sourcePath, sourceInodeNum, sourceInodo, destInodeNum, sourceName, session.UserID, session.GroupID)
	}

	return nil
}

func copyFile(fileManager *System.EXT2FileManager, sourceInodeNum int32, sourceInodo *Models.Inodo, destDirInodeNum int32, fileName string, uid, gid int) {
	content, _ := readFileContent(fileManager, sourceInodo)
	newInodeNum, _ := findFreeInode(fileManager)

	newInodo := Models.Inodo{
		I_uid:   sourceInodo.I_uid,
		I_gid:   sourceInodo.I_gid,
		I_s:     sourceInodo.I_s,
		I_atime: float64(Models.GetCurrentUnixTime()),
		I_ctime: float64(Models.GetCurrentUnixTime()),
		I_mtime: float64(Models.GetCurrentUnixTime()),
		I_type:  Models.INODO_ARCHIVO,
		I_perm:  sourceInodo.I_perm,
	}

	for i := range newInodo.I_block {
		newInodo.I_block[i] = Models.FREE_BLOCK
	}

	writeMultipleBlocksForCopy(fileManager, &newInodo, content)
	writeInode(fileManager, newInodeNum, &newInodo)
	addEntryToDirectory(fileManager, destDirInodeNum, fileName, newInodeNum)
	updateInodeBitmap(fileManager, newInodeNum, true)
}

func copyDirectory(fileManager *System.EXT2FileManager, sourcePath string, sourceInodeNum int32, sourceInodo *Models.Inodo, destDirInodeNum int32, dirName string, uid, gid int) {
	newDirInodeNum, _ := findFreeInode(fileManager)
	newDirBlockNum, _ := findFreeBlock(fileManager)

	newDirInodo := Models.Inodo{
		I_uid:   sourceInodo.I_uid,
		I_gid:   sourceInodo.I_gid,
		I_s:     Models.BLOQUE_SIZE,
		I_atime: float64(Models.GetCurrentUnixTime()),
		I_ctime: float64(Models.GetCurrentUnixTime()),
		I_mtime: float64(Models.GetCurrentUnixTime()),
		I_type:  Models.INODO_DIRECTORIO,
		I_perm:  sourceInodo.I_perm,
	}

	for i := range newDirInodo.I_block {
		newDirInodo.I_block[i] = Models.FREE_BLOCK
	}
	newDirInodo.I_block[0] = newDirBlockNum

	dirBlock := Models.BloqueCarpeta{}
	for i := range dirBlock.B_content {
		dirBlock.B_content[i].B_inodo = int32(Models.FREE_INODE)
		for j := range dirBlock.B_content[i].B_name {
			dirBlock.B_content[i].B_name[j] = 0
		}
	}

	dirBlock.B_content[0].B_inodo = newDirInodeNum
	copy(dirBlock.B_content[0].B_name[:], ".")
	dirBlock.B_content[1].B_inodo = destDirInodeNum
	copy(dirBlock.B_content[1].B_name[:], "..")

	writeDirectoryBlock(fileManager, newDirBlockNum, &dirBlock)
	writeInode(fileManager, newDirInodeNum, &newDirInodo)
	addEntryToDirectory(fileManager, destDirInodeNum, dirName, newDirInodeNum)
	updateInodeBitmap(fileManager, newDirInodeNum, true)
	updateBlockBitmap(fileManager, newDirBlockNum, true)

	for i := 0; i < 12; i++ {
		if sourceInodo.I_block[i] == Models.FREE_BLOCK {
			break
		}

		sourceDirBlock, _ := readDirectoryBlock(fileManager, sourceInodo.I_block[i])

		for _, entry := range sourceDirBlock.B_content {
			if entry.B_inodo == Models.FREE_INODE {
				continue
			}

			entryName := strings.TrimRight(string(entry.B_name[:]), "\x00")

			if entryName == "." || entryName == ".." || entryName == "" {
				continue
			}

			entryInodo, err := readInode(fileManager, entry.B_inodo)
			if err != nil {
				continue
			}

			hasReadPermission := System.ValidateFileReadPermission(
				entryInodo.I_uid,
				entryInodo.I_gid,
				entryInodo.I_perm,
				uid,
				gid,
			)

			if !hasReadPermission {
				continue
			}

			if entryInodo.I_type == Models.INODO_ARCHIVO {
				copyFile(fileManager, entry.B_inodo, entryInodo, newDirInodeNum, entryName, uid, gid)
			} else if entryInodo.I_type == Models.INODO_DIRECTORIO {
				subSourcePath := sourcePath + "/" + entryName
				copyDirectory(fileManager, subSourcePath, entry.B_inodo, entryInodo, newDirInodeNum, entryName, uid, gid)
			}
		}
	}
}

func readFileContent(fileManager *System.EXT2FileManager, inodo *Models.Inodo) ([]byte, error) {
	content := make([]byte, 0, inodo.I_s)
	bytesRead := int32(0)

	for i := 0; i < 12 && i < len(inodo.I_block); i++ {
		if inodo.I_block[i] == Models.FREE_BLOCK {
			break
		}

		bytesRemaining := inodo.I_s - bytesRead
		if bytesRemaining <= 0 {
			break
		}

		blockContent, _ := readFileBlock(fileManager, inodo.I_block[i])

		bytesToTake := int32(Models.BLOQUE_SIZE)
		if bytesToTake > bytesRemaining {
			bytesToTake = bytesRemaining
		}

		content = append(content, blockContent[:bytesToTake]...)
		bytesRead += bytesToTake

		if bytesRead >= inodo.I_s {
			break
		}
	}

	return content, nil
}

func readFileBlock(fileManager *System.EXT2FileManager, blockNumber int32) ([]byte, error) {
	manager := fileManager.GetManager()
	diskPath := manager.GetDiskPath()
	partitionInfo := manager.GetPartitionInfo()
	superBloque := manager.GetSuperBlock()

	file, _ := os.Open(diskPath)
	defer file.Close()

	blockPos := partitionInfo.PartStart + int64(superBloque.S_block_start) + int64(blockNumber*Models.BLOQUE_SIZE)
	file.Seek(blockPos, 0)

	var fileBlock Models.BloqueArchivos
	binary.Read(file, binary.LittleEndian, &fileBlock)

	return fileBlock.GetContent(), nil
}

func writeMultipleBlocksForCopy(fileManager *System.EXT2FileManager, inodo *Models.Inodo, content []byte) {
	totalBytes := len(content)
	blocksNeeded := (totalBytes + Models.BLOQUE_SIZE - 1) / Models.BLOQUE_SIZE

	if blocksNeeded > 12 {
		blocksNeeded = 12
		totalBytes = 12 * Models.BLOQUE_SIZE
		content = content[:totalBytes]
	}

	for i := 0; i < blocksNeeded; i++ {
		blockNum, _ := findFreeBlock(fileManager)

		start := i * Models.BLOQUE_SIZE
		end := start + Models.BLOQUE_SIZE
		if end > len(content) {
			end = len(content)
		}

		blockContent := content[start:end]
		writeFileBlock(fileManager, blockNum, blockContent)

		inodo.I_block[i] = blockNum
		updateBlockBitmap(fileManager, blockNum, true)
	}
}

func addEntryToDirectory(fileManager *System.EXT2FileManager, dirInodeNum int32, fileName string, fileInodeNum int32) {
	dirInodo, _ := readInode(fileManager, dirInodeNum)

	for i := 0; i < 12; i++ {
		if dirInodo.I_block[i] == Models.FREE_BLOCK {
			break
		}

		dirBlock, _ := readDirectoryBlock(fileManager, dirInodo.I_block[i])

		for j := 0; j < len(dirBlock.B_content); j++ {
			if dirBlock.B_content[j].B_inodo == Models.FREE_INODE {
				dirBlock.B_content[j].B_inodo = fileInodeNum
				copy(dirBlock.B_content[j].B_name[:], fileName)
				writeDirectoryBlock(fileManager, dirInodo.I_block[i], dirBlock)
				return
			}
		}
	}

	for i := 0; i < 12; i++ {
		if dirInodo.I_block[i] == Models.FREE_BLOCK {
			newBlockNum, _ := findFreeBlock(fileManager)
			dirInodo.I_block[i] = newBlockNum

			newDirBlock := &Models.BloqueCarpeta{}
			for j := 0; j < len(newDirBlock.B_content); j++ {
				newDirBlock.B_content[j].B_inodo = Models.FREE_INODE
			}

			newDirBlock.B_content[0].B_inodo = fileInodeNum
			copy(newDirBlock.B_content[0].B_name[:], fileName)

			writeDirectoryBlock(fileManager, newBlockNum, newDirBlock)
			updateBlockBitmap(fileManager, newBlockNum, true)
			writeInode(fileManager, dirInodeNum, dirInodo)
			return
		}
	}
}
