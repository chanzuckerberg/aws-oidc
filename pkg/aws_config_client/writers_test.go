package aws_config_client

import (
	"bytes"
	"crypto/rand"
	"io"
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAWSConfigFileWriter(t *testing.T) {
	r := require.New(t)

	dir, err := ioutil.TempDir("", "")
	r.NoError(err)
	defer os.RemoveAll(dir) //cleanup

	awsConfigPath := path.Join(dir, "config")

	writer := NewAWSConfigFileWriter(awsConfigPath)

	expectedData := make([]byte, 256)
	_, err = rand.Read(expectedData)
	r.NoError(err)

	_, err = io.Copy(writer, bytes.NewBuffer(expectedData))
	r.NoError(err)

	err = writer.Finalize()
	r.NoError(err)

	readData, err := ioutil.ReadFile(awsConfigPath)
	r.NoError(err)

	r.Equal(expectedData, readData)
}
