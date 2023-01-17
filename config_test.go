//go:generate

package config

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithKeySeparator(t *testing.T) {
	wantKeySeparator := "."

	m := &Manager{}
	WithKeySeparator(wantKeySeparator)(m)
	assert.Equal(t, wantKeySeparator, *m.keySeparator)
}

type MyTestConfig struct {
	DSN      string        `config:"dsn,required"`
	Password string        `config:"password,required,secret"`
	Timeout  time.Duration `config:"timeout"`
}

type MyTestWithNestedConfigDatabase struct {
	DSN     string        `config:"dsn,required"`
	Timeout time.Duration `config:"timeout"`
}

type MyTestWithNestedConfigTokens struct {
	AccessToken string `config:"access_token,required,secret"`
}

type MyTestWithNestedConfig struct {
	Database MyTestWithNestedConfigDatabase `config:"database"`
	Tokens   MyTestWithNestedConfigTokens   `config:"tokens"`
}

type MyTestConfigAllTypes struct {
	KeyString      string    `config:"key_string"`
	KeyStringSlice []string  `config:"key_string_slice"`
	KeyInt         int       `config:"key_int"`
	KeyIntSlice    []int     `config:"key_int_slice"`
	KeyBool        bool      `config:"key_bool"`
	KeyBoolSlice   []bool    `config:"key_bool_slice"`
	KeyFloat       float64   `config:"key_float"`
	KeyFloatSlice  []float64 `config:"key_float_slice"`
}

type MyTestConfigWithValidation struct {
	N int `config:"n"`
}

var (
	errMustBePositive = fmt.Errorf("n must be positive")
)

func (m MyTestConfigWithValidation) Validate() error {
	if m.N < 0 {
		return errMustBePositive
	}
	return nil
}

func TestManager_Populate(t *testing.T) {
	wantDSN := "postgres://user@host:port/database"
	wantPassword := "12345"

	t.Run("fail when the given config is not a pointer", func(t *testing.T) {
		m := NewManager()
		var cfg MyTestConfig
		err := m.Populate(cfg)
		require.ErrorIs(t, err, ErrConfigNotPointer)
	})

	t.Run("no secret engine defined", func(t *testing.T) {
		m := NewManager()
		var cfg MyTestConfig
		m.AddPlainEngine(NewMapEngine(map[string]interface{}{
			"dsn": wantDSN,
		}))
		err := m.Populate(&cfg)
		require.ErrorIs(t, err, ErrNoSecretEngineDefined)
	})

	t.Run("no plain engine defined", func(t *testing.T) {
		m := NewManager()
		var cfg MyTestConfig
		m.AddSecretEngine(NewMapEngine(map[string]interface{}{
			"dsn": wantDSN,
		}))
		err := m.Populate(&cfg)
		require.ErrorIs(t, err, ErrNoPlainEngineDefined)
	})

	t.Run("success", func(t *testing.T) {
		manager := NewManager()

		wantTimeout := time.Second * 10

		mapEngine := NewMapEngine(map[string]interface{}{
			"dsn":      wantDSN,
			"password": wantPassword,
			"timeout":  wantTimeout,
		})

		manager.AddPlainEngine(mapEngine)
		manager.AddSecretEngine(mapEngine)

		var cfg MyTestConfig
		err := manager.Populate(&cfg)
		require.NoError(t, err)
		assert.Equal(t, wantDSN, cfg.DSN)
		assert.Equal(t, wantPassword, cfg.Password)
		assert.Equal(t, wantTimeout, cfg.Timeout)
	})

	t.Run("success with all supported data types", func(t *testing.T) {
		manager := NewManager()

		l := NewFileLoader("testdata/config1.yaml")
		yEngine := NewYAMLEngine(l)

		require.NoError(t, yEngine.Load())

		manager.AddPlainEngine(yEngine)
		manager.AddSecretEngine(yEngine)

		var cfg MyTestConfigAllTypes
		err := manager.Populate(&cfg)
		require.NoError(t, err)
	})

	t.Run("success with nested", func(t *testing.T) {
		manager := NewManager()

		wantTimeout := time.Second * 10
		wantToken := "token value"

		mapEngine := NewMapEngine(map[string]interface{}{
			"database": map[string]interface{}{
				"dsn":     wantDSN,
				"timeout": wantTimeout,
			},
			"tokens": map[string]interface{}{
				"access_token": wantToken,
			},
		})

		manager.AddPlainEngine(mapEngine)
		manager.AddSecretEngine(mapEngine)

		var cfg MyTestWithNestedConfig
		err := manager.Populate(&cfg)
		require.NoError(t, err)
		assert.Equal(t, wantDSN, cfg.Database.DSN)
		assert.Equal(t, wantTimeout, cfg.Database.Timeout)
		assert.Equal(t, wantToken, cfg.Tokens.AccessToken)
	})

	t.Run("success with non required fields", func(t *testing.T) {
		manager := NewManager()

		manager.AddPlainEngine(NewMapEngine(map[string]interface{}{
			"dsn": wantDSN,
		}))
		manager.AddSecretEngine(NewMapEngine(map[string]interface{}{
			"password": wantDSN,
		}))

		var cfg MyTestConfig
		err := manager.Populate(&cfg)
		require.NoError(t, err)
		assert.Equal(t, wantDSN, cfg.DSN)
		assert.Zero(t, cfg.Timeout)
	})

	t.Run("required fields not provided", func(t *testing.T) {
		manager := NewManager()

		manager.AddPlainEngine(NewMapEngine(map[string]interface{}{
			"timeout": time.Second,
		}))

		manager.AddSecretEngine(NewMapEngine(map[string]interface{}{
			"timeout": time.Second,
		}))

		var cfg MyTestConfig
		err := manager.Populate(&cfg)
		require.ErrorIs(t, err, ErrKeyNotFound)
	})

	t.Run("config with validation", func(t *testing.T) {
		t.Run("should validate the given config", func(t *testing.T) {
			manager := NewManager()

			manager.AddPlainEngine(NewMapEngine(map[string]interface{}{
				"n": 1,
			}))

			var cfg MyTestConfigWithValidation
			err := manager.Populate(&cfg)
			require.NoError(t, err)
			assert.Equal(t, 1, cfg.N)
		})

		t.Run("should fail due to invalid config", func(t *testing.T) {
			manager := NewManager()

			manager.AddPlainEngine(NewMapEngine(map[string]interface{}{
				"n": -1,
			}))

			var cfg MyTestConfigWithValidation
			err := manager.Populate(&cfg)
			require.ErrorIs(t, err, errMustBePositive)
		})
	})
}
