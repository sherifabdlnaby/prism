package config

import "os"

var (
	// EnvPrismConfigDir is Environment Variable that points to the configuration directory (default: ./config)
	EnvPrismConfigDir = Env{
		env:  "PRISM_CONFIG_DIR",
		def:  "./config",
		desc: "Path for configuration files' directory",
	}

	// EnvPrismDataDir is Environment Variable that points to the configuration directory (default: ./config)
	EnvPrismDataDir = Env{
		env:  "PRISM_DATA_DIR",
		def:  "./data",
		desc: "Path for persistence db files' directory",
	}

	// EnvPrism is Environment Variable that represents either prod or dev environment
	EnvPrism = Env{
		env:  "PRISM_ENV",
		def:  "prod",
		desc: "Environment that is either Development(dev) or Production(prod)",
	}
)

//Env An Environment Variable Config
type Env struct {
	env, def, desc string
}

//Lookup Get the environment variable and evaluate it
func (e *Env) Lookup() string {
	environment, isset := os.LookupEnv(e.env)
	if !isset {
		return e.def
	}
	return environment
}

//Name Get the NAME of the environment variable
func (e *Env) Name() string {
	environment, isset := os.LookupEnv(e.env)
	if !isset {
		return e.def
	}
	return environment
}
