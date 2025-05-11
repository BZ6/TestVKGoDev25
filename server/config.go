package main

import (
    "log"
    "os"

    "gopkg.in/yaml.v2"
)

type ServerConfig struct {
    Port string `yaml:"port"`
}

type Config struct {
    Server ServerConfig `yaml:"server"`
}

func setDefaultConfig() Config {
    return Config{
        Server: ServerConfig{
            Port: "50051",
        },
    }
}

func loadConfig() Config {
    file, err := os.Open("config.yaml")
    if err != nil {
        log.Printf("Config file not found, using default values")
        return setDefaultConfig()
    }
    defer file.Close()

    var config Config
    decoder := yaml.NewDecoder(file)
    if err := decoder.Decode(&config); err != nil {
        log.Fatalf("Failed to parse config file: %v", err)
    }

    return setDefaultConfigWithOverrides(config)
}

func setDefaultConfigWithOverrides(config Config) Config {
	defaultConfig := setDefaultConfig()
	
    if config.Server.Port == "" {
        log.Printf("Port not specified in config, using default: %s", defaultConfig.Server.Port)
        config.Server.Port = defaultConfig.Server.Port
    }
    return config
}
