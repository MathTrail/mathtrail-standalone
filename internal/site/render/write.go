package render

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	dirPerm  = 0o755
	filePerm = 0o644
)

// Write replaces dir with the rendered files. The directory is emptied first,
// so a page that no longer exists stops being published instead of lingering
// beside the ones that replaced it.
func Write(dir string, files []File) error {
	if err := os.RemoveAll(dir); err != nil {
		return fmt.Errorf("render: clear %s: %w", dir, err)
	}
	for _, file := range files {
		target := filepath.Join(dir, filepath.FromSlash(file.Path))
		if err := os.MkdirAll(filepath.Dir(target), dirPerm); err != nil {
			return fmt.Errorf("render: create directory for %s: %w", file.Path, err)
		}
		if err := os.WriteFile(target, file.Data, filePerm); err != nil {
			return fmt.Errorf("render: write %s: %w", file.Path, err)
		}
	}
	return nil
}
