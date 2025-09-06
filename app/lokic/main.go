package main

import (
	"fmt"
	"log/slog"

	"github.com/joho/godotenv"
	"github.com/takevox/loki/app/lokic/config"
	"github.com/takevox/loki/lib"
)

func main() {
	var (
		err error
	)
	err = godotenv.Load()
	if err != nil {
		slog.Error(err.Error())
	}
	err = config.Load()
	if err != nil {
		slog.Error(err.Error())
	}

	plugin_manager, err := lib.NewPluginManager(config.PLUGINS_DIR)
	if err != nil {
		slog.Error(err.Error())
	}

	err = plugin_manager.LoadPlugins()
	if err != nil {
		slog.Error(err.Error())
	}

	err = plugin_manager.Initialize()
	if err != nil {
		slog.Error(err.Error())
		return
	}

	slog.Debug(fmt.Sprintf("plugin_dir = %s", plugin_manager.PluginDir))
}
