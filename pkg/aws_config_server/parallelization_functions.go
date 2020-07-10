package aws_config_server

import (
	"context"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"
	"github.com/chanzuckerberg/aws-oidc/pkg/okta"
	"github.com/hashicorp/go-multierror"
	"github.com/honeycombio/beeline-go"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var shouldSkipTags = func(tags []*iam.Tag) bool {
	for _, tag := range tags {
		if tag != nil && tag.Key != nil && *tag.Key == skipRolesTagKey {
			return true
		}
	}
	return false
}

func (a *ClientIDToAWSRoles) populateMapping(
	ctx context.Context,
	oidcProvider string,
) error {
	ctx, span := beeline.StartSpan(ctx, "server_fetch_assumable_roles")
	defer span.Send()

	aggregateMappings := func(ctx context.Context, arnPair *roleARNMatch) (map[okta.ClientID][]ConfigProfile, error) {
		accountName, roleARN := arnPair.accountName, arnPair.accountARN
		workerAWSConfig := &aws.Config{
			Credentials:                   stscreds.NewCredentials(a.awsSession, roleARN.String()),
			CredentialsChainVerboseErrors: aws.Bool(true),
			Retryer: &client.DefaultRetryer{
				NumMaxRetries:    10,
				MinRetryDelay:    time.Millisecond,
				MinThrottleDelay: time.Millisecond,
				MaxThrottleDelay: time.Second,
				MaxRetryDelay:    time.Second,
			},
		}
		iamClient := a.awsClient.WithIAM(workerAWSConfig).IAM.Svc
		workerRoles, err := listRoles(ctx, iamClient)
		if err != nil {
			return nil, errors.Wrapf(err, "error listing roles for %s", accountName)
		}
		// account aliases will be used to determine profile names
		// by the completer in cli
		accountAlias, err := getAcctAlias(ctx, iamClient)
		if err != nil {
			return nil, errors.Wrapf(err, "error listing account aliases for %s", accountName)
		}
		currentRoleMapping, err := getRoleMappings(ctx, accountName, accountAlias, workerRoles, oidcProvider)
		if err != nil {
			return nil, errors.Wrapf(err, "error generating role mapping for %s", accountName)
		}

		return currentRoleMapping, nil
	}

	flattenRoleARNs := func(roleARNMap map[string]arn.ARN) []*roleARNMatch {
		flattenedARNs := []*roleARNMatch{}
		for accountName, arn := range roleARNMap {
			flattenedARNs = append(flattenedARNs, &roleARNMatch{
				accountName: accountName,
				accountARN:  arn,
			})
		}
		return flattenedARNs
	}

	queue := flattenRoleARNs(a.roleARNs)
	mappingsList, err := parallelizeAggregateMapping(ctx, 5, queue, aggregateMappings)
	if err != nil {
		return errors.Wrap(err, "Unable to parallelize mapping generation process")
	}

	for _, mapping := range mappingsList {

		for clientID, configList := range mapping {
			for _, config := range configList {
				if _, ok := a.clientRoleMapping[clientID]; !ok {
					a.clientRoleMapping[clientID] = []ConfigProfile{config}
					continue
				}
				a.clientRoleMapping[clientID] = append(a.clientRoleMapping[clientID], config)
			}
		}
	}

	return nil
}

// We can skip over roles with specific tags
func filterRoles(
	ctx context.Context,
	svc iamiface.IAMAPI,
	roles []*iam.Role) ([]*iam.Role, error) {

	ctx, span := beeline.StartSpan(ctx, "filtering AWS roles")
	defer span.Send()

	filterRolesFunc := func(ctx context.Context, role *iam.Role) (*iam.Role, error) {
		tags, err := listRoleTags(ctx, svc, role.RoleName)
		if err != nil {
			return nil, errors.Wrapf(err, "error listing tags for %s", *role.RoleName)
		}

		if shouldSkipTags(tags) {
			return nil, nil
		}

		// After all of this... return the role if it fulfills all requirements
		return role, nil
	}

	return parallelizeFilterRoles(ctx, 5, roles, filterRolesFunc)
}

func parallelizeFilterRoles(ctx context.Context,
	concurrencyLimit int,
	queue []*iam.Role,
	action func(context.Context, *iam.Role) (*iam.Role, error),
) ([]*iam.Role, error) {
	logrus.Debug("in parallelize function")

	// wg to track goroutines
	wg := sync.WaitGroup{}
	// to aggregate errors, make it buffered so we don't block on errors
	errs := make(chan error, len(queue))

	// // how much concurrent work we're allowed to do
	scheduledQueue := make(chan *iam.Role, concurrencyLimit)

	// // outputRoles to aggregate all our roles
	outputChannel := make(chan *iam.Role, len(queue))
	logrus.Debug("made all the channels")

	// the goroutine that will process one element at a time
	processor := func(scheduledQueue <-chan *iam.Role, outputList chan<- *iam.Role) {
		defer wg.Done()

		for element := range scheduledQueue {
			if element == nil {
				return
			}

			output, err := action(ctx, element)
			if err != nil {
				errs <- err
			}
			if output != nil {
				outputList <- output
			}
		}
	}

	// start the role processors
	for i := 0; i < concurrencyLimit; i++ {
		logrus.Debug("start role processors (for loop)")
		wg.Add(1)
		go processor(scheduledQueue, outputChannel)
	}

	// // schedule all the work
	for _, element := range queue {
		scheduledQueue <- element
	}
	close(scheduledQueue) // signal processors they can stop

	wg.Wait()   // wait for all processors to be done
	close(errs) // no more errors at this point
	close(outputChannel)

	logrus.Debug("closed all the parallelizeFilterRoles channels")
	allErrs := &multierror.Error{}
	for err := range errs {
		allErrs = multierror.Append(allErrs, err)
	}
	logrus.Debug("added errors to error channel")

	// small lesson: don't set a length for the outputList!
	// Or else we'll get nil pointers in the output,
	// 	which will cause segmentation violations
	outputList := []*iam.Role{}
	for element := range outputChannel {
		outputList = append(outputList, element)
	}
	logrus.Debug("end of parallelize function")
	return outputList, allErrs.ErrorOrNil()
}

func parallelizeAggregateMapping(ctx context.Context,
	concurrencyLimit int,
	queue []*roleARNMatch,
	action func(context.Context, *roleARNMatch) (map[okta.ClientID][]ConfigProfile, error),
) ([]map[okta.ClientID][]ConfigProfile, error) {
	logrus.Debug("in parallelize function")

	// wg to track goroutines
	wg := sync.WaitGroup{}
	// to aggregate errors, make it buffered so we don't block on errors
	errs := make(chan error, len(queue))

	// // how much concurrent work we're allowed to do
	scheduledQueue := make(chan *roleARNMatch, concurrencyLimit)

	// // outputRoles to aggregate all our roles
	outputChannel := make(chan map[okta.ClientID][]ConfigProfile, len(queue))
	logrus.Debug("made all the channels")

	// the goroutine that will process one element at a time
	processor := func(scheduledQueue <-chan *roleARNMatch, outputList chan<- map[okta.ClientID][]ConfigProfile) {
		defer wg.Done()

		for element := range scheduledQueue {
			if element == nil {
				return
			}

			output, err := action(ctx, element)
			if err != nil {
				errs <- err
			}
			if output != nil {
				outputList <- output
			}
		}
	}

	// start the role processors
	for i := 0; i < concurrencyLimit; i++ {
		logrus.Debug("start role processors (for loop)")
		wg.Add(1)
		go processor(scheduledQueue, outputChannel)
	}

	// // schedule all the work
	for _, element := range queue {
		scheduledQueue <- element
	}
	close(scheduledQueue) // signal processors they can stop

	wg.Wait()   // wait for all processors to be done
	close(errs) // no more errors at this point
	close(outputChannel)

	logrus.Debug("closed all the parallelizeAggregateMapping channels")
	allErrs := &multierror.Error{}
	for err := range errs {
		allErrs = multierror.Append(allErrs, err)
	}
	logrus.Debug("added errors to error channel")

	// small lesson: don't set a length for the outputList!
	// Or else we'll get nil pointers in the output,
	// 	which will cause segmentation violations
	outputList := []map[okta.ClientID][]ConfigProfile{}
	for element := range outputChannel {
		outputList = append(outputList, element)
	}
	logrus.Debug("end of parallelize function")
	return outputList, allErrs.ErrorOrNil()
}
