package util

import (
	"fmt"
	"os"
	"strings"
)

// https://golang.org/src/os/env_test.go
func ResetEnv(origEnv []string) error {
	for _, pair := range origEnv {
		// Environment variables on Windows can begin with =
		// https://blogs.msdn.com/b/oldnewthing/archive/2010/05/06/10008132.aspx
		i := strings.Index(pair[1:], "=") + 1
		if err := os.Setenv(pair[:i], pair[i+1:]); err != nil {
			return fmt.Errorf("Setenv(%q, %q) failed during reset: %v", pair[:i], pair[i+1:], err)
		}
	}
	return nil
}
