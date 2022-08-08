package server

type Config struct {
	Port  int  `mapstructure:"port"`
	Debug bool `mapstructure:"debug"`
}
