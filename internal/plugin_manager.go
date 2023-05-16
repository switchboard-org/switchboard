package internal

import (
	"errors"
	"fmt"
	"github.com/hashicorp/go-plugin"
	"github.com/switchboard-org/plugin-sdk/sbsdk"
	"os/exec"
)

type PluginManager interface {
	LoadPlugin(RequiredProviderBlock) error
	PluginClient(string) (*plugin.Client, error)
	ProviderInstance(string) (sbsdk.Provider, error)
	KillPlugin(string) error
	KillAllPlugins()
	LoadedPlugins() []string
}

type PluginConfig struct {
	Name    string
	Source  string
	Version string
	Client  *plugin.Client
}

type DefaultPluginManager struct {
	plugins       []PluginConfig
	providerCache map[string]sbsdk.Provider
}

func NewDefaultPluginManager() PluginManager {
	return &DefaultPluginManager{
		providerCache: make(map[string]sbsdk.Provider),
	}
}

func (pm *DefaultPluginManager) LoadPlugin(provider RequiredProviderBlock) error {
	for _, plug := range pm.plugins {
		if plug.Name == provider.Name {
			return errors.New("plugin is already loaded")
		}
	}
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: sbsdk.HandshakeConfig,
		Plugins:         pluginMap,
		Cmd:             exec.Command(fmt.Sprintf("./.switchboard/packages/%s/%s/switchboard_plugin", provider.Source, provider.Version)),
	})
	pm.plugins = append(pm.plugins, PluginConfig{
		Name:    provider.Name,
		Source:  provider.Source,
		Version: provider.Version,
		Client:  client,
	})

	return nil
}

func (pm *DefaultPluginManager) PluginClient(name string) (*plugin.Client, error) {
	for _, plug := range pm.plugins {
		if plug.Name == name {
			return plug.Client, nil
		}
	}
	return nil, errors.New("plugin is not available")
}

func (pm *DefaultPluginManager) ProviderInstance(name string) (sbsdk.Provider, error) {
	if existingProvider, ok := pm.providerCache[name]; ok {
		return existingProvider, nil
	}
	client, err := pm.PluginClient(name)
	if err != nil {
		return nil, err
	}
	rpcClient, err := client.Client()
	if err != nil {
		return nil, err
	}
	raw, err := rpcClient.Dispense("provider")
	if err != nil {
		return nil, err
	}
	provider := raw.(sbsdk.Provider)
	pm.providerCache[name] = provider
	return provider, nil
}

func (pm *DefaultPluginManager) KillPlugin(name string) error {
	for i, plug := range pm.plugins {
		if plug.Name == name {
			plug.Client.Kill()
			pm.plugins = append(pm.plugins[:i], pm.plugins[i+1:]...)
			return nil
		}
	}
	return errors.New("plugin is not loaded")
}

func (pm *DefaultPluginManager) KillAllPlugins() {
	for _, plug := range pm.plugins {
		plug.Client.Kill()
	}
	pm.plugins = []PluginConfig{}
}

func (pm *DefaultPluginManager) LoadedPlugins() []string {
	var outputList []string
	for _, plug := range pm.plugins {
		outputList = append(outputList, fmt.Sprintf("%s (%s@%s)", plug.Name, plug.Source, plug.Version))
	}
	return outputList
}

var pluginMap = map[string]plugin.Plugin{
	"provider": &sbsdk.ProviderPlugin{},
}
