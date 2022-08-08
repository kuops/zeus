package config

import (
	"github.com/spf13/viper"
	"zeus/internal/server"
	"zeus/pkg/database"
)

type Config struct {
	Server   *server.Config
	Database *database.Config
}

func (cfg *Config) ParseFromFile(configFile string) error {
	viper.SetConfigFile(configFile)
	err := viper.ReadInConfig()
	if err != nil {
		return err
	}
	err = viper.Unmarshal(cfg)
	if err != nil {
		return err
	}
	return nil
}
