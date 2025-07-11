// internal/conf/conf.go
package conf

import (
	"encoding/json"
	"log"
	"os"
)

type ServerConf struct {
	Port string `json:"port"`
}

type DatabaseConf struct {
	Type     string `json:"type"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

// Config 结构体已简化，不再包含 BanResponses
type Config struct {
	Server   ServerConf   `json:"server"`
	Database DatabaseConf `json:"database"`
}

var Conf *Config

func Init() {
	log.Println("Loading configuration...")
	bytes, err := os.ReadFile("configs/config.json")
	if err != nil {
		log.Fatalf("Failed to read config file: %v", err)
	}
	Conf = &Config{}
	if err := json.Unmarshal(bytes, Conf); err != nil {
		log.Fatalf("Failed to parse config file: %v", err)
	}
	log.Println("Configuration loaded.")
}