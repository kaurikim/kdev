package storage

import (
	"context"
)

// type Serializer interface {
type Object interface {
	Encode() (string, error)
	Decode([]byte) error
	Clone() Object
}

type Interface interface {
	Create(ctx context.Context, key string, obj Object) error
	//Read(ctx context.Context, key string) (storage.Object, error)
	Get(ctx context.Context, key string, out Object) error
	GetList(ctx context.Context, refobj Object, objlist *[]Object) error
	Update(ctx context.Context, key string, obj Object) error
	Delete(ctx context.Context, key string) error
}

type GroupResource struct {
	Group    string
	Resource string
}

func (gr GroupResource) String() string {
	return gr.Group + "/" + gr.Resource
}

type LeaseManagerConfig struct{}
