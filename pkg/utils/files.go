package utils

import (
	"fmt"
	"io"
	"io/fs"
	"strings"
)

// ReadDirFiles read all file in input fs and dir
func ReadDirFiles(fsys fs.FS, dir string, handler func(data []byte) error) error {
	d, err := fsys.Open(dir)
	if err != nil {
		return fmt.Errorf("failed to open path %s with error: %w", dir, err)
	}
	defer d.Close()
	entries, err := d.(fs.ReadDirFile).ReadDir(-1)
	if err != nil {
		return fmt.Errorf("read dir %s failed with error: %w", dir, err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			// skip dir
			continue
		}
		if !HasSuffixIn(entry.Name(), []string{"yaml", "yml"}) {
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

// HasSuffixIn check input string a end with one of slice b
func HasSuffixIn(a string, b []string) bool {
	for _, suffix := range b {
		if strings.HasSuffix(a, suffix) {
			return true
		}
	}
	return false
}
