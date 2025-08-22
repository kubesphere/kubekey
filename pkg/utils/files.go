package utils

import (
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"sort"
)

// ReadDirFiles read all file in input fs and dir
func ReadDirFiles(fsys fs.FS, dir string, handler func(data []byte) error) error {
	entries, err := fs.ReadDir(fsys, dir)
	if err != nil {
		return fmt.Errorf("failed to read dir %q: %w", dir, err)
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})
	for _, entry := range entries {
		if entry.IsDir() {
			// skip dir
			continue
		}
		if filepath.Ext(entry.Name()) != ".yaml" && filepath.Ext(entry.Name()) != ".yml" {
			continue
		}
		filePath := dir + "/" + entry.Name()
		if dir == "." {
			filePath = entry.Name()
		}
		// open file
		file, err := fsys.Open(filePath)
		if err != nil {
			return fmt.Errorf("failed to open file %q with error: %w", filePath, err)
		}
		// read file content
		content, err := io.ReadAll(file)
		if err != nil {
			return fmt.Errorf("read file %q failed with error: %w", filePath, err)
		}
		err = file.Close()
		if err != nil {
			return fmt.Errorf("close file %q failed with error: %w", filePath, err)
		}
		err = handler(content)
		if err != nil {
			return fmt.Errorf("handle file %q failed with error: %w", filePath, err)
		}
	}
	return nil
}
