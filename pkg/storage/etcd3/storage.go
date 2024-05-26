package etcd3

import (
	"context"
	"fmt"
	"kdev/pkg/storage"
	"log"
	"path"
	"strings"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

type Store struct {
	Client              *clientv3.Client
	PathPrefix          string
	GroupResourceString string
}

func NewEtcdStorage(
	cli *clientv3.Client,
	prefix, resourcePrefix string,
) (storage.Interface, error) {
	log.Println("etcd 클라이언트 생성 성공")
	fullPathPrefix := ensureTrailingSlash(path.Join("/", prefix))

	store := &Store{
		Client:     cli,
		PathPrefix: fullPathPrefix,
	}

	log.Printf("새로운 저장소 생성: %s", fullPathPrefix)
	return store, nil
}

func ensureTrailingSlash(s string) string {
	if !strings.HasSuffix(s, "/") {
		return s + "/"
	}
	return s
}

// Create 메서드 구현 (동시성 보장, 추적 추가)
func (s *Store) Create(ctx context.Context, key string, obj storage.Object) error {
	encodedValue, err := obj.Encode()
	if err != nil {
		log.Printf("인코딩 실패: %v", err)
		return err
	}
	key = path.Join(s.PathPrefix, key)
	log.Printf("Create 호출: key=%s, encodedValue=%s", key, encodedValue)

	txn := s.Client.Txn(ctx)
	txn.If(clientv3.Compare(clientv3.CreateRevision(key), "=", 0)).
		Then(clientv3.OpPut(key, encodedValue))

	txnResp, err := txn.Commit()
	if err != nil {
		log.Printf("트랜잭션 실패: %v", err)
		return err
	}
	if !txnResp.Succeeded {
		log.Println("트랜잭션 실패: 키가 이미 존재함")
		return fmt.Errorf("key already exists")
	}
	log.Println("트랜잭션 성공")
	return nil
}

// s.PathPrefix 와 매칭되는 전체 데이터를 반환하는 함수
func (s *Store) GetList(ctx context.Context, refobj storage.Object, objlist *[]storage.Object) error {
	resp, err := s.Client.Get(ctx, s.PathPrefix, clientv3.WithPrefix())
	if err != nil {
		log.Printf("리스트 가져오기 실패: %v", err)
		return err
	}

	for _, kv := range resp.Kvs {
		obj := refobj.Clone()
		if err := obj.Decode(kv.Value); err != nil {
			log.Printf("디코딩 실패: %v", err)
			continue
		}
		*objlist = append(*objlist, obj)
	}
	log.Printf("리스트 가져오기 성공: %+v", *objlist)
	return nil
}

// Read 메서드 수정 (storage.Object 반환)
func (s *Store) Get(ctx context.Context, key string, obj storage.Object) error {
	key = path.Join(s.PathPrefix, key)
	log.Printf("Read 호출: key=%s", key)
	resp, err := s.Client.Get(ctx, key)
	if err != nil {
		log.Printf("키 읽기 실패: %v", err)
		return err
	}
	if len(resp.Kvs) == 0 {
		log.Println("키를 찾을 수 없음")
		return fmt.Errorf("key not found")
	}
	err = obj.Decode(resp.Kvs[0].Value)
	if err != nil {
		log.Printf("디코딩 실패: %v", err)
		return err
	}
	log.Printf("키 읽기 성공: object=%+v", obj)
	return nil
}

// Update 메서드 구현 (동시성 보장, 추적 추가, 재시도 횟수 제한)
func (s *Store) Update(ctx context.Context, key string, obj storage.Object) error {
	key = path.Join(s.PathPrefix, key)
	encodedValue, err := obj.Encode()
	if err != nil {
		log.Printf("값 인코딩 실패: %v", err)
		return err
	}
	log.Printf("Update 호출: key=%s, value=%s", key, encodedValue)
	retryLimit := 5 // 재시도 횟수 제한
	for attempt := 0; attempt < retryLimit; attempt++ {
		resp, err := s.Client.Get(ctx, key)
		if err != nil {
			log.Printf("키 업데이트 중 에러: %v", err)
			return err
		}
		if len(resp.Kvs) == 0 {
			log.Println("업데이트할 키를 찾을 수 없음")
			return fmt.Errorf("key not found")
		}
		kv := resp.Kvs[0]
		txn := s.Client.Txn(ctx)
		txn.If(clientv3.Compare(clientv3.ModRevision(key), "=", kv.ModRevision)).
			Then(clientv3.OpPut(key, string(encodedValue)))

		txnResp, err := txn.Commit()
		if err != nil {
			log.Printf("업데이트 트랜잭션 실패: %v", err)
			return err
		}
		if txnResp.Succeeded {
			log.Println("업데이트 트랜잭션 성공")
			return nil
		}
		log.Printf("업데이트 트랜잭션 실패, 재시도 %d번", attempt+1)
		time.Sleep(time.Duration(attempt+1) * 10 * time.Millisecond) // 지수 백오프
	}
	log.Println("업데이트 최대 재시도 횟수 초과")
	return fmt.Errorf("update failed after retries")
}

// Delete 메서드 구현 (동시성 보장, 추적 추가)
func (s *Store) Delete(ctx context.Context, key string) error {
	key = path.Join(s.PathPrefix, key)
	log.Printf("Delete 호출: key=%s", key)
	for {
		resp, err := s.Client.Get(ctx, key)
		if err != nil {
			log.Printf("삭제 중 에러: %v", err)
			return err
		}
		if len(resp.Kvs) == 0 {
			log.Println("삭제할 키를 찾을 수 없음")
			return fmt.Errorf("key not found")
		}
		kv := resp.Kvs[0]
		txn := s.Client.Txn(ctx)
		txn.If(clientv3.Compare(clientv3.ModRevision(key), "=", kv.ModRevision)).
			Then(clientv3.OpDelete(key))

		txnResp, err := txn.Commit()
		if err != nil {
			log.Printf("삭제 트랜잭션 실패: %v", err)
			return err
		}
		if txnResp.Succeeded {
			log.Println("삭제 트랜잭션 성공")
			return nil
		}
		log.Println("삭제 트랜잭션 재시도")
		time.Sleep(10 * time.Millisecond)
	}
}
