package Operations

import (
	"MIA_2S2025_P1_202105668/Logica/Disk"
	"MIA_2S2025_P1_202105668/Logica/System"
	"MIA_2S2025_P1_202105668/Logica/Users"
	"MIA_2S2025_P1_202105668/Models"
	"errors"
	"io/ioutil"
)

func Edit(params map[string]string) error {
	path := params["path"]
	contenido := params["contenido"]

	session := Users.GetCurrentSession()
	contentBytes, _ := ioutil.ReadFile(contenido)
	newContent := string(contentBytes)

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
		return errors.New("ERROR: No tiene permisos de lectura y escritura sobre el archivo")
	}

	inodo, _ := readInode(fileManager, inodeNum)

	if inodo.I_type != Models.INODO_ARCHIVO {
		return errors.New("ERROR: No tiene permisos de lectura y escritura sobre el archivo")
	}

	hasReadPermission := System.ValidateFileReadPermission(
		inodo.I_uid,
		inodo.I_gid,
		inodo.I_perm,
		session.UserID,
		session.GroupID,
	)

	hasWritePermission := System.ValidateFileWritePermission(
		inodo.I_uid,
		inodo.I_gid,
		inodo.I_perm,
		session.UserID,
		session.GroupID,
	)

	if !hasReadPermission || !hasWritePermission {
		return errors.New("ERROR: No tiene permisos de lectura y escritura sobre el archivo")
	}

	editFileContent(fileManager, inodeNum, inodo, newContent)
	return nil
}

func editFileContent(fileManager *System.EXT2FileManager, inodeNum int32, inodo *Models.Inodo, newContent string) {
	freeInodeBlocks(fileManager, inodo)
	writeMultipleBlocks(fileManager, inodo, []byte(newContent))

	inodo.I_s = int32(len(newContent))
	inodo.I_mtime = float64(Models.GetCurrentUnixTime())
	inodo.I_atime = float64(Models.GetCurrentUnixTime())

	writeInode(fileManager, inodeNum, inodo)
}

func writeMultipleBlocks(fileManager *System.EXT2FileManager, inodo *Models.Inodo, content []byte) {
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

func freeInodeBlocks(fileManager *System.EXT2FileManager, inodo *Models.Inodo) {
	for i := 0; i < 12; i++ {
		if inodo.I_block[i] != Models.FREE_BLOCK {
			updateBlockBitmap(fileManager, inodo.I_block[i], false)
			inodo.I_block[i] = Models.FREE_BLOCK
		}
	}
}
