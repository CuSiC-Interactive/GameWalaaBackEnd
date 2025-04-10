package config

import (
	"github.com/spf13/viper"
)

func LoadConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yml")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		panic("Unable to read config file: " + err.Error())
	}
}

func GetString(key string) string {
	return viper.GetString(key)
}
