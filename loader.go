package config

import (
	"io"
)

// Loader is the interface for loading configuration from a source.
type Loader interface {
	// Load returns an io.Reader for reading the data to be unmarshaled by the
	// Engine implementation.
	Load() (io.Reader, error)
	// Unload is called when the configuration is no longer needed and all the
	// resources should be released.
	Unload() error
}
