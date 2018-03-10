package configutils

import (
	"fmt"
	"io/ioutil"

	"github.com/koofr/envigo"
	yaml "gopkg.in/yaml.v2"
)

type LoadConfigOptions struct {
	EnvOverride         bool
	EnvPrefix           string
	EnvGetter           envigo.EnvGetter
	OverrideConfigFiles []string
}

func DisableEnvOverride() func(*LoadConfigOptions) {
	return func(opts *LoadConfigOptions) {
		opts.EnvOverride = false
	}
}

func EnvPrefix(envPrefix string) func(*LoadConfigOptions) {
	return func(opts *LoadConfigOptions) {
		opts.EnvPrefix = envPrefix
	}
}

func EnvGetter(envGetter envigo.EnvGetter) func(*LoadConfigOptions) {
	return func(opts *LoadConfigOptions) {
		opts.EnvGetter = envGetter
	}
}

func OverrideConfigFile(configFile string) func(*LoadConfigOptions) {
	return func(opts *LoadConfigOptions) {
		opts.OverrideConfigFiles = append(opts.OverrideConfigFiles, configFile)
	}
}

func LoadConfigFile(configFile string, config interface{}) (err error) {
	configBytes, err := ioutil.ReadFile(configFile)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(configBytes, config)
}

func LoadConfig(configFile string, config interface{}, optFuncs ...func(*LoadConfigOptions)) (err error) {
	opts := &LoadConfigOptions{
		EnvOverride:         true,
		EnvPrefix:           "",
		EnvGetter:           envigo.EnvironGetter(),
		OverrideConfigFiles: []string{},
	}

	for _, optFunc := range optFuncs {
		optFunc(opts)
	}

	if err = LoadConfigFile(configFile, config); err != nil {
		return fmt.Errorf("LoadConfig error: %s", err)
	}

	for _, overrideConfigFile := range opts.OverrideConfigFiles {
		if err = LoadConfigFile(overrideConfigFile, config); err != nil {
			return fmt.Errorf("LoadConfig override error: %s", err)
		}
	}

	if opts.EnvOverride {
		if err = envigo.Envigo(config, opts.EnvPrefix, opts.EnvGetter); err != nil {
			return fmt.Errorf("LoadConfig envigo error: %s", err)
		}
	}

	return nil
}
