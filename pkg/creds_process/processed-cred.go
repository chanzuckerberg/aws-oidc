package cred

import (
	"encoding/base64"
	"encoding/json"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// Used to store the parsed results from running assumeRole
type ProcessedCred struct {
	Version         int       `json:"Version"`
	AccessKeyID     string    `json:"AccessKeyId"`
	SecretAccessKey string    `json:"SecretAccessKey"`
	SessionToken    string    `json:"SessionToken"`
	Expiration      string    `json:"Expiration"`
	CacheExpiry     time.Time `json:"CacheExpiration"`
}

func (pc *ProcessedCred) IsFresh() bool {
	if pc == nil {
		return false
	}
	return pc.CacheExpiry.After(time.Now().Add(timeSkew))
}

const (
	timeSkew             = 5 * time.Minute
	ProcessedCredVersion = 1
)

func CredFromString(credString *string, opts ...MarshalOpts) (*ProcessedCred, error) {
	if credString == nil {
		logrus.Debug("nil cred string")
		return nil, nil
	}
	credBytes, err := base64.StdEncoding.DecodeString(*credString)
	if err != nil {
		return nil, errors.Wrap(err, "error b64 decoding token")
	}
	pc := &ProcessedCred{
		Version: ProcessedCredVersion,
	}
	err = json.Unmarshal(credBytes, pc)
	if err != nil {
		return nil, errors.Wrap(err, "could not json unmarshal cred")
	}

	for _, opt := range opts {
		opt(pc)
	}
	return pc, nil
}

func (pc *ProcessedCred) Marshal(opts ...MarshalOpts) (string, error) {
	if pc == nil {
		return "", errors.New("error Marshalling nil token")
	}

	// apply any processing to the token
	for _, opt := range opts {
		opt(pc)
	}

	credBytes, err := json.Marshal(pc)
	if err != nil {
		return "", errors.Wrap(err, "could not marshal token")
	}

	b64 := base64.StdEncoding.EncodeToString(credBytes)
	return b64, nil
}

// MarshalOpts changes a token for marshaling
type MarshalOpts func(*ProcessedCred)
