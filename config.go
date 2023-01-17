package config

import (
	"errors"
	"reflect"
	"strings"
	"sync"
)

const (
	defaultKeySeparator = "."
)

type Validator interface {
	Validate() error
}

type Manager struct {
	init         sync.Once
	keySeparator *string
	secrets      []Engine
	plains       []Engine
}

type Option func(*Manager)

func NewManager(opts ...Option) *Manager {
	r := &Manager{}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

// WithKeySeparator sets the key separator for the manager.
func WithKeySeparator(separator string) Option {
	return func(m *Manager) {
		m.keySeparator = &separator
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

func (m *Manager) AddSecretEngine(engine Engine) {
	m.init.Do(m.initialize)
	m.secrets = append(m.secrets, engine)
}

func (m *Manager) AddPlainEngine(engine Engine) {
	m.init.Do(m.initialize)
	m.plains = append(m.plains, engine)
}

func (m *Manager) Populate(cfg interface{}) error {
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

		switch {
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
				err := readFromEnginesInSequence(engines, func(engine Engine) error {
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
				err := readFromEnginesInSequence(engines, func(engine Engine) error {
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
				err := readFromEnginesInSequence(engines, func(engine Engine) error {
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
				err := readFromEnginesInSequence(engines, func(engine Engine) error {
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
				err := readFromEnginesInSequence(engines, func(engine Engine) error {
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
			err := readFromEnginesInSequence(engines, func(engine Engine) error {
				value, err := engine.GetString(key)
				if !isRequired && errors.Is(err, ErrKeyNotFound) {
					return nil
				} else if err != nil {
					return err
				}
				fieldValue.SetString(value)
				return nil
			})
			if err != nil {
				return err
			}
		case fieldValue.Kind() == reflect.Int || fieldValue.Kind() == reflect.Int8 || fieldValue.Kind() == reflect.Int16 || fieldValue.Kind() == reflect.Int32:
			err := readFromEnginesInSequence(engines, func(engine Engine) error {
				value, err := engine.GetInt(key)
				if !isRequired && errors.Is(err, ErrKeyNotFound) {
					return nil
				} else if err != nil {
					return err
				}
				fieldValue.SetInt(int64(value))
				return nil
			})
			if err != nil {
				return err
			}
		case fieldValue.Kind() == reflect.Int64:
			err := readFromEnginesInSequence(engines, func(engine Engine) error {
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
			err := readFromEnginesInSequence(engines, func(engine Engine) error {
				value, err := engine.GetUint(key)
				if !isRequired && errors.Is(err, ErrKeyNotFound) {
					return nil
				} else if err != nil {
					return err
				}
				fieldValue.SetUint(uint64(value))
				return nil
			})
			if err != nil {
				return err
			}
		case fieldValue.Kind() == reflect.Int64:
			err := readFromEnginesInSequence(engines, func(engine Engine) error {
				value, err := engine.GetInt64(key)
				if !isRequired && errors.Is(err, ErrKeyNotFound) {
					return nil
				} else if err != nil {
					return err
				}
				fieldValue.SetInt(value)
				return nil
			})
			if err != nil {
				return err
			}
		case fieldValue.Kind() == reflect.Uint64:
			err := readFromEnginesInSequence(engines, func(engine Engine) error {
				value, err := engine.GetUint64(key)
				if !isRequired && errors.Is(err, ErrKeyNotFound) {
					return nil
				} else if err != nil {
					return err
				}
				fieldValue.SetUint(value)
				return nil
			})
			if err != nil {
				return err
			}
		case fieldValue.Kind() == reflect.Float64 || fieldValue.Kind() == reflect.Float32:
			err := readFromEnginesInSequence(engines, func(engine Engine) error {
				value, err := engine.GetFloat(key)
				if !isRequired && errors.Is(err, ErrKeyNotFound) {
					return nil
				} else if err != nil {
					return err
				}
				fieldValue.SetFloat(value)
				return nil
			})
			if err != nil {
				return err
			}
		case fieldValue.Kind() == reflect.Bool:
			err := readFromEnginesInSequence(engines, func(engine Engine) error {
				value, err := engine.GetBool(key)
				if !isRequired && errors.Is(err, ErrKeyNotFound) {
					return nil
				} else if err != nil {
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

func readFromEnginesInSequence(engines []Engine, f func(engine Engine) error) error {
	var lastErr error
	for _, engine := range engines {
		err := f(engine)
		if errors.Is(err, ErrKeyNotFound) {
			lastErr = err
			continue
		} else if err != nil {
			return err
		}
		lastErr = nil // nolint
		return nil
	}

	if lastErr != nil {
		return lastErr
	}

	return ErrKeyNotFound
}
