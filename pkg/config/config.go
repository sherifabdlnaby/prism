package config

import (
	"fmt"
	"github.com/sherifabdlnaby/objx"
	"github.com/sherifabdlnaby/prism/pkg/transaction"
	"regexp"
	"strings"
)

// TODO evaluate defaults. (add ability to add default value)
var fieldsRegex = regexp.MustCompile(`@{([\w@.]+)}`)

//Value Contains a value in the config, this value can be static or dynamic, dynamic values must be get using Evaluate()
type Value struct {
	isDynamic bool
	value     objx.Value
	parts     []part
}

type part struct {
	string string
	eval   bool
}

// Config used to ease getting values from config using dot notation (obj.Value.array[0].Value), and used to resolve
// dynamic values.
type Config struct {
	config objx.Map
	cache  map[string]Value
}

// NewConfig construct new Config from map[string]interface{}
func NewConfig(config map[string]interface{}) *Config {
	return &Config{config: objx.Map(config), cache: make(map[string]Value)}
}

// Get gets value from config based on key, key access config using dot-notation (obj.Value.array[0].Value).
// Get will also evaluate dynamic fields in config ( @{dynamic.Value} ) using data, pass nill if you're sure that this
// Value is constant. returns error if key or dynamic Value doesn't exist.
func (cw *Config) Get(key string, data transaction.ImageData) (Value, error) {
	// Check cache
	if val, ok := cw.cache[key]; ok {
		return val, nil
	}

	val := cw.config.Get(key)
	if val.IsNil() {
		return Value{}, fmt.Errorf("value [%s] is not found", key)
	}

	str := val.String()
	parts := splitToParts(str)
	isDynamic := false

	if parts != nil {
		isDynamic = true
	}

	cacheField := Value{
		value:     *val,
		isDynamic: isDynamic,
		parts:     parts,
	}

	cw.cache["key"] = cacheField

	return cacheField, nil
}

func splitToParts(str string) []part {
	parts := make([]part, 0)
	matches := fieldsRegex.FindAllStringSubmatch(str, -1)
	if len(matches) == 0 {
		return nil
	}
	for i, submatches := range matches {

		idx := strings.Index(str, submatches[0])

		subStr := str[:idx]

		if len(subStr) > 0 {
			parts = append(parts, part{subStr, false})
		}

		parts = append(parts, part{submatches[1], true})

		str = str[idx+len(submatches[0]):]

		// Append the rest of the string
		if i == len(matches)-1 && len(str) > 0 {
			parts = append(parts, part{str, false})

		}
	}
	return parts
}

//Evaluate Evaluate dynamic values of config such as `@{image.title}`, return error if it doesn't exist in supplied
// ImageData. (Returned values still must be checked for its type)
func (v *Value) Evaluate(data transaction.ImageData) (objx.Value, error) {

	// No need to evaluate
	if !v.isDynamic {
		return v.value, nil
	}

	dataMap := objx.Map(data)

	// Return it directly
	if len(v.parts) == 1 {
		return *objx.NewValue(dataMap.Get(v.parts[0].string)), nil
	}

	var builder strings.Builder
	var partValue *objx.Value
	for _, part := range v.parts {
		if !part.eval {
			builder.WriteString(part.string)
			continue
		}

		partValue = dataMap.Get(part.string)
		if partValue.IsNil() {
			return objx.Value{}, fmt.Errorf("value [%s] is not found in transaction", part.string)
		}

		builder.WriteString(partValue.String())
	}

	return *objx.NewValue(builder.String()), nil

}

//Get Return Config Values (must be used for static config values only)
func (v *Value) Get() *objx.Value {
	return &v.value
}
