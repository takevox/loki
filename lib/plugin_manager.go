package lib

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"time"

	"connectrpc.com/connect"
	"github.com/goccy/go-yaml"
	"github.com/takevox/loki/gen/loki/v1/lokiv1connect"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
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
		Version  string
		Endpoint string
	}
	HttpClient *http.Client
	Client     lokiv1connect.PluginServiceClient
	Command    *exec.Cmd
	Context    context.Context
	Cancel     context.CancelFunc
}

type PluginManager struct {
	PluginDir string
	Plugins   []*PluginInfo
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

func (pm *PluginManager) InitializePlugins() error {
	for _, plugin_info := range pm.Plugins {
		err := pm.initializePlugin(plugin_info)
		if err != nil {
			slog.Error(err.Error())
		}
	}
	return nil
}

func (pm *PluginManager) TerminatePlugins() error {
	for _, plugin_info := range pm.Plugins {
		err := pm.terminatePlugin(plugin_info)
		if err != nil {
			slog.Error(err.Error())
		}
	}
	return nil
}

func (pm *PluginManager) loadPlugin(pluginYamlPath string) error {
	var plugin_info *PluginInfo
	var err error

	plugin_info = &PluginInfo{}

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

	pluginInfo.Context, pluginInfo.Cancel = signal.NotifyContext(context.Background(), os.Interrupt)

	cmd := exec.CommandContext(pluginInfo.Context, local_path)
	cmd.Dir = pluginInfo.DirPath
	cmd.Cancel = func() error {
		return cmd.Process.Kill()
	}

	pluginInfo.Command = cmd
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
			pluginInfo.Meta.Endpoint = scanner.Text()
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
	client := lokiv1connect.NewPluginServiceClient(pluginInfo.HttpClient, pluginInfo.Meta.Endpoint)

	_, err = client.Initialize(context.Background(), connect.NewRequest(&emptypb.Empty{}))
	if err != nil {
		return err
	}

	pluginInfo.Client = client
	return nil
}

func (pm *PluginManager) terminatePlugin(pluginInfo *PluginInfo) error {
	pluginInfo.Client.Terminate(context.Background(), connect.NewRequest(&emptypb.Empty{}))
	pluginInfo.Cancel()
	pluginInfo.Command.Process.Wait()

	return nil
}
