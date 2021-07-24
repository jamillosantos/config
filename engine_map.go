package config

import (
	"fmt"
	"time"
)

type MapEngine struct {
	data map[string]interface{}
}

// NewMapEngine returns a new instance of MapEngine with the given data.
//
// Internally, it will flatten the data map before storing for future use.
func NewMapEngine(data map[string]interface{}) *MapEngine {
	return &MapEngine{flattenMap(data)}
}

func (engine *MapEngine) Load() error {
	return nil
}

func (engine *MapEngine) Unload() error {
	engine.data = nil
	return nil
}

func (engine *MapEngine) GetString(key string) (string, error) {
	if engine.data == nil {
		return "", ErrEngineNotLoaded
	}
	value, ok := engine.data[key]
	if !ok {
		return "", fmt.Errorf("%w: %s", ErrKeyNotFound, key)
	}
	switch t := value.(type) {
	case string:
		return t, nil
	case *string:
		if t == nil {
			return "", nil
		}
		return *t, nil
	}
	return "", newErrTypeMismatch(value)
}

func (engine *MapEngine) GetStringSlice(key string) ([]string, error) {
	if engine.data == nil {
		return nil, ErrEngineNotLoaded
	}
	value, ok := engine.data[key]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrKeyNotFound, key)
	}
	switch t := value.(type) {
	case []string:
		return t, nil
	case []interface{}:
		return convertInterfaceSliceToStringSlice(t)
	}
	return nil, newErrTypeMismatch(value)
}

func (engine *MapEngine) GetInt(key string) (int, error) {
	if engine.data == nil {
		return 0, ErrEngineNotLoaded
	}
	value, ok := engine.data[key]
	if !ok {
		return 0, fmt.Errorf("%w: %s", ErrKeyNotFound, key)
	}
	switch t := value.(type) {
	case int:
		return t, nil
	case *int:
		if t == nil {
			return 0, nil
		}
		return *t, nil
	}
	return 0, newErrTypeMismatch(value)
}

func (engine *MapEngine) GetIntSlice(key string) ([]int, error) {
	if engine.data == nil {
		return nil, ErrEngineNotLoaded
	}
	value, ok := engine.data[key]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrKeyNotFound, key)
	}
	switch t := value.(type) {
	case []int:
		return t, nil
	case []interface{}:
		return convertInterfaceSliceToIntSlice(t)
	}
	return nil, newErrTypeMismatch(value)
}

func (engine *MapEngine) GetUint(key string) (uint, error) {
	if engine.data == nil {
		return 0, ErrEngineNotLoaded
	}
	value, ok := engine.data[key]
	if !ok {
		return 0, fmt.Errorf("%w: %s", ErrKeyNotFound, key)
	}
	switch t := value.(type) {
	case uint:
		return t, nil
	case *uint:
		if t == nil {
			return 0, nil
		}
		return *t, nil
	}
	return 0, newErrTypeMismatch(value)
}

func (engine *MapEngine) GetUintSlice(key string) ([]uint, error) {
	if engine.data == nil {
		return nil, ErrEngineNotLoaded
	}
	value, ok := engine.data[key]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrKeyNotFound, key)
	}
	switch t := value.(type) {
	case []uint:
		return t, nil
	case []interface{}:
		return convertInterfaceSliceToUintSlice(t)
	}
	return nil, newErrTypeMismatch(value)
}

func (engine *MapEngine) GetInt64(key string) (int64, error) {
	if engine.data == nil {
		return 0, ErrEngineNotLoaded
	}
	value, ok := engine.data[key]
	if !ok {
		return 0, fmt.Errorf("%w: %s", ErrKeyNotFound, key)
	}
	switch t := value.(type) {
	case int64:
		return t, nil
	case *int64:
		if t == nil {
			return 0, nil
		}
		return *t, nil
	}
	return 0, newErrTypeMismatch(value)
}

func (engine *MapEngine) GetInt64Slice(key string) ([]int64, error) {
	if engine.data == nil {
		return nil, ErrEngineNotLoaded
	}
	value, ok := engine.data[key]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrKeyNotFound, key)
	}
	switch t := value.(type) {
	case []int64:
		return t, nil
	case []interface{}:
		return convertInterfaceSliceToInt64Slice(t)
	}
	return nil, newErrTypeMismatch(value)
}

func (engine *MapEngine) GetUint64(key string) (uint64, error) {
	if engine.data == nil {
		return 0, ErrEngineNotLoaded
	}
	value, ok := engine.data[key]
	if !ok {
		return 0, fmt.Errorf("%w: %s", ErrKeyNotFound, key)
	}
	switch t := value.(type) {
	case uint64:
		return t, nil
	case *uint64:
		if t == nil {
			return 0, nil
		}
		return *t, nil
	}
	return 0, newErrTypeMismatch(value)
}

func (engine *MapEngine) GetUint64Slice(key string) ([]uint64, error) {
	if engine.data == nil {
		return nil, ErrEngineNotLoaded
	}
	value, ok := engine.data[key]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrKeyNotFound, key)
	}
	switch t := value.(type) {
	case []uint64:
		return t, nil
	case []interface{}:
		return convertInterfaceSliceToUint64Slice(t)
	}
	return nil, newErrTypeMismatch(value)
}

func (engine *MapEngine) GetBool(key string) (bool, error) {
	if engine.data == nil {
		return false, ErrEngineNotLoaded
	}
	value, ok := engine.data[key]
	if !ok {
		return false, fmt.Errorf("%w: %s", ErrKeyNotFound, key)
	}
	switch t := value.(type) {
	case bool:
		return t, nil
	case *bool:
		if t == nil {
			return false, nil
		}
		return *t, nil
	}
	return false, newErrTypeMismatch(value)
}

func (engine *MapEngine) GetBoolSlice(key string) ([]bool, error) {
	if engine.data == nil {
		return nil, ErrEngineNotLoaded
	}
	value, ok := engine.data[key]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrKeyNotFound, key)
	}
	switch t := value.(type) {
	case []bool:
		return t, nil
	case []interface{}:
		return convertInterfaceSliceToBoolSlice(t)
	}
	return nil, newErrTypeMismatch(value)
}

func (engine *MapEngine) GetFloat(key string) (float64, error) {
	if engine.data == nil {
		return 0, ErrEngineNotLoaded
	}
	value, ok := engine.data[key]
	if !ok {
		return 0, fmt.Errorf("%w: %s", ErrKeyNotFound, key)
	}
	switch t := value.(type) {
	case float64:
		return t, nil
	case *float64:
		if t == nil {
			return 0, nil
		}
		return *t, nil
	}
	return 0, newErrTypeMismatch(value)
}

func (engine *MapEngine) GetFloatSlice(key string) ([]float64, error) {
	if engine.data == nil {
		return nil, ErrEngineNotLoaded
	}
	value, ok := engine.data[key]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrKeyNotFound, key)
	}
	switch t := value.(type) {
	case []float64:
		return t, nil
	case []interface{}:
		return convertInterfaceSliceToFloat64Slice(t)
	}
	return nil, newErrTypeMismatch(value)
}

func (engine *MapEngine) GetDuration(key string) (time.Duration, error) {
	if engine.data == nil {
		return 0, ErrEngineNotLoaded
	}
	value, ok := engine.data[key]
	if !ok {
		return 0, fmt.Errorf("%w: %s", ErrKeyNotFound, key)
	}
	switch t := value.(type) {
	case string:
		return time.ParseDuration(t)
	case *string:
		if t == nil {
			return 0, nil
		}
		return time.ParseDuration(*t)
	case time.Duration:
		return t, nil
	case *time.Duration:
		if t == nil {
			return 0, nil
		}
		return *t, nil
	}
	return 0, newErrTypeMismatch(value)
}

func flattenMap(data map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range data {
		switch v := v.(type) {
		case map[string]interface{}:
			for kk, vv := range flattenMap(v) {
				result[k+"."+kk] = vv
			}
		default:
			result[k] = v
		}
	}
	return result
}

func convertInterfaceSliceToStringSlice(data []interface{}) ([]string, error) {
	result := make([]string, len(data))
	for i, v := range data {
		switch v := v.(type) {
		case string:
			result[i] = v
		case *string:
			if v != nil {
				result[i] = *v
			}
		default:
			return nil, newErrTypeMismatch(v)
		}
	}
	return result, nil
}

func convertInterfaceSliceToIntSlice(data []interface{}) ([]int, error) {
	result := make([]int, len(data))
	for i, v := range data {
		switch v := v.(type) {
		case int:
			result[i] = v
		case *int:
			if v != nil {
				result[i] = *v
			}
		default:
			return nil, newErrTypeMismatch(v)
		}
	}
	return result, nil
}

func convertInterfaceSliceToUintSlice(data []interface{}) ([]uint, error) {
	result := make([]uint, len(data))
	for i, v := range data {
		switch v := v.(type) {
		case uint:
			result[i] = v
		case *uint:
			if v != nil {
				result[i] = *v
			}
		default:
			return nil, newErrTypeMismatch(v)
		}
	}
	return result, nil
}

func convertInterfaceSliceToInt64Slice(data []interface{}) ([]int64, error) {
	result := make([]int64, len(data))
	for i, v := range data {
		switch v := v.(type) {
		case int64:
			result[i] = v
		case *int64:
			if v != nil {
				result[i] = *v
			}
		default:
			return nil, newErrTypeMismatch(v)
		}
	}
	return result, nil
}

func convertInterfaceSliceToUint64Slice(data []interface{}) ([]uint64, error) {
	result := make([]uint64, len(data))
	for i, v := range data {
		switch v := v.(type) {
		case uint64:
			result[i] = v
		case *uint64:
			if v != nil {
				result[i] = *v
			}
		default:
			return nil, newErrTypeMismatch(v)
		}
	}
	return result, nil
}

func convertInterfaceSliceToBoolSlice(data []interface{}) ([]bool, error) {
	result := make([]bool, len(data))
	for i, v := range data {
		switch v := v.(type) {
		case bool:
			result[i] = v
		case *bool:
			if v != nil {
				result[i] = *v
			}
		default:
			return nil, newErrTypeMismatch(v)
		}
	}
	return result, nil
}

func convertInterfaceSliceToFloat64Slice(data []interface{}) ([]float64, error) {
	result := make([]float64, len(data))
	for i, v := range data {
		switch v := v.(type) {
		case float64:
			result[i] = v
		case *float64:
			if v != nil {
				result[i] = *v
			}
		default:
			return nil, newErrTypeMismatch(v)
		}
	}
	return result, nil
}
