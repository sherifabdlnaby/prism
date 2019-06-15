package config

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/sherifabdlnaby/objx"
	"gopkg.in/go-playground/validator.v9"
)

// TODO evaluate defaults. (add ability to add default value)
var fieldsRegex = regexp.MustCompile(`@{([\w@.]+)}`)

// Config used to ease getting values from config using dot notation (obj.Selector.array[0].Selector), and used to resolve
// dynamic values.
type Config struct {
	config objx.Map
}

// NewConfig construct new Config from map[string]interface{}
func NewConfig(config map[string]interface{}) *Config {
	return &Config{config: objx.Map(config)}
}

//NewValue creates a temporary value and it doesn't cache it
//Populate func Populates the config
func (cw *Config) Populate(def interface{}) error {

	config := &mapstructure.DecoderConfig{
		WeaklyTypedInput: true,
		Result:           def,
	}

	decoder, err := mapstructure.NewDecoder(config)

	if err != nil {
		return err
	}

	err = decoder.Decode(cw.config)

	if err != nil {
		return err
	}

	// Validate
	err = validator.New().Struct(def)

	return err
}

//NewSelector func for dynamic configs
func (cw *Config) NewSelector(base interface{}) (Selector, error) {
	return NewSelector(base)
}

//NewSelector func
func NewSelector(base interface{}) (Selector, error) {
	val := objx.NewValue(base)

	if val.IsNil() {
		return Selector{}, fmt.Errorf("value to selector is nil")
	}

	str := val.String()
	parts := splitToParts(str)
	isDynamic := false

	if parts != nil {
		isDynamic = true
	}

	return Selector{
		value:     *val,
		isDynamic: isDynamic,
		parts:     parts,
	}, nil
}

//Evaluate Evaluate dynamic values of config such as `@{image.title}`, return error if it doesn't exist in supplied
// Data. (Returned values still must be checked for its type)
func (v *Selector) Evaluate(data map[string]interface{}) (string, error) {

	// No need to evaluate
	if !v.isDynamic {
		return v.value.String(), nil
	}

	dataMap := objx.Map(data)

	var retValue string

	// Return it directly
	if len(v.parts) == 1 {
		val := dataMap.Get(v.parts[0].string)
		if val.IsNil() {
			return "", fmt.Errorf("value [%s] is not found in transaction", v.parts[0].string)
		}
		retValue = val.String()

	} else {
		var builder strings.Builder
		var partValue *objx.Value
		for _, part := range v.parts {
			if !part.eval {
				builder.WriteString(part.string)
				continue
			}

			partValue = dataMap.Get(part.string)
			if partValue.IsNil() {
				return "", fmt.Errorf("value [%s] is not found in transaction", part.string)
			}

			builder.WriteString(partValue.String())
		}
		retValue = builder.String()
	}
	return retValue, nil

}

//Selector Contains a value in the config, this value can be static or dynamic, dynamic values must be get using Evaluate()
type Selector struct {
	isDynamic bool
	value     objx.Value
	parts     []part
}

type part struct {
	string string
	eval   bool
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

//CheckInValues func
func CheckInValues(data interface{}, values ...interface{}) error {
	for _, value := range values {
		if data == value {
			return nil
		}
	}
	return fmt.Errorf("value must be any of %v, found %v instead", values, data)
}

//// Get gets value from config based on key, key access config using dot-notation (obj.Selector.array[0].Selector).
//// Get will also evaluate dynamic fields in config ( @{dynamic.Selector} ) using data, pass nill if you're sure that this
//// Selector is constant. returns error if key or dynamic Selector doesn't exist.
//func (cw *Config) Get(key string) (Selector, error) {
//	// Check cache
//	if val, ok := cw.cache[key]; ok {
//		return val, nil
//	}
//
//	val := cw.config.Get(key)
//	if val.IsNil() {
//		return Selector{}, fmt.Errorf("value [%s] is not found", key)
//	}
//
//	str := val.String()
//	parts := splitToParts(str)
//	isDynamic := false
//
//	if parts != nil {
//		isDynamic = true
//	}
//
//	cacheField := Selector{
//		value:     *val,
//		isDynamic: isDynamic,
//		parts:     parts,
//	}
//
//	cw.cache["key"] = cacheField
//
//	return cacheField, nil
//}
