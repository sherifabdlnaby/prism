package types

import (
	"fmt"
	"github.com/sherifabdlnaby/objx"
	"regexp"
	"strings"
)

// TODO evaluate defaults. (add ability to add default value)
var fieldsRegex = regexp.MustCompile(`@{([\w@.]+)}`)

type field struct {
	isDynamic bool
	value     objx.Value
	parts     []part
}

type part struct {
	string string
	eval   bool
}

// Config used to ease getting values from config using dot notation (obj.field.array[0].field), and used to resolve
// dynamic values.
type Config struct {
	config objx.Map
	cache  map[string]field
}

// NewConfig construct new Config from map[string]interface{}
func NewConfig(config map[string]interface{}) *Config {
	return &Config{config: objx.Map(config), cache: make(map[string]field)}
}

// Get gets value from config based on key, key access config using dot-notation (obj.field.array[0].field).
// Get will also evaluate dynamic fields in config ( @{dynamic.field} ) using data, pass nill if you're sure that this
// field is constant. returns error if key or dynamic field doesn't exist.
func (cw *Config) Get(key string, data ImageData) (objx.Value, error) {
	// Check cache
	if val, ok := cw.cache[key]; ok {
		return evaluate(&val, data)
	}

	val := cw.config.Get(key)
	if val.IsNil() {
		return objx.Value{}, fmt.Errorf("field \"%s\" is not found", key)
	}

	str := val.String()
	parts := splitToParts(str)
	isDynamic := false

	if parts != nil {
		isDynamic = true
	}

	cacheField := field{
		value:     *val,
		isDynamic: isDynamic,
		parts:     parts,
	}

	cw.cache["key"] = cacheField

	return evaluate(&cacheField, data)
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

func evaluate(field *field, data ImageData) (objx.Value, error) {

	// No need to evaluate
	if !field.isDynamic {
		return field.value, nil
	}

	dataMap := objx.Map(data)

	// Return it directly
	if len(field.parts) == 1 {
		return *objx.NewValue(dataMap.Get(field.parts[0].string)), nil
	}

	var builder strings.Builder

	var partValue *objx.Value
	for _, part := range field.parts {
		if !part.eval {
			builder.WriteString(part.string)
			continue
		}

		partValue = dataMap.Get(part.string)
		if partValue.IsNil() {
			return objx.Value{}, fmt.Errorf("field \"%s\" is not found", part.string)
		}

		builder.WriteString(partValue.String())
	}

	return *objx.NewValue(builder.String()), nil

}
