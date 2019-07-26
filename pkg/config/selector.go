package config

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/sherifabdlnaby/objx"
)

//Selector Contains a base in the config, this base can be static or dynamic, dynamic values must be get using Evaluate()
type Selector struct {
	isDynamic bool
	base      string
	parts     []part
}

//NewSelector Returns a new Selector used to evaluate dynamic fields in a config
func NewSelector(base string) (Selector, error) {
	isDynamic := false
	parts := splitToParts(base)

	if parts != nil {
		isDynamic = true
	}

	return Selector{
		base:      base,
		isDynamic: isDynamic,
		parts:     parts,
	}, nil
}

// Evaluate Evaluate dynamic values of config such as `image-@{image.title}.jpg` as a string, return error if it doesn't exist in supplied
// Data.
// TODO differentiate between not found in data, and being evaluated to 0 in a better way.
func (v *Selector) Evaluate(data map[string]interface{}) (string, error) {

	// No need to evaluate
	if !v.isDynamic {
		return v.base, nil
	}

	dataMap := objx.Map(data)

	// Return it directly
	if len(v.parts) == 1 {
		val := dataMap.Get(v.parts[0].string)
		if val.IsNil() {
			return "", fmt.Errorf("base [%s] is not found in job", v.parts[0].string)
		}
		return val.String(), nil
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
			return "", fmt.Errorf("base [%s] is not found in job", part.string)
		}

		builder.WriteString(partValue.String())
	}

	return builder.String(), nil

}

// EvaluateInt64 Evaluate dynamic values of config such as `@{image.width}` as Int64, return error if it doesn't exist in supplied
// Data. if base evaluated value is "" it returns zero value of Int64 = 0
func (v *Selector) EvaluateInt64(data map[string]interface{}) (int64, error) {
	str, err := v.Evaluate(data)

	if err != nil || str == "" {
		return 0, err
	}

	return strconv.ParseInt(str, 10, 64)

}

// EvaluateUint8 Evaluate dynamic values of config such as `@{image.width}` as uint8, return error if it doesn't exist in supplied
// Data. if base evaluated value is "" it returns zero value of Uint8 = 0
func (v *Selector) EvaluateUint8(data map[string]interface{}) (uint8, error) {
	str, err := v.Evaluate(data)

	if err != nil || str == "" {
		return 0, err
	}

	int8x, err := strconv.ParseUint(str, 10, 8)
	if err != nil {
		return 0, nil
	}

	return uint8(int8x), nil

}

// EvaluateFloat64 Evaluate dynamic values of config such as `@{image.width}` as float64, return error if it doesn't exist in supplied
// Data. if base evaluated value is "" it returns zero value of Float64 = 0.0
func (v *Selector) EvaluateFloat64(data map[string]interface{}) (float64, error) {
	str, err := v.Evaluate(data)

	if err != nil || str == "" {
		return 0, err
	}

	return strconv.ParseFloat(str, 10)

}

// EvaluateBool Evaluate dynamic values of config such as `@{image.doFlip}` as bool, return error if it doesn't exist in supplied
// Data. values such as 0,1,T,F,True,False are all possible values,if base evaluated value is "" it returns zero value
// of Float64 = 0.0
func (v *Selector) EvaluateBool(data map[string]interface{}) (bool, error) {
	str, err := v.Evaluate(data)

	if err != nil || str == "" {
		return false, err
	}

	return strconv.ParseBool(str)

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

//// Get gets base from config based on key, key access config using dot-notation (obj.Selector.array[0].Selector).
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
//		return Selector{}, fmt.Errorf("base [%s] is not found", key)
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
//		base:     *val,
//		isDynamic: isDynamic,
//		parts:     parts,
//	}
//
//	cw.cache["key"] = cacheField
//
//	return cacheField, nil
//}
