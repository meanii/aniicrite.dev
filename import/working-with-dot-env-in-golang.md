---
title: Loading .env files in Go with Viper
slug: working-with-dot-env-in-golang
date: 2024-03-31T00:00:00Z
tags: Go, Viper, Config
status: published
summary: The small Viper setup I copy into Go projects to read a .env file into a typed struct.
---
I do not like `os.Getenv` scattered through a codebase. The variable names end up duplicated in a dozen places and a typo in one of them is a runtime surprise. In most Go projects I read the `.env` file once at startup into a struct using [Viper](https://github.com/spf13/viper), and the rest of the code takes the struct. This is the setup I copy from project to project.

```bash
go get github.com/spf13/viper
```

One package holds the struct and the loader:

```go
package configs

import (
    "log"
    "github.com/spf13/viper"
)

// Env holds the environment variables
type Env struct {
    Port          string `mapstructure:"PORT"`
    MongoURL      string `mapstructure:"MONGO_URI"`
    RedisURL      string `mapstructure:"REDIS_URI"`
    SecretToken   string `mapstructure:"SECRET_TOKEN"`
    RefreshToken  string `mapstructure:"REFRESH_TOKEN"`
}

// LoadConfig reads the .env file and unmarshals it into Env
func LoadConfig() *Env {
    var envs Env

    viper.AddConfigPath(".")
    viper.SetConfigName(".env")
    viper.SetConfigType("env")

    if err := viper.ReadInConfig(); err != nil {
        log.Fatalf("Error reading config file, %s", err)
    }

    if err := viper.Unmarshal(&envs); err != nil {
        log.Fatalf("Unable to decode into struct, %v", err)
    }

    return &envs
}
```

Then main loads it once and passes it along:

```go
package main

import (
    "fmt"
    "myapp/configs"
)

func main() {
    envs := configs.LoadConfig()
    fmt.Printf("Port: %s\n", envs.Port)
    fmt.Printf("MongoURL: %s\n", envs.MongoURL)
    fmt.Printf("RedisURL: %s\n", envs.RedisURL)
    fmt.Printf("SecretToken: %s\n", envs.SecretToken)
    fmt.Printf("RefreshToken: %s\n", envs.RefreshToken)
}
```

The `mapstructure` tags map each field to a variable name, so there is one place where the names are spelled out. If a key is misspelled, the field is empty when the config is printed at startup, not somewhere deep inside a request handler.
