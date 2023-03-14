package configutils

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"reflect"

	"github.com/koofr/envigo"
	yaml "gopkg.in/yaml.v3"
)

type LoadConfigOptions struct {
	EnvOverride         bool
	EnvPrefix           string
	EnvGetter           envigo.EnvGetter
	OverrideConfigFiles []string
	OverrideConfigBytes [][]byte
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

func OverrideConfigBytes(configBytes []byte) func(*LoadConfigOptions) {
	return func(opts *LoadConfigOptions) {
		opts.OverrideConfigBytes = append(opts.OverrideConfigBytes, configBytes)
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

	return loadConfigBytesOpts(configBytes, config, opts)
}

func loadConfigBytesOpts(configBytes []byte, config interface{}, opts *LoadConfigOptions) (err error) {
	if opts.YAMLPatchBytes != nil {
		configBytes = opts.YAMLPatchBytes(configBytes)
	}

	decoder := yaml.NewDecoder(bytes.NewReader(configBytes))

	if opts.YAMLValidateKeys {
		decoder.KnownFields(true)
	}

	return decoder.Decode(config)
}

func getOpts(optFuncs ...func(*LoadConfigOptions)) *LoadConfigOptions {
	opts := &LoadConfigOptions{
		EnvOverride:         true,
		EnvPrefix:           "",
		EnvGetter:           envigo.EnvironGetter(),
		OverrideConfigFiles: nil,
		OverrideConfigBytes: nil,
		YAMLValidateKeys:    true,
		YAMLPatchBytes:      nil,
	}

	for _, optFunc := range optFuncs {
		optFunc(opts)
	}

	return opts
}

func applyOpts(config interface{}, opts *LoadConfigOptions) error {
	for _, overrideConfigFile := range opts.OverrideConfigFiles {
		if err := loadConfigFileOpts(overrideConfigFile, config, opts); err != nil {
			return fmt.Errorf("override error: %w", err)
		}
	}

	for _, overrideConfigBytes := range opts.OverrideConfigBytes {
		if err := loadConfigBytesOpts(overrideConfigBytes, config, opts); err != nil {
			return fmt.Errorf("override error: %w", err)
		}
	}

	if opts.EnvOverride {
		if err := envigo.Envigo(config, opts.EnvPrefix, opts.EnvGetter); err != nil {
			return fmt.Errorf("envigo error: %w", err)
		}
	}

	return nil
}

func LoadConfig(configFile string, config interface{}, optFuncs ...func(*LoadConfigOptions)) (err error) {
	opts := getOpts(optFuncs...)

	for reflect.ValueOf(config).Kind() == reflect.Ptr && reflect.ValueOf(config).Elem().Kind() == reflect.Ptr {
		config = reflect.ValueOf(config).Elem().Interface()
	}

	if err := loadConfigFileOpts(configFile, config, opts); err != nil {
		return fmt.Errorf("LoadConfig error: %w", err)
	}

	if err := applyOpts(config, opts); err != nil {
		return fmt.Errorf("LoadConfig error: %w", err)
	}

	return nil
}

func LoadConfigBytes(configFileBytes []byte, config interface{}, optFuncs ...func(*LoadConfigOptions)) (err error) {
	opts := getOpts(optFuncs...)

	for reflect.ValueOf(config).Kind() == reflect.Ptr && reflect.ValueOf(config).Elem().Kind() == reflect.Ptr {
		config = reflect.ValueOf(config).Elem().Interface()
	}

	if err = loadConfigBytesOpts(configFileBytes, config, opts); err != nil {
		return fmt.Errorf("LoadConfigBytes error: %w", err)
	}

	if err := applyOpts(config, opts); err != nil {
		return fmt.Errorf("LoadConfigBytes error: %w", err)
	}

	return nil
}
