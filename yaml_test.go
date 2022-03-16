package configutils_test

import (
	. "github.com/koofr/go-configutils"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("YAML", func() {
	Describe("YAMLRemoveRootKeys", func() {
		It("should remove keys and keep the lines", func() {
			Expect(string(YAMLRemoveRootKeys([]byte(`
key1:
  key11: 11

key2: true

key3:
  key31: 31

key4:
  key41: 41
`), "key2", "key3", "key3x"))).To(Equal(`
key1:
  key11: 11






key4:
  key41: 41
`))
		})
	})

	Describe("YAMLKeepRootKeys", func() {
		It("should keep keys and keep the lines", func() {
			Expect(string(YAMLKeepRootKeys([]byte(`
key1:
  key11: 11

key2: true

key3:
  key31: 31

key4:
  key41: 41
`), "key2", "key3", "key3x"))).To(Equal(`



key2: true

key3:
  key31: 31



`))
		})
	})
})
