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
	"gopkg.in/ini.v1"
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

func TestMergeAWSConfigs(t *testing.T) {
	r := require.New(t)

	old := `
[profile source]
region = us-west-2
output = json


[profile foo]
role_arn           = arn:aws:iam::01234567890:role/foo
source_profile     = czi-id
region             = us-west-2
output             = json
credential_process = aws-oidc creds-process --issuer-url=foo --client-id=foo --aws-role-arn=arn:aws:iam::01234567890:role/foo
`

	new := `
[profile foo]
region             = us-west-2
output             = json
credential_process = aws-oidc creds-process --issuer-url=foo --client-id=foo --aws-role-arn=arn:aws:iam::01234567890:role/foo
`

	expected := `[profile source]
region = us-west-2
output = json

[profile foo]
region             = us-west-2
output             = json
credential_process = aws-oidc creds-process --issuer-url=foo --client-id=foo --aws-role-arn=arn:aws:iam::01234567890:role/foo
`

	oldINI, err := ini.Load([]byte(old))
	r.NoError(err)

	newINI, err := ini.Load([]byte(new))
	r.NoError(err)

	resultINI, err := mergeAWSConfigs(newINI, oldINI)
	r.NoError(err)

	result := bytes.NewBuffer(nil)
	_, err = resultINI.WriteTo(result)
	r.NoError(err)

	r.Equal(expected, result.String())
}
