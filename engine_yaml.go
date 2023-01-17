package config

import (
	yamlv3 "gopkg.in/yaml.v3"
)

type YAMLEngine struct {
	*MapEngine
	loader Loader
}

func NewYAMLEngine(loader Loader) *YAMLEngine {
	return &YAMLEngine{nil, loader}
}

// Load loads the YAML file defined by the filePath set on the NewYAMLEngine saving the data into a internal map.
func (engine *YAMLEngine) Load() error {
	reader, err := engine.loader.Load()
	if err != nil {
		return err
	}
	defer func() {
		_ = engine.loader.Unload()
	}()

	decoder := yamlv3.NewDecoder(reader)
	data := make(map[string]interface{})
	if err = decoder.Decode(&data); err != nil {
		return err
	}

	engine.MapEngine = NewMapEngine(data)
	return engine.MapEngine.Load()
}
