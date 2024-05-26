package etcd3

import (
	"context"
	"fmt"
	"kdev/pkg/service"
	"kdev/pkg/storage"
	"testing"
	"time"

	"kdev/pkg/storage/model"

	clientv3 "go.etcd.io/etcd/client/v3"
)

var (
	GEndpints = []string{"localhost:49855", "localhost:49856", "localhost:49857"}
)

// storge test Create(), Delete() 함수 테스트
func TestStore_Create_Delete(t *testing.T) {
	ctx := context.TODO()
	key := "kauri"
	prefix := "user"
	v := &model.Entity[*service.LoginRequest]{}

	pw := time.Now().Format("2006-01-02T15:04:05Z07:00") // ISO 8601 형식
	v.Set(&service.LoginRequest{
		Username: key,
		Password: pw,
	})

	// Create a new etcd client
	client, err := clientv3.New(clientv3.Config{
		//Endpoints: []string{"localhost:2379"},
		Endpoints: GEndpints,
	})
	if err != nil {
		t.Fatalf("Failed to create etcd client: %v", err)
	}
	defer client.Close()
	// Create a new etcd client
	s, err := NewEtcdStorage(client, prefix, "")
	if err != nil {
		t.Fatalf("Failed to create etcd client: %v", err)
	}

	err = s.Create(ctx, key, v)
	if err != nil {
		t.Errorf("Create failed: %v", err)
	}
	out := model.NewEntity[*service.LoginRequest](&service.LoginRequest{})
	//out := &model.Entity[*service.LoginRequest]{}
	//out.Set(&service.LoginRequest{})

	err = s.Get(ctx, key, out)
	if err != nil {
		t.Errorf("Get failed: %v", err)
	}

	v.Entry.Password = "123321"
	err = s.Update(ctx, key, v)
	if err != nil {
		t.Errorf("Update failed: %v", err)
	}

	objlist := []storage.Object{}
	err = s.GetList(ctx, out, &objlist)
	if err != nil {
		t.Errorf("GetList failed: %v", err)
	}

	for _, v := range objlist {
		fmt.Println(v)
		s.Delete(ctx, key)
	}
}
