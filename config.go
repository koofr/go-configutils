package configutils

import (
	"fmt"
	"github.com/koofr/envigo"
	"gopkg.in/yaml.v1"
	"io/ioutil"
)

func LoadConfigFile(configFile string, config interface{}) (err error) {
	configBytes, err := ioutil.ReadFile(configFile)

	if err != nil {
		return
	}

	err = yaml.Unmarshal(configBytes, config)

	return
}

func LoadConfig(configFile string, config interface{}) (err error) {
	err = LoadConfigFile(configFile, config)

	if err != nil {
		err = fmt.Errorf("LoadConfig error: %s", err)
		return
	}

	err = envigo.Envigo(config, "", envigo.EnvironGetter())

	if err != nil {
		err = fmt.Errorf("LoadConfig envigo error: %s", err)
		return
	}

	return
}
