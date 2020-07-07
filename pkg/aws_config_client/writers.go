package aws_config_client

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
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
	awsConfigPath string
	buffer        *bytes.Buffer
}

func NewAWSConfigFileWriter(awsConfigPath string) *AWSConfigFile {
	return &AWSConfigFile{
		buffer:        bytes.NewBuffer(nil),
		awsConfigPath: awsConfigPath,
	}
}

func (a *AWSConfigFile) Finalize() error {
	tmpfile, err := ioutil.TempFile("", "")
	if err != nil {
		return errors.Wrap(err, "could not create temporary file for aws config")
	}

	defer func() {
		// if we have an error, attempt cleanup best we can
		if err != nil {
			tmpfile.Close()
			os.Remove(tmpfile.Name())
		}
	}()

	_, err = io.Copy(tmpfile, a.buffer)
	if err != nil {
		return errors.Wrap(err, "could not write contents to temporary aws config")
	}

	err = tmpfile.Sync()
	if err != nil {
		return errors.Wrap(err, "could not Sync temporary aws credentials file")
	}

	err = tmpfile.Close()
	if err != nil {
		return errors.Wrap(err, "could not close temporary aws credentials file")
	}

	err = os.Rename(tmpfile.Name(), a.awsConfigPath)
	return errors.Wrapf(err, "could not move aws config to %s", err)
}

func (a *AWSConfigFile) Write(p []byte) (int, error) {
	return a.buffer.Write(p)
}

// mergeConfig will merge the new config with the existing aws config
//             giving precedence to the new aws config blocks
func (a *AWSConfigFile) MergeAWSConfigs(new *ini.File, old *ini.File) (*ini.File, error) {
	return mergeAWSConfigs(new, old)
}

type AWSConfigSTDOUTWriter struct {
	preamble sync.Once
}

// stdout writer only returns the new config
//        up to users to figure out how to merge
func (a *AWSConfigSTDOUTWriter) MergeAWSConfigs(new *ini.File, old *ini.File) (*ini.File, error) {
	return new, nil
}

func (a *AWSConfigSTDOUTWriter) Write(p []byte) (int, error) {
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
