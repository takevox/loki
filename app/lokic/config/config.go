package config

import "os"

var (
	PLUGINS_DIR string = "./plugins"
)

func _GetEnv(key string, empty_value string) string {
	value, eixsts := os.LookupEnv(key)
	if !eixsts {
		return empty_value
	}
	return value
}

func Load() error {
	PLUGINS_DIR = _GetEnv("PLUGINS_DIR", PLUGINS_DIR)

	return nil
}
