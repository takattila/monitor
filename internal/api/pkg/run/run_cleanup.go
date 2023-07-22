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
	dirname := "./cmd/"

	d, err := os.Open(dirname)
	if err != nil {
		L.Error(err)
		return
	}
	defer d.Close()

	files, err := d.Readdir(-1)
	if err != nil {
		L.Error(err)
		return
	}

	L.Info("reading", dirname)

	for _, file := range files {
		if file.Mode().IsRegular() {
			if filepath.Ext(file.Name()) == ext {
				filepath := dirname + file.Name()
				_ = os.Remove(filepath)
				L.Info("deleted", filepath)
			}
		}
	}
}
