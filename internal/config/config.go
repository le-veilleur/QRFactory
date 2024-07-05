package config

import (
    "os"
)

type Config struct {
    ServerAddress string
}

func LoadConfig() *Config {
    return &Config{
        ServerAddress: getEnv("SERVER_ADDRESS", ":8080"),
    }
}

func getEnv(key, fallback string) string {
    if value, exists := os.LookupEnv(key); exists {
        return value
    }
    return fallback
}
