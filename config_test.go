package configutils_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	. "github.com/koofr/go-configutils"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type Section struct {
	SectionKey string
}

type Config struct {
	Key     string
	Do      bool
	Pi      float64
	Section *Section
}

const TestConfig = `
key: "value"
do: true
pi: 3.14
section:
  sectionkey: "sectionvalue"
`

const OverrideConfig1 = `
do: false
`

const OverrideConfig2 = `
section:
  sectionkey: sectionvalueoverride
`

const InvalidConfig = `
key
`

func writeConfig(configFile string, content string) {
	err := ioutil.WriteFile(configFile, []byte(content), 0600)
	Expect(err).NotTo(HaveOccurred())
}

var _ = Describe("LoadConfig", func() {
	var tmp string

	BeforeEach(func() {
		var err error
		tmp, err = ioutil.TempDir(os.TempDir(), "configutils-")
		Expect(err).NotTo(HaveOccurred())
	})

	It("should load config", func() {
		configFile := filepath.Join(tmp, "config.yaml")
		writeConfig(configFile, TestConfig)

		cfg := &Config{}

		err := LoadConfig(configFile, cfg)
		Expect(err).NotTo(HaveOccurred())

		Expect(cfg).To(Equal(&Config{
			Key: "value",
			Do:  true,
			Pi:  3.14,
			Section: &Section{
				SectionKey: "sectionvalue",
			},
		}))
	})

	It("should load config with env override", func() {
		configFile := filepath.Join(tmp, "config.yaml")
		writeConfig(configFile, TestConfig)

		getenv := func(key string) (string, bool) {
			switch key {
			case "DO":
				return "false", true
			case "SECTION_SECTIONKEY":
				return "sectionvalueoverride", true
			default:
				return "", false
			}
		}

		cfg := &Config{}

		err := LoadConfig(configFile, cfg, EnvGetter(getenv))
		Expect(err).NotTo(HaveOccurred())

		Expect(cfg).To(Equal(&Config{
			Key: "value",
			Do:  false,
			Pi:  3.14,
			Section: &Section{
				SectionKey: "sectionvalueoverride",
			},
		}))
	})

	It("should load config with env override and prefix", func() {
		configFile := filepath.Join(tmp, "config.yaml")
		writeConfig(configFile, TestConfig)

		getenv := func(key string) (string, bool) {
			switch key {
			case "MYAPP_DO":
				return "false", true
			case "MYAPP_SECTION_SECTIONKEY":
				return "sectionvalueoverride", true
			default:
				return "", false
			}
		}

		cfg := &Config{}

		err := LoadConfig(configFile, cfg, EnvGetter(getenv), EnvPrefix("MYAPP"))
		Expect(err).NotTo(HaveOccurred())

		Expect(cfg).To(Equal(&Config{
			Key: "value",
			Do:  false,
			Pi:  3.14,
			Section: &Section{
				SectionKey: "sectionvalueoverride",
			},
		}))
	})

	It("should load config without env override", func() {
		configFile := filepath.Join(tmp, "config.yaml")
		writeConfig(configFile, TestConfig)

		getenv := func(key string) (string, bool) {
			switch key {
			case "DO":
				return "false", true
			case "SECTION_SECTIONKEY":
				return "sectionvalueoverride", true
			default:
				return "", false
			}
		}

		cfg := &Config{}

		err := LoadConfig(configFile, cfg, EnvGetter(getenv), DisableEnvOverride())
		Expect(err).NotTo(HaveOccurred())

		Expect(cfg).To(Equal(&Config{
			Key: "value",
			Do:  true,
			Pi:  3.14,
			Section: &Section{
				SectionKey: "sectionvalue",
			},
		}))
	})

	It("should load config with file override", func() {
		configFile := filepath.Join(tmp, "config.yaml")
		writeConfig(configFile, TestConfig)

		overrideConfigFile1 := filepath.Join(tmp, "override1.yaml")
		writeConfig(overrideConfigFile1, OverrideConfig1)

		overrideConfigFile2 := filepath.Join(tmp, "override2.yaml")
		writeConfig(overrideConfigFile2, OverrideConfig2)

		cfg := &Config{}

		err := LoadConfig(configFile, cfg, OverrideConfigFile(overrideConfigFile1), OverrideConfigFile(overrideConfigFile2))
		Expect(err).NotTo(HaveOccurred())

		Expect(cfg).To(Equal(&Config{
			Key: "value",
			Do:  false,
			Pi:  3.14,
			Section: &Section{
				SectionKey: "sectionvalueoverride",
			},
		}))
	})

	It("should not load non-existent config", func() {
		configFile := filepath.Join(tmp, "config.yaml")

		cfg := &Config{}

		err := LoadConfig(configFile, cfg)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal(fmt.Sprintf("LoadConfig error: open %s: no such file or directory", configFile)))
	})

	It("should not load invalid config", func() {
		configFile := filepath.Join(tmp, "invalidconfig.yaml")
		writeConfig(configFile, InvalidConfig)

		cfg := &Config{}

		err := LoadConfig(configFile, cfg)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("yaml: unmarshal errors"))
	})

	It("should not load invalid override config", func() {
		configFile := filepath.Join(tmp, "config.yaml")
		writeConfig(configFile, TestConfig)

		overrideConfigFile := filepath.Join(tmp, "invalidconfig.yaml")
		writeConfig(overrideConfigFile, InvalidConfig)

		cfg := &Config{}

		err := LoadConfig(configFile, cfg, OverrideConfigFile(overrideConfigFile))
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("yaml: unmarshal errors"))
	})

	It("should not load config with error in env variable", func() {
		configFile := filepath.Join(tmp, "config.yaml")
		writeConfig(configFile, TestConfig)

		getenv := func(key string) (string, bool) {
			switch key {
			case "PI":
				return "3,14", true
			default:
				return "", false
			}
		}

		cfg := &Config{}

		err := LoadConfig(configFile, cfg, EnvGetter(getenv))
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("LoadConfig envigo error: envigo PI parse float error: strconv.ParseFloat: parsing \"3,14\": invalid syntax"))
	})
})
