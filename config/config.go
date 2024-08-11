package config

import (
    "log"
    "github.com/spf13/viper"
)

type Config struct {
    Database DatabaseConfig
	JWT      JWTConfig
	Server	 ServerConfig
}

type DatabaseConfig struct {
    Host     string
    Port     int
    User     string
    Password string
    DBName   string
}

type JWTConfig struct {
	Secret string
}

type ServerConfig struct {
	Port string
}

var AppConfig Config

func LoadConfig() {
    viper.SetConfigFile("config.yaml")
    if err := viper.ReadInConfig(); err != nil {
        log.Fatalf("Error reading config file, %s", err)
    }

    err := viper.Unmarshal(&AppConfig)
    if err != nil {
        log.Fatalf("Unable to decode into struct, %v", err)
    }
}
