package types

import "github.com/stretchr/objx"

// Config is the basic `static` config json.
type Config map[string]interface{}

type field struct {
	isDynamic bool
	value     objx.Value
}

type ConfigWrapper struct {
	Config objx.Map
	cache  map[string]field
}

func (cw *ConfigWrapper) Get(key string) objx.Value {
	// Check cache
	if val, ok := cw.cache[key]; ok {
		return val.value
	}

	val := cw.Config.Get(key)

	cacheField := field{
		isDynamic: false,
		value:     *val,
	}

	cw.cache["key"] = cacheField

	return cw.cache["key"].value
}
