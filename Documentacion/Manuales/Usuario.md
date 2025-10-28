# Manual de Usuario – Proyecto 2 (MIA File System)
#### Daved Ejcalon Chonay - 202105668 - Lab MIA
#### Sistema de Archivos EXT2/EXT3 con Interfaz Web

---

## Requisitos del Sistema

### Requisitos Mínimos
- **Sistema Operativo:** Windows 10 / Linux Ubuntu 20.04 o superior
- **Procesador:** Intel i3 o equivalente
- **Memoria RAM:** 4 GB
- **Espacio en disco:** 500 MB libres
- **Navegador Web:** Chrome 90+, Firefox 88+, Edge 90+
- **Dependencias:**
  - Go 1.18+
  - Node.js 16+ y npm
  - Graphviz instalado y accesible desde la línea de comandos

### Requisitos Recomendados
- **Sistema Operativo:** Windows 11 / Linux Ubuntu 22.04
- **Procesador:** Intel i5 o superior
- **Memoria RAM:** 8 GB o más
- **Espacio en disco:** 1 GB libres
- **Dependencias adicionales:**
  - Conexión a internet estable para actualizaciones

---

## Descripción General

**MIA File System** es una aplicación web full-stack que simula el funcionamiento de un sistema de archivos **EXT2/EXT3** sobre discos virtuales, utilizando archivos binarios con extensión `.mia` como contenedores.

La aplicación está desarrollada en **Go** (backend) y **React** (frontend), empleando:
- **Graphviz** para la generación de reportes visuales
- **Journaling EXT3** para recuperación ante fallos
- **API REST** para comunicación entre frontend y backend
- **Interfaz web moderna** para navegación visual del sistema de archivos

### Novedades del Proyecto 2

✨ **Sistema EXT3 con Journaling** - Registro de transacciones para recuperación

✨ **Interfaz Web Interactiva** - Explorador visual de archivos y carpetas

✨ **Operaciones Avanzadas** - Copiar, mover, renombrar, editar archivos

✨ **Gestión de Permisos** - CHMOD, CHOWN, CHGRP estilo Unix

✨ **Recuperación de Sistema** - Comando RECOVERY desde 

✨ **Eliminación de Particiones** - FDISK -delete y FDISK -add

---
## Instalación y Configuración

#### Paso 1: Compilar el Backend
#### Paso 2: Instalar Dependencias del Frontend
#### Paso 3: Iniciar el Sistema


## Interfaz de Usuario


### Explorador Visual de Archivos (Nuevo en P2)

Interfaz gráfica moderna para navegar el sistema de archivos.

**Niveles de navegación:**

#### Nivel 1: Selección de Disco

Muestra tarjetas con información de todos los discos creados.

**Información mostrada:**
- 💾 Nombre del disco (ej: Disco1.mia)
- Capacidad total
- Algoritmo de ajuste (FF, BF, WF)
- Número de particiones montadas

![Vista de Discos](https://i.ibb.co/nMfcv00k/PI-MIA-1.png)

#### Nivel 2: Selección de Partición

Al hacer clic en un disco, muestra sus particiones.

**Información mostrada:**
- 📦 Nombre de la partición (ej: Part1)
- Estado: MONTADA / NO MONTADA
- ID de montaje (ej: 681a)
- Tamaño
- Algoritmo de ajuste


#### Nivel 3: Explorador de Archivos

Al hacer clic en una partición montada, muestra su contenido.

**Componentes:**
- **Breadcrumb** - Ruta actual (ej: Raíz / calificacion / U2025)
- **Información de partición** - Nombre e ID en la esquina superior
- **Grid de archivos** - Carpetas 📁 y archivos 📄

**Información de cada elemento:**
- Nombre
- Tipo (carpeta/archivo)
- Permisos (ej: 664, 755)
- Tamaño (solo archivos)

![Explorador de Archivos](https://i.ibb.co/WN4FHxZf/PI-MIA-4.png)

**Navegación:**
- **Click en carpeta** → Entra a la carpeta
- **Botón "Atrás"** → Regresa al nivel anterior
- **Botón "Raíz"** → Vuelve al directorio raíz
- **Botón "Volver a Particiones"** → Regresa a la vista de particiones

![Login](https://i.ibb.co/9mbcz7fC/PI-MIA-2.png)
---

## Comandos del Sistema

### Gestión de Discos

#### MKDISK - Crear Disco Virtual

Crea un archivo binario que simula un disco duro.

**Parámetros:**
- `-size` - Tamaño del disco (requerido)
- `-unit` - Unidad: `K` (KB), `M` (MB), `G` (GB). Default: M
- `-fit` - Ajuste: `FF` (First Fit), `BF` (Best Fit), `WF` (Worst Fit). Default: FF
- `-path` - Ruta donde crear el disco (requerido)

**Ejemplos:**
```bash
mkdisk -size=60 -unit=M -fit=FF -path=C:/Discos/Disco1.mia
mkdisk -size=2 -unit=K -path=C:/Discos/Disco2.mia
```

#### RMDISK - Eliminar Disco

Elimina un disco virtual del sistema.

**Ejemplo:**
```bash
rmdisk -path=C:/Discos/Disco1.mia
```

---

### Gestión de Particiones

#### FDISK - Crear Partición

Crea particiones primarias, extendidas o lógicas en un disco.

**Parámetros:**
- `-size` - Tamaño de la partición (requerido para crear)
- `-unit` - Unidad: K, M, B (bytes). Default: K
- `-path` - Ruta del disco (requerido)
- `-type` - Tipo: `P` (Primaria), `E` (Extendida), `L` (Lógica). Default: P
- `-fit` - Ajuste: FF, BF, WF. Default: WF
- `-name` - Nombre de la partición (requerido)

**Ejemplos:**
```bash
fdisk -type=P -unit=b -name=Part1 -size=20971520 -path=C:/Discos/Disco1.mia -fit=BF
fdisk -type=E -unit=M -name=PartExt -size=30 -path=C:/Discos/Disco1.mia
fdisk -type=L -unit=M -name=PartLog1 -size=10 -path=C:/Discos/Disco1.mia
```

#### FDISK -delete - Eliminar Partición [NUEVO P2]

Elimina una partición existente.

**Parámetros:**
- `-delete` - Modo: `fast` (rápida) o `full` (completa con borrado)
- `-path` - Ruta del disco
- `-name` - Nombre de la partición a eliminar

**Ejemplos:**
```bash
fdisk -delete=fast -name=Part3 -path=C:/Discos/Disco1.mia
fdisk -delete=full -name=Part4 -path=C:/Discos/Disco1.mia
```

#### FDISK -add - Modificar Tamaño [NUEVO P2]

Aumenta o reduce el tamaño de una partición.

**Parámetros:**
- `-add` - Cantidad a agregar (positivo) o quitar (negativo)
- `-unit` - Unidad del valor
- `-path` - Ruta del disco
- `-name` - Nombre de la partición

**Ejemplos:**
```bash
fdisk -add=-500 -unit=k -path=C:/Discos/Disco1.mia -name=Part2
fdisk -add=200 -unit=k -path=C:/Discos/Disco1.mia -name=Part2
```

---

### Montaje de Particiones

#### MOUNT - Montar Partición

Monta una partición para poder usarla. Genera un ID único.

**Ejemplo:**
```bash
mount -path=C:/Discos/Disco1.mia -name=Part1
```

**Formato del ID:** `[últimos 2 dígitos del carnet][correlativo][letra del disco]`
- Ejemplo: **681a** → carnet termina en 68, correlativo 1, disco A

#### UNMOUNT - Desmontar Partición [NUEVO P2]

Desmonta una partición previamente montada.

**Ejemplo:**
```bash
unmount -id=682a
```

#### MOUNTED - Ver Particiones Montadas

Muestra todas las particiones actualmente montadas.

**Ejemplo:**
```bash
mounted
```

---

### Sistema de Archivos

#### MKFS - Formatear Partición

Crea un sistema de archivos EXT2 o EXT3 en una partición montada.

**Parámetros:**
- `-type` - Tipo de formateo: `full` (completo). Default: full
- `-id` - ID de la partición montada (requerido)
- `-fs` - Sistema: `2fs` (EXT2) o `3fs` (EXT3). Default: 2fs

**Ejemplos:**
```bash
mkfs -type=full -id=681a -fs=2fs
mkfs -type=full -id=681a -fs=3fs
```

**¿Cuándo usar EXT3?**
- ✅ Si necesita recuperación ante fallos (comando RECOVERY)
- ✅ Si quiere registro de transacciones (journaling)
- ⚠️ Ocupa 8 KB adicionales para el journal

---

### Gestión de Usuarios

#### LOGIN - Iniciar Sesión

Autentica un usuario en el sistema.

**Ejemplo:**
```bash
login -user=root -pass=123 -id=681a
```

**Usuario por defecto:** `root` / `123`

#### LOGOUT - Cerrar Sesión [NUEVO P2]

Cierra la sesión actual del usuario.

**Ejemplo:**
```bash
logout
```



#### MKGRP - Crear Grupo

Crea un nuevo grupo de usuarios (solo root).

**Sintaxis:**
```bash
mkgrp -name=<nombre>
```

**Ejemplo:**
```bash
mkgrp -name=desarrolladores
```

#### MKUSR - Crear Usuario

Crea un nuevo usuario en el sistema (solo root).

**Sintaxis:**
```bash
mkusr -user=<usuario> -pass=<contraseña> -grp=<grupo>
```

**Ejemplo:**
```bash
mkusr -user=pedro -pass=abc123 -grp=desarrolladores
```



#### RMGRP - Eliminar Grupo

Elimina un grupo existente (solo root).

**Sintaxis:**
```bash
rmgrp -name=<nombre>
```

#### RMUSR - Eliminar Usuario

Elimina un usuario existente (solo root).

**Sintaxis:**
```bash
rmusr -user=<usuario>
```

---

### Directorios y Archivos

#### MKDIR - Crear Directorio

Crea un directorio en el sistema de archivos.

**Sintaxis:**
```bash
mkdir [-p] -path=<ruta>
```

**Parámetros:**
- `-p` - Crea directorios padres si no existen (opcional)
- `-path` - Ruta del directorio a crear (requerido)

**Ejemplos:**
```bash
# Crear directorio simple
mkdir -path=/documentos

# Crear estructura completa de directorios
mkdir -p -path=/calificacion/U2025/6toSemestre/MIA
```



#### MKFILE - Crear Archivo

Crea un archivo con contenido generado automáticamente.

**Sintaxis:**
```bash
mkfile -path=<ruta> -size=<tamaño> [-cont=<ruta_archivo>]
```

**Parámetros:**
- `-path` - Ruta completa del archivo (requerido)
- `-size` - Tamaño en bytes (requerido)
- `-cont` - Archivo local para copiar contenido (opcional)

**Ejemplos:**
```bash
# Crear archivo con contenido generado (números 0-9)
mkfile -path=/documentos/archivo.txt -size=100

# Crear archivo con contenido de un archivo local
mkfile -path=/documentos/tarea.txt -cont=C:/Users/local/tarea.txt
```



#### CAT - Mostrar Contenido

Muestra el contenido de uno o varios archivos.

**Sintaxis:**
```bash
cat -file<n>=<ruta> ...
```

**Parámetros:**
- `-file1`, `-file2`, ... - Rutas de archivos a mostrar

**Ejemplos:**
```bash
# Mostrar un archivo
cat -file1=/documentos/archivo.txt

# Mostrar múltiples archivos
cat -file1=/docs/a.txt -file2=/docs/b.txt -file3=/docs/c.txt
```



---

### Operaciones Avanzadas de Archivos [NUEVO P2]

#### COPY - Copiar Archivo

Copia un archivo a otra ubicación.

**Sintaxis:**
```bash
copy -path=<origen> -dest=<destino>
```

**Ejemplo:**
```bash
copy -path=/documentos/original.txt -dest=/respaldo/copia.txt
```



#### MOVE - Mover Archivo

Mueve un archivo a otra ubicación (cortar y pegar).

**Sintaxis:**
```bash
move -path=<origen> -dest=<destino>
```

**Ejemplo:**
```bash
move -path=/temporal/archivo.txt -dest=/permanente/
```



#### RENAME - Renombrar

Cambia el nombre de un archivo o carpeta.

**Sintaxis:**
```bash
rename -path=<ruta> -name=<nuevo_nombre>
```

**Ejemplo:**
```bash
rename -path=/documentos/viejo.txt -name=nuevo.txt
```



#### REMOVE - Eliminar Archivo

Elimina un archivo del sistema.

**Sintaxis:**
```bash
remove -path=<ruta>
```

**Ejemplo:**
```bash
remove -path=/temporal/basura.txt
```

**⚠️ Advertencia:** Esta acción no se puede deshacer (a menos que use RECOVERY en EXT3).



#### EDIT - Editar Contenido [NO IMPLEMENTADO AÚN]

Edita el contenido de un archivo.

**Sintaxis:**
```bash
edit -path=<ruta> -content=<contenido>
```



#### FIND - Buscar Archivos

Busca archivos por nombre o patrón.

**Sintaxis:**
```bash
find -path=<directorio> -name=<patrón>
```

**Ejemplos:**
```bash
# Buscar archivos .txt
find -path=/documentos -name=*.txt

# Buscar archivo específico
find -path=/ -name=tarea.txt
```



---

### Gestión de Permisos [NUEVO P2]

#### CHMOD - Cambiar Permisos

Modifica los permisos de un archivo o directorio.

**Sintaxis:**
```bash
chmod -path=<ruta> -ugo=<permisos>
```

**Parámetros:**
- `-path` - Ruta del archivo/carpeta
- `-ugo` - Permisos en formato octal (ej: 664, 755)

**Formato de permisos:**
```
7 = rwx (lectura, escritura, ejecución)
6 = rw- (lectura, escritura)
5 = r-x (lectura, ejecución)
4 = r-- (solo lectura)
0 = --- (sin permisos)
```

**Estructura:** `[owner][group][others]`

**Ejemplos:**
```bash
# Dar todos los permisos al owner, lectura al grupo y otros
chmod -path=/archivo.txt -ugo=744

# Permisos típicos de archivo
chmod -path=/documento.txt -ugo=664

# Permisos típicos de directorio
chmod -path=/carpeta -ugo=755
```

**Restricción:** Solo el propietario o root puede ejecutar este comando.



#### CHOWN - Cambiar Propietario

Cambia el propietario de un archivo o directorio (solo root).

**Sintaxis:**
```bash
chown -path=<ruta> -user=<usuario>
```

**Ejemplo:**
```bash
chown -path=/documentos/archivo.txt -user=pedro
```



#### CHGRP - Cambiar Grupo

Cambia el grupo de un archivo o directorio.

**Sintaxis:**
```bash
chgrp -path=<ruta> -grp=<grupo>
```

**Ejemplo:**
```bash
chgrp -path=/proyecto -grp=desarrolladores
```



---

### Recuperación y Journaling [NUEVO P2 - Solo EXT3]

#### JOURNALING - Ver Journal

Muestra el contenido del journal de transacciones (solo EXT3).

**Sintaxis:**
```bash
journaling -id=<id>
```

**Ejemplo:**
```bash
journaling -id=681a
```





#### LOSS - Simular Pérdida

Simula una pérdida catastrófica del sistema (para probar RECOVERY).

**Sintaxis:**
```bash
loss -id=<id>
```

**Ejemplo:**
```bash
loss -id=681a
```

**⚠️ Advertencia:** Este comando corrompe las estructuras del sistema de archivos. Úselo solo para demostrar la recuperación.



#### RECOVERY - Recuperar Sistema

Recupera el sistema de archivos desde el journal (solo EXT3).

**Sintaxis:**
```bash
recovery -id=<id>
```

**Ejemplo:**
```bash
recovery -id=681a
```

**Proceso:**
1. Lee el journal desde el disco
2. Reproduce cada operación registrada
3. Restaura archivos y directorios
4. Reporta operaciones recuperadas





---

### Reportes

#### REP - Generar Reporte

Genera reportes visuales usando Graphviz.

**Sintaxis:**
```bash
rep -id=<id> -path=<ruta_salida> -name=<tipo> [-path_file_ls=<ruta>]
```

**Tipos de reportes:**

##### 1. MBR Report
Muestra la estructura del MBR y particiones.

```bash
rep -id=681a -path=C:/Reportes/mbr.jpg -name=mbr
```



##### 2. DISK Report
Visualiza el uso del disco con gráfico de barras.

```bash
rep -id=681a -path=C:/Reportes/disk.jpg -name=disk
```



##### 3. SuperBlock Report
Muestra información del SuperBloque (EXT2/EXT3).

```bash
rep -id=681a -path=C:/Reportes/sb.jpg -name=sb
```



##### 4. Inode Report
Muestra la estructura de un inodo específico.

```bash
rep -id=681a -path=C:/Reportes/inode.jpg -name=inode
```



##### 5. File Report
Muestra el contenido de un archivo con formato tabular.

```bash
rep -id=681a -path=C:/Reportes/file.jpg -path_file_ls=/archivo.txt -name=file
```



##### 6. LS Report
Lista el contenido de un directorio (árbol recursivo).

```bash
rep -id=681a -path=C:/Reportes/ls.jpg -path_file_ls=/documentos -name=ls
```



##### 7. EBR Report [NUEVO P2]
Muestra la estructura de particiones lógicas (EBR).

```bash
rep -id=681a -path=C:/Reportes/ebr.jpg -name=ebr
```



---

## Casos de Uso Prácticos

### Caso 1: Crear y Configurar un Disco Completo

**Objetivo:** Crear un disco, particionar, formatear y montar.

**Pasos:**

1. **Crear disco de 60 MB:**
```bash
mkdisk -size=60 -unit=M -fit=FF -path=C:/Discos/MiDisco.mia
```

2. **Crear partición primaria de 20 MB:**
```bash
fdisk -type=P -unit=M -name=Datos -size=20 -path=C:/Discos/MiDisco.mia -fit=BF
```

3. **Montar la partición:**
```bash
mount -path=C:/Discos/MiDisco.mia -name=Datos
```
*Salida: ID asignado.*

4. **Formatear con EXT3:**
```bash
mkfs -type=full -id=681a -fs=3fs
```

5. **Iniciar sesión:**
```bash
login -user=root -pass=123 -id=681a
```



---

### Caso 2: Crear Estructura de Archivos Académica

**Objetivo:** Crear estructura de carpetas para un estudiante universitario.

**Pasos:**

1. **Crear estructura de directorios:**
```bash
mkdir -p -path=/Universidad/2025/Semestre6/MIA
mkdir -p -path=/Universidad/2025/Semestre6/Arqui
mkdir -p -path=/Universidad/2025/Semestre6/Compi
```

2. **Crear archivos de tareas:**
```bash
mkfile -path=/Universidad/2025/Semestre6/MIA/proyecto1.txt -size=500
mkfile -path=/Universidad/2025/Semestre6/MIA/proyecto2.txt -size=800
mkfile -path=/Universidad/2025/Semestre6/Arqui/practica.txt -size=300
```

3. **Visualizar en interfaz web:**
   - Abrir http://localhost:3000
   - Navegar al disco
   - Seleccionar la partición
   - Explorar carpetas visualmente



---

### Caso 3: Gestión de Usuarios y Permisos

**Objetivo:** Crear usuarios, grupos y asignar permisos.

**Pasos:**

1. **Crear grupos:**
```bash
mkgrp -name=estudiantes
mkgrp -name=profesores
```

2. **Crear usuarios:**
```bash
mkusr -user=juan -pass=123 -grp=estudiantes
mkusr -user=maria -pass=456 -grp=profesores
```

3. **Crear archivo compartido:**
```bash
mkfile -path=/compartido/notas.txt -size=100
```

4. **Asignar permisos:**
```bash
# Profesores pueden escribir, estudiantes solo leer
chmod -path=/compartido/notas.txt -ugo=644
chgrp -path=/compartido/notas.txt -grp=profesores
```



---

### Caso 4: Recuperación ante Fallos (EXT3)

**Objetivo:** Demostrar la recuperación del sistema usando journaling.

**Pasos:**

1. **Formatear con EXT3:**
```bash
mkfs -type=full -id=681a -fs=3fs
login -user=root -pass=123 -id=681a
```

2. **Crear archivos y carpetas:**
```bash
mkdir -p -path=/importante/documentos
mkfile -path=/importante/documentos/datos.txt -size=200
mkfile -path=/importante/respaldo.txt -size=150
```

3. **Verificar journal:**
```bash
journaling -id=681a
```

4. **Simular pérdida del sistema:**
```bash
loss -id=681a
```
*El sistema ahora está corrupto*

5. **Recuperar desde journal:**
```bash
recovery -id=681a
```

6. **Verificar que los archivos están de vuelta:**
```bash
cat -file1=/importante/documentos/datos.txt
```



---

### Caso 5: Operaciones Avanzadas de Archivos

**Objetivo:** Usar comandos COPY, MOVE, RENAME.

**Pasos:**

1. **Crear estructura inicial:**
```bash
mkdir -path=/temporal
mkdir -path=/definitivo
mkfile -path=/temporal/borrador.txt -size=100
```

2. **Copiar archivo:**
```bash
copy -path=/temporal/borrador.txt -dest=/definitivo/borrador.txt
```

3. **Renombrar:**
```bash
rename -path=/definitivo/borrador.txt -name=final.txt
```

4. **Mover a otra ubicación:**
```bash
mkdir -path=/archivo
move -path=/definitivo/final.txt -dest=/archivo/
```

5. **Eliminar temporal:**
```bash
remove -path=/temporal/borrador.txt
```
