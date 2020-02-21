package stepcurry

import (
	"cloud.google.com/go/datastore"
	"context"
	"google.golang.org/api/option"
	"io"
)

const (
	// Try operations that could fail at most twice. The first time is assummed to potentially fail because
	// of authentication errors when credentials have expired. The second time, a failure is probably
	// something to report back
	maxAttemptCount = 2
)

// TestValue represents an test entity to validate a datastore connection
type TestValue struct {
	Value string `datastore:",noindex"`
}

// gcdatastore wraps an actual google cloud datastore Client for real/production datastore interaction
type gcdatastore struct {
	*datastore.Client
	gcloudProjectID  string
	gcloudClientOpts []option.ClientOption
}

type retryableOperation func() (err error)

type retryableOperationWithkey func() (key *datastore.Key, err error)

// NewDatastorer creates a new instance of a DataStorer backed by a real datastore client
func NewDatastorer(gcloudProjectID string, gcloudOpts ...option.ClientOption) (ds *gcdatastore, err error) {
	ds = new(gcdatastore)
	ds.gcloudProjectID = gcloudProjectID
	ds.gcloudClientOpts = gcloudOpts

	err = ds.Connect()
	if err != nil {
		return nil, err
	}

	return ds, nil
}

// connect creates a new client instance from the initial gcloud project id and client options
// If the client options can be updated during the course of a process (such as option.WithCredentialsFile),
// connect should be able to reflect changes in those when it lazily reconnects on error
func (ds *gcdatastore) Connect() (err error) {
	ctx := context.Background()

	ds.Client, err = datastore.NewClient(ctx, ds.gcloudProjectID, ds.gcloudClientOpts...)
	if err != nil {
		return err
	}

	err = ds.testDB()
	if err = ds.testDB(); err != nil {
		ds.Close()
		return err
	}

	return nil
}

// testDB makes a lightweight call to the datastore to validate connectivity and credentials
func (ds *gcdatastore) testDB() (err error) {
	err = ds.Client.Get(context.Background(), datastore.NameKey("test", "testConnectivity", nil), &TestValue{})

	if err != nil && err != datastore.ErrNoSuchEntity {
		return err
	}

	return nil
}

// datastorer is implemented by any value that implements all of its methods. It is meant
// to allow easier testing decoupled from an actual datastore to interact with and
// the methods defined are method implemented by the datastore.Client that this package
// uses
type Datastorer interface {
	Connecter
	io.Closer
	Delete(c context.Context, k *datastore.Key) (err error)
	Get(c context.Context, k *datastore.Key, dest interface{}) (err error)
	Run(ctx context.Context, q *datastore.Query) *datastore.Iterator
	Put(c context.Context, k *datastore.Key, v interface{}) (key *datastore.Key, err error)
}

// Delete deletes the entity for the given key. See https://godoc.org/cloud.google.com/go/datastore#Client.Delete
func (ds *gcdatastore) Delete(c context.Context, k *datastore.Key) (err error) {
	return ds.tryWithRecovery(func() (err error) {
		return ds.Client.Delete(c, k)
	})
}

func (ds *gcdatastore) tryWithRecovery(operation retryableOperation) (err error) {
	err = operation()
	for attempt := 1; attempt < maxAttemptCount && err != nil && shouldRetry(err); attempt = attempt + 1 {
		ds.Connect()

		err = operation()
	}

	return err
}

func (ds *gcdatastore) tryKeyOperationWithRecovery(operation retryableOperationWithkey) (key *datastore.Key, err error) {
	key, err = operation()
	for attempt := 1; attempt < maxAttemptCount && err != nil && shouldRetry(err); attempt = attempt + 1 {
		ds.Connect()

		key, err = operation()
	}

	return key, err
}

// shouldRetry returns true if the given error should be retried or false if not.
// In order to determine this, one approach would be to only retry on a
// statusError (https://github.com/grpc/grpc-go/blob/master/status/status.go#L43)
// with code Unauthenticated (https://godoc.org/google.golang.org/grpc/codes) but that's made
// trickier by the statusError not being promoted outside the package (checking for the Error string
// would be reasonable but a bit dirty).
// Alternatively, and what's done here is to be a little conservative and retry on everything except
// ErrNoSuchEntity, ErrInvalidEntityType and ErrInvalidKey which are not things retries would help
// with. This means we could still retry when it's pointless to do so at the expense of added latency.
func shouldRetry(err error) bool {
	return err != datastore.ErrNoSuchEntity && err != datastore.ErrInvalidEntityType && err != datastore.ErrInvalidKey
}

// Get loads the entity stored for key into dst. See https://godoc.org/cloud.google.com/go/datastore#Client.Get
func (ds *gcdatastore) Get(c context.Context, k *datastore.Key, dest interface{}) (err error) {
	return ds.tryWithRecovery(func() (err error) {
		return ds.Client.Get(c, k, dest)
	})
}

// GetAll runs the provided query in the given context and returns all keys that match that query.
// See https://godoc.org/cloud.google.com/go/datastore#Client.GetAll
func (ds *gcdatastore) Run(c context.Context, q *datastore.Query) *datastore.Iterator {
	return ds.Client.Run(c, q)
}

// Put saves the entity src into the datastore with the given key. See https://godoc.org/cloud.google.com/go/datastore#Client.Put
func (ds *gcdatastore) Put(c context.Context, k *datastore.Key, v interface{}) (key *datastore.Key, err error) {
	return ds.tryKeyOperationWithRecovery(func() (key *datastore.Key, err error) {
		return ds.Client.Put(c, k, v)
	})
}

// NewKeyWithNamespace returns a new NameKey with a namespace
func NewKeyWithNamespace(kind string, namespace string, id string, parent *datastore.Key) (key *datastore.Key) {
	key = datastore.NameKey(kind, id, parent)
	key.Namespace = namespace

	return key
}
