package config

import (
	"bytes"
	"io"
)

// BytesLoader receives a byte slice and implements the Loader interface.
type BytesLoader struct {
	bytes []byte
}

// NewBytesLoader creates a new BytesLoader.
func NewBytesLoader(bytes []byte) *BytesLoader {
	return &BytesLoader{bytes}
}

// Load returns a io.Reader from the given bytes (check NewBytesLoader).
func (loader *BytesLoader) Load() (io.Reader, error) {
	return bytes.NewReader(loader.bytes), nil
}

// Unload sets the internal pointer to the given slice to nil.
func (loader *BytesLoader) Unload() error {
	loader.bytes = nil
	return nil
}
