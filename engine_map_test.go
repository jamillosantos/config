package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_flattenMap(t *testing.T) {
	tests := []struct {
		name string
		data map[string]interface{}
		want map[string]interface{}
	}{
		{
			"receiving a shallow map",
			map[string]interface{}{
				"a": "a",
				"b": 1,
				"c": true,
			}, map[string]interface{}{
				"a": "a",
				"b": 1,
				"c": true,
			},
		},
		{
			"receiving a l2 map",
			map[string]interface{}{
				"a": map[string]interface{}{
					"d": 2,
					"e": 3,
				},
				"b": 1,
				"c": true,
			}, map[string]interface{}{
				"a.d": 2,
				"a.e": 3,
				"b":   1,
				"c":   true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := flattenMap(tt.data)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMapEngine_Load(t *testing.T) {
	mapEngine := MapEngine{}
	assert.NoError(t, mapEngine.Load())
}

func TestMapEngine_Unload(t *testing.T) {
	mapEngine := NewMapEngine(map[string]interface{}{
		"123": 456,
	})
	assert.NoError(t, mapEngine.Unload())
	assert.Nil(t, mapEngine.data)
}

func TestMapEngine_GetString(t *testing.T) {
	t.Run("not loaded", func(t *testing.T) {
		m := MapEngine{}
		_, err := m.GetString("a")
		assert.ErrorIs(t, err, ErrEngineNotLoaded)
	})

	str := "string ptr"
	var strPtrNil *string
	mapEngine := NewMapEngine(map[string]interface{}{
		"string":       "string value",
		"stringptr":    &str,
		"stringptrnil": strPtrNil,
		"nonstring":    false,
	})

	t.Run("get string", func(t *testing.T) {
		value, err := mapEngine.GetString("string")
		assert.NoError(t, err)
		assert.Equal(t, "string value", value)
	})

	t.Run("get string ptr", func(t *testing.T) {
		value, err := mapEngine.GetString("stringptr")
		assert.NoError(t, err)
		assert.Equal(t, "string ptr", value)
	})

	t.Run("get string ptr nil", func(t *testing.T) {
		value, err := mapEngine.GetString("stringptrnil")
		assert.NoError(t, err)
		assert.Equal(t, "", value)
	})

	t.Run("fail with wrong type", func(t *testing.T) {
		value, err := mapEngine.GetString("nonstring")
		assert.ErrorIs(t, err, ErrTypeMismatch)
		assert.Zero(t, value)
	})

	t.Run("fail when key does not exists", func(t *testing.T) {
		value, err := mapEngine.GetString("non existing key")
		assert.ErrorIs(t, err, ErrKeyNotFound)
		assert.Zero(t, value)
	})
}

func TestMapEngine_GetStringSlice(t *testing.T) {
	t.Run("not loaded", func(t *testing.T) {
		m := MapEngine{}
		_, err := m.GetStringSlice("a")
		assert.ErrorIs(t, err, ErrEngineNotLoaded)
	})

	wantStringSlice := []string{"string1", "string2"}
	wantInterfaceSlice := []interface{}{"string1", "string2"}
	mapEngine := NewMapEngine(map[string]interface{}{
		"stringslice":    wantStringSlice,
		"interfaceslice": wantInterfaceSlice,
		"nonstringslice": "string value",
	})

	t.Run("get string slice", func(t *testing.T) {
		value, err := mapEngine.GetStringSlice("stringslice")
		assert.NoError(t, err)
		assert.Equal(t, wantStringSlice, value)
	})

	t.Run("get string slice from an interface slice", func(t *testing.T) {
		value, err := mapEngine.GetStringSlice("interfaceslice")
		assert.NoError(t, err)
		assert.Equal(t, wantStringSlice, value)
	})

	t.Run("fail with wrong type", func(t *testing.T) {
		value, err := mapEngine.GetStringSlice("nonstringslice")
		assert.ErrorIs(t, err, ErrTypeMismatch)
		assert.Zero(t, value)
	})

	t.Run("fail when key does not exists", func(t *testing.T) {
		value, err := mapEngine.GetStringSlice("non existing key")
		assert.ErrorIs(t, err, ErrKeyNotFound)
		assert.Zero(t, value)
	})
}

func TestMapEngine_GetInt(t *testing.T) {
	t.Run("not loaded", func(t *testing.T) {
		m := MapEngine{}
		_, err := m.GetInt("a")
		assert.ErrorIs(t, err, ErrEngineNotLoaded)
	})

	wantInt := 1
	var intPtrNil *int
	mapEngine := NewMapEngine(map[string]interface{}{
		"int":       wantInt,
		"intptr":    &wantInt,
		"intptrnil": intPtrNil,
		"nonint":    "non int value",
	})

	t.Run("get int", func(t *testing.T) {
		value, err := mapEngine.GetInt("int")
		assert.NoError(t, err)
		assert.Equal(t, wantInt, value)
	})

	t.Run("get int ptr", func(t *testing.T) {
		value, err := mapEngine.GetInt("intptr")
		assert.NoError(t, err)
		assert.Equal(t, wantInt, value)
	})

	t.Run("get int ptr nil", func(t *testing.T) {
		value, err := mapEngine.GetInt("intptrnil")
		assert.NoError(t, err)
		assert.Zero(t, value)
	})

	t.Run("fail with wrong type", func(t *testing.T) {
		value, err := mapEngine.GetInt("nonint")
		assert.ErrorIs(t, err, ErrTypeMismatch)
		assert.Zero(t, value)
	})

	t.Run("fail when key does not exists", func(t *testing.T) {
		value, err := mapEngine.GetInt("non existing key")
		assert.ErrorIs(t, err, ErrKeyNotFound)
		assert.Zero(t, value)
	})
}

func TestMapEngine_GetIntSlice(t *testing.T) {
	t.Run("not loaded", func(t *testing.T) {
		m := MapEngine{}
		_, err := m.GetIntSlice("a")
		assert.ErrorIs(t, err, ErrEngineNotLoaded)
	})

	wantIntSlice := []int{1, 2}
	mapEngine := NewMapEngine(map[string]interface{}{
		"intslice":    wantIntSlice,
		"nonintslice": "string value",
	})

	t.Run("get int slice", func(t *testing.T) {
		value, err := mapEngine.GetIntSlice("intslice")
		assert.NoError(t, err)
		assert.Equal(t, wantIntSlice, value)
	})

	t.Run("fail with wrong type", func(t *testing.T) {
		value, err := mapEngine.GetIntSlice("nonintslice")
		assert.ErrorIs(t, err, ErrTypeMismatch)
		assert.Zero(t, value)
	})

	t.Run("fail when key does not exists", func(t *testing.T) {
		value, err := mapEngine.GetIntSlice("non existing key")
		assert.ErrorIs(t, err, ErrKeyNotFound)
		assert.Zero(t, value)
	})
}

func TestMapEngine_GetUint(t *testing.T) {
	t.Run("not loaded", func(t *testing.T) {
		m := MapEngine{}
		_, err := m.GetUint("a")
		assert.ErrorIs(t, err, ErrEngineNotLoaded)
	})

	wantUint := uint(1)
	var uintPtrNil *uint
	mapEngine := NewMapEngine(map[string]interface{}{
		"uint":       wantUint,
		"uintptr":    &wantUint,
		"uintptrnil": uintPtrNil,
		"nonuint":    "non uint value",
	})

	t.Run("get uint", func(t *testing.T) {
		value, err := mapEngine.GetUint("uint")
		assert.NoError(t, err)
		assert.Equal(t, wantUint, value)
	})

	t.Run("get uint ptr", func(t *testing.T) {
		value, err := mapEngine.GetUint("uintptr")
		assert.NoError(t, err)
		assert.Equal(t, wantUint, value)
	})

	t.Run("get uint ptr nil", func(t *testing.T) {
		value, err := mapEngine.GetUint("uintptrnil")
		assert.NoError(t, err)
		assert.Zero(t, value)
	})

	t.Run("fail with wrong type", func(t *testing.T) {
		value, err := mapEngine.GetUint("nonuint")
		assert.ErrorIs(t, err, ErrTypeMismatch)
		assert.Zero(t, value)
	})

	t.Run("fail when key does not exists", func(t *testing.T) {
		value, err := mapEngine.GetUint("non existing key")
		assert.ErrorIs(t, err, ErrKeyNotFound)
		assert.Zero(t, value)
	})
}

func TestMapEngine_GetUintSlice(t *testing.T) {
	t.Run("not loaded", func(t *testing.T) {
		m := MapEngine{}
		_, err := m.GetUintSlice("a")
		assert.ErrorIs(t, err, ErrEngineNotLoaded)
	})

	wantUintSlice := []uint{1, 2}
	mapEngine := NewMapEngine(map[string]interface{}{
		"uintslice":    wantUintSlice,
		"nonuintslice": "string value",
	})

	t.Run("get uint slice", func(t *testing.T) {
		value, err := mapEngine.GetUintSlice("uintslice")
		assert.NoError(t, err)
		assert.Equal(t, wantUintSlice, value)
	})

	t.Run("fail with wrong type", func(t *testing.T) {
		value, err := mapEngine.GetUintSlice("nonuintslice")
		assert.ErrorIs(t, err, ErrTypeMismatch)
		assert.Zero(t, value)
	})

	t.Run("fail when key does not exists", func(t *testing.T) {
		value, err := mapEngine.GetUintSlice("non existing key")
		assert.ErrorIs(t, err, ErrKeyNotFound)
		assert.Zero(t, value)
	})
}

func TestMapEngine_GetInt64(t *testing.T) {
	t.Run("not loaded", func(t *testing.T) {
		m := MapEngine{}
		_, err := m.GetInt64("a")
		assert.ErrorIs(t, err, ErrEngineNotLoaded)
	})

	wantInt64 := int64(1)
	var int64PtrNil *int64
	mapEngine := NewMapEngine(map[string]interface{}{
		"int64":       wantInt64,
		"int64ptr":    &wantInt64,
		"int64ptrnil": int64PtrNil,
		"nonint64":    "non int64 value",
	})

	t.Run("get int64", func(t *testing.T) {
		value, err := mapEngine.GetInt64("int64")
		assert.NoError(t, err)
		assert.Equal(t, wantInt64, value)
	})

	t.Run("get int64 ptr", func(t *testing.T) {
		value, err := mapEngine.GetInt64("int64ptr")
		assert.NoError(t, err)
		assert.Equal(t, wantInt64, value)
	})

	t.Run("get int64 ptr nil", func(t *testing.T) {
		value, err := mapEngine.GetInt64("int64ptrnil")
		assert.NoError(t, err)
		assert.Zero(t, value)
	})

	t.Run("fail with wrong type", func(t *testing.T) {
		value, err := mapEngine.GetInt64("nonint64")
		assert.ErrorIs(t, err, ErrTypeMismatch)
		assert.Zero(t, value)
	})

	t.Run("fail when key does not exists", func(t *testing.T) {
		value, err := mapEngine.GetInt64("non existing key")
		assert.ErrorIs(t, err, ErrKeyNotFound)
		assert.Zero(t, value)
	})
}

func TestMapEngine_GetInt64Slice(t *testing.T) {
	t.Run("not loaded", func(t *testing.T) {
		m := MapEngine{}
		_, err := m.GetInt64Slice("a")
		assert.ErrorIs(t, err, ErrEngineNotLoaded)
	})

	wantInt64Slice := []int64{1, 2}
	mapEngine := NewMapEngine(map[string]interface{}{
		"int64slice":    wantInt64Slice,
		"nonint64slice": "string value",
	})

	t.Run("get int64 slice", func(t *testing.T) {
		value, err := mapEngine.GetInt64Slice("int64slice")
		assert.NoError(t, err)
		assert.Equal(t, wantInt64Slice, value)
	})

	t.Run("fail with wrong type", func(t *testing.T) {
		value, err := mapEngine.GetInt64Slice("nonint64slice")
		assert.ErrorIs(t, err, ErrTypeMismatch)
		assert.Zero(t, value)
	})

	t.Run("fail when key does not exists", func(t *testing.T) {
		value, err := mapEngine.GetInt64Slice("non existing key")
		assert.ErrorIs(t, err, ErrKeyNotFound)
		assert.Zero(t, value)
	})
}

func TestMapEngine_GetUint64(t *testing.T) {
	t.Run("not loaded", func(t *testing.T) {
		m := MapEngine{}
		_, err := m.GetUint64("a")
		assert.ErrorIs(t, err, ErrEngineNotLoaded)
	})

	wantUint64 := uint64(1)
	var uint64PtrNil *uint64
	mapEngine := NewMapEngine(map[string]interface{}{
		"uint64":       wantUint64,
		"uint64ptr":    &wantUint64,
		"uint64ptrnil": uint64PtrNil,
		"nonuint64":    "non uint64 value",
	})

	t.Run("get uint64", func(t *testing.T) {
		value, err := mapEngine.GetUint64("uint64")
		assert.NoError(t, err)
		assert.Equal(t, wantUint64, value)
	})

	t.Run("get uint64 ptr", func(t *testing.T) {
		value, err := mapEngine.GetUint64("uint64ptr")
		assert.NoError(t, err)
		assert.Equal(t, wantUint64, value)
	})

	t.Run("get uint64 ptr nil", func(t *testing.T) {
		value, err := mapEngine.GetUint64("uint64ptrnil")
		assert.NoError(t, err)
		assert.Zero(t, value)
	})

	t.Run("fail with wrong type", func(t *testing.T) {
		value, err := mapEngine.GetUint64("nonuint64")
		assert.ErrorIs(t, err, ErrTypeMismatch)
		assert.Zero(t, value)
	})

	t.Run("fail when key does not exists", func(t *testing.T) {
		value, err := mapEngine.GetUint64("non existing key")
		assert.ErrorIs(t, err, ErrKeyNotFound)
		assert.Zero(t, value)
	})
}

func TestMapEngine_GetUint64Slice(t *testing.T) {
	t.Run("not loaded", func(t *testing.T) {
		m := MapEngine{}
		_, err := m.GetUint64Slice("a")
		assert.ErrorIs(t, err, ErrEngineNotLoaded)
	})

	wantUint64Slice := []uint64{1, 2}
	mapEngine := NewMapEngine(map[string]interface{}{
		"uint64slice":    wantUint64Slice,
		"nonuint64slice": "string value",
	})

	t.Run("get uint64 slice", func(t *testing.T) {
		value, err := mapEngine.GetUint64Slice("uint64slice")
		assert.NoError(t, err)
		assert.Equal(t, wantUint64Slice, value)
	})

	t.Run("fail with wrong type", func(t *testing.T) {
		value, err := mapEngine.GetUint64Slice("nonuint64slice")
		assert.ErrorIs(t, err, ErrTypeMismatch)
		assert.Zero(t, value)
	})

	t.Run("fail when key does not exists", func(t *testing.T) {
		value, err := mapEngine.GetUint64Slice("non existing key")
		assert.ErrorIs(t, err, ErrKeyNotFound)
		assert.Zero(t, value)
	})
}

func TestMapEngine_GetBool(t *testing.T) {
	t.Run("not loaded", func(t *testing.T) {
		m := MapEngine{}
		_, err := m.GetBool("a")
		assert.ErrorIs(t, err, ErrEngineNotLoaded)
	})

	wantBool := true
	var boolPtrNil *bool
	mapEngine := NewMapEngine(map[string]interface{}{
		"bool":       wantBool,
		"boolptr":    &wantBool,
		"boolptrnil": boolPtrNil,
		"nonbool":    "string value",
	})

	t.Run("get bool", func(t *testing.T) {
		value, err := mapEngine.GetBool("bool")
		assert.NoError(t, err)
		assert.Equal(t, wantBool, value)
	})

	t.Run("get bool ptr", func(t *testing.T) {
		value, err := mapEngine.GetBool("boolptr")
		assert.NoError(t, err)
		assert.Equal(t, wantBool, value)
	})

	t.Run("get bool ptr nil", func(t *testing.T) {
		value, err := mapEngine.GetBool("boolptrnil")
		assert.NoError(t, err)
		assert.Zero(t, value)
	})

	t.Run("fail with wrong type", func(t *testing.T) {
		value, err := mapEngine.GetBool("nonbool")
		assert.ErrorIs(t, err, ErrTypeMismatch)
		assert.Zero(t, value)
	})

	t.Run("fail when key does not exists", func(t *testing.T) {
		value, err := mapEngine.GetBool("non existing key")
		assert.ErrorIs(t, err, ErrKeyNotFound)
		assert.Zero(t, value)
	})
}

func TestMapEngine_GetBoolSlice(t *testing.T) {
	t.Run("not loaded", func(t *testing.T) {
		m := MapEngine{}
		_, err := m.GetBoolSlice("a")
		assert.ErrorIs(t, err, ErrEngineNotLoaded)
	})

	wantBoolSlice := []bool{true, false}
	mapEngine := NewMapEngine(map[string]interface{}{
		"boolslice":    wantBoolSlice,
		"nonboolslice": "string value",
	})

	t.Run("get bool slice", func(t *testing.T) {
		value, err := mapEngine.GetBoolSlice("boolslice")
		assert.NoError(t, err)
		assert.Equal(t, wantBoolSlice, value)
	})

	t.Run("fail with wrong type", func(t *testing.T) {
		value, err := mapEngine.GetBoolSlice("nonboolslice")
		assert.ErrorIs(t, err, ErrTypeMismatch)
		assert.Zero(t, value)
	})

	t.Run("fail when key does not exists", func(t *testing.T) {
		value, err := mapEngine.GetBoolSlice("non existing key")
		assert.ErrorIs(t, err, ErrKeyNotFound)
		assert.Zero(t, value)
	})
}

func TestMapEngine_GetFloat(t *testing.T) {
	t.Run("not loaded", func(t *testing.T) {
		m := MapEngine{}
		_, err := m.GetFloat("a")
		assert.ErrorIs(t, err, ErrEngineNotLoaded)
	})

	wantFloat := 1.1
	var floatPtrNil *float64
	mapEngine := NewMapEngine(map[string]interface{}{
		"float":       wantFloat,
		"floatptr":    &wantFloat,
		"floatptrnil": floatPtrNil,
		"nonfloat":    "string value",
	})

	t.Run("get float", func(t *testing.T) {
		value, err := mapEngine.GetFloat("float")
		assert.NoError(t, err)
		assert.Equal(t, wantFloat, value)
	})

	t.Run("get float ptr", func(t *testing.T) {
		value, err := mapEngine.GetFloat("floatptr")
		assert.NoError(t, err)
		assert.Equal(t, wantFloat, value)
	})

	t.Run("get float ptr nil", func(t *testing.T) {
		value, err := mapEngine.GetFloat("floatptrnil")
		assert.NoError(t, err)
		assert.Zero(t, value)
	})

	t.Run("fail with wrong type", func(t *testing.T) {
		value, err := mapEngine.GetFloat("nonfloat")
		assert.ErrorIs(t, err, ErrTypeMismatch)
		assert.Zero(t, value)
	})

	t.Run("fail when key does not exists", func(t *testing.T) {
		value, err := mapEngine.GetFloat("non existing key")
		assert.ErrorIs(t, err, ErrKeyNotFound)
		assert.Zero(t, value)
	})
}

func TestMapEngine_GetFloatSlice(t *testing.T) {
	t.Run("not loaded", func(t *testing.T) {
		m := MapEngine{}
		_, err := m.GetFloatSlice("a")
		assert.ErrorIs(t, err, ErrEngineNotLoaded)
	})

	wantFloatSlice := []float64{1.1, 2.2}
	mapEngine := NewMapEngine(map[string]interface{}{
		"floatslice":    wantFloatSlice,
		"nonfloatslice": "string value",
	})

	t.Run("get float slice", func(t *testing.T) {
		value, err := mapEngine.GetFloatSlice("floatslice")
		assert.NoError(t, err)
		assert.Equal(t, wantFloatSlice, value)
	})

	t.Run("fail with wrong type", func(t *testing.T) {
		value, err := mapEngine.GetFloatSlice("nonfloatslice")
		assert.ErrorIs(t, err, ErrTypeMismatch)
		assert.Zero(t, value)
	})

	t.Run("fail when key does not exists", func(t *testing.T) {
		value, err := mapEngine.GetFloatSlice("non existing key")
		assert.ErrorIs(t, err, ErrKeyNotFound)
		assert.Zero(t, value)
	})
}

func TestMapEngine_GetDuration(t *testing.T) {
	t.Run("not loaded", func(t *testing.T) {
		m := MapEngine{}
		_, err := m.GetDuration("a")
		assert.ErrorIs(t, err, ErrEngineNotLoaded)
	})

	wantDuration := time.Second
	durationStr := "1s"
	var durationStrNil *string
	var durationPtrNil *time.Duration
	mapEngine := NewMapEngine(map[string]interface{}{
		"duration":          wantDuration,
		"durationptr":       &wantDuration,
		"durationptrnil":    durationPtrNil,
		"durationstr":       durationStr,
		"durationstrptr":    &durationStr,
		"durationstrptrnil": durationStrNil,
		"nonduration":       false,
	})

	t.Run("get duration", func(t *testing.T) {
		value, err := mapEngine.GetDuration("duration")
		assert.NoError(t, err)
		assert.Equal(t, wantDuration, value)
	})

	t.Run("get duration ptr", func(t *testing.T) {
		value, err := mapEngine.GetDuration("durationptr")
		assert.NoError(t, err)
		assert.Equal(t, wantDuration, value)
	})

	t.Run("get duration ptr nil", func(t *testing.T) {
		value, err := mapEngine.GetDuration("durationptrnil")
		assert.NoError(t, err)
		assert.Zero(t, value)
	})

	t.Run("get duration string", func(t *testing.T) {
		value, err := mapEngine.GetDuration("durationstr")
		assert.NoError(t, err)
		assert.Equal(t, wantDuration, value)
	})

	t.Run("get duration string pointer", func(t *testing.T) {
		value, err := mapEngine.GetDuration("durationstrptr")
		assert.NoError(t, err)
		assert.Equal(t, wantDuration, value)
	})

	t.Run("get duration string pointer nil", func(t *testing.T) {
		value, err := mapEngine.GetDuration("durationstrptrnil")
		assert.NoError(t, err)
		assert.Zero(t, value)
	})

	t.Run("fail with wrong type", func(t *testing.T) {
		value, err := mapEngine.GetDuration("nonduration")
		assert.ErrorIs(t, err, ErrTypeMismatch)
		assert.Zero(t, value)
	})

	t.Run("fail when key does not exists", func(t *testing.T) {
		value, err := mapEngine.GetDuration("non existing key")
		assert.ErrorIs(t, err, ErrKeyNotFound)
		assert.Zero(t, value)
	})
}
