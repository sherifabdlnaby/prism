package config

import (
	"regexp"

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

func (cw *Config) NewSelector(base interface{}) (Selector, error) {
	return NewSelector(base)
}
