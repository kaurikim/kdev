package main

import (
	"context"
	"fmt"
	"kdev/pkg/auth"
	"kdev/pkg/storage"
	"kdev/pkg/storage/etcd3"
	"kdev/pkg/storage/model"
	"time"

	"github.com/sirupsen/logrus"
	clientv3 "go.etcd.io/etcd/client/v3"
)

var (
	GEndpints = []string{"localhost:49855", "localhost:49856", "localhost:49857"}
)

func main() {
	ctx := context.TODO()
	key := "kauri"
	prefix := "user"
	v := &model.Entity[*auth.Account]{}
	//resolver.SetDefaultScheme("dns")

	pw := time.Now().Format("2006-01-02T15:04:05Z07:00") // ISO 8601 형식
	v.Set(&auth.Account{
		ID:       key,
		Password: pw,
	})

	// Create a new etcd client
	client, err := clientv3.New(clientv3.Config{
		//Endpoints: []string{"localhost:2379"},
		Endpoints: GEndpints,
	})
	if err != nil {
		logrus.Fatalf("Failed to create etcd client: %v", err)
	}
	defer client.Close()
	// Create a new etcd client
	s, err := etcd3.NewEtcdStorage(client, prefix, "")
	if err != nil {
		logrus.Fatalf("Failed to create etcd client: %v", err)
	}

	err = s.Create(ctx, key, v)
	if err != nil {
		logrus.Errorf("Create failed: %v", err)
	}
	out := model.NewEntity[*auth.Account](&auth.Account{})
	//out := &model.Entity[*service.LoginRequest]{}
	//out.Set(&service.LoginRequest{})

	err = s.Get(ctx, key, out)
	if err != nil {
		logrus.Errorf("Get failed: %v", err)
	}

	v.Entry.Password = "newpassword"
	err = s.Update(ctx, key, v)
	if err != nil {
		logrus.Errorf("Update failed: %v", err)
	}

	objlist := []storage.Object{}
	err = s.GetList(ctx, out, &objlist)
	if err != nil {
		logrus.Errorf("GetList failed: %v", err)
	}

	for _, v := range objlist {
		fmt.Println(v)
		s.Delete(ctx, key)
	}
}
