# config


# Usage

## Configuration

```go
type MyConfig struct {
    Host string               `config:"hostname,secret,required"`
    User string               `config:"user,secret,required"`
    Password string           `config:"user,secret,required"`
    ConnTimeout time.Duration `config:"conn_timeout"`
}
```

## Loading

```go
manager := NewManager()

plainEngine := NewYAMLEngine(NewConfigLoader(".config.yaml"))
secretEngine := NewYAMLEngine(NewConfigLoader(".config-secrets.yaml"))

manager.AddPlainEngine(mapEngine)
manager.AddSecretEngine(mapEngine)

var cfg MyConfig
err := manager.Populate(&cfg)
if err != nil {
    panic(err)
}
```