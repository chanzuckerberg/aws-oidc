package aws_config_client

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"sync"

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
		return fmt.Errorf("could not create temporary file for aws config: %w", err)
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
		return fmt.Errorf("could not write contents to temporary aws config: %w", err)
	}

	err = tmpfile.Sync()
	if err != nil {
		return fmt.Errorf("could not Sync temporary aws credentials file: %w", err)
	}

	err = tmpfile.Close()
	if err != nil {
		return fmt.Errorf("could not close temporary aws credentials file: %w", err)
	}

	err = os.Rename(tmpfile.Name(), a.awsConfigPath)
	if err != nil {
		return fmt.Errorf("could not move aws config to %s: %w", a.awsConfigPath, err)
	}
	return nil
}

func (a *AWSConfigFile) Write(p []byte) (int, error) {
	return a.buffer.Write(p)
}

// mergeConfig will merge the new config with the existing aws config
//
//	giving precedence to the new aws config blocks
func (a *AWSConfigFile) MergeAWSConfigs(new *ini.File, old *ini.File) (*ini.File, error) {
	return mergeAWSConfigs(new, old)
}

type AWSConfigSTDOUTWriter struct {
	preamble sync.Once
}

// stdout writer only returns the new config
//
//	up to users to figure out how to merge
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
	// first, delete all overlapping sections
	for _, section := range new.Sections() {
		// skip over the default section
		if section.Name() == "DEFAULT" {
			continue
		}

		old.DeleteSection(section.Name())
	}

	baseBytes := bytes.NewBuffer(nil)
	newAWSProfileBytes := bytes.NewBuffer(nil)
	_, err := new.WriteTo(newAWSProfileBytes)
	if err != nil {
		return nil, fmt.Errorf("Unable to write AWS Profiles: %w", err)
	}
	_, err = old.WriteTo(baseBytes)
	if err != nil {
		return nil, fmt.Errorf("Unable to incorporate original AWS config file with new config changes: %w", err)
	}

	mergedConfig, err := ini.Load(baseBytes, newAWSProfileBytes)
	if err != nil {
		return nil, fmt.Errorf("Unable to merge old and new config files: %w", err)
	}
	return mergedConfig, nil
}
