# config

A Go library for loading typed configuration into structs from multiple sources (YAML, environment variables) with support for secrets.

## Quick Start

```go
go get github.com/jamillosantos/config
```

```go
type MyConfig struct {
    Host        string        `config:"host,required"`
    Password    string        `config:"password,secret,required"`
    ConnTimeout time.Duration `config:"conn_timeout"`
}

func main() {
    secretEngine := config.NewEnvEngine()

    manager := config.NewManager()
    manager.AddPlainEngine(config.NewYAMLEngine(config.NewFileLoader("config.yaml")), &secretEngine)
    manager.AddSecretEngine(&secretEngine) // secrets always from env vars

    var cfg MyConfig
    if err := manager.Populate(&cfg); err != nil {
        panic(err)
    }
}
```

## Struct tags

`config:"key,secret,required"` — `secret` routes the field to secret engines; `required` returns an error if no engine has the value; use `-` to skip a field.

## Engines

| Engine | Description |
|--------|-------------|
| `NewYAMLEngine(loader)` | Reads from a YAML source via a `Loader` |
| `NewEnvEngine()` | Reads from env vars; `foo.bar` → `FOO_BAR`; slices are comma-separated |
| `NewMapEngine(map)` | Reads from an in-memory map |

Engines are tried in registration order; the first to return a value wins.

## Dynamic engine selection

Pass `WithLoadOptionsEnv("CONFIG_LOAD_OPTIONS")` to override which engines are used at runtime via an environment variable containing a JSON object:

```json
{"plain": ["yamlfile:/etc/app/config.yaml", "env"], "secrets": ["env"]}
```

Supported tokens:

| Token | Description |
|-------|-------------|
| `env` | Creates a new `EnvEngine` |
| `yamlfile:<path>` | Creates a new `YAMLEngine` reading from the given file path |
| `yamlfileenv:<ENV>` | Creates a new `YAMLEngine` reading from the file path stored in the named env var |

## License

MIT
