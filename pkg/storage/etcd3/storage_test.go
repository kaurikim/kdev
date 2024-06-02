package etcd3

import (
	"context"
	"encoding/json"
	"fmt"
	"kdev/pkg/auth"
	"kdev/pkg/storage"
	"sync"
	"testing"
	"time"

	"kdev/pkg/runtime"
	"kdev/pkg/storage/model"
	"kdev/pkg/storage/watch"

	"github.com/stretchr/testify/assert"
	clientv3 "go.etcd.io/etcd/client/v3"
)

var (
	GEndpoints = []string{"localhost:52988", "localhost:52989", "localhost:52990"}
)

type TestObject struct {
	Name  string
	Value string
}

func (t *TestObject) Clone() runtime.Object {
	return &TestObject{}
}

func (t *TestObject) Decode(data []byte) error {
	return json.Unmarshal(data, t)
}

func (t *TestObject) Encode() (string, error) {
	d, err := json.Marshal(t)
	if err != nil {
		return "", err
	}
	return string(d), nil
}

// storge test Create(), Delete() 함수 테스트
func TestStore_Create_Delete(t *testing.T) {
	ctx := context.TODO()
	key := "kauri"
	prefix := "user"
	v := &model.Entity[*auth.Account]{}

	pw := time.Now().Format("2006-01-02T15:04:05Z07:00") // ISO 8601 형식
	v.Set(&auth.Account{
		ID:       key,
		Password: pw,
	})

	// Create a new etcd client
	client, err := clientv3.New(clientv3.Config{
		//Endpoints: []string{"localhost:2379"},
		Endpoints: GEndpoints,
	})
	if err != nil {
		t.Fatalf("Failed to create etcd client: %v", err)
	}
	defer client.Close()
	// Create a new etcd client
	s, err := NewEtcdStorage(client, prefix)
	if err != nil {
		t.Fatalf("Failed to create etcd client: %v", err)
	}

	err = s.Create(ctx, key, v)
	if err != nil {
		t.Errorf("Create failed: %v", err)
	}
	out := model.NewEntity[*auth.Account](&auth.Account{})
	//out := &model.Entity[*service.LoginRequest]{}
	//out.Set(&service.LoginRequest{})

	err = s.Get(ctx, key, out)
	if err != nil {
		t.Errorf("Get failed: %v", err)
	}

	v.Entry.Password = "newpassword"
	err = s.Update(ctx, key, v)
	if err != nil {
		t.Errorf("Update failed: %v", err)
	}

	objlist := []runtime.Object{}
	err = s.GetList(ctx, out, &objlist)
	if err != nil {
		t.Errorf("GetList failed: %v", err)
	}

	for _, v := range objlist {
		fmt.Println(v)
		s.Delete(ctx, key)
	}
}

func setupStore(t *testing.T) storage.Interface {
	client, err := clientv3.New(clientv3.Config{
		Endpoints: GEndpoints,
	})
	if err != nil {
		t.Fatalf("Failed to create etcd client: %v", err)
	}

	store, err := NewEtcdStorage(client, "/test")
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}

	//return store.(*Store)
	return store
}

func TestCreate(t *testing.T) {
	store := setupStore(t)
	ctx := context.TODO()

	obj := &TestObject{Name: "test", Value: "value"}
	err := store.Create(ctx, "key", obj)
	assert.NoError(t, err)

	var got TestObject
	err = store.Get(ctx, "key", &got)
	assert.NoError(t, err)
	assert.Equal(t, obj.Name, got.Name)
	assert.Equal(t, obj.Value, got.Value)

	err = store.Delete(ctx, "key")
	assert.NoError(t, err)
}

func TestGet(t *testing.T) {
	store := setupStore(t)
	ctx := context.TODO()

	obj := &TestObject{Name: "test", Value: "value"}
	err := store.Create(ctx, "key", obj)
	assert.NoError(t, err)

	var got TestObject
	err = store.Get(ctx, "key", &got)
	assert.NoError(t, err)
	assert.Equal(t, obj.Name, got.Name)
	assert.Equal(t, obj.Value, got.Value)

	err = store.Delete(ctx, "key")
	assert.NoError(t, err)
}

func TestGetList(t *testing.T) {
	store := setupStore(t)
	ctx := context.TODO()

	refobj := &TestObject{}
	obj1 := &TestObject{Name: "test1", Value: "value1"}
	obj2 := &TestObject{Name: "test2", Value: "value2"}

	err := store.Create(ctx, "hello/key1", obj1)
	assert.NoError(t, err)
	err = store.Create(ctx, "key2", obj2)
	assert.NoError(t, err)

	var objlist []runtime.Object
	err = store.GetList(ctx, refobj, &objlist)
	assert.NoError(t, err)
	assert.Len(t, objlist, 2)

	err = store.Delete(ctx, "hello/key1")
	assert.NoError(t, err)
	err = store.Delete(ctx, "key2")
	assert.NoError(t, err)
}

func TestUpdate(t *testing.T) {
	store := setupStore(t)
	ctx := context.TODO()

	obj := &TestObject{Name: "test", Value: "value"}
	err := store.Create(ctx, "key", obj)
	assert.NoError(t, err)

	obj.Value = "new_value"
	err = store.Update(ctx, "key", obj)
	assert.NoError(t, err)

	var got TestObject
	err = store.Get(ctx, "key", &got)
	assert.NoError(t, err)
	assert.Equal(t, "new_value", got.Value)

	err = store.Delete(ctx, "key")
	assert.NoError(t, err)
}

func TestDelete(t *testing.T) {
	store := setupStore(t)
	ctx := context.TODO()

	obj := &TestObject{Name: "test", Value: "value"}
	err := store.Create(ctx, "key", obj)
	assert.NoError(t, err)

	err = store.Delete(ctx, "key")
	assert.NoError(t, err)

	var got TestObject
	err = store.Get(ctx, "key", &got)
	assert.Error(t, err)
}

func TestWatch(t *testing.T) {
	store := setupStore(t)
	ctx := context.TODO()
	key := "/test/user"

	watcher, err := store.Watch(ctx, "", &TestObject{})
	assert.NoError(t, err)

	obj := &TestObject{Name: "test", Value: "value"}
	err = store.Create(ctx, key, obj)
	assert.NoError(t, err)

	// Expecting the "Added" event
	select {
	case event := <-watcher.ResultChan():
		assert.Equal(t, watch.Added, event.Type, "Expected Added event")
		assert.Equal(t, "test", event.Object.(*TestObject).Name)
		assert.Equal(t, "value", event.Object.(*TestObject).Value)
	case <-time.After(5 * time.Second):
		t.Fatal("Watch timed out while waiting for Added event")
	}

	obj.Value = "new_test"
	err = store.Update(ctx, key, obj)
	assert.NoError(t, err)

	// Expecting the "Modified" event
	select {
	case event := <-watcher.ResultChan():
		assert.Equal(t, watch.Modified, event.Type, "Expected Modified event")
		assert.Equal(t, "test", event.Object.(*TestObject).Name)
		assert.Equal(t, "new_test", event.Object.(*TestObject).Value)
	case <-time.After(5 * time.Second):
		t.Fatal("Watch timed out while waiting for Modified event")
	}

	err = store.Delete(ctx, key)
	assert.NoError(t, err)

	// Expecting the "Deleted" event
	select {
	case event := <-watcher.ResultChan():
		assert.Equal(t, watch.Deleted, event.Type, "Expected Deleted event")
	case <-time.After(5 * time.Second):
		t.Fatal("Watch timed out while waiting for Deleted event")
	}
	TestConcurrency(t)
	time.Sleep(1 * time.Second)

	watcher.Stop()
	time.Sleep(1 * time.Second)
}
func TestDeleteAll(t *testing.T) {
	store := setupStore(t)
	ctx := context.TODO()

	// 키를 먼저 생성합니다.
	obj := &TestObject{Name: "test", Value: "value"}

	objlist := []runtime.Object{}
	err := store.GetList(ctx, obj, &objlist)
	if err != nil {
		t.Errorf("GetList failed: %v", err)
	}

	for _, v := range objlist {
		fmt.Println(v)
	}

	store.DeleteAll(ctx)
}

func TestConcurrency(t *testing.T) {
	store := setupStore(t)
	ctx := context.TODO()
	key := "/user/test/"

	// 키를 먼저 생성합니다.
	obj := &TestObject{Name: "test", Value: "value"}

	store.DeleteAll(ctx)
	err := store.Create(ctx, key, obj)
	assert.NoError(t, err)

	var wg sync.WaitGroup
	const goroutines = 10

	// Create a channel to capture errors
	errChan := make(chan error, goroutines)

	// Create a channel to capture the final values
	finalValues := make(chan string, goroutines)

	// Update function to be run concurrently
	updateFunc := func(i int) {
		defer wg.Done()
		updateObj := &TestObject{Name: "test", Value: fmt.Sprintf("value_%d", i)}
		if err := store.Update(ctx, key, updateObj); err != nil {
			errChan <- fmt.Errorf("update error: %w", err)
		} else {
			finalValues <- updateObj.Value
		}
	}

	// Start concurrent updates
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go updateFunc(i)
	}

	// Wait for all goroutines to finish
	wg.Wait()
	close(errChan)
	close(finalValues)

	// Check for errors
	for err := range errChan {
		if err != nil {
			t.Errorf("Concurrency test failed: %v", err)
		}
	}

	// Read the final value from the store
	var got TestObject
	err = store.Get(ctx, key, &got)
	assert.NoError(t, err)

	// Verify that the final value is one of the expected values
	expectedValues := map[string]bool{}
	for value := range finalValues {
		expectedValues[value] = true
	}

	if !expectedValues[got.Value] {
		t.Errorf("Final value %s is not in the expected values %v", got.Value, expectedValues)
	}
	store.DeleteAll(ctx)
}
