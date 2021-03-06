package config

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"regexp"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cast"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

//TODO Rename

// WithDefaults interface is the interface for default values
type WithDefaults interface {
	ApplyDefault() error
}

// Config is the collection of config needed for the engine to start.
type Config struct {
	App        App
	Components Components
	Pipelines  Pipelines
	Logger     zap.SugaredLogger
}

// Component used for YAML decoding
type Component struct {
	Plugin      string                 `yaml:"plugin"`
	Concurrency int                    `yaml:"concurrency"`
	Config      map[string]interface{} `yaml:"config"`
}

// Input used for YAML decoding
type Input struct {
	Component `yaml:",inline" mapstructure:",squash"`
}

// Processor used for YAML decoding
type Processor struct {
	Component `yaml:",inline" mapstructure:",squash"`
}

// Output used for YAML decoding
type Output struct {
	Component `yaml:",inline" mapstructure:",squash"`
}

type Components struct {
	Inputs     Inputs
	Processors Processors
	Outputs    Outputs
}

// inputs used for YAML decoding
type Inputs struct {
	Inputs map[string]*Input `yaml:"inputs"`
}

// App used for YAML decoding
type App struct {
	Logger string `yaml:"logger"`
}

// Processors used for YAML decoding
type Processors struct {
	Processors map[string]*Processor `yaml:"processors"`
}

// outputs used for YAML decoding
type Outputs struct {
	Outputs map[string]*Output `yaml:"outputs"`
}

// Node is a single node in a pipeline
type Node struct {
	Async bool             `yaml:"async"`
	Next  map[string]*Node `yaml:",inline"`
}

// Pipelines used for YAML decoding
type Pipelines struct {
	Pipelines map[string]*Pipeline `yaml:"pipelines"`
	Hash      string
}

// Pipelines used for YAML decoding
type Pipeline struct {
	Concurrency int              `yaml:"concurrency"`
	Pipeline    map[string]*Node `yaml:"pipeline"`
}

// TODO support default values
var envRegex = regexp.MustCompile(`\${([\w@.]+)}`)

// Load loads a .yaml file into out. resolveEnv will replace ${ENV_VAR} with value of env variable "ENV_VAR"
func Load(filePath string, out WithDefaults, resolveEnv bool) (hash string, err error) {

	fileBytes, err := ioutil.ReadFile(filePath)

	if err != nil {
		return "", err
	}

	if resolveEnv {
		fileBytes, err = resolveEnvFromBytes(fileBytes)
		if err != nil {
			return "", err
		}
	}

	err = unmarshal(fileBytes, out)
	if err != nil {
		return "", err
	}

	err = out.ApplyDefault()
	if err != nil {
		return "", err
	}

	// calculate hash
	h := sha256.New()
	_, _ = h.Write(fileBytes)
	hash = hex.EncodeToString(h.Sum(nil))

	return hash, nil
}

func unmarshal(bytes []byte, out interface{}) error {
	mapRaw := make(map[interface{}]interface{})

	err := yaml.Unmarshal(bytes, &mapRaw)

	if err != nil {
		return err
	}

	mapString := RecursivelyTurnYAMLMaps(mapRaw)

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

// RecursivelyTurnYAMLMaps turns nested YAML maps from map[interface{}]interface{} to map[string]interface{}
// inspired by solutions in https://github.com/go-yaml/yaml/issues/139 especially elastic's #issuecomment-183937598
func RecursivelyTurnYAMLMaps(in interface{}) interface{} {
	switch in := in.(type) {
	case []interface{}:
		res := make([]interface{}, len(in))
		for i, v := range in {
			res[i] = RecursivelyTurnYAMLMaps(v)
		}
		return res
	case map[interface{}]interface{}:
		res := make(map[string]interface{})
		for k, v := range in {
			res[fmt.Sprintf("%v", k)] = RecursivelyTurnYAMLMaps(v)
		}
		return res
	case Component:
		res := make(map[string]interface{})
		for k, v := range in.Config {
			res[fmt.Sprintf("%v", k)] = RecursivelyTurnYAMLMaps(v)
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
			return nil, fmt.Errorf("environment variable [%s] is not set", envKey)
		}
		toString = strings.Replace(toString, submatches[0], value, -1)
	}
	bytes = []byte(toString)

	return bytes, nil
}
