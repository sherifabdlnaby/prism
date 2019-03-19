package config

import (
	"github.com/mitchellh/mapstructure"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"reflect"
	"regexp"
	"strings"
)

type Plugin struct {
	Plugin string                 `yaml:"plugin"`
	Number int                    `yaml:"number"`
	Config map[string]interface{} `yaml:"config"`
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
	Inputs []map[string]Input `yaml:"inputs"`
}

type ProcessorsConfig struct {
	Processors []map[string]Input `yaml:"processors"`
}

type OutputsConfig struct {
	Outputs []map[string]Input `yaml:"outputs"`
}

//type Node struct {
//
//}
//
//type Pipeline struct {
//
//}

var envRegex = regexp.MustCompile(`\${([\w@.]+)}`)

func Load(filePath string, out interface{}, resloveEnv bool) error {

	fileBytes, err := ioutil.ReadFile(filePath)

	if err != nil {
		return err
	}

	mapRaw := make(map[string]interface{})

	err = yaml.Unmarshal(fileBytes, mapRaw)

	config := &mapstructure.DecoderConfig{
		WeaklyTypedInput: true,
		Result:           &out,
	}

	if resloveEnv {
		config.DecodeHook = ResolveEnvCallback
	}

	decoder, err := mapstructure.NewDecoder(config)

	if err != nil {
		return err
	}

	err = decoder.Decode(mapRaw)

	if err != nil {
		return err
	}

	return err
}

func ResolveEnvCallback(in reflect.Kind, _ reflect.Kind, val interface{}) (interface{}, error) {

	if in != reflect.String {
		return val, nil
	}

	stringVal := val.(string)

	matches := envRegex.FindAllStringSubmatch(stringVal, -1)

	for _, submatches := range matches {
		envKey := submatches[1]

		value, isset := os.LookupEnv(envKey)

		if isset {
			stringVal = strings.Replace(stringVal, submatches[0], value, -1)
		}
	}

	return stringVal, nil
}
