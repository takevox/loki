package lib

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/goccy/go-yaml"
)

type PluginConfig struct {
	Version string `yaml:"version"`
	Plugin  struct {
		Local  string `yaml:"local"`
		Remote string `yaml:"remote"`
	} `yaml:"plugin"`
}

type PluginInfo struct {
	Config  PluginConfig
	DirPath string
	Stdin   io.WriteCloser
	Stdout  io.ReadCloser
	Meta    struct {
		Version string
		Uri     string
	}
}

type PluginManager struct {
	PluginDir string
	Plugins   []PluginInfo
}

func NewPluginManager(plugin_dir string) (*PluginManager, error) {

	abs_plugin_dir, err := filepath.Abs(plugin_dir)
	if err != nil {
		return nil, err
	}
	plugin_manager := &PluginManager{
		PluginDir: abs_plugin_dir,
	}
	return plugin_manager, nil
}

func (pm *PluginManager) LoadPlugins() error {
	var err error
	plugin_fs := os.DirFS(pm.PluginDir)
	plugin_yaml_list, err := fs.Glob(plugin_fs, "**/plugin.yaml")
	if err != nil {
		return err
	}
	for _, plugin_yaml_path := range plugin_yaml_list {
		pm.loadPlugin(fmt.Sprintf("%s/%s", pm.PluginDir, plugin_yaml_path))
	}

	return nil
}

func (pm *PluginManager) Startup() error {
	for _, plugin_info := range pm.Plugins {
		pm.startupPlugin(&plugin_info)
	}
	return nil
}

func (pm *PluginManager) loadPlugin(plugin_yaml_path string) error {
	var plugin_info PluginInfo
	var err error

	yml, err := os.ReadFile(plugin_yaml_path)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(yml, &plugin_info.Config)
	if err != nil {
		return err
	}

	plugin_info.DirPath, _ = filepath.Split(plugin_yaml_path)

	pm.Plugins = append(pm.Plugins, plugin_info)

	return nil
}

func (pm *PluginManager) startupPlugin(plugin_info *PluginInfo) error {

	var (
		local_path string = plugin_info.Config.Plugin.Local
	)

	if !filepath.IsAbs(local_path) {
		local_path = filepath.Join(plugin_info.DirPath, local_path)
	}

	cmd := exec.Command(plugin_info.Config.Plugin.Local)
	cmd.Dir = plugin_info.DirPath

	plugin_info.Stdin, _ = cmd.StdinPipe()
	plugin_info.Stdout, _ = cmd.StdoutPipe()

	cmd

	return nil
}
