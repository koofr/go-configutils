package configutils_test

import (
	. "github.com/koofr/go-configutils"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"os"
)

type Config struct {
	Key string
}

var _ = Describe("LoadConfig", func() {
	It("should load config from YAML file", func() {
		dir, err := ioutil.TempDir(os.TempDir(), "configutils-")
		Expect(err).NotTo(HaveOccurred())

		configFile := dir + "/config.yaml"

		err = ioutil.WriteFile(configFile, []byte(`key: "value"`), 0600)
		Expect(err).NotTo(HaveOccurred())

		cfg := &Config{}

		err = LoadConfig(configFile, cfg)
		Expect(err).NotTo(HaveOccurred())

		Expect(cfg.Key).To(Equal("value"))
	})
})
