package config

import (
	"fmt"
	"log/slog"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

func SetConfig(filePath string) (*viper.Viper, error) {
	c := viper.New()
	c.SetConfigName("config")
	c.SetConfigType("yaml")

	c.AddConfigPath(filePath)
	err := c.ReadInConfig()
	if err != nil {
		slog.Error(err.Error())
	}

	c.OnConfigChange(func(e fsnotify.Event) {
		fmt.Println("Config file changed and Reloaded:", e.Name)
		err := c.ReadInConfig()
		if err != nil {
			slog.Error(err.Error())
		}
	})
	c.WatchConfig()

	return c, nil
}
