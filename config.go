package config

import (
	"encoding" // nolint
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
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
	keySeparator   string
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
		keySeparator:   defaultKeySeparator,
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
func WithKeySeparator(value string) Option {
	return func(m *Manager) {
		m.keySeparator = value
	}
}

// WithLoadOptionsEnv sets the name of the environment variable from which the manager will read
// load options (JSON) at creation time. If envName is empty, this feature is disabled.
func WithLoadOptionsEnv(envName string) Option {
	return func(m *Manager) {
		m.loadOptionsEnv = envName
	}
}

func (m *Manager) initializeEngines() {
	if m.secrets != nil || m.plains != nil {
		m.secrets = make([]Engine, 0)
		m.plains = make([]Engine, 0)
	}
}

func (m *Manager) AddSecretEngine(engines ...Engine) {
	m.init.Do(m.initializeEngines)
	m.secrets = append(m.secrets, engines...)
}

func (m *Manager) AddPlainEngine(engines ...Engine) {
	m.init.Do(m.initializeEngines)
	m.plains = append(m.plains, engines...)
}

func (m *Manager) Populate(cfg interface{}) error {
	if m.loadOptionsErr != nil {
		return m.loadOptionsErr
	}

	if m.loadOptions != nil {
		m.initializeEngines()

		if m.loadOptions.Plain != nil {
			plains, err := buildEnginesFromOptions(m.loadOptions.Plain)
			if err != nil {
				return err
			}
			m.plains = append(m.plains, plains...)
		}
		if m.loadOptions.Secrets != nil {
			secrets, err := buildEnginesFromOptions(m.loadOptions.Secrets)
			if err != nil {
				return err
			}
			m.secrets = append(m.secrets, secrets...)
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
			key += m.keySeparator
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
		case fieldValue.Kind() == reflect.Map:
			if err := m.unmarshalMap(key, fieldValue, isRequired); err != nil {
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

func (m *Manager) unmarshalMap(prefix string, fieldValue reflect.Value, isRequired bool) error {
	mapType := fieldValue.Type()
	if mapType.Elem().Kind() != reflect.Struct {
		return nil
	}
	indexes := m.collectMapIndexes(prefix)
	if len(indexes) == 0 {
		if isRequired {
			return fmt.Errorf("%w: %s", ErrKeyNotFound, prefix)
		}
		return nil
	}
	keyType := mapType.Key()
	elemType := mapType.Elem()
	out := reflect.MakeMapWithSize(mapType, len(indexes))
	for _, idx := range indexes {
		mapKey, err := convertMapKey(keyType, idx)
		if err != nil {
			return fmt.Errorf("map %s: invalid key %q: %w", prefix, idx, err)
		}
		elem := reflect.New(elemType)
		subKey := prefix + m.keySeparator + idx
		if err := m.unmarshalObj(subKey, elem.Interface()); err != nil {
			return err
		}
		out.SetMapIndex(mapKey, elem.Elem())
	}
	fieldValue.Set(out)
	return nil
}

func (m *Manager) collectMapIndexes(prefix string) []string {
	sep := m.keySeparator
	full := prefix + sep
	seen := make(map[string]struct{})
	out := make([]string, 0)
	collect := func(engines []Engine) {
		for _, eng := range engines {
			lister, ok := eng.(EngineKeyLister)
			if !ok {
				continue
			}
			for _, k := range lister.Keys() {
				if !strings.HasPrefix(k, full) {
					continue
				}
				rest := k[len(full):]
				seg := rest
				if i := strings.Index(rest, sep); i >= 0 {
					seg = rest[:i]
				}
				if seg == "" {
					continue
				}
				if _, ok := seen[seg]; ok {
					continue
				}
				seen[seg] = struct{}{}
				out = append(out, seg)
			}
		}
	}
	collect(m.plains)
	collect(m.secrets)
	return out
}

func convertMapKey(t reflect.Type, s string) (reflect.Value, error) {
	ptr := reflect.New(t)
	if u, ok := ptr.Interface().(encoding.TextUnmarshaler); ok {
		if err := u.UnmarshalText([]byte(s)); err != nil {
			return reflect.Value{}, err
		}
		return ptr.Elem(), nil
	}
	v := ptr.Elem()
	switch t.Kind() {
	case reflect.String:
		v.SetString(s)
		return v, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := strconv.ParseInt(s, 10, t.Bits())
		if err != nil {
			return reflect.Value{}, err
		}
		v.SetInt(n)
		return v, nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		n, err := strconv.ParseUint(s, 10, t.Bits())
		if err != nil {
			return reflect.Value{}, err
		}
		v.SetUint(n)
		return v, nil
	}
	return reflect.Value{}, fmt.Errorf("unsupported map key type %s", t.String())
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
