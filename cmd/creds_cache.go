package cmd

import (
	"context"

	"github.com/chanzuckerberg/go-misc/oidc_cli/storage"
	"github.com/chanzuckerberg/go-misc/pidlock"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/chanzuckerberg/aws-oidc/pkg/aws_config_client"
)

type Cache struct {
	storage storage.Storage
	lock    *pidlock.Lock

	updateCred func(context.Context, *aws_config_client.AWSOIDCConfiguration) (*processedCred, error)
}

func NewCache(
	storage storage.Storage,
	credGetter func(context.Context, *aws_config_client.AWSOIDCConfiguration) (*processedCred, error),
	lock *pidlock.Lock,
) *Cache {
	return &Cache{
		storage:      storage,
		updateCred: credGetter,
		lock:         lock,
	}
}

// Read will attempt to read a cred from the cache
//      if not present or expired, will refresh
func (c *Cache) Read(ctx context.Context, config *aws_config_client.AWSOIDCConfiguration) (*processedCred, error) {
	cachedCred, err := c.readFromStorage(ctx)
	if err != nil {
		return nil, err
	}
	// if we have a valid cred, use it
	if cachedCred.IsFresh() {
		return cachedCred, nil
	}

	// otherwise, try refreshing
	return c.refresh(ctx, config)
}

func (c *Cache) refresh(ctx context.Context, config *aws_config_client.AWSOIDCConfiguration) (*processedCred, error) {
	err := c.lock.Lock()
	if err != nil {
		return nil, err
	}
	defer c.lock.Unlock() //nolint:errcheck

	// acquire lock, try reading from cache again just in case
	// someone else got here first
	cachedCred, err := c.readFromStorage(ctx)
	if err != nil {
		return nil, err
	}
	// if we have a valid cred, use it
	if cachedCred.IsFresh() {
		return cachedCred, nil
	}

	// ok, at this point we have the lock and there are no good creds around
	// fetch a new one and save it
	cred, err := c.updateCred(ctx, config)
	if err != nil {
		return nil, err
	}

	// check the new cred is good to use
	if !cred.IsFresh() {
		return nil, errors.New("invalid cred fetched")
	}

	strCred, err := cred.Marshal()
	
	if err != nil {
		return nil, errors.Wrap(err, "unable to marshall cred")
	}
	// save cred to storage
	err = c.storage.Set(ctx, strCred)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to cache the strCred")
	}

	return cred, nil
}

// reads cred from storage, potentially returning a nil/expired cred
// users must call IsFresh to check cred validty
func (c *Cache) readFromStorage(ctx context.Context) (*processedCred, error) {
	cached, err := c.storage.Read(ctx)
	if err != nil {
		return nil, err
	}
	cachedCred, err := CredFromString(cached)
	if err != nil {
		logrus.WithError(err).Debug("error fetching stored cred")
		err = c.storage.Delete(ctx) // can't read it, so attempt to purge it
		if err != nil {
			logrus.WithError(err).Debug("error clearing cred from storage")
		}
	}
	return cachedCred, nil
}
