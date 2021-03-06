package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

type DatabaseConfig struct {
	Address    string `json:"address"`
	Database   string `json:"database"`
	Collection string `json:"collection"`
}

type Settings struct {
	ListenAddress string         `json:"listen_address"`
	Database      DatabaseConfig `json:"database"`
}

func Init(cfgFile string) (*Settings, error) {
	conf := &Settings{}

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		pwd, _ := os.Getwd()
		viper.SetConfigName("configs/example_config")
		viper.AddConfigPath(pwd)
		viper.AutomaticEnv()
		viper.SetConfigType("json")
	}

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("cannot read config file %s: %w", cfgFile, err)
	}

	if err := viper.Unmarshal(conf); err != nil {
		return nil, fmt.Errorf("cannot parse config file %s: %w", cfgFile, err)
	}

	return conf, nil
}
