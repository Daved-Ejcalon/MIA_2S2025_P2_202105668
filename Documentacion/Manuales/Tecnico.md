
# Manual Técnico - Sistema de Archivos EXT2/EXT3
**Daved Abshalon Ejcalon Chonay - 202105668**
**Proyecto 2 - MIA**

## Requisitos del Sistema

### Requisitos Mínimos
- **Sistema Operativo:** Windows 10 / Linux Ubuntu 20.04 o superior
- **Procesador:** Intel i3 o equivalente
- **Memoria RAM:** 4 GB
- **Espacio en disco:** 500 MB libres
- **Dependencias:**
  - Go 1.18+
  - Node.js 16+ y npm
  - React 18+
  - Graphviz instalado y accesible desde la línea de comandos

### Requisitos Recomendados
- **Sistema Operativo:** Windows 11 / Linux Ubuntu 22.04
- **Procesador:** Intel i5 o superior
- **Memoria RAM:** 8 GB o más
- **Espacio en disco:** 1 GB libres
- **Dependencias adicionales:**
  - Conexión a internet estable para actualizaciones
  - Navegador web moderno (Chrome, Firefox, Edge)

---

## 1. Arquitectura General

El proyecto implementa un sistema de archivos **EXT2/EXT3** con soporte para journaling, gestión avanzada de usuarios, operaciones de archivos complejas y una **interfaz web moderna**. La arquitectura está dividida en dos capas principales:

```
MIA_2S2025_P2_202105668/
├── Backend/                    # Servidor Go
│   ├── Models/                 # Estructuras de datos
│   ├── Logica/
│   │   ├── Disk/              # Gestión de discos y particiones
│   │   ├── System/            # Sistema de archivos EXT2/EXT3
│   │   │   ├── ext2_manager.go
│   │   │   ├── ext3_manager.go    [NUEVO P2]
│   │   │   ├── journaling.go      [NUEVO P2]
│   │   │   ├── recovery.go        [NUEVO P2]
│   │   │   └── loss.go            [NUEVO P2]
│   │   ├── Users/             # Gestión de usuarios y permisos
│   │   │   ├── Operations/    # Operaciones de archivos [NUEVO P2]
│   │   │   └── Root/          # Comandos privilegiados
│   │   └── Reportes/          # Generación de reportes
│   └── Utils/                 # Utilidades generales
│
└── frontend/                   # Interfaz Web React [NUEVO P2]
    ├── src/
    │   ├── components/
    │   │   ├── Terminal.js
    │   │   ├── FileSystemVisualizer.js  [NUEVO P2]
    │   │   └── FileSystemVisualizer.css
    │   └── App.js
    └── package.json
```

### **1.1 Nuevas Características del Proyecto 2**

✨ **Sistema EXT3 con Journaling**

✨ **Interfaz Web React** para visualización de archivos

✨ **Operaciones avanzadas:** COPY, MOVE, RENAME, FIND, EDIT, REMOVE

✨ **Gestión de permisos:** CHMOD, CHOWN, CHGRP

✨ **Recuperación de sistema:** RECOVERY comando

✨ **Simulación de pérdidas:** LOSS comando

✨ **Visualización de journal:** JOURNALING comando

✨ **API REST** para comunicación frontend-backend

---

## 2. Modelos de Datos Principales

### **2.1 MBR (Master Boot Record)**
```go
type MBR struct {
    MbrSize        int32        // Tamaño del MBR
    MbrDate        float64      // Fecha de creación
    MbrDskSignature int32       // Firma del disco
    DskFit         byte         // Algoritmo de ajuste (FF, BF, WF)
    Partitions     [4]Partition // 4 particiones primarias/extendidas
}
```
**Funciones principales:**
- Almacena información del disco y particiones
- Controla algoritmos de ajuste de espacio
- **[NUEVO P2]** Soporta eliminación y modificación de particiones

### **2.2 SuperBloque - NÚCLEO DEL SISTEMA**
```go
type SuperBloque struct {
    S_filesystem_type   int32   // Tipo de sistema (2 = EXT2, 3 = EXT3) [ACTUALIZADO P2]
    S_inodes_count      int32   // Total de inodos
    S_blocks_count      int32   // Total de bloques
    S_free_blocks_count int32   // Bloques libres
    S_free_inodes_count int32   // Inodos libres
    S_mtime             float64 // Última modificación
    S_umtime            float64 // Último montaje
    S_mnt_count         int32   // Número de montajes
    S_magic             int32   // Número mágico (0xEF53)
    S_inode_s           int32   // Tamaño de inodo (128 bytes)
    S_block_s           int32   // Tamaño de bloque (64 bytes)
    S_firts_ino         int32   // Primer inodo libre
    S_first_blo         int32   // Primer bloque libre
    S_bm_inode_start    int32   // Inicio bitmap inodos
    S_bm_block_start    int32   // Inicio bitmap bloques
    S_inode_start       int32   // Inicio tabla inodos
    S_block_start       int32   // Inicio área bloques
}
```
**Importancia:** Es el **corazón del sistema EXT2/EXT3**, controla toda la metadata del filesystem.

### **2.3 Journal - Sistema de Transacciones [NUEVO P2]**
```go
type Journal struct {
    J_count   int32           // Contador de entradas
    J_content [64]Information // 64 entradas de información
}

type Information struct {
    I_operation [10]byte  // Operación realizada (MKDIR, MKFILE, etc.)
    I_path      [32]byte  // Ruta donde se realizó
    I_content   [64]byte  // Contenido (para archivos)
    I_date      float64   // Fecha de la operación
}
```

**Características clave:**
- **Registro de transacciones** para EXT3
- **Máximo 64 entradas** históricas
- **Recuperación ante fallos** del sistema
- **Auditoría completa** de operaciones

**Constantes:**
```go
const (
    JOURNAL_SIZE        = 8192  // 8 KB para el journal
    JOURNAL_MAX_ENTRIES = 64    // Máximo de entradas
)
```

### **2.4 Inodo - Gestión de Archivos**
```go
type Inodo struct {
    I_uid   int32     // ID usuario propietario
    I_gid   int32     // ID grupo propietario
    I_s     int32     // Tamaño del archivo
    I_atime float64   // Último acceso
    I_ctime float64   // Creación
    I_mtime float64   // Última modificación
    I_block [15]int32 // Punteros a bloques (12 directos + 3 indirectos)
    I_type  byte      // Tipo (0=directorio, 1=archivo) [ACTUALIZADO P2]
    I_perm  [3]byte   // Permisos en formato [owner, group, others] [ACTUALIZADO P2]
}
```
**Funciones principales:**
- Almacena metadata de archivos y directorios
- Controla permisos y propiedades
- **[NUEVO P2]** Permisos en formato de bytes [7,5,5] = 755

### **2.5 Bloques de Datos**
```go
// Bloque de carpeta (4 entradas)
type BloqueCarpeta struct {
    B_content [4]B_content
}

type B_content struct {
    B_name  [12]byte  // Nombre del archivo/carpeta
    B_inodo int32     // Número de inodo
}

// Bloque de contenido de archivo (64 bytes)
type BloqueArchivos struct {
    B_content [64]byte
}

// Bloque de apuntadores indirectos
type BloqueApuntadores struct {
    B_pointers [16]int32
}
```

---

## 3. Clases Core del Sistema

### **3.1 EXT2Manager - GESTOR PRINCIPAL**
**Ubicación:** `Backend/Logica/System/ext2_manager.go`

```go
type EXT2Manager struct {
    mountInfo   *MountInfo
    superblock  *Models.SuperBloque
    diskFile    *os.File
}
```

**Responsabilidades principales:**
- **Inicialización del sistema de archivos**
- **Gestión de bitmaps** (inodos y bloques)
- **Lectura/escritura de inodos**
- **Asignación de espacio libre**

**Métodos críticos:**
- `FormatPartition()` - Formatea partición con EXT2
- `AllocateInode()` - Asigna nuevo inodo
- `AllocateBlock()` - Asigna nuevo bloque
- `ReadInode()` - Lee inodo desde disco
- `WriteInode()` - Escribe inodo a disco

### **3.2 EXT3Manager - GESTOR CON JOURNALING [NUEVO P2]**
**Ubicación:** `Backend/Logica/System/ext3_manager.go`

```go
type EXT3Manager struct {
    ext2Manager *EXT2Manager
    journal     *Models.Journal
    journalPos  int32  // Posición del journal en disco
}
```

**Responsabilidades:**
- **Gestión de journaling** (bitácora de transacciones)
- **Escritura atómica** de operaciones
- **Recuperación ante fallos**

**Métodos importantes:**
- `FormatEXT3()` - Formatea partición con EXT3
- `LogOperation()` - Registra operación en journal
- `GetJournal()` - Lee journal desde disco
- `WriteJournal()` - Escribe journal a disco

### **3.3 EXT2FileManager - Gestión de Archivos**
**Ubicación:** `Backend/Logica/System/ext2_files.go`

**Responsabilidades:**
- **Creación y lectura de archivos**
- **Navegación del sistema de archivos**
- **Resolución de rutas**
- **[NUEVO P2]** Edición de contenido de archivos
- **[NUEVO P2]** Búsqueda de archivos

**Métodos importantes:**
- `CreateFile()` - Crea nuevos archivos
- `ReadFileContent()` - Lee contenido completo
- `findFileInode()` - Busca archivos por ruta
- `WriteFileContent()` - Escribe contenido [NUEVO P2]
- `SearchFiles()` - Busca archivos por patrón [NUEVO P2]

### **3.4 EXT2DirectoryManager - Gestión de Directorios**
**Ubicación:** `Backend/Logica/System/ext2_directories.go`

**Responsabilidades:**
- **Creación de directorios**
- **Listado de contenido**
- **Gestión de entradas de directorio**
- **[NUEVO P2]** Navegación recursiva

**Métodos importantes:**
- `CreateDirectory()` - Crea directorios
- `ListDirectory()` - Lista contenido (retorna DirectoryEntry)
- `addDirectoryEntry()` - Agrega entradas
- `RemoveDirectory()` - Elimina directorios vacíos [NUEVO P2]

**Estructura de retorno:**
```go
type DirectoryEntry struct {
    Name        string
    InodeNumber int32
    Type        byte    // 0=directorio, 1=archivo
    Size        int32
    Permissions int32
    UID         int32
    GID         int32
    ATime       float64
    CTime       float64
    MTime       float64
}
```

### **3.5 RecoveryManager - Recuperación del Sistema [NUEVO P2]**
**Ubicación:** `Backend/Logica/System/recovery.go`

**Responsabilidades:**
- **Lectura del journal**
- **Reproducción de operaciones**
- **Restauración del estado del sistema**

**Métodos:**
- `RecoverFileSystem()` - Ejecuta recuperación completa
- `replayOperation()` - Re-ejecuta operaciones del journal

### **3.6 LossSimulator - Simulador de Pérdidas [NUEVO P2]**
**Ubicación:** `Backend/Logica/System/loss.go`

**Responsabilidades:**
- **Corrupción controlada** de estructuras
- **Limpieza de journal**
- **Simulación de fallos** del sistema

**Métodos:**
- `SimulateSystemLoss()` - Simula pérdida del sistema
- `CorruptStructures()` - Corrompe datos

---

## 4. Gestión de Usuarios y Permisos

### **4.1 UserManager**
**Ubicación:** `Backend/Logica/Users/user_manager.go`

**Responsabilidades:**
- **Autenticación de usuarios**
- **Gestión de grupos**
- **Control de permisos**
- **[NUEVO P2]** Sesiones de usuario activas

**Funciones principales:**
- `CreateUser()` - Crea nuevos usuarios
- `DeleteUser()` - Elimina usuarios
- `Login()` - Inicia sesión [ACTUALIZADO P2]
- `Logout()` - Cierra sesión [NUEVO P2]
- `GetCurrentSession()` - Obtiene sesión activa [NUEVO P2]

**Estructura de sesión:**
```go
type UserSession struct {
    Username  string
    UID       int32
    GID       int32
    MountID   string
    IsActive  bool
    LoginTime time.Time
}
```

### **4.2 Sistema de Permisos**
```go
type UserRecord struct {
    ID       int
    Group    string
    Username string
    Password string
}
```

**Características:**
- **Usuarios únicos por sistema**
- **Grupos de usuarios**
- **Permisos estilo Unix (rwx)** - formato octal [664, 755, etc.]
- **[NUEVO P2]** CHMOD - Cambiar permisos
- **[NUEVO P2]** CHOWN - Cambiar propietario
- **[NUEVO P2]** CHGRP - Cambiar grupo

---

## 5. Operaciones Avanzadas de Archivos [NUEVO P2]

### **5.1 COPY - Copiar Archivos**
**Ubicación:** `Backend/Logica/Users/Operations/copy.go`

**Funcionamiento:**
1. Lee contenido del archivo origen
2. Crea nuevo archivo en destino
3. Copia contenido bloque por bloque
4. Preserva permisos si se especifica

**Comando:**
```bash
copy -path=/origen.txt -dest=/destino.txt -id=681A
```

### **5.2 MOVE - Mover Archivos**
**Ubicación:** `Backend/Logica/Users/Operations/move.go`

**Funcionamiento:**
1. Copia archivo a nueva ubicación
2. Elimina archivo original
3. Actualiza entradas de directorio

**Comando:**
```bash
move -path=/archivo.txt -dest=/nueva/ubicacion/ -id=681A
```

### **5.3 RENAME - Renombrar**
**Ubicación:** `Backend/Logica/Users/Operations/rename.go`

**Funcionamiento:**
1. Busca entrada en directorio padre
2. Actualiza nombre en BloqueCarpeta
3. Mantiene mismo inodo

**Comando:**
```bash
rename -path=/archivo.txt -name=nuevo_nombre.txt -id=681A
```

### **5.4 REMOVE - Eliminar Archivos**
**Ubicación:** `Backend/Logica/Users/Operations/remove.go`

**Funcionamiento:**
1. Libera bloques de contenido
2. Libera inodo
3. Actualiza bitmaps
4. Elimina entrada de directorio

**Comando:**
```bash
remove -path=/archivo.txt -id=681A
```

### **5.5 EDIT - Editar Contenido**
**Ubicación:** `Backend/Logica/Users/Operations/edit.go`

**Funcionamiento:**
1. Lee contenido actual
2. Busca línea específica
3. Reemplaza contenido
4. Reescribe bloques

**Comando:**
```bash
edit -path=/archivo.txt -content="nuevo contenido" -id=681A
```

### **5.6 FIND - Buscar Archivos**
**Ubicación:** `Backend/Logica/Users/Operations/find.go`

**Funcionamiento:**
1. Recorre árbol de directorios recursivamente
2. Busca por nombre o patrón
3. Retorna rutas completas

**Comando:**
```bash
find -path=/directorio -name="*.txt" -id=681A
```

### **5.7 CHMOD - Cambiar Permisos**
**Ubicación:** `Backend/Logica/Users/Operations/chmod.go`

**Funcionamiento:**
1. Valida permisos de usuario (solo owner o root)
2. Lee inodo del archivo
3. Actualiza campo I_perm
4. Escribe inodo modificado

**Comando:**
```bash
chmod -path=/archivo.txt -ugo=664 -id=681A
```

### **5.8 CHOWN - Cambiar Propietario**
**Ubicación:** `Backend/Logica/Users/Operations/chown.go`

**Funcionamiento:**
1. Valida permisos (solo root)
2. Verifica que usuario existe
3. Actualiza I_uid en inodo

**Comando:**
```bash
chown -path=/archivo.txt -user=usuario1 -id=681A
```

---

## 6. Sistema de Reportes

### **6.1 Arquitectura de Reportes**
**Ubicación:** `Backend/Logica/Reportes/`

**Tipos implementados:**
- **MBR Report** - Visualiza estructura de particiones
- **Disk Report** - Muestra uso del disco
- **EBR Report** - Particiones extendidas [NUEVO P2]
- **SuperBlock Report** - Información del filesystem
- **Inode Report** - Estructura de inodos
- **File Report** - Contenido de archivos (con tabulación)
- **Ls Report** - Listado de directorios

### **6.2 Generación con Graphviz**
**Ubicación:** `Backend/Logica/Reportes/Graphviz/`

**Características:**
- **Tablas HTML** para formato profesional
- **Colores temáticos** consistentes
- **Responsive sizing** según contenido
- **Tabulación automática** para archivos grandes
- **[NUEVO P2]** Soporte para journal report

**Ejemplo de uso:**
```bash
rep -id=681A -path="reporte.jpg" -name=sb
rep -id=681A -path="archivo.jpg" -path_file_ls="/archivo.txt" -name=file
rep -id=681A -path="inodo.jpg" -name=inode
rep -id=681A -path="ls.jpg" -path_file_ls="/directorio" -name=ls
```

---


## 9. Comandos de Gestión de Particiones [NUEVO P2]

### **9.1 FDISK -delete**
Elimina particiones del disco.

**Sintaxis:**
```bash
fdisk -delete=fast -name=Part1 -path=disco.mia
fdisk -delete=full -name=Part2 -path=disco.mia
```

**Tipos de eliminación:**
- **fast** - Marca como libre en tabla MBR
- **full** - Sobrescribe con ceros

### **9.2 FDISK -add**
Modifica el tamaño de una partición.

**Sintaxis:**
```bash
# Reducir 500 KB
fdisk -add=-500 -unit=k -path=disco.mia -name=Part1

# Aumentar 200 KB
fdisk -add=200 -unit=k -path=disco.mia -name=Part1
```

**Comportamiento:**
- **-add** tiene prioridad sobre **-size**
- Valor negativo = reducir
- Valor positivo = aumentar

### **9.3 UNMOUNT** [NUEVO P2]
Desmonta una partición.

**Sintaxis:**
```bash
unmount -id=681a
```

**Funcionamiento:**
- Cierra sesiones activas
- Libera recursos
- Resetea correlativo de la letra

---

## 10. Comandos de Sistema EXT3 [NUEVO P2]

### **10.1 MKFS -fs=3fs**
Formatea partición con EXT3 (incluye journaling).

**Sintaxis:**
```bash
mkfs -type=full -id=681a -fs=3fs
```


### **10.2 RECOVERY**
Recupera el sistema de archivos desde el journal.

**Sintaxis:**
```bash
recovery -id=681a
```

**Proceso:**
1. Lee journal desde disco
2. Por cada entrada del journal:
   - Reproduce operación (MKDIR, MKFILE, etc.)
   - Restaura contenido
3. Reporta operaciones recuperadas

**Salida:**
```
=== INICIANDO RECUPERACIÓN DEL SISTEMA ===
Journal encontrado: 15 operaciones registradas

[1] MKDIR /calificacion (2025-01-15 10:30:00) ✓
[2] MKDIR /calificacion/U2025 (2025-01-15 10:30:01) ✓
[3] MKFILE /user.txt (2025-01-15 10:31:00) ✓
...

✓ Recuperación completada: 15 operaciones restauradas
```

### **10.3 LOSS**
Simula pérdida del sistema (corrompe estructuras).

**Sintaxis:**
```bash
loss -id=681a
```

**Proceso:**
1. Limpia bitmaps de inodos y bloques
2. Reinicia contadores del SuperBloque
3. Preserva journal intacto
4. Simula fallo catastrófico

**Uso:** Para probar comando RECOVERY

### **10.4 JOURNALING**
Muestra el contenido del journal.

**Sintaxis:**
```bash
journaling -id=681a
```



---

## 11. Flujo de Comandos Principales

### **11.1 Formateo de Partición con EXT3**
```bash
mkfs -type=full -id=681A -fs=3fs
```
**Proceso:**
1. Busca partición por ID
2. Crea SuperBloque (S_filesystem_type = 3)
3. **Inicializa Journal** con 64 entradas vacías
4. Inicializa bitmaps
5. Crea inodo raíz (/)
6. Configura archivos de usuarios
7. Escribe todas las estructuras a disco



### **11.2 Creación de Archivos con Journaling**
```bash
mkfile -path="/archivo.txt" -size=100 -id=681A
```
**Proceso:**
1. Valida permisos de usuario
2. **[NUEVO P2] Registra en Journal:** MKFILE operation
3. Busca directorio padre
4. Asigna inodo y bloques
5. Actualiza entrada de directorio
6. Escribe contenido
7. **[NUEVO P2] Actualiza Journal** en disco

### **11.3 Gestión de Usuarios**
```bash
mkusr -user=usuario1 -pwd=123 -grp=grupo1 -id=681A
login -user=usuario1 -pwd=123 -id=681A
logout
```
**Proceso:**
1. Valida sesión root (para mkusr)
2. Verifica usuario único
3. Actualiza archivo users.txt
4. **[NUEVO P2] Crea sesión activa** (login)
5. **[NUEVO P2] Guarda en memoria** (UserSession)
6. Sincroniza con disco

### **11.4 Operación de Copia** [NUEVO P2]
```bash
copy -path=/origen.txt -dest=/destino.txt -id=681A
```

---

## 14. Diagrama de Arquitectura del Sistema

```

**Diagrama:**  

![Ventana Principal] (https://i.ibb.co/NgKfQb00/Diagrama.png)

