package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	ErrUnterminatedString = fmt.Errorf("unterminated string")
)

type EnvEngine struct {
	prefix string
}

type EnvOption func(engine *EnvEngine)

func WithPrefix(prefix string) EnvOption {
	return func(engine *EnvEngine) {
		engine.prefix = prefix
	}

}

func NewEnvEngine(opts ...EnvOption) EnvEngine {
	engine := EnvEngine{}
	for _, opt := range opts {
		opt(&engine)
	}
	return engine
}

func (e *EnvEngine) Load() error {
	return nil
}

func (e *EnvEngine) Unload() error {
	return nil
}

func (e *EnvEngine) getKey(key string) string {
	return strings.ToUpper(strings.ReplaceAll(e.prefix+key, ".", "_"))
}

func (e *EnvEngine) GetString(key string) (string, error) {
	value, ok := os.LookupEnv(e.getKey(key))
	if !ok {
		return "", ErrKeyNotFound
	}
	return value, nil
}

func (e *EnvEngine) GetStringSlice(key string) ([]string, error) {
	value, ok := os.LookupEnv(e.getKey(key))
	if !ok {
		return nil, ErrKeyNotFound
	}
	if value == "" {
		return nil, nil
	}
	return strings.Split(value, ","), nil
}

func (e *EnvEngine) GetInt(key string) (int, error) {
	value, ok := os.LookupEnv(e.getKey(key))
	if !ok {
		return 0, ErrKeyNotFound
	}
	return strconv.Atoi(value)
}

func (e *EnvEngine) GetIntSlice(key string) ([]int, error) {
	value, ok := os.LookupEnv(e.getKey(key))
	if !ok {
		return nil, ErrKeyNotFound
	}
	if value == "" {
		return nil, nil
	}
	values := strings.Split(value, ",")
	result := make([]int, 0, len(values))
	for i, v := range values {
		u, err := strconv.Atoi(strings.TrimSpace(v))
		if err != nil {
			return nil, fmt.Errorf("element %d: %w", i, err)
		}
		result = append(result, u)
	}
	return result, nil
}

func (e *EnvEngine) GetUint(key string) (uint, error) {
	value, ok := os.LookupEnv(e.getKey(key))
	if !ok {
		return 0, ErrKeyNotFound
	}
	v, err := strconv.ParseUint(value, 10, 32)
	return uint(v), err
}

func (e *EnvEngine) GetUintSlice(key string) ([]uint, error) {
	value, err := os.LookupEnv(e.getKey(key))
	if !err {
		return nil, ErrKeyNotFound
	}
	if value == "" {
		return nil, nil
	}
	values := strings.Split(value, ",")
	result := make([]uint, 0, len(values))
	for i, v := range values {
		u, err := strconv.ParseUint(strings.TrimSpace(v), 10, 32)
		if err != nil {
			return nil, fmt.Errorf("element %d: %w", i, err)
		}
		result = append(result, uint(u))
	}
	return result, nil
}

func (e *EnvEngine) GetInt64(key string) (int64, error) {
	value, ok := os.LookupEnv(e.getKey(key))
	if !ok {
		return 0, ErrKeyNotFound
	}
	return strconv.ParseInt(value, 10, 64)
}

func (e *EnvEngine) GetInt64Slice(key string) ([]int64, error) {
	value, ok := os.LookupEnv(e.getKey(key))
	if !ok {
		return nil, ErrKeyNotFound
	}
	if value == "" {
		return nil, nil
	}
	values := strings.Split(value, ",")
	result := make([]int64, 0, len(values))
	for i, v := range values {
		u, err := strconv.ParseInt(strings.TrimSpace(v), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("element %d: %w", i, err)
		}
		result = append(result, u)
	}
	return result, nil
}

func (e *EnvEngine) GetUint64(key string) (uint64, error) {
	value, ok := os.LookupEnv(e.getKey(key))
	if !ok {
		return 0, ErrKeyNotFound
	}
	return strconv.ParseUint(value, 10, 64)
}

func (e *EnvEngine) GetUint64Slice(key string) ([]uint64, error) {
	value, ok := os.LookupEnv(e.getKey(key))
	if !ok {
		return nil, ErrKeyNotFound
	}
	if value == "" {
		return nil, nil
	}
	values := strings.Split(value, ",")
	result := make([]uint64, 0, len(values))
	for i, v := range values {
		u, err := strconv.ParseUint(strings.TrimSpace(v), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("element %d: %w", i, err)
		}
		result = append(result, u)
	}
	return result, nil
}

func (e *EnvEngine) GetBool(key string) (bool, error) {
	value, ok := os.LookupEnv(e.getKey(key))
	if !ok {
		return false, ErrKeyNotFound
	}
	return strconv.ParseBool(value)
}

func (e *EnvEngine) GetBoolSlice(key string) ([]bool, error) {
	value, ok := os.LookupEnv(e.getKey(key))
	if !ok {
		return nil, ErrKeyNotFound
	}
	if value == "" {
		return nil, nil
	}
	values := strings.Split(value, ",")
	result := make([]bool, 0, len(values))
	for i, v := range values {
		b, err := strconv.ParseBool(v)
		if err != nil {
			return nil, fmt.Errorf("element %d: %w", i, err)
		}
		result = append(result, b)
	}
	return result, nil
}

func (e *EnvEngine) GetFloat(key string) (float64, error) {
	value, ok := os.LookupEnv(e.getKey(key))
	if !ok {
		return 0, ErrKeyNotFound
	}
	return strconv.ParseFloat(value, 64)
}

func (e *EnvEngine) GetFloatSlice(key string) ([]float64, error) {
	value, ok := os.LookupEnv(e.getKey(key))
	if !ok {
		return nil, ErrKeyNotFound
	}
	if value == "" {
		return nil, nil
	}
	values := strings.Split(value, ",")
	result := make([]float64, 0, len(values))
	for i, v := range values {
		f, err := strconv.ParseFloat(strings.TrimSpace(v), 64)
		if err != nil {
			return nil, fmt.Errorf("element %d: %w", i, err)
		}
		result = append(result, f)
	}
	return result, nil
}

func (e *EnvEngine) GetDuration(key string) (time.Duration, error) {
	value, ok := os.LookupEnv(e.getKey(key))
	if !ok {
		return 0, ErrKeyNotFound
	}
	return time.ParseDuration(value)
}
