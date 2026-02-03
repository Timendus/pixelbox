package server

import (
	"encoding/json"
	"io"
	"os"
)

type Config struct {
	Server  ConfigServer `json:"server"`
	Devices []Device     `json:"devices"`
}

type ConfigServer struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

type Device struct {
	Name    string `json:"name"`
	Mac     string `json:"mac"`
	Channel int    `json:"channel"`
}

var config Config

func InitConfig() error {
	configFile, err := os.Open("config.json")
	if err != nil {
		return err
	}
	configBytes, _ := io.ReadAll(configFile)
	err = json.Unmarshal(configBytes, &config)
	if err != nil {
		return err
	}
	if config.Server.Port == 0 {
		config.Server.Port = 3000
	}
	if config.Server.Host == "" {
		config.Server.Host = "127.0.0.1"
	}
	return nil
}

func GetConfig() Config {
	return config
}
