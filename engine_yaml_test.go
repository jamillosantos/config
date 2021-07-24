package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestYAMLEngine_Load(t *testing.T) {
	yamlEngine := NewYAMLEngine(NewBytesLoader([]byte(`string1: "string1"
stringslice:
 - "string1"
 - "string2"
 - "string3"
nested:
  string2: "string2"
  nested2:
    string3: "string3"
`)))
	err := yamlEngine.Load()
	require.NoError(t, err)

	value, err := yamlEngine.GetString("string1")
	assert.NoError(t, err)
	assert.Equal(t, "string1", value)

	valueSlice, err := yamlEngine.GetStringSlice("stringslice")
	assert.NoError(t, err)
	assert.Equal(t, []string{"string1", "string2", "string3"}, valueSlice)

	valueNested1, err := yamlEngine.GetString("nested.string2")
	assert.NoError(t, err)
	assert.Equal(t, "string2", valueNested1)

	valueNested2, err := yamlEngine.GetString("nested.nested2.string3")
	assert.NoError(t, err)
	assert.Equal(t, "string3", valueNested2)
}
