package database

import (
	"fmt"
	"time"
)

type Config struct {
	Host                  string        `mapstructure:"host"`
	Port                  string        `mapstructure:"port"`
	Username              string        `mapstructure:"username"`
	Password              string        `mapstructure:"password"`
	DBName                string        `mapstructure:"db_name"`
	MaxConnectionLifetime time.Duration `mapstructure:"max_connection_lifetime"`
	MaxConnectionIdleTime time.Duration `mapstructure:"max_connection_idle_time"`
}

func (c *Config) ConnectionURL() string {
	return fmt.Sprintf("%v:%v@tcp(%v:%v)/%v?charset=utf8mb4&parseTime=True&loc=Local",
		c.Username,
		c.Password,
		c.Host,
		c.Port,
		c.DBName)
}
