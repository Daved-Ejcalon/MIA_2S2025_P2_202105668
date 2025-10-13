package System

import (
	"MIA_2S2025_P1_202105668/Models"
	"encoding/binary"
	"fmt"
	"os"
	"strings"
	"time"
)

type JournalingViewer struct {
	diskPath      string
	partitionInfo *Models.Partition
}

func NewJournalingViewer(diskPath string, partitionInfo *Models.Partition) *JournalingViewer {
	return &JournalingViewer{
		diskPath:      diskPath,
		partitionInfo: partitionInfo,
	}
}

func (jv *JournalingViewer) ShowJournal() error {
	file, err := os.Open(jv.diskPath)
	if err != nil {
		return err
	}
	defer file.Close()

	var sb Models.SuperBloque
	file.Seek(jv.partitionInfo.PartStart, 0)
	err = binary.Read(file, binary.LittleEndian, &sb)
	if err != nil {
		return err
	}

	if sb.S_filesystem_type != 3 {
		return fmt.Errorf("ERROR: La partición no tiene sistema de archivos EXT3")
	}

	journalManager := NewJournalManager(jv.diskPath, jv.partitionInfo)
	entries, err := journalManager.GetJournalEntries()
	if err != nil {
		return err
	}

	if len(entries) == 0 {
		fmt.Println("No hay transacciones registradas en el journal")
		return nil
	}

	fmt.Println("╔════════════════════════════════════════════════════════════════════════════╗")
	fmt.Println("║                         JOURNAL - TRANSACCIONES EXT3                       ║")
	fmt.Println("╠════════════════════════════════════════════════════════════════════════════╣")
	fmt.Printf("║ Total de transacciones: %-52d║\n", len(entries))
	fmt.Println("╚════════════════════════════════════════════════════════════════════════════╝")
	fmt.Println()

	for i, entry := range entries {
		operation := strings.TrimRight(string(entry.I_operation[:]), "\x00")
		path := strings.TrimRight(string(entry.I_path[:]), "\x00")
		content := strings.TrimRight(string(entry.I_content[:]), "\x00")
		date := time.Unix(int64(entry.I_date), 0)

		fmt.Println("┌────────────────────────────────────────────────────────────────────────────┐")
		fmt.Printf("│ Transacción #%-65d│\n", i+1)
		fmt.Println("├────────────────────────────────────────────────────────────────────────────┤")
		fmt.Printf("│ Operación:  %-66s│\n", operation)
		fmt.Printf("│ Ruta:       %-66s│\n", path)

		if content != "" {
			if len(content) > 60 {
				fmt.Printf("│ Contenido:  %-66s│\n", content[:60]+"...")
			} else {
				fmt.Printf("│ Contenido:  %-66s│\n", content)
			}
		} else {
			fmt.Printf("│ Contenido:  %-66s│\n", "(vacío)")
		}

		fmt.Printf("│ Fecha/Hora: %-66s│\n", date.Format("2006-01-02 15:04:05"))
		fmt.Println("└────────────────────────────────────────────────────────────────────────────┘")
		fmt.Println()
	}

	return nil
}
