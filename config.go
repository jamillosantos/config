package config

import (
	"encoding" // nolint
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
	"sync"
)

type configLoadOptions struct {
	Plain   []string `json:"plain"`
	Secrets []string `json:"secrets"`
}

// buildEnginesFromOptions builds a list of engines from a list of token strings.
//
// Supported tokens:
//   - "env": creates a new EnvEngine.
//   - "yamlfile:<filepath>": creates a new YAMLEngine backed by the given file path.
//   - "yamlfileenv:<ENV>": creates a new YAMLEngine backed by the file path read from the named env var.
func buildEnginesFromOptions(options []string) ([]Engine, error) {
	result := make([]Engine, 0, len(options))
	for _, opt := range options {
		switch {
		case opt == "env":
			eng := NewEnvEngine()
			result = append(result, &eng)
		case strings.HasPrefix(opt, "yamlfile:"):
			filePath := strings.TrimPrefix(opt, "yamlfile:")
			result = append(result, NewYAMLEngine(NewFileLoader(filePath)))
		case strings.HasPrefix(opt, "yamlfileenv:"):
			envName := strings.TrimPrefix(opt, "yamlfileenv:")
			filePath, ok := os.LookupEnv(envName)
			if !ok || filePath == "" {
				return nil, fmt.Errorf("yamlfileenv: environment variable %q is not set", envName)
			}
			result = append(result, NewYAMLEngine(NewFileLoader(filePath)))
		}
	}
	return result, nil
}

const (
	defaultKeySeparator = "."
)

type Validator interface {
	Validate() error
}

type Manager struct {
	init           sync.Once
	keySeparator   *string
	secrets        []Engine
	plains         []Engine
	loadOptionsEnv string
	loadOptions    *configLoadOptions
	loadOptionsErr error
}

type Option func(*Manager)

func NewManager(opts ...Option) *Manager {
	r := &Manager{
		loadOptionsEnv: "CONFIG_LOAD_OPTIONS",
	}
	for _, opt := range opts {
		opt(r)
	}
	if r.loadOptionsEnv != "" {
		if raw, ok := os.LookupEnv(r.loadOptionsEnv); ok && raw != "" {
			var parsed configLoadOptions
			if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
				r.loadOptionsErr = fmt.Errorf("parsing %s: %w", r.loadOptionsEnv, err)
			} else {
				r.loadOptions = &parsed
			}
		}
	}
	return r
}

// WithKeySeparator sets the key separator for the manager.
func WithKeySeparator(separator string) Option {
	return func(m *Manager) {
		m.keySeparator = &separator
	}
}

// WithLoadOptionsEnv sets the name of the environment variable from which the manager will read
// load options (JSON) at creation time. If envName is empty, this feature is disabled.
func WithLoadOptionsEnv(envName string) Option {
	return func(m *Manager) {
		m.loadOptionsEnv = envName
	}
}

func (m *Manager) initialize() {
	if m.keySeparator == nil {
		keySeparator := defaultKeySeparator
		m.keySeparator = &keySeparator
	}
	m.secrets = make([]Engine, 0)
	m.plains = make([]Engine, 0)
}

func (m *Manager) AddSecretEngine(engines ...Engine) {
	m.init.Do(m.initialize)
	m.secrets = append(m.secrets, engines...)
}

func (m *Manager) AddPlainEngine(engines ...Engine) {
	m.init.Do(m.initialize)
	m.plains = append(m.plains, engines...)
}

func (m *Manager) Populate(cfg interface{}) error {
	if m.loadOptionsErr != nil {
		return m.loadOptionsErr
	}

	if m.loadOptions != nil {
		if m.loadOptions.Plain != nil {
			plains, err := buildEnginesFromOptions(m.loadOptions.Plain)
			if err != nil {
				return err
			}
			m.plains = plains
		}
		if m.loadOptions.Secrets != nil {
			secrets, err := buildEnginesFromOptions(m.loadOptions.Secrets)
			if err != nil {
				return err
			}
			m.secrets = secrets
		}
	}

	for _, eng := range m.plains {
		if err := eng.Load(); err != nil {
			return err
		}
	}
	for _, eng := range m.secrets {
		if err := eng.Load(); err != nil {
			return err
		}
	}
	if reflect.ValueOf(cfg).Kind() != reflect.Ptr {
		return ErrConfigNotPointer
	}
	return m.unmarshalObj("", cfg)
}

func (m *Manager) unmarshalObj(keyPrefix string, obj interface{}) error {
	v := reflect.ValueOf(obj)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	t := v.Type()
	for f := 0; f < v.NumField(); f++ {
		fieldValue, fieldType := v.Field(f), t.Field(f)
		configTag := fieldType.Tag.Get("config")
		configTagTokens := strings.Split(configTag, ",")
		propName := configTagTokens[0]
		if propName == "-" || propName == "" {
			continue
		}
		isRequired := false
		isSecret := false
		for _, tok := range configTagTokens[1:] {
			switch tok {
			case "required":
				isRequired = true
			case "secret":
				isSecret = true
			}
		}

		if isSecret && len(m.secrets) == 0 {
			return ErrNoSecretEngineDefined
		}

		if !isSecret && len(m.plains) == 0 {
			return ErrNoPlainEngineDefined
		}

		engines := m.secrets // Default to secrets
		if !isSecret {
			engines = m.plains
		}

		if configTag != "" && len(m.plains) == 0 {
			return ErrNoPlainEngineDefined
		}

		key := keyPrefix
		if key != "" {
			key += *m.keySeparator
		}
		key += propName

		fieldTextUnmarshalerValue, okTextUnmarshalerValue := fieldValue.Addr().Interface().(encoding.TextUnmarshaler)
		if okTextUnmarshalerValue {
			err := readFromEnginesInSequence(engines, key, isRequired, func(engine Engine) error {
				value, err := engine.GetString(key)
				if err != nil {
					return fmt.Errorf("%w: %s", err, key)
				}
				return fieldTextUnmarshalerValue.UnmarshalText([]byte(value))
			})
			switch {
			case errors.Is(err, ErrTypeMismatch):
				okTextUnmarshalerValue = false
			case err != nil:
				return fmt.Errorf("%w: %s", err, key)
			}
		}
		switch {
		case okTextUnmarshalerValue:
			// Do nothing
		case fieldValue.Kind() == reflect.Struct:
			if err := m.unmarshalObj(key, fieldValue.Addr().Interface()); err != nil {
				return err
			}
		case fieldValue.Kind() == reflect.Slice:
			switch fieldValue.Type().Elem().Kind() {
			case reflect.Struct:
				if err := m.unmarshalObj(key, fieldValue.Interface()); err != nil {
					return err
				}
			case reflect.Int:
				err := readFromEnginesInSequence(engines, key, isRequired, func(engine Engine) error {
					value, err := engine.GetIntSlice(key)
					if err != nil {
						return err
					}
					fieldValue.Set(reflect.ValueOf(value))
					return nil
				})
				if err != nil {
					return err
				}
			case reflect.Int64:
				err := readFromEnginesInSequence(engines, key, isRequired, func(engine Engine) error {
					value, err := engine.GetInt64Slice(key)
					if err != nil {
						return err
					}
					fieldValue.Set(reflect.ValueOf(value))
					return nil
				})
				if err != nil {
					return err
				}
			case reflect.String:
				err := readFromEnginesInSequence(engines, key, isRequired, func(engine Engine) error {
					value, err := engine.GetStringSlice(key)
					if err != nil {
						return err
					}
					fieldValue.Set(reflect.ValueOf(value))
					return nil
				})
				if err != nil {
					return err
				}
			case reflect.Bool:
				err := readFromEnginesInSequence(engines, key, isRequired, func(engine Engine) error {
					value, err := engine.GetBoolSlice(key)
					if err != nil {
						return err
					}
					fieldValue.Set(reflect.ValueOf(value))
					return nil
				})
				if err != nil {
					return err
				}
			case reflect.Float64:
				err := readFromEnginesInSequence(engines, key, isRequired, func(engine Engine) error {
					value, err := engine.GetFloatSlice(key)
					if err != nil {
						return err
					}
					fieldValue.Set(reflect.ValueOf(value))
					return nil
				})
				if err != nil {
					return err
				}
			}
		case fieldValue.Kind() == reflect.String:
			err := readFromEnginesInSequence(engines, key, isRequired, func(engine Engine) error {
				value, err := engine.GetString(key)
				if err != nil {
					return err
				}
				fieldValue.SetString(value)
				return nil
			})
			if err != nil {
				return err
			}
		case fieldValue.Kind() == reflect.Int || fieldValue.Kind() == reflect.Int8 || fieldValue.Kind() == reflect.Int16 || fieldValue.Kind() == reflect.Int32:
			err := readFromEnginesInSequence(engines, key, isRequired, func(engine Engine) error {
				value, err := engine.GetInt(key)
				if err != nil {
					return err
				}
				fieldValue.SetInt(int64(value))
				return nil
			})
			if err != nil {
				return err
			}
		case fieldValue.Kind() == reflect.Int64:
			err := readFromEnginesInSequence(engines, key, isRequired, func(engine Engine) error {
				switch fieldValue.Type().String() {
				case "time.Duration":
					value, err := engine.GetDuration(key)
					if !isRequired && errors.Is(err, ErrKeyNotFound) {
						return nil
					} else if err != nil {
						return err
					}
					fieldValue.Set(reflect.ValueOf(value))
				default:
					value, err := engine.GetInt64(key)
					if !isRequired && errors.Is(err, ErrKeyNotFound) {
						return nil
					} else if err != nil {
						return err
					}
					fieldValue.SetInt(value)
				}
				return nil
			})
			if err != nil {
				return err
			}
		case fieldValue.Kind() == reflect.Uint || fieldValue.Kind() == reflect.Uint8 || fieldValue.Kind() == reflect.Uint16 || fieldValue.Kind() == reflect.Uint32:
			err := readFromEnginesInSequence(engines, key, isRequired, func(engine Engine) error {
				value, err := engine.GetUint(key)
				if err != nil {
					return err
				}
				fieldValue.SetUint(uint64(value))
				return nil
			})
			if err != nil {
				return err
			}
		case fieldValue.Kind() == reflect.Int64:
			err := readFromEnginesInSequence(engines, key, isRequired, func(engine Engine) error {
				value, err := engine.GetInt64(key)
				if err != nil {
					return err
				}
				fieldValue.SetInt(value)
				return nil
			})
			if err != nil {
				return err
			}
		case fieldValue.Kind() == reflect.Uint64:
			err := readFromEnginesInSequence(engines, key, isRequired, func(engine Engine) error {
				value, err := engine.GetUint64(key)
				if err != nil {
					return err
				}
				fieldValue.SetUint(value)
				return nil
			})
			if err != nil {
				return err
			}
		case fieldValue.Kind() == reflect.Float64 || fieldValue.Kind() == reflect.Float32:
			err := readFromEnginesInSequence(engines, key, isRequired, func(engine Engine) error {
				value, err := engine.GetFloat(key)
				if err != nil {
					return err
				}
				fieldValue.SetFloat(value)
				return nil
			})
			if err != nil {
				return err
			}
		case fieldValue.Kind() == reflect.Bool:
			err := readFromEnginesInSequence(engines, key, isRequired, func(engine Engine) error {
				value, err := engine.GetBool(key)
				if err != nil {
					return err
				}
				fieldValue.SetBool(value)
				return nil
			})
			if err != nil {
				return err
			}
		}
	}

	if validator, ok := obj.(Validator); ok {
		if err := validator.Validate(); err != nil {
			return err
		}
	}

	return nil
}

func readFromEnginesInSequence(engines []Engine, key string, isRequired bool, f func(engine Engine) error) error {
	var err error
	for _, engine := range engines {
		err = f(engine)
		if errors.Is(err, ErrKeyNotFound) {
			continue
		} else if err != nil {
			return err
		}
		return nil
	}

	if isRequired && errors.Is(err, ErrKeyNotFound) {
		return fmt.Errorf("%w: %s", ErrKeyNotFound, key)
	}

	return nil
}
