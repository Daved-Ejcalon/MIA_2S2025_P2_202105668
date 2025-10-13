package Operations

import (
	"MIA_2S2025_P1_202105668/Logica/Disk"
	"MIA_2S2025_P1_202105668/Logica/System"
	"MIA_2S2025_P1_202105668/Logica/Users"
	"MIA_2S2025_P1_202105668/Models"
	"fmt"
	"regexp"
	"strings"
)

func Find(params map[string]string) error {
	searchPath := params["path"]
	searchName := params["name"]
	session := Users.GetCurrentSession()
	mountInfo, _ := Disk.GetMountInfoByID(session.MountID)
	partitionInfo, superBloque, _ := Users.GetPartitionAndSuperBlock(mountInfo)

	manager := &System.EXT2Manager{}
	manager.SetPartitionInfo(partitionInfo)
	manager.SetSuperBlock(superBloque)
	manager.SetDiskPath(mountInfo.DiskPath)

	fileManager := System.NewEXT2FileManager(manager)
	searchPath = normalizePath(searchPath)
	searchInodeNum, _ := findFileInode(fileManager, searchPath)
	searchInodo, _ := readInode(fileManager, searchInodeNum)

	hasReadPermission := System.ValidateFileReadPermission(
		searchInodo.I_uid,
		searchInodo.I_gid,
		searchInodo.I_perm,
		session.UserID,
		session.GroupID,
	)

	if !hasReadPermission {
		return nil
	}

	pattern := wildcardToRegex(searchName)
	results := []FindResult{}
	searchRecursive(fileManager, searchPath, searchInodeNum, pattern, session.UserID, session.GroupID, &results, 0)

	if len(results) == 0 {
		fmt.Println("No se encontraron coincidencias")
		return nil
	}

	printFindResults(results, searchPath)
	return nil
}

type FindResult struct {
	Path        string
	Name        string
	IsDirectory bool
	Permissions int32
	Level       int
}

func searchRecursive(fileManager *System.EXT2FileManager, currentPath string, currentInodeNum int32, pattern *regexp.Regexp, userID, groupID int, results *[]FindResult, level int) {
	currentInodo, err := readInode(fileManager, currentInodeNum)
	if err != nil {
		return
	}

	hasReadPermission := System.ValidateFileReadPermission(
		currentInodo.I_uid,
		currentInodo.I_gid,
		currentInodo.I_perm,
		userID,
		groupID,
	)

	if !hasReadPermission {
		return
	}

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

			entryInodo, err := readInode(fileManager, entry.B_inodo)
			if err != nil {
				continue
			}

			hasEntryReadPermission := System.ValidateFileReadPermission(
				entryInodo.I_uid,
				entryInodo.I_gid,
				entryInodo.I_perm,
				userID,
				groupID,
			)

			if !hasEntryReadPermission {
				continue
			}

			var entryPath string
			if currentPath == "/" {
				entryPath = "/" + entryName
			} else {
				entryPath = currentPath + "/" + entryName
			}

			if pattern.MatchString(entryName) {
				result := FindResult{
					Path:        entryPath,
					Name:        entryName,
					IsDirectory: entryInodo.I_type == Models.INODO_DIRECTORIO,
					Permissions: Models.GetPermissions(entryInodo.I_perm),
					Level:       level,
				}
				*results = append(*results, result)
			}

			if entryInodo.I_type == Models.INODO_DIRECTORIO {
				searchRecursive(fileManager, entryPath, entry.B_inodo, pattern, userID, groupID, results, level+1)
			}
		}
	}
}

func wildcardToRegex(pattern string) *regexp.Regexp {
	pattern = regexp.QuoteMeta(pattern)
	pattern = strings.ReplaceAll(pattern, `\?`, `.`)
	pattern = strings.ReplaceAll(pattern, `\*`, `.+`)
	pattern = "^" + pattern + "$"
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return regexp.MustCompile("^$")
	}
	return regex
}

func printFindResults(results []FindResult, basePath string) {
	if len(results) == 0 {
		return
	}

	tree := make(map[string][]FindResult)

	for _, result := range results {
		parentPath := getParentPath(result.Path)
		tree[parentPath] = append(tree[parentPath], result)
	}

	fmt.Println(basePath)
	printTreeLevel(basePath, tree, "", true)
}

func printTreeLevel(currentPath string, tree map[string][]FindResult, prefix string, isRoot bool) {
	children, exists := tree[currentPath]
	if !exists {
		return
	}

	for i, child := range children {
		isLast := i == len(children)-1

		var linePrefix string
		if isRoot && i == 0 {
			linePrefix = prefix + "   |_ "
		} else if isLast {
			linePrefix = prefix + "      |_ "
		} else {
			linePrefix = prefix + "   |_ "
		}

		fmt.Printf("%s%s\t#%d\n", linePrefix, child.Name, child.Permissions)

		if child.IsDirectory {
			var newPrefix string
			if isLast {
				newPrefix = prefix + "      "
			} else {
				newPrefix = prefix + "   "
			}
			printTreeLevel(child.Path, tree, newPrefix, false)
		}
	}
}

func getParentPath(path string) string {
	if path == "/" {
		return ""
	}

	path = strings.TrimSuffix(path, "/")
	lastSlash := strings.LastIndex(path, "/")
	if lastSlash <= 0 {
		return "/"
	}

	return path[:lastSlash]
}
