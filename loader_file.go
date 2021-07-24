package config

import (
	"io"
	"os"
)

// FileLoader is a config loader that loads a file. This can be used to load
// JSON, YAML, or INI files, according with the Engine that is being used.
type FileLoader struct {
	filePath    string
	fileHandler *os.File
}

func NewFileLoader(filePath string) *FileLoader {
	return &FileLoader{filePath, nil}
}

// Load loads the given filePath (check NewFileLoader) saving the file handler
// for further use.
func (loader *FileLoader) Load() (io.Reader, error) {
	f, err := os.Open(loader.filePath)
	if err != nil {
		return nil, err
	}

	loader.fileHandler = f

	return f, nil
}

// Unload closes the file handler.
func (loader *FileLoader) Unload() error {
	if loader.fileHandler == nil {
		return nil
	}
	return loader.fileHandler.Close()
}
