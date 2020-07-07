package aws_config_client

import (
	"bytes"
	"fmt"
	"os"
	"sync"

	"github.com/pkg/errors"
	"gopkg.in/ini.v1"
)

type AWSConfigWriter interface {
	MergeAWSConfigs(*ini.File, *ini.File) (*ini.File, error)
	Write([]byte) (int, error)
}

type AWSConfigFile struct {
	f *os.File
}

func (a *AWSConfigFile) Close() error {
	return a.f.Close()
}
func (a *AWSConfigFile) Open(awsConfigPath string) error {
	f, err := os.OpenFile(awsConfigPath, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return errors.Wrapf(err, "could not open %s", awsConfigPath)
	}
	a.f = f
	return nil
}
func (a *AWSConfigFile) Write(p []byte) (int, error) {
	return a.f.Write(p)
}

// mergeConfig will merge the new config with the existing aws config
//             giving precedence to the new aws config blocks
func (a *AWSConfigFile) MergeAWSConfigs(new *ini.File, old *ini.File) (*ini.File, error) {
	return mergeAWSConfigs(new, old)
}

type AWSConfigSTDOUT struct {
	preamble sync.Once
}

// stdout writer only returns the new config
//        up to users to figure out how to merge
func (a *AWSConfigSTDOUT) MergeAWSConfigs(new *ini.File, old *ini.File) (*ini.File, error) {
	return new, nil
}

func (a *AWSConfigSTDOUT) Write(p []byte) (int, error) {
	w := os.Stdout
	// we only write the preamble once
	// a bit Hacky, but ok since just stdout
	a.preamble.Do(func() {
		fmt.Fprintf(w, "Please add the following to your AWS Config:\n")
	})
	return w.Write(p)
}

func mergeAWSConfigs(new *ini.File, old *ini.File) (*ini.File, error) {
	baseBytes := bytes.NewBuffer(nil)
	newAWSProfileBytes := bytes.NewBuffer(nil)
	_, err := new.WriteTo(newAWSProfileBytes)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to write AWS Profiles")
	}
	_, err = old.WriteTo(baseBytes)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to incorporate original AWS config file with new config changes")
	}

	mergedConfig, err := ini.Load(baseBytes, newAWSProfileBytes)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to merge old and new config files")
	}
	return mergedConfig, nil
}
