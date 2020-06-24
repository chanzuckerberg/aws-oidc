package aws_config_server

import (
	"context"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/chanzuckerberg/aws-oidc/pkg/okta"
	cziAWS "github.com/chanzuckerberg/go-misc/aws"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type CachedGetClientIDToProfiles struct {
	mu sync.RWMutex

	clientIDToProfiles map[okta.ClientID][]ConfigProfile
}

func NewCachedGetClientIDToProfiles(
	ctx context.Context,
	configParams *AWSConfigGenerationParams,
	awsSession *session.Session,
) (*CachedGetClientIDToProfiles, error) {
	c := &CachedGetClientIDToProfiles{}
	// We manually run the first refresh to make sure something is there
	err := c.refresh(ctx, configParams, awsSession)
	if err != nil {
		return nil, err
	}

	// TODO(el): Currently no way to stop this ticker gracefully
	//           need to figure out something cleaner later
	//           but so be it for now.
	ticker := time.NewTicker(10 * time.Minute)
	go func() {
		for t := range ticker.C {
			logrus.WithTime(t).Infof("Refreshing AWS role mapping at %s", t.Format(time.RFC3339))
			err := c.refresh(ctx, configParams, awsSession)
			if err != nil {
				logrus.WithTime(t).WithError(err).Error("could not refresh aws role mapping")
			}
		}
	}()
	return c, nil
}

// Get returns the cached values
func (c *CachedGetClientIDToProfiles) Get(ctx context.Context) (map[okta.ClientID][]ConfigProfile, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.clientIDToProfiles == nil {
		return nil, errors.New("nil mapping of client_ids to aws profiles")
	}
	// return the cached value
	return c.clientIDToProfiles, nil
}

func (c *CachedGetClientIDToProfiles) refresh(
	ctx context.Context,
	configParams *AWSConfigGenerationParams,
	awsSession *session.Session,
) error {
	configData := &ClientIDToAWSRoles{
		awsSession:        awsSession,
		clientRoleMapping: map[okta.ClientID][]ConfigProfile{},
		roleARNs:          map[string]arn.ARN{},
		awsClient:         cziAWS.New(awsSession),
	}
	err := configData.getWorkerRoles(ctx, configParams.AWSMasterRoles, configParams.AWSWorkerRole)
	if err != nil {
		return errors.Wrap(err, "Unable to get list of RoleARNs accessible by the Master Roles")
	}

	err = configData.fetchAssumableRoles(ctx, configParams.OIDCProvider)
	if err != nil {
		return errors.Wrap(err, "Unable to create mapping needed for config generation")
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.clientIDToProfiles = configData.clientRoleMapping
	return nil
}
