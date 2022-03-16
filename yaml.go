package configutils

import (
	"strings"
	"unicode"
)

func YAMLRemoveRootKeys(b []byte, keys ...string) []byte {
	return yamlFilterRootKeys(b, keys, false)
}

func YAMLKeepRootKeys(b []byte, keys ...string) []byte {
	return yamlFilterRootKeys(b, keys, true)
}

func yamlFilterRootKeys(b []byte, keys []string, reverse bool) []byte {
	lines := strings.Split(string(b), "\n")

	isIgnoring := false

	for i, line := range lines {
		isEmpty := len(strings.TrimSpace(line)) == 0
		isRoot := strings.TrimLeftFunc(line, unicode.IsSpace) == line && !isEmpty

		if isRoot {
			isIgnoring = false

			for _, key := range keys {
				if strings.HasPrefix(line, key+":") {
					isIgnoring = true
					break
				}
			}
		}

		if (isIgnoring && !reverse) || (!isIgnoring && reverse) {
			lines[i] = ""
		}
	}

	return []byte(strings.Join(lines, "\n"))
}
