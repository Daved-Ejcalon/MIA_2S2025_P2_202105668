package System

import (
	"MIA_2S2025_P1_202105668/Models"
	"encoding/binary"
	"fmt"
	"os"
)

type LossSimulator struct {
	diskPath      string
	partitionInfo *Models.Partition
	superBloque   *Models.SuperBloque
}

func NewLossSimulator(diskPath string, partitionInfo *Models.Partition) *LossSimulator {
	return &LossSimulator{
		diskPath:      diskPath,
		partitionInfo: partitionInfo,
	}
}

func (ls *LossSimulator) SimulateSystemLoss() error {
	file, err := os.OpenFile(ls.diskPath, os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	var sb Models.SuperBloque
	file.Seek(ls.partitionInfo.PartStart, 0)
	err = binary.Read(file, binary.LittleEndian, &sb)
	if err != nil {
		return err
	}

	ls.superBloque = &sb

	fmt.Println("Simulando pérdida del sistema de archivos...")
	fmt.Println("ADVERTENCIA: Esta operación destruirá todos los datos en la partición")

	err = ls.clearInodeBitmap(file)
	if err != nil {
		return err
	}
	fmt.Println("✓ Bitmap de Inodos limpiado")

	err = ls.clearBlockBitmap(file)
	if err != nil {
		return err
	}
	fmt.Println("✓ Bitmap de Bloques limpiado")

	err = ls.clearInodeArea(file)
	if err != nil {
		return err
	}
	fmt.Println("✓ Área de Inodos limpiada")

	err = ls.clearBlockArea(file)
	if err != nil {
		return err
	}
	fmt.Println("✓ Área de Bloques limpiada")

	fmt.Println("Pérdida del sistema simulada exitosamente")
	return nil
}

func (ls *LossSimulator) clearInodeBitmap(file *os.File) error {
	bitmapSize := ls.superBloque.S_inodes_count
	bitmapPos := ls.partitionInfo.PartStart + int64(ls.superBloque.S_bm_inode_start)

	zeros := make([]byte, bitmapSize)
	file.Seek(bitmapPos, 0)
	_, err := file.Write(zeros)
	return err
}

func (ls *LossSimulator) clearBlockBitmap(file *os.File) error {
	bitmapSize := ls.superBloque.S_blocks_count
	bitmapPos := ls.partitionInfo.PartStart + int64(ls.superBloque.S_bm_block_start)

	zeros := make([]byte, bitmapSize)
	file.Seek(bitmapPos, 0)
	_, err := file.Write(zeros)
	return err
}

func (ls *LossSimulator) clearInodeArea(file *os.File) error {
	inodeCount := ls.superBloque.S_inodes_count
	inodeSize := int64(Models.GetInodoSize())
	inodeAreaSize := int64(inodeCount) * inodeSize
	inodeAreaPos := ls.partitionInfo.PartStart + int64(ls.superBloque.S_inode_start)

	zeros := make([]byte, inodeAreaSize)
	file.Seek(inodeAreaPos, 0)
	_, err := file.Write(zeros)
	return err
}

func (ls *LossSimulator) clearBlockArea(file *os.File) error {
	blockCount := ls.superBloque.S_blocks_count
	blockSize := int64(Models.GetBloqueSize())
	blockAreaSize := int64(blockCount) * blockSize
	blockAreaPos := ls.partitionInfo.PartStart + int64(ls.superBloque.S_block_start)

	zeros := make([]byte, blockAreaSize)
	file.Seek(blockAreaPos, 0)
	_, err := file.Write(zeros)
	return err
}
