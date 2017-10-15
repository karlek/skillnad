package skillnad

import (
	"path/filepath"
	"strings"
)

func RemoveExt(path string) string {
	index := strings.LastIndex(path, filepath.Ext(path))
	if index == -1 {
		return path
	}
	return path[:index]
}
