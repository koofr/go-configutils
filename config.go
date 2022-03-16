package configutils

import (
	"bytes"
	"fmt"
	"io/ioutil"

	"github.com/koofr/envigo"
	yaml "gopkg.in/yaml.v3"
)

type LoadConfigOptions struct {
	EnvOverride         bool
	EnvPrefix           string
	EnvGetter           envigo.EnvGetter
	OverrideConfigFiles []string
	YAMLValidateKeys    bool
	YAMLPatchBytes      func(b []byte) []byte
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
		if configFile != "" {
			opts.OverrideConfigFiles = append(opts.OverrideConfigFiles, configFile)
		}
	}
}

func YAMLValidateKeys(validate bool) func(*LoadConfigOptions) {
	return func(opts *LoadConfigOptions) {
		opts.YAMLValidateKeys = validate
	}
}

func YAMLPatchBytes(patch func(b []byte) []byte) func(*LoadConfigOptions) {
	return func(opts *LoadConfigOptions) {
		opts.YAMLPatchBytes = patch
	}
}

func LoadConfigFile(configFile string, config interface{}) (err error) {
	configBytes, err := ioutil.ReadFile(configFile)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(configBytes, config)
}

func loadConfigFileOpts(configFile string, config interface{}, opts *LoadConfigOptions) (err error) {
	configBytes, err := ioutil.ReadFile(configFile)
	if err != nil {
		return err
	}

	if opts.YAMLPatchBytes != nil {
		configBytes = opts.YAMLPatchBytes(configBytes)
	}

	decoder := yaml.NewDecoder(bytes.NewReader(configBytes))

	if opts.YAMLValidateKeys {
		decoder.KnownFields(true)
	}

	return decoder.Decode(config)
}

func LoadConfig(configFile string, config interface{}, optFuncs ...func(*LoadConfigOptions)) (err error) {
	opts := &LoadConfigOptions{
		EnvOverride:         true,
		EnvPrefix:           "",
		EnvGetter:           envigo.EnvironGetter(),
		OverrideConfigFiles: []string{},
		YAMLValidateKeys:    true,
		YAMLPatchBytes:      nil,
	}

	for _, optFunc := range optFuncs {
		optFunc(opts)
	}

	if err = loadConfigFileOpts(configFile, config, opts); err != nil {
		return fmt.Errorf("LoadConfig error: %w", err)
	}

	for _, overrideConfigFile := range opts.OverrideConfigFiles {
		if err = loadConfigFileOpts(overrideConfigFile, config, opts); err != nil {
			return fmt.Errorf("LoadConfig override error: %w", err)
		}
	}

	if opts.EnvOverride {
		if err = envigo.Envigo(config, opts.EnvPrefix, opts.EnvGetter); err != nil {
			return fmt.Errorf("LoadConfig envigo error: %w", err)
		}
	}

	return nil
}
