package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func withEnvironment(env map[string]string, f func()) {
	remove := make([]string, 0)
	restore := make(map[string]string)
	for key, value := range env {
		if oldValue, exists := os.LookupEnv(key); exists {
			restore[key] = oldValue
		} else {
			remove = append(remove, key)
		}
		_ = os.Setenv(key, value)
	}
	defer func() {
		for key, value := range restore {
			_ = os.Setenv(key, value)
		}
		for _, key := range remove {
			_ = os.Unsetenv(key)
		}
	}()
	f()
}

func TestEnvEngine_GetString(t *testing.T) {
	e := NewEnvEngine()
	wantString := "test_value"
	withEnvironment(map[string]string{
		"TEST_KEY": wantString,
	}, func() {
		t.Run("when key exists using dot notation", func(t *testing.T) {
			value, err := e.GetString("test.key")
			require.NoError(t, err)
			assert.Equal(t, wantString, value)
		})

		t.Run("when key exists using underscore notation", func(t *testing.T) {
			value, err := e.GetString("test_key")
			require.NoError(t, err)
			assert.Equal(t, wantString, value)
		})

		t.Run("when key exists using UPPERCASE", func(t *testing.T) {
			value, err := e.GetString("TEST_KEY")
			require.NoError(t, err)
			assert.Equal(t, wantString, value)
		})

		t.Run("when key does not exist", func(t *testing.T) {
			value, err := e.GetString("non_existing_key")
			require.NoError(t, err)
			assert.Empty(t, value)
		})
	})
}

func TestEnvEngine_GetStringSlice(t *testing.T) {
	e := NewEnvEngine()
	wantSlice := []string{"value1", "value2"}
	withEnvironment(map[string]string{
		"TEST_KEY": "value1,value2",
	}, func() {
		value, err := e.GetStringSlice("test.key")
		require.NoError(t, err)
		assert.Equal(t, wantSlice, value)

		value, err = e.GetStringSlice("non_exiting_key")
		require.NoError(t, err)
		assert.Nil(t, value)
	})
}

func TestEnvEngine_GetInt(t *testing.T) {
	e := NewEnvEngine()
	wantInt := 42
	withEnvironment(map[string]string{
		"TEST_KEY":  "42",
		"TEST_KEY2": "invalid_int",
	}, func() {
		t.Run("when key exists and valid int", func(t *testing.T) {
			value, err := e.GetInt("test.key")
			require.NoError(t, err)
			assert.Equal(t, wantInt, value)
		})

		t.Run("when key does not exist", func(t *testing.T) {
			value, err := e.GetInt("non_existing_key")
			require.NoError(t, err)
			assert.Zero(t, value)
		})

		t.Run("when key exists but invalid int", func(t *testing.T) {
			_, err := e.GetInt("test.key2")
			require.Error(t, err)
			require.Contains(t, err.Error(), "invalid_int")
		})
	})
}

func TestEnvEngine_GetIntSlice(t *testing.T) {
	e := NewEnvEngine()
	wantSlice := []int{1, 2, 3}
	withEnvironment(map[string]string{
		"TEST_KEY":  "1,2 ,3",
		"TEST_KEY2": "1,invalid_int,3",
	}, func() {
		t.Run("when key exists and valid int slice", func(t *testing.T) {
			value, err := e.GetIntSlice("test.key")
			require.NoError(t, err)
			assert.Equal(t, wantSlice, value)
		})

		t.Run("when key does not exist", func(t *testing.T) {
			value, err := e.GetIntSlice("non_existing_key")
			require.NoError(t, err)
			assert.Nil(t, value)
		})

		t.Run("when key exists but invalid int slice", func(t *testing.T) {
			_, err := e.GetIntSlice("test.key2")
			require.Error(t, err)
			require.Contains(t, err.Error(), "element 1")
		})
	})
}

func TestEnvEngine_GetUint(t *testing.T) {
	e := NewEnvEngine()
	wantUint := uint(42)
	withEnvironment(map[string]string{
		"TEST_KEY":  "42",
		"TEST_KEY2": "invalid_uint",
	}, func() {
		t.Run("when key exists and valid int", func(t *testing.T) {
			value, err := e.GetUint("test.key")
			require.NoError(t, err)
			assert.Equal(t, wantUint, value)
		})

		t.Run("when key does not exist", func(t *testing.T) {
			value, err := e.GetUint("non_existing_key")
			require.NoError(t, err)
			assert.Zero(t, value)
		})

		t.Run("when key exists but invalid int", func(t *testing.T) {
			_, err := e.GetUint("test.key2")
			require.Error(t, err)
			require.Contains(t, err.Error(), "invalid_uint")
		})
	})
}

func TestEnvEngine_GetUintSlice(t *testing.T) {
	e := NewEnvEngine()
	wantSlice := []uint{1, 2, 3}
	withEnvironment(map[string]string{
		"TEST_KEY":  "1,2 ,3",
		"TEST_KEY2": "1,invalid_uint,3",
	}, func() {
		t.Run("when key exists and valid uint slice", func(t *testing.T) {
			value, err := e.GetUintSlice("test.key")
			require.NoError(t, err)
			assert.Equal(t, wantSlice, value)
		})

		t.Run("when key does not exist", func(t *testing.T) {
			value, err := e.GetUintSlice("non_existing_key")
			require.NoError(t, err)
			assert.Nil(t, value)
		})

		t.Run("when key exists but invalid int slice", func(t *testing.T) {
			_, err := e.GetUintSlice("test.key2")
			require.Error(t, err)
			require.Contains(t, err.Error(), "element 1")
		})
	})
}

func TestEnvEngine_GetInt64(t *testing.T) {
	e := NewEnvEngine()
	wantInt64 := int64(42)
	withEnvironment(map[string]string{
		"TEST_KEY":  "42",
		"TEST_KEY2": "invalid_int64",
	}, func() {
		t.Run("when key exists and valid int", func(t *testing.T) {
			value, err := e.GetInt64("test.key")
			require.NoError(t, err)
			assert.Equal(t, wantInt64, value)
		})

		t.Run("when key does not exist", func(t *testing.T) {
			value, err := e.GetInt64("non_existing_key")
			require.NoError(t, err)
			assert.Zero(t, value)
		})

		t.Run("when key exists but invalid int", func(t *testing.T) {
			_, err := e.GetInt64("test.key2")
			require.Error(t, err)
			require.Contains(t, err.Error(), "invalid_int64")
		})
	})
}

func TestEnvEngine_GetInt64Slice(t *testing.T) {
	e := NewEnvEngine()
	wantSlice := []int64{1, 2, 3}
	withEnvironment(map[string]string{
		"TEST_KEY":  "1,2 ,3",
		"TEST_KEY2": "1,invalid_int64,3",
	}, func() {
		t.Run("when key exists and valid int slice", func(t *testing.T) {
			value, err := e.GetInt64Slice("test.key")
			require.NoError(t, err)
			assert.Equal(t, wantSlice, value)
		})

		t.Run("when key does not exist", func(t *testing.T) {
			value, err := e.GetInt64Slice("non_existing_key")
			require.NoError(t, err)
			assert.Nil(t, value)
		})

		t.Run("when key exists but invalid int slice", func(t *testing.T) {
			_, err := e.GetInt64Slice("test.key2")
			require.Error(t, err)
			require.Contains(t, err.Error(), "element 1")
		})
	})
}

func TestEnvEngine_GetUint64(t *testing.T) {
	e := NewEnvEngine()
	wantUint64 := uint64(42)
	withEnvironment(map[string]string{
		"TEST_KEY":  "42",
		"TEST_KEY2": "invalid_uint64",
	}, func() {
		t.Run("when key exists and valid int", func(t *testing.T) {
			value, err := e.GetUint64("test.key")
			require.NoError(t, err)
			assert.Equal(t, wantUint64, value)
		})

		t.Run("when key does not exist", func(t *testing.T) {
			value, err := e.GetUint64("non_existing_key")
			require.NoError(t, err)
			assert.Zero(t, value)
		})

		t.Run("when key exists but invalid int", func(t *testing.T) {
			_, err := e.GetUint64("test.key2")
			require.Error(t, err)
			require.Contains(t, err.Error(), "invalid_uint64")
		})
	})
}

func TestEnvEngine_GetUint64Slice(t *testing.T) {
	e := NewEnvEngine()
	wantSlice := []uint64{1, 2, 3}
	withEnvironment(map[string]string{
		"TEST_KEY":  "1,2 ,3",
		"TEST_KEY2": "1,invalid_uint64,3",
	}, func() {
		t.Run("when key exists and valid uint slice", func(t *testing.T) {
			value, err := e.GetUint64Slice("test.key")
			require.NoError(t, err)
			assert.Equal(t, wantSlice, value)
		})

		t.Run("when key does not exist", func(t *testing.T) {
			value, err := e.GetUint64Slice("non_existing_key")
			require.NoError(t, err)
			assert.Nil(t, value)
		})

		t.Run("when key exists but invalid int slice", func(t *testing.T) {
			_, err := e.GetUint64Slice("test.key2")
			require.Error(t, err)
			require.Contains(t, err.Error(), "element 1")
		})
	})
}

func TestEnvEngine_GetBool(t *testing.T) {
	e := NewEnvEngine()
	withEnvironment(map[string]string{
		"TEST_KEY":  "true",
		"TEST_KEY2": "invalid_bool",
	}, func() {
		t.Run("when key exists and valid bool", func(t *testing.T) {
			value, err := e.GetBool("test.key")
			require.NoError(t, err)
			assert.True(t, value)
		})

		t.Run("when key does not exist", func(t *testing.T) {
			value, err := e.GetBool("non_existing_key")
			require.NoError(t, err)
			assert.False(t, value)
		})

		t.Run("when key exists but invalid bool", func(t *testing.T) {
			_, err := e.GetBool("test.key2")
			require.Error(t, err)
			require.Contains(t, err.Error(), "invalid_bool")
		})
	})
}

func TestEnvEngine_GetBoolSlice(t *testing.T) {
	e := NewEnvEngine()
	withEnvironment(map[string]string{
		"TEST_KEY":  "true,false,true",
		"TEST_KEY2": "true,invalid_bool,true",
	}, func() {
		t.Run("when key exists and valid bool slice", func(t *testing.T) {
			value, err := e.GetBoolSlice("test.key")
			require.NoError(t, err)
			assert.Equal(t, []bool{true, false, true}, value)
		})

		t.Run("when key does not exist", func(t *testing.T) {
			value, err := e.GetBoolSlice("non_existing_key")
			require.NoError(t, err)
			assert.Nil(t, value)
		})

		t.Run("when key exists but invalid bool slice", func(t *testing.T) {
			_, err := e.GetBoolSlice("test.key2")
			require.Error(t, err)
			require.Contains(t, err.Error(), "element 1")
		})
	})
}

func TestEnvEngine_GetFloat(t *testing.T) {
	e := NewEnvEngine()
	withEnvironment(map[string]string{
		"TEST_KEY":  "42.42",
		"TEST_KEY2": "invalid_float",
	}, func() {
		t.Run("when key exists and valid float", func(t *testing.T) {
			value, err := e.GetFloat("test.key")
			require.NoError(t, err)
			assert.Equal(t, 42.42, value)
		})

		t.Run("when key does not exist", func(t *testing.T) {
			value, err := e.GetFloat("non_existing_key")
			require.NoError(t, err)
			assert.Zero(t, value)
		})

		t.Run("when key exists but invalid float", func(t *testing.T) {
			_, err := e.GetFloat("test.key2")
			require.Error(t, err)
			require.Contains(t, err.Error(), "invalid_float")
		})
	})
}

func TestEnvEngine_GetFloatSlice(t *testing.T) {
	e := NewEnvEngine()
	withEnvironment(map[string]string{
		"TEST_KEY":  "42.42,42.42,42.42",
		"TEST_KEY2": "42.42,invalid_float,42.42",
	}, func() {
		t.Run("when key exists and valid float slice", func(t *testing.T) {
			value, err := e.GetFloatSlice("test.key")
			require.NoError(t, err)
			assert.Equal(t, []float64{42.42, 42.42, 42.42}, value)
		})

		t.Run("when key does not exist", func(t *testing.T) {
			value, err := e.GetFloatSlice("non_existing_key")
			require.NoError(t, err)
			assert.Nil(t, value)
		})

		t.Run("when key exists but invalid float slice", func(t *testing.T) {
			_, err := e.GetFloatSlice("test.key2")
			require.Error(t, err)
			require.Contains(t, err.Error(), "element 1")
		})
	})
}

func TestEnvEngine_GetDuration(t *testing.T) {
	e := NewEnvEngine()
	withEnvironment(map[string]string{
		"TEST_KEY":  "1s",
		"TEST_KEY2": "invalid_duration",
	}, func() {
		t.Run("when key exists and valid duration", func(t *testing.T) {
			value, err := e.GetDuration("test.key")
			require.NoError(t, err)
			assert.Equal(t, time.Second, value)
		})

		t.Run("when key does not exist", func(t *testing.T) {
			value, err := e.GetDuration("non_existing_key")
			require.NoError(t, err)
			assert.Zero(t, value)
		})

		t.Run("when key exists but invalid duration", func(t *testing.T) {
			_, err := e.GetDuration("test.key2")
			require.Error(t, err)
			require.Contains(t, err.Error(), "invalid_duration")
		})
	})
}
