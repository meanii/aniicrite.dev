---
title: Simplify Configuration Management in Golang with Viper
slug: working-with-dot-env-in-golang
date: 2024-03-31T00:00:00Z
tags: Go, Viper, Config
status: published
summary: Learn how to efficiently load .env variables in Golang using the Viper package.
---
Managing configuration in a Golang project can often be a challenging task, especially when dealing with multiple environment variables. The `viper` package offers a simple and effective solution for loading and managing these variables. In this post, we'll walk through how to use `viper` to load `.env` variables in your Golang application.

### Setting Up Viper

First, you need to install the `viper` package. You can do this using `go get`:

```bash
go get github.com/spf13/viper
```

### Loading Configuration

Let's create a configuration file to load environment variables. Here's an example of how you can structure your code:

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

// LoadConfig reads and loads the environment variables from a .env file
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

In this code:

- We define an `Env` struct to map our environment variables.
- The `LoadConfig` function reads the `.env` file, loads the environment variables using `viper`, and unmarshals them into the `Env` struct.

### Using the Configuration

You can now use the `LoadConfig` function to load your environment variables at the start of your application:

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

### Conclusion

Using `viper` to manage environment variables in Golang is an efficient way to handle configuration. It simplifies the process of loading variables and ensures your application is easy to configure and maintain. Give it a try in your next Golang project and enjoy a more streamlined configuration management process!

**Happy coding!**
