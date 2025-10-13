package Operations

import (
	"MIA_2S2025_P1_202105668/Logica/System"
	"MIA_2S2025_P1_202105668/Models"
	"bytes"
	"encoding/binary"
	"errors"
	"os"
	"strings"
)

// ========== FUNCIONES AUXILIARES COMPARTIDAS ==========

func normalizePath(path string) string {
	if path == "" {
		return "/"
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return path
}

func splitPath(filePath string) (string, string) {
	filePath = strings.Trim(filePath, "/")
	parts := strings.Split(filePath, "/")

	if len(parts) <= 1 {
		return "/", parts[0]
	}

	fileName := parts[len(parts)-1]
	parentPath := "/" + strings.Join(parts[:len(parts)-1], "/")

	return parentPath, fileName
}

func findFileInode(fileManager *System.EXT2FileManager, filePath string) (int32, error) {
	if filePath == "/" {
		return Models.ROOT_INODE, nil
	}

	pathParts := strings.Split(strings.Trim(filePath, "/"), "/")
	currentInode := Models.ROOT_INODE

	for _, part := range pathParts {
		if part == "" {
			continue
		}

		inodo, err := readInode(fileManager, int32(currentInode))
		if err != nil {
			return -1, err
		}

		if inodo.I_type != Models.INODO_DIRECTORIO {
			return -1, errors.New("no es un directorio")
		}

		nextInode, err := findInDirectory(fileManager, inodo, part)
		if err != nil {
			return -1, err
		}

		currentInode = int(nextInode)
	}

	return int32(currentInode), nil
}

func findInDirectory(fileManager *System.EXT2FileManager, dirInodo *Models.Inodo, filename string) (int32, error) {
	for i := 0; i < 12; i++ {
		if dirInodo.I_block[i] == Models.FREE_BLOCK {
			break
		}

		dirBlock, err := readDirectoryBlock(fileManager, dirInodo.I_block[i])
		if err != nil {
			return -1, err
		}

		for _, entry := range dirBlock.B_content {
			if entry.B_inodo != Models.FREE_INODE {
				entryName := strings.TrimRight(string(entry.B_name[:]), "\x00")
				if entryName == filename {
					return entry.B_inodo, nil
				}
			}
		}
	}

	return -1, errors.New("archivo no encontrado")
}

func readInode(fileManager *System.EXT2FileManager, inodeNumber int32) (*Models.Inodo, error) {
	manager := fileManager.GetManager()
	diskPath := manager.GetDiskPath()
	partitionInfo := manager.GetPartitionInfo()
	superBloque := manager.GetSuperBlock()

	file, err := os.Open(diskPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	inodoPos := partitionInfo.PartStart + int64(superBloque.S_inode_start) + int64(inodeNumber*Models.INODO_SIZE)
	_, err = file.Seek(inodoPos, 0)
	if err != nil {
		return nil, err
	}

	var inodo Models.Inodo
	err = binary.Read(file, binary.LittleEndian, &inodo)
	if err != nil {
		return nil, err
	}

	return &inodo, nil
}

func readDirectoryBlock(fileManager *System.EXT2FileManager, blockNumber int32) (*Models.BloqueCarpeta, error) {
	manager := fileManager.GetManager()
	diskPath := manager.GetDiskPath()
	partitionInfo := manager.GetPartitionInfo()
	superBloque := manager.GetSuperBlock()

	file, err := os.Open(diskPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	blockPos := partitionInfo.PartStart + int64(superBloque.S_block_start) + int64(blockNumber*Models.BLOQUE_SIZE)
	_, err = file.Seek(blockPos, 0)
	if err != nil {
		return nil, err
	}

	var dirBlock Models.BloqueCarpeta
	err = binary.Read(file, binary.LittleEndian, &dirBlock)
	if err != nil {
		return nil, err
	}

	return &dirBlock, nil
}

func writeDirectoryBlock(fileManager *System.EXT2FileManager, blockNumber int32, dirBlock *Models.BloqueCarpeta) error {
	manager := fileManager.GetManager()
	diskPath := manager.GetDiskPath()
	partitionInfo := manager.GetPartitionInfo()
	superBloque := manager.GetSuperBlock()

	file, err := os.OpenFile(diskPath, os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	blockPos := partitionInfo.PartStart + int64(superBloque.S_block_start) + int64(blockNumber*Models.BLOQUE_SIZE)
	_, err = file.Seek(blockPos, 0)
	if err != nil {
		return err
	}

	buffer := new(bytes.Buffer)
	err = binary.Write(buffer, binary.LittleEndian, dirBlock)
	if err != nil {
		return err
	}

	_, err = file.Write(buffer.Bytes())
	return err
}

func updateInodeBitmap(fileManager *System.EXT2FileManager, inodeNumber int32, used bool) error {
	manager := fileManager.GetManager()
	diskPath := manager.GetDiskPath()
	partitionInfo := manager.GetPartitionInfo()
	superBloque := manager.GetSuperBlock()

	file, err := os.OpenFile(diskPath, os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Leer bitmap
	bitmapPos := partitionInfo.PartStart + int64(superBloque.S_bm_inode_start)
	_, err = file.Seek(bitmapPos, 0)
	if err != nil {
		return err
	}

	bitmapSize := int(superBloque.S_inodes_count/8 + 1)
	bitmap := make([]byte, bitmapSize)
	_, err = file.Read(bitmap)
	if err != nil {
		return err
	}

	// Actualizar bit
	if used {
		Models.SetBitmapBit(bitmap, int(inodeNumber))
	} else {
		Models.ClearBitmapBit(bitmap, int(inodeNumber))
	}

	// Escribir bitmap
	_, err = file.Seek(bitmapPos, 0)
	if err != nil {
		return err
	}

	_, err = file.Write(bitmap)
	return err
}

func updateBlockBitmap(fileManager *System.EXT2FileManager, blockNumber int32, used bool) error {
	manager := fileManager.GetManager()
	diskPath := manager.GetDiskPath()
	partitionInfo := manager.GetPartitionInfo()
	superBloque := manager.GetSuperBlock()

	file, err := os.OpenFile(diskPath, os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Leer bitmap
	bitmapPos := partitionInfo.PartStart + int64(superBloque.S_bm_block_start)
	_, err = file.Seek(bitmapPos, 0)
	if err != nil {
		return err
	}

	bitmapSize := int(superBloque.S_blocks_count/8 + 1)
	bitmap := make([]byte, bitmapSize)
	_, err = file.Read(bitmap)
	if err != nil {
		return err
	}

	// Actualizar bit
	if used {
		Models.SetBitmapBit(bitmap, int(blockNumber))
	} else {
		Models.ClearBitmapBit(bitmap, int(blockNumber))
	}

	// Escribir bitmap
	_, err = file.Seek(bitmapPos, 0)
	if err != nil {
		return err
	}

	_, err = file.Write(bitmap)
	return err
}

// writeInode escribe un inodo en el disco
func writeInode(fileManager *System.EXT2FileManager, inodeNumber int32, inodo *Models.Inodo) error {
	manager := fileManager.GetManager()
	diskPath := manager.GetDiskPath()
	partitionInfo := manager.GetPartitionInfo()
	superBloque := manager.GetSuperBlock()

	file, err := os.OpenFile(diskPath, os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	inodoPos := partitionInfo.PartStart + int64(superBloque.S_inode_start) + int64(inodeNumber*Models.INODO_SIZE)
	_, err = file.Seek(inodoPos, 0)
	if err != nil {
		return err
	}

	buffer := new(bytes.Buffer)
	err = binary.Write(buffer, binary.LittleEndian, inodo)
	if err != nil {
		return err
	}

	_, err = file.Write(buffer.Bytes())
	return err
}

// findFreeBlock busca el primer bloque libre
func findFreeBlock(fileManager *System.EXT2FileManager) (int32, error) {
	manager := fileManager.GetManager()
	diskPath := manager.GetDiskPath()
	partitionInfo := manager.GetPartitionInfo()
	superBloque := manager.GetSuperBlock()

	file, err := os.Open(diskPath)
	if err != nil {
		return -1, err
	}
	defer file.Close()

	bitmapPos := partitionInfo.PartStart + int64(superBloque.S_bm_block_start)
	_, err = file.Seek(bitmapPos, 0)
	if err != nil {
		return -1, err
	}

	bitmapSize := int(superBloque.S_blocks_count/8 + 1)
	bitmap := make([]byte, bitmapSize)
	_, err = file.Read(bitmap)
	if err != nil {
		return -1, err
	}

	freeIndex := Models.FindFreeBitmapBit(bitmap)
	if freeIndex == -1 {
		return -1, errors.New("no hay bloques libres")
	}

	return int32(freeIndex), nil
}

// findFreeInode busca el primer inodo libre
func findFreeInode(fileManager *System.EXT2FileManager) (int32, error) {
	manager := fileManager.GetManager()
	diskPath := manager.GetDiskPath()
	partitionInfo := manager.GetPartitionInfo()
	superBloque := manager.GetSuperBlock()

	file, err := os.Open(diskPath)
	if err != nil {
		return -1, err
	}
	defer file.Close()

	bitmapPos := partitionInfo.PartStart + int64(superBloque.S_bm_inode_start)
	_, err = file.Seek(bitmapPos, 0)
	if err != nil {
		return -1, err
	}

	bitmapSize := int(superBloque.S_inodes_count/8 + 1)
	bitmap := make([]byte, bitmapSize)
	_, err = file.Read(bitmap)
	if err != nil {
		return -1, err
	}

	freeIndex := Models.FindFreeBitmapBit(bitmap)
	if freeIndex == -1 {
		return -1, errors.New("no hay inodos libres")
	}

	return int32(freeIndex), nil
}

// writeFileBlock escribe un bloque de archivo
func writeFileBlock(fileManager *System.EXT2FileManager, blockNumber int32, content []byte) error {
	manager := fileManager.GetManager()
	diskPath := manager.GetDiskPath()
	partitionInfo := manager.GetPartitionInfo()
	superBloque := manager.GetSuperBlock()

	file, err := os.OpenFile(diskPath, os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	blockPos := partitionInfo.PartStart + int64(superBloque.S_block_start) + int64(blockNumber*Models.BLOQUE_SIZE)
	_, err = file.Seek(blockPos, 0)
	if err != nil {
		return err
	}

	fileBlock := Models.BloqueArchivos{}
	fileBlock.SetContent(content)

	buffer := new(bytes.Buffer)
	err = binary.Write(buffer, binary.LittleEndian, &fileBlock)
	if err != nil {
		return err
	}

	_, err = file.Write(buffer.Bytes())
	return err
}

// checkNameExistsInDirectory verifica si existe un archivo/carpeta con el nombre especificado
func checkNameExistsInDirectory(fileManager *System.EXT2FileManager, dirInodo *Models.Inodo, name string) (bool, error) {
	for i := 0; i < 12; i++ {
		if dirInodo.I_block[i] == Models.FREE_BLOCK {
			break
		}

		dirBlock, err := readDirectoryBlock(fileManager, dirInodo.I_block[i])
		if err != nil {
			return false, err
		}

		for _, entry := range dirBlock.B_content {
			if entry.B_inodo != Models.FREE_INODE {
				entryName := strings.TrimRight(string(entry.B_name[:]), "\x00")
				if entryName == "." || entryName == ".." {
					continue
				}
				if entryName == name {
					return true, nil
				}
			}
		}
	}

	return false, nil
}
