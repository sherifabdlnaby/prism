package manager

import (
	"fmt"
	"github.com/sherifabdlnaby/prism/pkg/types"
)

var plugins = make(map[string]types.Component)

func Load(name, plugin string) error {
	_, ok := plugins[name]
	if ok {
		return fmt.Errorf("plugin instance with name [%s] is already loaded", name)
	}

	component, ok := registered[plugin]
	if !ok {
		return fmt.Errorf("plugin type [%s] doesn't exist", plugin)
	}

	pluginInstance := component()

	plugins[name] = pluginInstance

	return nil
}

func Get(name string) types.Component {
	return plugins[name]
}
