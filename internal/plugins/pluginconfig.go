package plugins

import (
	"fmt"

	"github.com/volte6/gomud/internal/configs"
)

type PluginConfig struct {
	pluginName string
}

func (p *PluginConfig) Set(name string, val any) {
	configs.SetVal(fmt.Sprintf(`Modules.%s.%s`, p.pluginName, name), fmt.Sprintf(`%v`, val))
}

func (p *PluginConfig) Get(name string) any {
	m := configs.Flatten(configs.GetModulesConfig())
	return m[fmt.Sprintf(`%s.%s`, p.pluginName, name)]
}
