---
title: Loading .env files in Go with Viper
slug: working-with-dot-env-in-golang
date: 2024-03-31T00:00:00Z
tags: Go, Viper, Config
status: published
summary: How I load .env variables into a typed struct in Go using Viper.
---
I use [Viper](https://github.com/spf13/viper) to read a `.env` file into a typed struct so the rest of the app never touches `os.Getenv` directly. Here's the setup I copy into most projects.

Install it:

```bash
go get github.com/spf13/viper
```

Define a struct for the variables you expect and load them in one place:

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

Then load it once at startup and pass the struct around:

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

That's the whole thing. The nice part is the `mapstructure` tags — the config is typed, so a missing or misspelled key shows up at load time instead of somewhere deep in a request.
