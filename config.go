package main

import (
	"fmt"
	"io"
	"os"

	"github.com/pelletier/go-toml/v2"
)

const StartTypeAuto = "auto"
const StartTypeManual = "manual"
const StartTypeDisabled = "disabled"

type ServiceInfo struct {
	Name         string `toml:"name"`
	DisplayName  string `toml:"display_name"`
	Description  string `toml:"description"`
	StartType    string `toml:"start_type"`
	Interactive  bool   `toml:"interactive"`
	ExecRetry    bool   `toml:"exec_retry"`
	ExecMaxRetry int    `toml:"exec_max_retry"`
}

type ExecInfo struct {
	Command string   `toml:"command"`
	Args    []string `toml:"args"`
	Envs    []string `toml:"envs"`
}

type LogInfo struct {
	Path      string `toml:"path"`
	MaxSize   int    `toml:"max_size"`
	MaxBackup int    `toml:"max_backup"`
}

type Config struct {
	Service ServiceInfo
	Exec    ExecInfo
	Log     LogInfo
}

func loadConfig(path string) (*Config, error) {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("config file not found: %v", path)
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open config file failed: %w", err)
	}
	defer file.Close()

	config := &Config{
		Service: ServiceInfo{
			StartType:   StartTypeAuto,
			Interactive: false,
		},
		Log: LogInfo{
			MaxBackup: 5,
			MaxSize:   5,
		},
	}
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("load config file failed: %w", err)
	}

	err = toml.Unmarshal(data, config)
	if err != nil {
		return nil, fmt.Errorf("parse config file failed: %w", err)
	}

	return config, nil
}
