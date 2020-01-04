package stepcurry

import (
	"cloud.google.com/go/datastore"
	"context"
	"google.golang.org/api/option"
	"io"
)

// gcdatastore wraps an actual google cloud datastore Client for real/production datastore interaction
type gcdatastore struct {
	*datastore.Client
	gcloudProjectID  string
	gcloudClientOpts []option.ClientOption
}

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
	return ds.Client.Delete(c, k)
}

// Get loads the entity stored for key into dst. See https://godoc.org/cloud.google.com/go/datastore#Client.Get
func (ds *gcdatastore) Get(c context.Context, k *datastore.Key, dest interface{}) (err error) {
	return ds.Client.Get(c, k, dest)
}

// GetAll runs the provided query in the given context and returns all keys that match that query.
// See https://godoc.org/cloud.google.com/go/datastore#Client.GetAll
func (ds *gcdatastore) Run(c context.Context, q *datastore.Query) *datastore.Iterator {
	return ds.Client.Run(c, q)
}

// Put saves the entity src into the datastore with the given key. See https://godoc.org/cloud.google.com/go/datastore#Client.Put
func (ds *gcdatastore) Put(c context.Context, k *datastore.Key, v interface{}) (key *datastore.Key, err error) {
	return ds.Client.Put(c, k, v)
}

// NewKeyWithNamespace returns a new NameKey with a namespace
func NewKeyWithNamespace(kind string, namespace string, id string, parent *datastore.Key) (key *datastore.Key) {
	key = datastore.NameKey(kind, id, parent)
	key.Namespace = namespace

	return key
}
