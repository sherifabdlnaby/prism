package config

import (
	"fmt"
	"github.com/mitchellh/mapstructure"
	"github.com/sherifabdlnaby/prism/pkg/types"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
)

type Plugin struct {
	Plugin string       `yaml:"plugin"`
	Number int          `yaml:"number"`
	Config types.Config `yaml:"config"`
}

type Input struct {
	Plugin `yaml:",inline" mapstructure:",squash"`
}

type Processor struct {
	Plugin `yaml:",inline" mapstructure:",squash"`
}

type Output struct {
	Plugin `yaml:",inline"`
}

type InputsConfig struct {
	Inputs map[string]Input `yaml:"inputs"`
}

type ProcessorsConfig struct {
	Processors map[string]Input `yaml:"processors"`
}

type OutputsConfig struct {
	Outputs map[string]Input `yaml:"outputs"`
}

var envRegex = regexp.MustCompile(`\${([\w@.]+)}`)

func Load(filePath string, out interface{}, resloveEnv bool) error {

	fileBytes, err := ioutil.ReadFile(filePath)

	if err != nil {
		return err
	}

	if resloveEnv {
		fileBytes = resolveEnvFromBytes(fileBytes)
	}

	return unmarshal(fileBytes, out)
}

func resolveEnvFromBytes(bytes []byte) []byte {
	toString := string(bytes)
	matches := envRegex.FindAllStringSubmatch(toString, -1)
	for _, submatches := range matches {
		envKey := submatches[1]

		value, isset := os.LookupEnv(envKey)

		if isset {
			toString = strings.Replace(toString, submatches[0], value, -1)
		}
	}
	bytes = []byte(toString)
	return bytes
}

func unmarshal(bytes []byte, out interface{}) error {
	mapRaw := make(map[interface{}]interface{})

	err := yaml.Unmarshal(bytes, mapRaw)

	mapString := recursivelyTurnYAMLMaps(mapRaw)

	config := &mapstructure.DecoderConfig{
		WeaklyTypedInput: true,
		Result:           &out,
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
