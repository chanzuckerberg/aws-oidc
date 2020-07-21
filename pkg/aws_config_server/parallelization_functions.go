package aws_config_server

import (
	"context"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/service/iam"
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
	configParams *AWSConfigGenerationParams,
) error {
	ctx, span := beeline.StartSpan(ctx, "server_fetch_assumable_roles")
	defer span.Send()

	if configParams.MappingConcurrency == 0 {
		return errors.Errorf("Set configParams.MappingConcurrency to a value > 0")
	}

	oidcProvider := configParams.OIDCProvider

	aggregateMappings := func(ctx context.Context, arnPair *roleARNMatch) (map[okta.ClientID][]ConfigProfile, error) {
		accountName, roleARN := arnPair.accountName, arnPair.accountARN
		workerAWSConfig := &aws.Config{
			Credentials:                   stscreds.NewCredentials(a.awsSession, roleARN.String()),
			CredentialsChainVerboseErrors: aws.Bool(true),
			Retryer:                       a.awsSession.Config.Retryer,
		}
		iamClient := a.awsClient.WithIAM(workerAWSConfig).IAM.Svc
		workerRoles, err := listRoles(ctx, iamClient, configParams)
		if processAWSErr(err) != nil {
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
	mappingsList, err := parallelizeAggregateMapping(ctx, configParams.MappingConcurrency, queue, aggregateMappings)
	if err != nil {
		return errors.Wrap(err, "Unable to parallelize mapping generation process")
	}

	// For each of those roles in the mappingsList, filter them out using filterRoles
	for _, mapping := range mappingsList {
		for clientID, configList := range mapping {
			filteredConfigs, err := a.awsTagFilter(ctx, configList, configParams.RolesConcurrency)
			if err != nil {
				return errors.Wrapf(err, "Unable to filter these configs: %v", filteredConfigs)
			}

			for _, config := range filteredConfigs {
				if _, ok := a.clientRoleMapping[clientID]; !ok {
					a.clientRoleMapping[clientID] = []ConfigProfile{*config}
					continue
				}
				a.clientRoleMapping[clientID] = append(a.clientRoleMapping[clientID], *config)
			}
		}
	}

	return nil
}

// We can skip over roles with specific tags
func (a *ClientIDToAWSRoles) awsTagFilter(
	ctx context.Context,
	configList []ConfigProfile,
	rolesConcurrency int) ([]*ConfigProfile, error) {

	ctx, span := beeline.StartSpan(ctx, "filtering AWS Configs")
	defer span.Send()

	// Despite filtering AWS Config objects, we implemented concurrency for _role-based_actions_
	if rolesConcurrency == 0 {
		return nil, errors.Errorf("Set rolesConcurrency to a value > 0")
	}

	filterConfigsFunc := func(ctx context.Context, config ConfigProfile) (*ConfigProfile, error) {
		workerAWSConfig := &aws.Config{
			Credentials:                   stscreds.NewCredentials(a.awsSession, config.RoleARN.String()),
			CredentialsChainVerboseErrors: aws.Bool(true),
			Retryer:                       a.awsSession.Config.Retryer,
		}
		svc := a.awsClient.WithIAM(workerAWSConfig).IAM.Svc

		tags, err := listRoleTags(ctx, svc, &config.RoleName)
		if processAWSErr(err) != nil {
			return nil, errors.Wrapf(err, "error listing tags for %s", config.RoleName)
		}

		if shouldSkipTags(tags) {
			return nil, nil
		}

		// After all of this... return the role if it fulfills all requirements
		return &config, nil
	}

	return parallelizeFilterConfigs(ctx, rolesConcurrency, configList, filterConfigsFunc)
}

func parallelizeFilterConfigs(ctx context.Context,
	concurrencyLimit int,
	queue []ConfigProfile,
	action func(context.Context, ConfigProfile) (*ConfigProfile, error)) ([]*ConfigProfile, error) {
	logrus.Debug("start of parallelize function")

	// wg to track goroutines
	wg := sync.WaitGroup{}
	// to aggregate errors, make it buffered so we don't block on errors
	errs := make(chan error, len(queue))

	// how much concurrent work we're allowed to do
	scheduledQueue := make(chan *ConfigProfile, concurrencyLimit)

	// outputRoles to aggregate all our roles
	outputChannel := make(chan *ConfigProfile, len(queue))
	logrus.Debug("made all the channels")

	// the goroutine that will process one element at a time
	processor := func(scheduledQueue <-chan *ConfigProfile, outputList chan<- *ConfigProfile) {
		defer wg.Done()

		for element := range scheduledQueue {
			if element == nil {
				continue
			}

			output, err := action(ctx, *element)
			if err != nil {
				errs <- err
				continue
			}
			if output != nil {
				outputList <- output
				continue
			}
		}
	}

	// start the role processors
	for i := 0; i < concurrencyLimit; i++ {
		wg.Add(1)
		go processor(scheduledQueue, outputChannel)
	}

	// // schedule all the work
	for _, element := range queue {
		scheduledQueue <- &element
	}
	close(scheduledQueue) // signal processors they can stop

	wg.Wait()            // wait for all processors to be done
	close(errs)          // no more errors at this point
	close(outputChannel) // no more outputs at this point

	logrus.Debug("closed all the parallelizeFilterRoles channels")
	allErrs := &multierror.Error{}
	for err := range errs {
		allErrs = multierror.Append(allErrs, err)
	}
	logrus.Debug("added errors to error channel")

	// small lesson: don't set a length for the outputList!
	// Or else we'll get nil pointers in the output,
	// 	which will cause segmentation violations
	outputList := []*ConfigProfile{}
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

	// how much concurrent work we're allowed to do
	scheduledQueue := make(chan *roleARNMatch, concurrencyLimit)

	// outputRoles to aggregate all our roles
	outputChannel := make(chan map[okta.ClientID][]ConfigProfile, len(queue))
	logrus.Debug("made all the channels")

	// the goroutine that will process one element at a time
	processor := func(scheduledQueue <-chan *roleARNMatch, outputList chan<- map[okta.ClientID][]ConfigProfile) {
		defer wg.Done()

		for element := range scheduledQueue {
			if element == nil {
				continue
			}

			output, err := action(ctx, element)
			if err != nil {
				errs <- err
				continue
			}
			if output != nil {
				outputList <- output
				continue
			}
		}
	}

	// start the role processors
	for i := 0; i < concurrencyLimit; i++ {
		wg.Add(1)
		go processor(scheduledQueue, outputChannel)
	}

	// schedule all the work
	for _, element := range queue {
		scheduledQueue <- element
	}
	close(scheduledQueue) // signal processors they can stop

	wg.Wait()            // wait for all processors to be done
	close(errs)          // no more errors at this point
	close(outputChannel) // no more responses at this point

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
	logrus.Debugf("end of parallelize function. Errors: %v", allErrs.ErrorOrNil())
	return outputList, allErrs.ErrorOrNil()
}
