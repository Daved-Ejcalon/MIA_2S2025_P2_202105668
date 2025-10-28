package main

import (
	"MIA_2S2025_P1_202105668/Logica/Disk"
	"MIA_2S2025_P1_202105668/Logica/Reportes"
	"MIA_2S2025_P1_202105668/Logica/System"
	"MIA_2S2025_P1_202105668/Logica/Users"
	"MIA_2S2025_P1_202105668/Logica/Users/Comandos"
	"MIA_2S2025_P1_202105668/Logica/Users/Operations"
	"MIA_2S2025_P1_202105668/Logica/Users/Root"
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func main() {
	// Verificar si se debe ejecutar como servidor web
	if len(os.Args) > 1 && os.Args[1] == "server" {
		startServer()
		return
	}

	// Modo consola tradicional
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("MIA> ")

		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())

		// Ignorar líneas que empiecen con # (comentarios)
		if strings.HasPrefix(input, "#") || input == "" {
			continue
		}

		// Remover comentarios inline (después del comando)
		if commentIndex := strings.Index(input, "#"); commentIndex != -1 {
			input = strings.TrimSpace(input[:commentIndex])
		}

		// Si después de remover comentarios queda vacío, continuar
		if input == "" {
			continue
		}

		if input == "exit" {
			fmt.Println("saliendo del sistema...")
			break
		}

		func() {
			defer func() {
				if r := recover(); r != nil {
					fmt.Printf("PANIC: %v\n", r)
				}
			}()
			err := processCommand(input)
			if err != nil {
				fmt.Printf("error: %s\n", err.Error())
			}
		}()
	}
}

func processCommand(input string) error {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return fmt.Errorf("comando vacio")
	}

	command := strings.ToLower(parts[0])
	params := parseParameters(parts[1:])

	switch command {
	case "mkdisk":
		return processMkdisk(params)
	case "rmdisk":
		return processRmdisk(params)
	case "fdisk":
		return processFdisk(params)
	case "mount":
		return processMount(params)
	case "unmount":
		return processUnmount(params)
	case "mounted":
		Disk.Mounted()
		return nil
	case "mkfs":
		return processMkfs(params)
	case "cat":
		return processCat(params)
	case "showdisk":
		return Disk.ShowDisk(params)
	case "login":
		return Users.Login(params)
	case "logout":
		return Users.Logout()
	case "mkgrp":
		return Comandos.MkGrp(params)
	case "rmgrp":
		return Comandos.RmGrp(params)
	case "mkusr":
		return Comandos.MkUsr(params)
	case "rmusr":
		return Comandos.RmUsr(params)
	case "chgrp":
		return Comandos.ChGrp(params)
	case "mkdir":
		return Root.MkDir(params)
	case "mkfile":
		return Root.MkFile(params)
	case "remove":
		return Operations.Remove(params)
	case "edit":
		return Operations.Edit(params)
	case "rename":
		return Operations.Rename(params)
	case "copy":
		return Operations.Copy(params)
	case "move":
		return Operations.Move(params)
	case "find":
		return Operations.Find(params)
	case "chown":
		return Operations.Chown(params)
	case "chmod":
		return Operations.Chmod(params)
	case "recovery":
		return processRecovery(params)
	case "loss":
		return processLoss(params)
	case "journaling":
		return processJournaling(params)
	case "rep":
		return processRep(params)
	default:
		return fmt.Errorf("comando '%s' no reconocido", command)
	}
}

func processMkdisk(params map[string]string) error {
	// Validar que solo se usen parámetros permitidos
	validParams := map[string]bool{
		"size": true,
		"unit": true,
		"fit":  true,
		"path": true,
	}

	for param := range params {
		if !validParams[param] {
			return fmt.Errorf("parametro -%s no es valido para mkdisk", param)
		}
	}

	sizeStr, hasSize := params["size"]
	if !hasSize {
		return fmt.Errorf("parametro -size requerido")
	}

	size, err := strconv.ParseInt(sizeStr, 10, 64)
	if err != nil {
		return fmt.Errorf("size invalido: %v", err)
	}

	unit := params["unit"]
	if unit == "" {
		unit = "M"
	}

	fit := params["fit"]
	if fit == "" {
		fit = "WF"
	}

	path := params["path"]
	if path == "" {
		return fmt.Errorf("parametro -path requerido")
	}

	return Disk.MkDisk(size, unit, fit, path)
}

func processRmdisk(params map[string]string) error {
	// Validar que solo se usen parámetros permitidos
	validParams := map[string]bool{
		"path": true,
	}

	for param := range params {
		if !validParams[param] {
			return fmt.Errorf("parametro -%s no es valido para rmdisk", param)
		}
	}

	path := params["path"]
	if path == "" {
		return fmt.Errorf("parametro -path requerido")
	}

	return Disk.RmDisk(path)
}

func processFdisk(params map[string]string) error {
	// Validar que solo se usen parámetros permitidos
	validParams := map[string]bool{
		"size":   true,
		"unit":   true,
		"fit":    true,
		"path":   true,
		"type":   true,
		"name":   true,
		"delete": true,
		"add":    true,
	}

	for param := range params {
		if !validParams[param] {
			return fmt.Errorf("parametro -%s no es valido para fdisk", param)
		}
	}

	// Verificar si es operación de eliminación
	deleteMode := params["delete"]

	// Si es eliminación, solo requiere path, name y delete
	if deleteMode != "" {
		path := params["path"]
		if path == "" {
			return fmt.Errorf("parametro -path requerido")
		}

		name := params["name"]
		if name == "" {
			return fmt.Errorf("parametro -name requerido")
		}

		return Disk.Fdisk(0, "", "", path, "", name, deleteMode, 0)
	}

	// Verificar si es operación de redimensionamiento
	addStr := params["add"]
	if addStr != "" {
		path := params["path"]
		if path == "" {
			return fmt.Errorf("parametro -path requerido")
		}

		name := params["name"]
		if name == "" {
			return fmt.Errorf("parametro -name requerido")
		}

		add, err := strconv.ParseInt(addStr, 10, 64)
		if err != nil {
			return fmt.Errorf("add invalido: %v", err)
		}

		unit := params["unit"]
		if unit == "" {
			unit = "K"
		}

		return Disk.Fdisk(0, unit, "", path, "", name, "", add)
	}

	// Operación de creación - validar parámetros requeridos
	sizeStr, hasSize := params["size"]
	if !hasSize {
		return fmt.Errorf("parametro -size requerido")
	}

	size, err := strconv.ParseInt(sizeStr, 10, 64)
	if err != nil {
		return fmt.Errorf("size invalido: %v", err)
	}

	unit := params["unit"]
	if unit == "" {
		unit = "K"
	}

	fit := params["fit"]
	if fit == "" {
		fit = "WF"
	}

	path := params["path"]
	if path == "" {
		return fmt.Errorf("parametro -path requerido")
	}

	ptype := params["type"]
	if ptype == "" {
		ptype = "P"
	}

	name := params["name"]
	if name == "" {
		return fmt.Errorf("parametro -name requerido")
	}

	return Disk.Fdisk(size, unit, fit, path, ptype, name, "", 0)
}

func processMount(params map[string]string) error {
	// Validar que solo se usen parámetros permitidos
	validParams := map[string]bool{
		"path": true,
		"name": true,
	}

	for param := range params {
		if !validParams[param] {
			return fmt.Errorf("parametro -%s no es valido para mount", param)
		}
	}

	path := params["path"]
	if path == "" {
		return fmt.Errorf("parametro -path requerido")
	}

	name := params["name"]
	if name == "" {
		return fmt.Errorf("parametro -name requerido")
	}

	return Disk.Mount(path, name)
}

func processUnmount(params map[string]string) error {
	// Validar que solo se usen parámetros permitidos
	validParams := map[string]bool{
		"id": true,
	}

	for param := range params {
		if !validParams[param] {
			return fmt.Errorf("parametro -%s no es valido para unmount", param)
		}
	}

	id := params["id"]
	if id == "" {
		return fmt.Errorf("parametro -id requerido")
	}

	return Disk.UnmountPartition(id)
}

func processMkfs(params map[string]string) error {
	// Validar que solo se usen parámetros permitidos
	validParams := map[string]bool{
		"id":   true,
		"type": true,
		"fs":   true,
	}

	for param := range params {
		if !validParams[param] {
			return fmt.Errorf("parametro -%s no es valido para mkfs", param)
		}
	}

	id := params["id"]
	if id == "" {
		return fmt.Errorf("parametro -id requerido")
	}

	// Parámetro type (opcional, default "full")
	formatType := params["type"]
	if formatType == "" {
		formatType = "full"
	}

	// Parámetro fs (opcional, default "2fs" para EXT2)
	fs := params["fs"]
	if fs == "" {
		fs = "2fs"
	}

	return Disk.Mkfs(id, formatType, fs)
}

func processRecovery(params map[string]string) error {
	id := params["id"]
	if id == "" {
		return fmt.Errorf("parametro -id requerido")
	}

	mountInfo, err := Disk.GetMountInfoByID(id)
	if err != nil {
		return fmt.Errorf("particion con id '%s' no esta montada", id)
	}

	systemMountInfo := &System.MountInfo{
		DiskPath:      mountInfo.DiskPath,
		PartitionName: mountInfo.PartitionName,
		MountID:       mountInfo.MountID,
		DiskLetter:    mountInfo.DiskLetter,
		PartNumber:    mountInfo.PartNumber,
	}

	ext2Manager := System.NewEXT2Manager(systemMountInfo)
	if ext2Manager == nil {
		return fmt.Errorf("error inicializando gestor de archivos")
	}

	partitionInfo := ext2Manager.GetPartitionInfo()
	recoveryManager := System.NewRecoveryManager(mountInfo.DiskPath, partitionInfo)
	return recoveryManager.RecoverFileSystem()
}

func processLoss(params map[string]string) error {
	id := params["id"]
	if id == "" {
		return fmt.Errorf("parametro -id requerido")
	}

	mountInfo, err := Disk.GetMountInfoByID(id)
	if err != nil {
		return fmt.Errorf("particion con id '%s' no esta montada", id)
	}

	systemMountInfo := &System.MountInfo{
		DiskPath:      mountInfo.DiskPath,
		PartitionName: mountInfo.PartitionName,
		MountID:       mountInfo.MountID,
		DiskLetter:    mountInfo.DiskLetter,
		PartNumber:    mountInfo.PartNumber,
	}

	ext2Manager := System.NewEXT2Manager(systemMountInfo)
	if ext2Manager == nil {
		return fmt.Errorf("error inicializando gestor de archivos")
	}

	partitionInfo := ext2Manager.GetPartitionInfo()
	lossSimulator := System.NewLossSimulator(mountInfo.DiskPath, partitionInfo)
	return lossSimulator.SimulateSystemLoss()
}

func processJournaling(params map[string]string) error {
	id := params["id"]
	if id == "" {
		return fmt.Errorf("parametro -id requerido")
	}

	mountInfo, err := Disk.GetMountInfoByID(id)
	if err != nil {
		return fmt.Errorf("particion con id '%s' no esta montada", id)
	}

	systemMountInfo := &System.MountInfo{
		DiskPath:      mountInfo.DiskPath,
		PartitionName: mountInfo.PartitionName,
		MountID:       mountInfo.MountID,
		DiskLetter:    mountInfo.DiskLetter,
		PartNumber:    mountInfo.PartNumber,
	}

	ext2Manager := System.NewEXT2Manager(systemMountInfo)
	if ext2Manager == nil {
		return fmt.Errorf("error inicializando gestor de archivos")
	}

	partitionInfo := ext2Manager.GetPartitionInfo()
	journalingViewer := System.NewJournalingViewer(mountInfo.DiskPath, partitionInfo)
	return journalingViewer.ShowJournal()
}

func parseParameters(args []string) map[string]string {
	params := make(map[string]string)

	for i := 0; i < len(args); i++ {
		arg := args[i]

		if strings.HasPrefix(arg, "-") {
			if strings.Contains(arg, "=") {
				parts := strings.SplitN(arg, "=", 2)
				key := strings.TrimPrefix(parts[0], "-")
				value := strings.Trim(parts[1], "\"")
				params[key] = value
			} else {
				key := strings.TrimPrefix(arg, "-")
				params[key] = "true"
			}
		}
	}

	return params
}

func processCat(params map[string]string) error {
	// Verificar sesión activa
	session := Users.GetCurrentSession()
	if session == nil || !session.IsActive {
		return fmt.Errorf("ERROR: Debe iniciar sesión para usar este comando")
	}

	// Pasar la sesión al comando Cat
	return Disk.CatWithSession(params, session.MountID)
}

func processRep(params map[string]string) error {
	// Validar que solo se usen parámetros permitidos
	validParams := map[string]bool{
		"name":         true,
		"path":         true,
		"id":           true,
		"path_file_ls": true,
	}

	for param := range params {
		if !validParams[param] {
			return fmt.Errorf("parametro -%s no es valido para rep", param)
		}
	}

	// Validar parámetros obligatorios
	name := params["name"]
	if name == "" {
		return fmt.Errorf("parametro -name requerido")
	}

	path := params["path"]
	if path == "" {
		return fmt.Errorf("parametro -path requerido")
	}

	id := params["id"]
	if id == "" {
		return fmt.Errorf("parametro -id requerido")
	}

	// Validar valores válidos para name
	validNames := map[string]bool{
		"mbr":   true,
		"disk":  true,
		"ebr":   true,
		"inode": true,
		"sb":    true,
		"file":  true,
		"ls":    true,
	}

	if !validNames[name] {
		return fmt.Errorf("valor de -name debe ser: mbr, disk, ebr, inode, sb, file o ls")
	}

	// Llamar al generador de reportes correspondiente
	return Reportes.GenerateReport(name, id, path, params["path_file_ls"])
}

// === SERVIDOR WEB ===

type CommandRequest struct {
	Command string `json:"command"`
}

type CommandResponse struct {
	Output string `json:"output"`
	Error  string `json:"error,omitempty"`
}

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
}

func executeCommandHandler(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)

	if r.Method == "OPTIONS" {
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	var req CommandRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Error al decodificar JSON", http.StatusBadRequest)
		return
	}

	// Capturar la salida estándar
	oldStdout := os.Stdout
	r_out, w_out, _ := os.Pipe()
	os.Stdout = w_out

	// Capturar errores
	var cmdError error

	// Ejecutar el comando en una goroutine para capturar la salida
	done := make(chan bool)
	var output []byte

	go func() {
		defer func() {
			if r := recover(); r != nil {
				cmdError = fmt.Errorf("PANIC: %v", r)
			}
			done <- true
		}()

		// Procesar el comando usando la función existente
		cmdError = processCommand(strings.TrimSpace(req.Command))
	}()

	// Esperar a que termine y cerrar el pipe
	go func() {
		<-done
		w_out.Close()
	}()

	// Leer la salida
	output, _ = io.ReadAll(r_out)
	os.Stdout = oldStdout

	// Preparar la respuesta
	resp := CommandResponse{
		Output: string(output),
	}

	if cmdError != nil {
		resp.Error = cmdError.Error()
		w.WriteHeader(http.StatusBadRequest)
	}

	// Enviar respuesta JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func getDisksHandler(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)

	if r.Method == "OPTIONS" {
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	// Obtener información de todos los discos
	disks, err := Disk.GetAllDisksInfo()

	type DisksResponse struct {
		Disks []Disk.DiskInfo `json:"disks"`
		Error string          `json:"error,omitempty"`
	}

	resp := DisksResponse{
		Disks: disks,
	}

	if err != nil {
		resp.Error = err.Error()
		w.WriteHeader(http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func getFileSystemContentHandler(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)

	if r.Method == "OPTIONS" {
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	// Obtener parámetros de la URL
	partitionID := r.URL.Query().Get("partition_id")
	path := r.URL.Query().Get("path")

	if partitionID == "" {
		http.Error(w, "Parámetro partition_id requerido", http.StatusBadRequest)
		return
	}

	if path == "" {
		path = "/"
	}

	// Obtener información de la partición montada
	mountInfo, err := Disk.GetMountInfoByID(partitionID)
	if err != nil {
		type ErrorResponse struct {
			Error string `json:"error"`
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Partición no montada o no encontrada"})
		return
	}

	// Crear MountInfo para el sistema EXT2
	systemMountInfo := &System.MountInfo{
		DiskPath:      mountInfo.DiskPath,
		PartitionName: mountInfo.PartitionName,
		MountID:       mountInfo.MountID,
		DiskLetter:    mountInfo.DiskLetter,
		PartNumber:    mountInfo.PartNumber,
	}

	// Inicializar el gestor EXT2
	ext2Manager := System.NewEXT2Manager(systemMountInfo)
	if ext2Manager == nil {
		type ErrorResponse struct {
			Error string `json:"error"`
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Error al inicializar el sistema de archivos"})
		return
	}

	// Crear el gestor de directorios
	dirManager := System.NewEXT2DirectoryManager(ext2Manager)
	if dirManager == nil {
		type ErrorResponse struct {
			Error string `json:"error"`
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Error al crear gestor de directorios"})
		return
	}

	// Listar el contenido del directorio
	entries, err := dirManager.ListDirectory(path)
	if err != nil {
		type ErrorResponse struct {
			Error string `json:"error"`
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(ErrorResponse{Error: fmt.Sprintf("Error al listar directorio: %s", err.Error())})
		return
	}

	// Transformar las entradas al formato esperado por el frontend
	type FileSystemEntry struct {
		Name        string `json:"name"`
		Type        string `json:"type"`
		Size        int32  `json:"size"`
		Permissions string `json:"permissions"`
		Owner       string `json:"owner"`
		UID         int32  `json:"uid"`
		GID         int32  `json:"gid"`
	}

	// Inicializar response como array vacío para evitar null en JSON
	response := make([]FileSystemEntry, 0)
	for _, entry := range entries {
		// Filtrar "." y ".." para no mostrarlos en la interfaz
		if entry.Name == "." || entry.Name == ".." {
			continue
		}

		entryType := "file"
		if entry.Type == 0 { // INODO_DIRECTORIO
			entryType = "folder"
		}

		// Formatear permisos en formato octal (ej: "664")
		perms := fmt.Sprintf("%d", entry.Permissions)

		// TODO: Obtener el nombre del usuario desde el archivo users.txt
		// Por ahora, mostrar el UID como owner
		owner := fmt.Sprintf("uid:%d", entry.UID)

		response = append(response, FileSystemEntry{
			Name:        entry.Name,
			Type:        entryType,
			Size:        entry.Size,
			Permissions: perms,
			Owner:       owner,
			UID:         entry.UID,
			GID:         entry.GID,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func getFileContentHandler(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)

	if r.Method == "OPTIONS" {
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	// Obtener parámetros de la URL
	partitionID := r.URL.Query().Get("partition_id")
	filePath := r.URL.Query().Get("path")

	if partitionID == "" {
		http.Error(w, "Parámetro partition_id requerido", http.StatusBadRequest)
		return
	}

	if filePath == "" {
		http.Error(w, "Parámetro path requerido", http.StatusBadRequest)
		return
	}

	// Obtener información de la partición montada
	mountInfo, err := Disk.GetMountInfoByID(partitionID)
	if err != nil {
		type ErrorResponse struct {
			Error string `json:"error"`
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Partición no montada o no encontrada"})
		return
	}

	// Crear MountInfo para el sistema EXT2
	systemMountInfo := &System.MountInfo{
		DiskPath:      mountInfo.DiskPath,
		PartitionName: mountInfo.PartitionName,
		MountID:       mountInfo.MountID,
		DiskLetter:    mountInfo.DiskLetter,
		PartNumber:    mountInfo.PartNumber,
	}

	// Inicializar el gestor EXT2
	ext2Manager := System.NewEXT2Manager(systemMountInfo)
	if ext2Manager == nil {
		type ErrorResponse struct {
			Error string `json:"error"`
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Error al inicializar el sistema de archivos"})
		return
	}

	// Crear el gestor de archivos
	fileManager := System.NewEXT2FileManager(ext2Manager)
	if fileManager == nil {
		type ErrorResponse struct {
			Error string `json:"error"`
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Error al crear gestor de archivos"})
		return
	}

	// Leer el contenido del archivo
	content, err := fileManager.ReadFileContent(filePath)
	if err != nil {
		type ErrorResponse struct {
			Error string `json:"error"`
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(ErrorResponse{Error: fmt.Sprintf("Error al leer archivo: %s", err.Error())})
		return
	}

	// Respuesta con el contenido del archivo
	type FileContentResponse struct {
		Path    string `json:"path"`
		Content string `json:"content"`
		Size    int    `json:"size"`
	}

	response := FileContentResponse{
		Path:    filePath,
		Content: content,
		Size:    len(content),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Middleware CORS para permitir peticiones desde S3
func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Permitir peticiones desde cualquier origen (puedes restringir a tu bucket S3 específico)
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Manejar peticiones preflight (OPTIONS)
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

func startServer() {
	http.HandleFunc("/execute", corsMiddleware(executeCommandHandler))
	http.HandleFunc("/disks", corsMiddleware(getDisksHandler))
	http.HandleFunc("/filesystem", corsMiddleware(getFileSystemContentHandler))
	http.HandleFunc("/file-content", corsMiddleware(getFileContentHandler))

	fmt.Println("Servidor iniciado en http://localhost:8080")
	fmt.Println("Ctrl+C para detener")

	log.Fatal(http.ListenAndServe(":8080", nil))
}
