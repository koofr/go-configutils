package configutils_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"testing"
)

func TestConfigutils(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Configutils Suite")
}
