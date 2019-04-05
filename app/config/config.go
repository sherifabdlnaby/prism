package config

import (
	"fmt"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cast"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"reflect"
	"regexp"
	"strings"
)

// Plugin used for YAML decoding
type Plugin struct {
	Plugin string                 `yaml:"plugin"`
	Number int                    `yaml:"number"`
	Config map[string]interface{} `yaml:"config"`
}

// Input used for YAML decoding
type Input struct {
	Plugin `yaml:",inline" mapstructure:",squash"`
}

// Processor used for YAML decoding
type Processor struct {
	Plugin `yaml:",inline" mapstructure:",squash"`
}

// Output used for YAML decoding
type Output struct {
	Plugin `yaml:",inline"`
}

// InputsConfig used for YAML decoding
type InputsConfig struct {
	Inputs map[string]Input `yaml:"inputs"`
}

// ProcessorsConfig used for YAML decoding
type ProcessorsConfig struct {
	Processors map[string]Input `yaml:"processors"`
}

// OutputsConfig used for YAML decoding
type OutputsConfig struct {
	Outputs map[string]Input `yaml:"outputs"`
}

// TODO support default values
var envRegex = regexp.MustCompile(`\${([\w@.]+)}`)

// Load loads a .yaml file into out. resolveEnv will replace ${ENV_VAR} with value of env variable "ENV_VAR"
func Load(filePath string, out interface{}, resolveEnv bool) error {

	fileBytes, err := ioutil.ReadFile(filePath)

	if err != nil {
		return err
	}

	if resolveEnv {
		fileBytes, err = resolveEnvFromBytes(fileBytes)
		if err != nil {
			return err
		}
	}

	return unmarshal(fileBytes, out)
}

func unmarshal(bytes []byte, out interface{}) error {
	mapRaw := make(map[interface{}]interface{})

	err := yaml.Unmarshal(bytes, mapRaw)

	mapString := recursivelyTurnYAMLMaps(mapRaw)

	config := &mapstructure.DecoderConfig{
		WeaklyTypedInput: true,
		Result:           &out,
		DecodeHook:       resolveTypes,
	}

	decoder, err := mapstructure.NewDecoder(config)

	if err != nil {
		return err
	}

	err = decoder.Decode(mapString)

	if err != nil {
		return err
	}

	return nil
}

// recursivelyTurnYAMLMaps turns nested YAML maps from map[interface{}]interface{} to map[string]interface{}
// inspired by solutions in https://github.com/go-yaml/yaml/issues/139 especially elastic's #issuecomment-183937598
func recursivelyTurnYAMLMaps(in interface{}) interface{} {
	switch in := in.(type) {
	case []interface{}:
		res := make([]interface{}, len(in))
		for i, v := range in {
			res[i] = recursivelyTurnYAMLMaps(v)
		}
		return res
	case map[interface{}]interface{}:
		res := make(map[string]interface{})
		for k, v := range in {
			res[fmt.Sprintf("%v", k)] = recursivelyTurnYAMLMaps(v)
		}
		return res
	case Plugin:
		res := make(map[string]interface{})
		for k, v := range in.Config {
			res[fmt.Sprintf("%v", k)] = recursivelyTurnYAMLMaps(v)
		}
		return res
	default:
		return in
	}
}

func resolveTypes(in reflect.Kind, out reflect.Kind, val interface{}) (interface{}, error) {

	if in != reflect.String {
		return val, nil
	}

	// try cast to number if possible
	if out == reflect.Interface {
		// Try to Int
		toInt, err := cast.ToIntE(val)
		if err == nil {
			return toInt, nil
		}

		toIntFloat32, err := cast.ToFloat32E(val)
		if err == nil {
			return toIntFloat32, nil
		}

		toIntFloat64, err := cast.ToFloat64E(val)
		if err == nil {
			return toIntFloat64, nil
		}
	}

	return val, nil
}

func resolveEnvFromBytes(bytes []byte) ([]byte, error) {
	toString := string(bytes)
	matches := envRegex.FindAllStringSubmatch(toString, -1)
	for _, submatches := range matches {
		envKey := submatches[1]

		value, isset := os.LookupEnv(envKey)

		if !isset {
			return nil, fmt.Errorf("environment variable \"%s\" is not set", envKey)
		}
		toString = strings.Replace(toString, submatches[0], value, -1)
	}
	bytes = []byte(toString)

	return bytes, nil
}