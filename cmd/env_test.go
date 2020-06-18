package cmd

import (
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func TestResolveProfile(t *testing.T) {
	r := require.New(t)

	// https://golang.org/src/os/env_test.go
	defer func(origEnv []string) {
		for _, pair := range origEnv {
			// Environment variables on Windows can begin with =
			// https://blogs.msdn.com/b/oldnewthing/archive/2010/05/06/10008132.aspx
			i := strings.Index(pair[1:], "=") + 1
			if err := os.Setenv(pair[:i], pair[i+1:]); err != nil {
				t.Errorf("Setenv(%q, %q) failed during reset: %v", pair[:i], pair[i+1:], err)
			}
		}
	}(os.Environ())
	r.NoError(os.Unsetenv(envAWSProfile))

	// default
	r.Equal(defaultAWSProfile, resolveProfile(nil))

	// from env
	expectedProfile := "asdfasdfalkwq;e"
	os.Setenv(envAWSProfile, expectedProfile)
	r.Equal(expectedProfile, resolveProfile(nil))

	// flag
	expectedProfile = "flag-profile"
	cmd := &cobra.Command{}
	cmd.Flags().StringVar(
		&flagProfileName,
		flagProfile,
		"",
		"AWS Profile to fetch credentials from.")

	r.NoError(cmd.Flags().Set(
		flagProfile,
		expectedProfile))
	r.Equal(expectedProfile, resolveProfile(cmd))
}
