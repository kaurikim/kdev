package storage

import (
	"context"

	"kdev/pkg/storage/watch"

	"kdev/pkg/runtime"
)

type Interface interface {
	Create(ctx context.Context, key string, obj runtime.Object) error
	//Read(ctx context.Context, key string) (storage.runtime.Object, error)
	Get(ctx context.Context, key string, out runtime.Object) error
	GetList(ctx context.Context, refobj runtime.Object, objlist *[]runtime.Object) error
	Update(ctx context.Context, key string, obj runtime.Object) error
	Delete(ctx context.Context, key string) error
	Watch(ctx context.Context, key string, refobj runtime.Object) (watch.Interface, error)
}

type GroupResource struct {
	Group    string
	Resource string
}

func (gr GroupResource) String() string {
	return gr.Group + "/" + gr.Resource
}

type LeaseManagerConfig struct{}
