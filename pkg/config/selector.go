package config

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/sherifabdlnaby/objx"
)

//Selector Contains a value in the config, this value can be static or dynamic, dynamic values must be get using Evaluate()
type Selector struct {
	isDynamic bool
	value     objx.Value
	parts     []part
}

// TODO make string base not interface
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

	// Return it directly
	if len(v.parts) == 1 {
		val := dataMap.Get(v.parts[0].string)
		if val.IsNil() {
			return "", fmt.Errorf("value [%s] is not found in transaction", v.parts[0].string)
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
			return "", fmt.Errorf("value [%s] is not found in transaction", part.string)
		}

		builder.WriteString(partValue.String())
	}

	return builder.String(), nil

}

//Evaluate Evaluate dynamic values of config such as `@{image.title}`, return error if it doesn't exist in supplied
// Data. (Returned values still must be checked for its type)
func (v *Selector) EvaluateInt64(data map[string]interface{}) (int64, error) {
	str, err := v.Evaluate(data)

	if err != nil || str == "" {
		return 0, err
	}

	return strconv.ParseInt(str, 10, 64)

}

//Evaluate Evaluate dynamic values of config such as `@{image.title}`, return error if it doesn't exist in supplied
// Data. (Returned values still must be checked for its type)
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

//Evaluate Evaluate dynamic values of config such as `@{image.title}`, return error if it doesn't exist in supplied
// Data. (Returned values still must be checked for its type)
func (v *Selector) EvaluateFloat64(data map[string]interface{}) (float64, error) {
	str, err := v.Evaluate(data)

	if err != nil || str == "" {
		return 0, err
	}

	return strconv.ParseFloat(str, 10)

}

//Evaluate Evaluate dynamic values of config such as `@{image.title}`, return error if it doesn't exist in supplied
// Data. (Returned values still must be checked for its type)
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

func (v *Selector) IsDynamic() bool {
	return v.isDynamic
}

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
