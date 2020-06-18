package aws_config_client

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
	prof, err := resolveProfile(nil)
	r.NoError(err)
	r.Equal(defaultAWSProfile, prof)

	// from env
	expectedProfile := "asdfasdfalkwq;e"
	os.Setenv(envAWSProfile, expectedProfile)
	prof, err = resolveProfile(nil)
	r.NoError(err)
	r.Equal(expectedProfile, prof)

	// flag
	var flagVal string

	expectedProfile = "flag-profile"
	cmd := &cobra.Command{}
	cmd.Flags().StringVar(
		&flagVal,
		FlagProfile,
		"",
		"AWS Profile to fetch credentials from.")

	r.NoError(cmd.Flags().Set(
		FlagProfile,
		expectedProfile))

	prof, err = resolveProfile(cmd)
	r.NoError(err)
	r.Equal(expectedProfile, prof)
}
