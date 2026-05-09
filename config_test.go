//go:generate

package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	bs "github.com/inhies/go-bytesize"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestYAMLEngine() *YAMLEngine {
	return NewYAMLEngine(NewBytesLoader([]byte("dsn: from-yaml\npassword: yaml-pass")))
}

func TestWithLoadOptionsEnv(t *testing.T) {
	t.Run("when env var is not set, uses registered engines as-is", func(t *testing.T) {
		m := NewManager(WithLoadOptionsEnv("CONFIG_LOAD_OPTIONS_TEST"))
		m.AddPlainEngine(newTestYAMLEngine())
		m.AddSecretEngine(newTestYAMLEngine())
		var cfg MyTestConfig
		require.NoError(t, m.Populate(&cfg))
		assert.Equal(t, "from-yaml", cfg.DSN)
	})

	t.Run("when env var is set with invalid JSON, Populate returns error", func(t *testing.T) {
		os.Setenv("CONFIG_LOAD_OPTIONS_TEST", "not-json")
		defer os.Unsetenv("CONFIG_LOAD_OPTIONS_TEST")

		m := NewManager(WithLoadOptionsEnv("CONFIG_LOAD_OPTIONS_TEST"))
		var cfg MyTestConfig
		err := m.Populate(&cfg)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "CONFIG_LOAD_OPTIONS_TEST")
	})

	t.Run("when plain is overridden to env, reads from environment variables", func(t *testing.T) {
		withEnvironment(map[string]string{
			"DSN":      "from-env",
			"PASSWORD": "env-pass",
		}, func() {
			os.Setenv("CONFIG_LOAD_OPTIONS_TEST", `{"plain":["env"],"secrets":["env"]}`)
			defer os.Unsetenv("CONFIG_LOAD_OPTIONS_TEST")

			m := NewManager(WithLoadOptionsEnv("CONFIG_LOAD_OPTIONS_TEST"))

			var cfg MyTestConfig
			require.NoError(t, m.Populate(&cfg))
			assert.Equal(t, "from-env", cfg.DSN)
		})
	})

	t.Run("when plain is overridden to yamlfile, reads from the given file", func(t *testing.T) {
		withEnvironment(map[string]string{
			"DSN":      "from-env",
			"PASSWORD": "env-pass",
		}, func() {
			os.Setenv("CONFIG_LOAD_OPTIONS_TEST", `{"plain":["yamlfile:testdata/config_simple.yaml"],"secrets":["env"]}`)
			defer os.Unsetenv("CONFIG_LOAD_OPTIONS_TEST")

			m := NewManager(WithLoadOptionsEnv("CONFIG_LOAD_OPTIONS_TEST"))

			var cfg MyTestConfig
			require.NoError(t, m.Populate(&cfg))
			assert.Equal(t, "from-yaml", cfg.DSN)
			assert.Equal(t, "env-pass", cfg.Password)
		})
	})

	t.Run("when plain is overridden to yamlfileenv, reads file path from env var", func(t *testing.T) {
		withEnvironment(map[string]string{
			"PASSWORD":         "env-pass",
			"YAML_CONFIG_PATH": "testdata/config_simple.yaml",
		}, func() {
			os.Setenv("CONFIG_LOAD_OPTIONS_TEST", `{"plain":["yamlfileenv:YAML_CONFIG_PATH"],"secrets":["env"]}`)
			defer os.Unsetenv("CONFIG_LOAD_OPTIONS_TEST")

			m := NewManager(WithLoadOptionsEnv("CONFIG_LOAD_OPTIONS_TEST"))

			var cfg MyTestConfig
			require.NoError(t, m.Populate(&cfg))
			assert.Equal(t, "from-yaml", cfg.DSN)
			assert.Equal(t, "env-pass", cfg.Password)
		})
	})

	t.Run("when yamlfileenv references an unset env var, Populate returns error", func(t *testing.T) {
		os.Setenv("CONFIG_LOAD_OPTIONS_TEST", `{"plain":["yamlfileenv:MISSING_VAR"]}`)
		defer os.Unsetenv("CONFIG_LOAD_OPTIONS_TEST")

		m := NewManager(WithLoadOptionsEnv("CONFIG_LOAD_OPTIONS_TEST"))

		var cfg MyTestConfig
		err := m.Populate(&cfg)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "MISSING_VAR")
	})

	t.Run("when loadOptionsEnv is empty, feature is disabled", func(t *testing.T) {
		os.Setenv("CONFIG_LOAD_OPTIONS_TEST", `{"plain":["env"]}`)
		defer os.Unsetenv("CONFIG_LOAD_OPTIONS_TEST")

		m := NewManager() // no WithLoadOptionsEnv
		m.AddPlainEngine(newTestYAMLEngine())
		m.AddSecretEngine(newTestYAMLEngine())

		var cfg MyTestConfig
		require.NoError(t, m.Populate(&cfg))
		assert.Equal(t, "from-yaml", cfg.DSN)
	})
}

func TestWithKeySeparator(t *testing.T) {
	wantKeySeparator := "."

	m := &Manager{}
	WithKeySeparator(wantKeySeparator)(m)
	assert.Equal(t, wantKeySeparator, m.keySeparator)
}

type MyTestConfig struct {
	DSN      string        `config:"dsn,required"`
	Password string        `config:"password,required,secret"`
	Timeout  time.Duration `config:"timeout"`
	Size     bs.ByteSize   `config:"size"`
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

type MyTestMapItem struct {
	Name  string `config:"name"`
	Value int    `config:"value"`
}

type MyTestMapItemRequired struct {
	Name  string `config:"name,required"`
	Value int    `config:"value"`
}

type MyTestMapItemWithSecret struct {
	Name  string `config:"name"`
	Token string `config:"token,secret"`
}

type MyTestConfigWithIntMap struct {
	Items map[int]MyTestMapItem `config:"items"`
}

type MyTestConfigWithStringMap struct {
	Items map[string]MyTestMapItem `config:"items"`
}

type MyTestConfigWithIntMapRequiredSub struct {
	Items map[int]MyTestMapItemRequired `config:"items"`
}

type MyTestConfigWithIntMapSecret struct {
	Items map[int]MyTestMapItemWithSecret `config:"items"`
}

type MyTestConfigWithInt64Map struct {
	Items map[int64]MyTestMapItem `config:"items"`
}

type textKey struct {
	Region string
	ID     int
}

func (k *textKey) UnmarshalText(text []byte) error {
	parts := strings.SplitN(string(text), "-", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid textKey: %q", string(text))
	}
	id, err := strconv.Atoi(parts[1])
	if err != nil {
		return err
	}
	k.Region = parts[0]
	k.ID = id
	return nil
}

type MyTestConfigWithTextKeyMap struct {
	Items map[textKey]MyTestMapItem `config:"items"`
}

type MyTestConfigWithFloatKeyMap struct {
	Items map[float64]MyTestMapItem `config:"items"`
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
		wantSize := 10 * bs.MB

		mapEngine := NewMapEngine(map[string]interface{}{
			"dsn":      wantDSN,
			"password": wantPassword,
			"timeout":  wantTimeout,
			"size":     wantSize.String(),
		})

		manager.AddPlainEngine(mapEngine)
		manager.AddSecretEngine(mapEngine)

		var cfg MyTestConfig
		err := manager.Populate(&cfg)
		require.NoError(t, err)
		assert.Equal(t, wantDSN, cfg.DSN)
		assert.Equal(t, wantPassword, cfg.Password)
		assert.Equal(t, wantTimeout, cfg.Timeout)
		assert.Equal(t, wantSize, cfg.Size)
	})

	t.Run("WHEN config implements TextUnmarshaler but the given config value isn't string", func(t *testing.T) {
		t.Run("should parse the config as raw value", func(t *testing.T) {
			manager := NewManager()

			wantSize := 10 * bs.MB

			mapEngine := NewMapEngine(map[string]interface{}{
				"dsn":      wantDSN,
				"password": wantPassword,
				"size":     uint64(wantSize), // This will for the config to use the uint64 instead of the TextUnmarshaler
			})

			manager.AddPlainEngine(mapEngine)
			manager.AddSecretEngine(mapEngine)

			var cfg MyTestConfig
			err := manager.Populate(&cfg)
			require.NoError(t, err)
			assert.Equal(t, wantSize, cfg.Size)
		})
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

	t.Run("WHEN reading from multiple engines", func(t *testing.T) {
		t.Run("WHEN key is present on both engines", func(t *testing.T) {
			t.Run("should return the value from the second", func(t *testing.T) {
				manager := NewManager()

				manager.AddPlainEngine(
					NewMapEngine(map[string]interface{}{
						"dsn": wantDSN,
					}),
					NewMapEngine(map[string]interface{}{
						"dsn": "",
					}),
				)
				manager.AddSecretEngine(NewMapEngine(map[string]interface{}{
					"password": "",
				}))

				var cfg MyTestConfig
				err := manager.Populate(&cfg)
				require.NoError(t, err)
				assert.Equal(t, wantDSN, cfg.DSN)
				assert.Zero(t, cfg.Timeout)
			})
		})

		t.Run("WHEN key not present on the first but present on second engine", func(t *testing.T) {
			t.Run("should return the value from the second", func(t *testing.T) {
				manager := NewManager()

				manager.AddPlainEngine(
					NewMapEngine(map[string]interface{}{}),
					NewMapEngine(map[string]interface{}{
						"dsn": wantDSN,
					}),
				)
				manager.AddSecretEngine(NewMapEngine(map[string]interface{}{
					"password": wantDSN,
				}))

				var cfg MyTestConfig
				err := manager.Populate(&cfg)
				require.NoError(t, err)
				assert.Equal(t, wantDSN, cfg.DSN)
				assert.Zero(t, cfg.Timeout)
			})
		})
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

	t.Run("map of struct", func(t *testing.T) {
		t.Run("populates map[int]Struct from MapEngine", func(t *testing.T) {
			manager := NewManager()
			mapEngine := NewMapEngine(map[string]interface{}{
				"items": map[string]interface{}{
					"1": map[string]interface{}{
						"name":  "alpha",
						"value": 10,
					},
					"2": map[string]interface{}{
						"name":  "beta",
						"value": 20,
					},
				},
			})
			manager.AddPlainEngine(mapEngine)
			manager.AddSecretEngine(mapEngine)

			var cfg MyTestConfigWithIntMap
			err := manager.Populate(&cfg)
			require.NoError(t, err)
			assert.Len(t, cfg.Items, 2)
			assert.Equal(t, "alpha", cfg.Items[1].Name)
			assert.Equal(t, 10, cfg.Items[1].Value)
			assert.Equal(t, "beta", cfg.Items[2].Name)
			assert.Equal(t, 20, cfg.Items[2].Value)
		})

		t.Run("populates map[string]Struct from MapEngine", func(t *testing.T) {
			manager := NewManager()
			mapEngine := NewMapEngine(map[string]interface{}{
				"items": map[string]interface{}{
					"primary": map[string]interface{}{
						"name":  "alpha",
						"value": 10,
					},
					"secondary": map[string]interface{}{
						"name":  "beta",
						"value": 20,
					},
				},
			})
			manager.AddPlainEngine(mapEngine)
			manager.AddSecretEngine(mapEngine)

			var cfg MyTestConfigWithStringMap
			err := manager.Populate(&cfg)
			require.NoError(t, err)
			assert.Len(t, cfg.Items, 2)
			assert.Equal(t, "alpha", cfg.Items["primary"].Name)
			assert.Equal(t, 10, cfg.Items["primary"].Value)
			assert.Equal(t, "beta", cfg.Items["secondary"].Name)
			assert.Equal(t, 20, cfg.Items["secondary"].Value)
		})

		t.Run("populates map[int64]Struct from MapEngine", func(t *testing.T) {
			manager := NewManager()
			mapEngine := NewMapEngine(map[string]interface{}{
				"items": map[string]interface{}{
					"100": map[string]interface{}{
						"name":  "alpha",
						"value": 10,
					},
				},
			})
			manager.AddPlainEngine(mapEngine)
			manager.AddSecretEngine(mapEngine)

			var cfg MyTestConfigWithInt64Map
			err := manager.Populate(&cfg)
			require.NoError(t, err)
			assert.Len(t, cfg.Items, 1)
			assert.Equal(t, "alpha", cfg.Items[100].Name)
		})

		t.Run("leaves map nil when no keys present", func(t *testing.T) {
			manager := NewManager()
			mapEngine := NewMapEngine(map[string]interface{}{
				"unrelated": "value",
			})
			manager.AddPlainEngine(mapEngine)
			manager.AddSecretEngine(mapEngine)

			var cfg MyTestConfigWithIntMap
			err := manager.Populate(&cfg)
			require.NoError(t, err)
			assert.Nil(t, cfg.Items)
		})

		t.Run("returns error when required subfield missing", func(t *testing.T) {
			manager := NewManager()
			mapEngine := NewMapEngine(map[string]interface{}{
				"items": map[string]interface{}{
					"1": map[string]interface{}{
						"value": 10,
					},
				},
			})
			manager.AddPlainEngine(mapEngine)
			manager.AddSecretEngine(mapEngine)

			var cfg MyTestConfigWithIntMapRequiredSub
			err := manager.Populate(&cfg)
			require.ErrorIs(t, err, ErrKeyNotFound)
		})

		t.Run("routes secret subfield to secret engines", func(t *testing.T) {
			manager := NewManager()
			plain := NewMapEngine(map[string]interface{}{
				"items": map[string]interface{}{
					"1": map[string]interface{}{
						"name": "alpha",
					},
				},
			})
			secret := NewMapEngine(map[string]interface{}{
				"items": map[string]interface{}{
					"1": map[string]interface{}{
						"token": "s3cr3t",
					},
				},
			})
			manager.AddPlainEngine(plain)
			manager.AddSecretEngine(secret)

			var cfg MyTestConfigWithIntMapSecret
			err := manager.Populate(&cfg)
			require.NoError(t, err)
			assert.Equal(t, "alpha", cfg.Items[1].Name)
			assert.Equal(t, "s3cr3t", cfg.Items[1].Token)
		})

		t.Run("populates map[int]Struct from YAMLEngine", func(t *testing.T) {
			manager := NewManager()
			yEngine := NewYAMLEngine(NewFileLoader("testdata/config_intmap.yaml"))
			manager.AddPlainEngine(yEngine)
			manager.AddSecretEngine(yEngine)

			var cfg MyTestConfigWithIntMap
			err := manager.Populate(&cfg)
			require.NoError(t, err)
			assert.Len(t, cfg.Items, 2)
			assert.Equal(t, "alpha", cfg.Items[1].Name)
			assert.Equal(t, 10, cfg.Items[1].Value)
			assert.Equal(t, "beta", cfg.Items[2].Name)
			assert.Equal(t, 20, cfg.Items[2].Value)
		})

		t.Run("populates map[string]Struct from YAMLEngine", func(t *testing.T) {
			manager := NewManager()
			yEngine := NewYAMLEngine(NewFileLoader("testdata/config_stringmap.yaml"))
			manager.AddPlainEngine(yEngine)
			manager.AddSecretEngine(yEngine)

			var cfg MyTestConfigWithStringMap
			err := manager.Populate(&cfg)
			require.NoError(t, err)
			assert.Len(t, cfg.Items, 2)
			assert.Equal(t, "alpha", cfg.Items["primary"].Name)
			assert.Equal(t, 10, cfg.Items["primary"].Value)
			assert.Equal(t, "beta", cfg.Items["secondary"].Name)
			assert.Equal(t, 20, cfg.Items["secondary"].Value)
		})

		t.Run("populates map with TextUnmarshaler key from YAMLEngine", func(t *testing.T) {
			manager := NewManager()
			yEngine := NewYAMLEngine(NewFileLoader("testdata/config_textkeymap.yaml"))
			manager.AddPlainEngine(yEngine)
			manager.AddSecretEngine(yEngine)

			var cfg MyTestConfigWithTextKeyMap
			err := manager.Populate(&cfg)
			require.NoError(t, err)
			assert.Len(t, cfg.Items, 2)
			assert.Equal(t, "alpha", cfg.Items[textKey{Region: "us", ID: 1}].Name)
			assert.Equal(t, 10, cfg.Items[textKey{Region: "us", ID: 1}].Value)
			assert.Equal(t, "beta", cfg.Items[textKey{Region: "eu", ID: 2}].Name)
			assert.Equal(t, 20, cfg.Items[textKey{Region: "eu", ID: 2}].Value)
		})

		t.Run("returns error when map key type unsupported and not TextUnmarshaler", func(t *testing.T) {
			manager := NewManager()
			mapEngine := NewMapEngine(map[string]interface{}{
				"items": map[string]interface{}{
					"1.5": map[string]interface{}{
						"name": "alpha",
					},
				},
			})
			manager.AddPlainEngine(mapEngine)
			manager.AddSecretEngine(mapEngine)

			var cfg MyTestConfigWithFloatKeyMap
			err := manager.Populate(&cfg)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "unsupported map key type")
		})
	})
}
