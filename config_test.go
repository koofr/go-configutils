package configutils_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/koofr/go-configutils"
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

const TestConfigUnknownKey = `
key: "value"
do: true
pi: 3.14
section:
  sectionkey: "sectionvalue"
unknown: "value"
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

	It("should validate YAML keys", func() {
		configFile := filepath.Join(tmp, "config.yaml")
		writeConfig(configFile, TestConfigUnknownKey)

		cfg := &Config{}

		err := LoadConfig(configFile, cfg)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("field unknown not found"))
	})

	It("should ignore unknown YAML keys", func() {
		configFile := filepath.Join(tmp, "config.yaml")
		writeConfig(configFile, TestConfigUnknownKey)

		cfg := &Config{}

		err := LoadConfig(configFile, cfg, YAMLValidateKeys(false))
		Expect(err).NotTo(HaveOccurred())
	})

	It("should patch YAML file", func() {
		configFile := filepath.Join(tmp, "config.yaml")
		writeConfig(configFile, TestConfig)

		cfg := &Config{}

		err := LoadConfig(configFile, cfg, YAMLPatchBytes(func(b []byte) []byte {
			return bytes.ReplaceAll(b, []byte(`key: "value"`), []byte(`key: "patchedvalue"`))
		}))
		Expect(err).NotTo(HaveOccurred())

		Expect(cfg.Key).To(Equal("patchedvalue"))
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
		Expect(err.Error()).To(Equal("LoadConfig error: envigo error: envigo PI parse float error: strconv.ParseFloat: parsing \"3,14\": invalid syntax"))
	})
})

var _ = Describe("LoadConfigBytes", func() {
	It("should load config", func() {
		cfg := &Config{}

		err := LoadConfigBytes([]byte(TestConfig), cfg)
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

	It("should validate YAML keys", func() {
		cfg := &Config{}

		err := LoadConfigBytes([]byte(TestConfigUnknownKey), cfg)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("field unknown not found"))
	})

	It("should ignore unknown YAML keys", func() {
		cfg := &Config{}

		err := LoadConfigBytes([]byte(TestConfigUnknownKey), cfg, YAMLValidateKeys(false))
		Expect(err).NotTo(HaveOccurred())
	})

	It("should patch YAML file", func() {
		cfg := &Config{}

		err := LoadConfigBytes([]byte(TestConfig), cfg, YAMLPatchBytes(func(b []byte) []byte {
			return bytes.ReplaceAll(b, []byte(`key: "value"`), []byte(`key: "patchedvalue"`))
		}))
		Expect(err).NotTo(HaveOccurred())

		Expect(cfg.Key).To(Equal("patchedvalue"))
	})

	It("should load config with env override", func() {
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

		err := LoadConfigBytes([]byte(TestConfig), cfg, EnvGetter(getenv))
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

		err := LoadConfigBytes([]byte(TestConfig), cfg, EnvGetter(getenv), EnvPrefix("MYAPP"))
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

		err := LoadConfigBytes([]byte(TestConfig), cfg, EnvGetter(getenv), DisableEnvOverride())
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

	It("should load config with bytes override", func() {
		cfg := &Config{}

		err := LoadConfigBytes([]byte(TestConfig), cfg, OverrideConfigBytes([]byte(OverrideConfig1)), OverrideConfigBytes([]byte(OverrideConfig2)))
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

	It("should not load invalid config", func() {
		cfg := &Config{}

		err := LoadConfigBytes([]byte(InvalidConfig), cfg)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("yaml: unmarshal errors"))
	})

	It("should not load invalid override config", func() {
		cfg := &Config{}

		err := LoadConfigBytes([]byte(TestConfig), cfg, OverrideConfigBytes([]byte(InvalidConfig)))
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("yaml: unmarshal errors"))
	})

	It("should not load config with error in env variable", func() {
		getenv := func(key string) (string, bool) {
			switch key {
			case "PI":
				return "3,14", true
			default:
				return "", false
			}
		}

		cfg := &Config{}

		err := LoadConfigBytes([]byte(TestConfig), cfg, EnvGetter(getenv))
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("LoadConfigBytes error: envigo error: envigo PI parse float error: strconv.ParseFloat: parsing \"3,14\": invalid syntax"))
	})
})
