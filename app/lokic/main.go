package main

import (
	"log"

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
		log.Fatalln(err)
	}
	err = config.Load()
	if err != nil {
		log.Fatalln(err)
	}

	plugin_manager, err := lib.NewPluginManager(config.PLUGINS_DIR)
	if err != nil {
		log.Fatalln(err)
		return
	}

	plugin_manager.LoadPlugins()
	plugin_manager.Startup()

	log.Println("plugin_dir = ", plugin_manager.PluginDir)
}
