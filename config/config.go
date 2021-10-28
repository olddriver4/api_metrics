package config

import (
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func ReadConfig(parameter string) interface{} {
	path, err := os.Getwd()
	if err != nil {
		log.Error(err)
	}

	viper.AddConfigPath(path + "/config")
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatal("Reading config yaml is fail: ", err)
	}
	if viper.IsSet(parameter) {
		return viper.Get(parameter)
	} else {
		log.Fatal("Not is config: ", parameter)
	}

	return nil
}
