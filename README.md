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
import (
    "github.com/jamillosantos/config"
)

func main() {
    manager := config.NewManager()
    
    plainEngine := config.NewYAMLEngine(config.NewFileLoader("config.yaml"))
    secretEngine := config.NewEnvEngine()
    
    manager.AddPlainEngine(plainEngine, secretEngine) // Try to read from config.yaml, if not found, read from environment variables
    manager.AddSecretEngine(secretEngine) // Secrets will always come from environment variables
    
    var cfg MyConfig
    err := manager.Populate(&cfg)
    if err != nil {
        panic(err)
    }
}
```