package config

import (
	"os"

	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func ReadConfig(parameter string) interface{} {
	path, err := os.Getwd()
	if err != nil {
		log.Error(err)
	}

	v := viper.New()
	v.AddConfigPath(path + "/config")
	v.SetConfigName("config")
	v.SetConfigType("yaml")

	v.WatchConfig()
	v.OnConfigChange(func(e fsnotify.Event) {
		requestLogger := log.WithFields(log.Fields{
			"file": e.Name,
			"type": e.Op,
		})
		requestLogger.Info("Config file updated.")
	})

	if err := v.ReadInConfig(); err != nil {
		log.Fatal("Reading config yaml is fail: ", err)
	}
	if v.IsSet(parameter) {
		return v.Get(parameter)
	} else {
		log.Fatal("Not is config: ", parameter)
	}

	return nil
}
