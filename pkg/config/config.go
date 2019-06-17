package config

import (
	"regexp"

	"github.com/mitchellh/mapstructure"
	"github.com/sherifabdlnaby/objx"
	"gopkg.in/go-playground/validator.v9"
)

// TODO evaluate defaults. (add ability to add default base)
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

// Populate will populate 'dst' struct with the config field in the YAML configuration,
// structs can use two tags, `mapstructure` to map the config to
// the struct look up (github.com/mitchellh/mapstructure) docsfor more about its tags,
// and `validate` tag for quick validation lookup (gopkg.in/go-playground/validator.v9) validate tags
// for more about its tags.
func (cw *Config) Populate(dst interface{}) error {

	config := &mapstructure.DecoderConfig{
		WeaklyTypedInput: true,
		Result:           dst,
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
	err = validator.New().Struct(dst)

	return err
}

//NewSelector Returns a new Selector used to evaluate dynamic fields in a config
// (this receiver was made for easing refactoring)
func (cw *Config) NewSelector(base string) (Selector, error) {
	return NewSelector(base)
}
