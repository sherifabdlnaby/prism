package config

import "os"

var (
	// ConfigDirEnv is Environment Variable that points to the configuration directory (default: ./config)
	PRISM_CONFIG_DIR = Env{
		env:  "PRISM_CONFIG_DIR",
		def:  "./config",
		desc: "Path for configuration files' directory",
	}

	// ConfigDirEnv is Environment Variable that points to the configuration directory (default: ./config)
	PRISM_DATA_DIR = Env{
		env:  "PRISM_DATA_DIR",
		def:  "./data",
		desc: "Path for persistence db files' directory",
	}

	// ConfigDirEnv is Environment Variable that points to the configuration directory (default: ./config)
	PRISM_TMP_DIR = Env{
		env:  "PRISM_TMP_DIR",
		def:  "./data/images",
		desc: "Path for configuration files' directory",
	}

	// Env is Environment Variable that represents either prod or dev environment
	PRISM_ENV = Env{
		env:  "PRISM_ENV",
		def:  "prod",
		desc: "Environment that is either Development(dev) or Production(prod)",
	}
)

type Env struct {
	env, def, desc string
}

func (e *Env) Lookup() string {
	environment, isset := os.LookupEnv(e.env)
	if !isset {
		return e.def
	}
	return environment
}

func (e *Env) Name() string {
	environment, isset := os.LookupEnv(e.env)
	if !isset {
		return e.def
	}
	return environment
}
