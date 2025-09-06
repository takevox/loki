package lib

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"connectrpc.com/connect"
	"github.com/goccy/go-yaml"
	lokiv1 "github.com/takevox/loki/gen/loki/v1"
	"github.com/takevox/loki/gen/loki/v1/lokiv1connect"
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
	HttpClient *http.Client
	Client     lokiv1connect.PluginServiceClient
}

type PluginManager struct {
	PluginDir string
	Plugins   []PluginInfo
}

func NewPluginManager(pluginDir string) (*PluginManager, error) {

	abs_plugin_dir, err := filepath.Abs(pluginDir)
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

func (pm *PluginManager) Initialize() error {
	for _, plugin_info := range pm.Plugins {
		err := pm.initializePlugin(&plugin_info)
		if err != nil {
			log.Fatalln(err)
		}
	}
	return nil
}

func (pm *PluginManager) loadPlugin(pluginYamlPath string) error {
	var plugin_info PluginInfo
	var err error

	yml, err := os.ReadFile(pluginYamlPath)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(yml, &plugin_info.Config)
	if err != nil {
		return err
	}

	plugin_info.DirPath, _ = filepath.Split(pluginYamlPath)

	pm.Plugins = append(pm.Plugins, plugin_info)

	return nil
}

func (pm *PluginManager) initializePlugin(pluginInfo *PluginInfo) error {

	var (
		err        error
		local_path string = pluginInfo.Config.Plugin.Local
	)

	if !filepath.IsAbs(local_path) {
		local_path = filepath.Join(pluginInfo.DirPath, local_path)
	}

	cmd := exec.Command(local_path)
	cmd.Dir = pluginInfo.DirPath

	pluginInfo.Stdin, _ = cmd.StdinPipe()
	pluginInfo.Stdout, _ = cmd.StdoutPipe()

	err = cmd.Start()
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(pluginInfo.Stdout)
	scan_state := 0
	for scanner.Scan() {
		switch scan_state {
		case 0:
			pluginInfo.Meta.Version = scanner.Text()
			scan_state = 1
		case 1:
			pluginInfo.Meta.Uri = scanner.Text()
			scan_state = 2
		}
		if scan_state == 2 {
			break
		}
	}
	if scan_state != 2 {
		return fmt.Errorf("起動情報の読み取りに失敗(%s)", local_path)
	}

	pluginInfo.HttpClient = &http.Client{Timeout: 30 * time.Second}
	client := lokiv1connect.NewPluginServiceClient(pluginInfo.HttpClient, pluginInfo.Meta.Uri)

	_, err = client.Initialize(context.Background(), connect.NewRequest(&lokiv1.InitializeRequest{}))
	if err != nil {
		return err
	}

	pluginInfo.Client = client
	return nil
}
