package plugins

import (
	"fmt"

	"github.com/GoMudEngine/GoMud/internal/configs"
	"github.com/GoMudEngine/GoMud/internal/mudlog"
)

type PluginConfig struct {
	pluginName string
}

func (p *PluginConfig) Set(name string, val any) {
	if err := configs.SetVal(fmt.Sprintf(`Modules.%s.%s`, p.pluginName, name), fmt.Sprintf(`%v`, val)); err != nil {
		mudlog.Error(`PluginConfig.Set`, `plugin`, p.pluginName, `key`, name, `error`, err)
	}
}

func (p *PluginConfig) Get(name string) any {
	m := configs.Flatten(configs.GetModulesConfig())
	return m[fmt.Sprintf(`%s.%s`, p.pluginName, name)]
}
