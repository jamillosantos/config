package config

import "time"

// Engine is an interface that provides the contract for configuration engines
// to be able to read configuration from a variety of sources.
type Engine interface {
	Load() error
	Unload() error

	GetString(key string) (string, error)
	GetStringSlice(key string) ([]string, error)

	GetInt(key string) (int, error)
	GetIntSlice(key string) ([]int, error)

	GetUint(key string) (uint, error)
	GetUintSlice(key string) ([]uint, error)

	GetInt64(key string) (int64, error)
	GetInt64Slice(key string) ([]int64, error)

	GetUint64(key string) (uint64, error)
	GetUint64Slice(key string) ([]uint64, error)

	GetBool(key string) (bool, error)
	GetBoolSlice(key string) ([]bool, error)

	GetFloat(key string) (float64, error)
	GetFloatSlice(key string) ([]float64, error)

	GetDuration(key string) (time.Duration, error)
}
