package config

import (
	"errors"
	"fmt"
)

var (
	// ErrKeyNotFound is returned when a key is not found by an engine.
	ErrKeyNotFound = errors.New("key not found")

	// ErrTypeMismatch is returned when a type mismatch for Engine get operations.
	ErrTypeMismatch = errors.New("type mismatch")

	// ErrNoSecretEngineDefined is returned by Manager.Populate when a secret config is defined but there is no secret
	// engine defined.
	ErrNoSecretEngineDefined = errors.New("no secret engine defined")

	// ErrNoPlainEngineDefined is returned by Manager.Populate when a plain config is defined but there is no plain
	// engine defined.
	ErrNoPlainEngineDefined = errors.New("no plain engine defined")

	// ErrEngineNotLoaded is returned when trying to get a key from an Engine that is not loaded.
	ErrEngineNotLoaded = errors.New("engine not loaded")

	// ErrConfigNotPointer is returned by Manager.Populate when the config is not a pointer.
	ErrConfigNotPointer = errors.New("config not pointer")
)

func newErrTypeMismatch(key string, value interface{}) error {
	return fmt.Errorf("%w: %s: %T found", ErrTypeMismatch, key, value)
}
