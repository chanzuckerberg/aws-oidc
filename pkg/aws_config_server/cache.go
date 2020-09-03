package aws_config_server

import (
	"context"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"
	"github.com/aws/aws-sdk-go/service/organizations/organizationsiface"
	cziAWS "github.com/chanzuckerberg/go-misc/aws"
	"github.com/honeycombio/beeline-go"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type CachedGetClientIDToProfiles struct {
	mu sync.RWMutex

	clientIDToProfiles *oidcFederatedRoles
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
func (c *CachedGetClientIDToProfiles) Get(ctx context.Context) (*oidcFederatedRoles, error) {
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
	logrus.Info("Initiating AWS Roles refresh")
	start := time.Now()

	ctx, span := beeline.StartSpan(ctx, "refresh_client_mapping")
	defer span.Send()

	awsClient := cziAWS.New(awsSession)
	orgAssumer := func(config *aws.Config) organizationsiface.OrganizationsAPI {
		return awsClient.WithOrganizations(config).Organizations.Svc
	}
	iamAssumer := func(config *aws.Config) iamiface.IAMAPI {
		return awsClient.WithIAM(config).IAM.Svc
	}

	workerRoles, err := getWorkerRoles(
		ctx,
		awsSession,
		orgAssumer,
		configParams.AWSOrgRoles,
		configParams.AWSWorkerRole,
		configParams.SkipAccounts, // note to self: should be a string set
	)
	if err != nil {
		return err
	}

	allRoles, err := listRolesForAccounts(
		ctx,
		awsSession,
		iamAssumer,
		workerRoles,
		configParams.OIDCProvider,
		configParams.Concurrency,
	)
	if err != nil {
		return err
	}

	c.clientIDToProfiles = allRoles

	logrus.Infof("done refreshing aws roles %f", time.Since(start).Seconds())
	return nil
}
