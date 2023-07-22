package run

import (
	"os"
	"path/filepath"
)

// Cleanup removes all temporary files: *.stdout, *.finish
func Cleanup() {
	deleteFilesByExtension("stdout")
	deleteFilesByExtension("finish")
}

// deleteFilesByExtension removes all files under the ./cmd directory by extension.
func deleteFilesByExtension(ext string) {
	ext = "." + ext

	d, err := os.Open(CmdFolder)
	L.Error(err)
	defer d.Close()

	files, err := d.Readdir(-1)
	L.Error(err)

	L.Info("reading", CmdFolder)

	for _, file := range files {
		if file.Mode().IsRegular() {
			if filepath.Ext(file.Name()) == ext {
				filepath := CmdFolder + file.Name()
				_ = os.Remove(filepath)
				L.Info("deleted", filepath)
			}
		}
	}
}
