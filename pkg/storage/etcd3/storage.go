package etcd3

import (
	"context"
	"fmt"
	"log"
	"path"
	"strings"
	"time"

	"kdev/pkg/runtime"
	"kdev/pkg/storage"
	"kdev/pkg/storage/watch"

	"github.com/sirupsen/logrus"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type Store struct {
	Client              *clientv3.Client
	PathPrefix          string
	GroupResourceString string
}

func NewEtcdStorage(cli *clientv3.Client, prefix string) (storage.Interface, error) {
	fullPathPrefix := ensureTrailingSlash(path.Join("/", prefix))

	logrus.Infof("새로운 저장소 생성: %s", fullPathPrefix)
	return &Store{Client: cli, PathPrefix: fullPathPrefix}, nil
}

func ensureTrailingSlash(s string) string {
	if !strings.HasSuffix(s, "/") {
		return s + "/"
	}
	return s
}

// Create 메서드 구현 (동시성 보장, 추적 추가)
func (s *Store) Create(ctx context.Context, key string, obj runtime.Object) error {
	encodedValue, err := obj.Encode()
	if err != nil {
		logrus.Infof("인코딩 실패: %v", err)
		return err
	}
	key = path.Join(s.PathPrefix, key)
	logrus.Infof("Create 호출: key=%s, encodedValue=%s", key, encodedValue)

	txn := s.Client.Txn(ctx)
	txn.If(clientv3.Compare(clientv3.CreateRevision(key), "=", 0)).
		Then(clientv3.OpPut(key, encodedValue))

	txnResp, err := txn.Commit()
	if err != nil {
		logrus.Infof("트랜잭션 실패: %v", err)
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
func (s *Store) GetList(ctx context.Context, refobj runtime.Object, objlist *[]runtime.Object) error {
	resp, err := s.Client.Get(ctx, s.PathPrefix, clientv3.WithPrefix())
	if err != nil {
		logrus.Infof("리스트 가져오기 실패: %v", err)
		return err
	}

	for _, kv := range resp.Kvs {
		obj := refobj.Clone()
		log.Println("check", string(kv.Value))
		if err := obj.Decode(kv.Value); err != nil {
			logrus.Infof("디코딩 실패: %v", err)
			continue
		}
		*objlist = append(*objlist, obj)
	}
	logrus.Infof("리스트 가져오기 성공: %+v", *objlist)
	return nil
}

// Read 메서드 수정 (runtime.Object 반환)
func (s *Store) Get(ctx context.Context, key string, obj runtime.Object) error {
	key = path.Join(s.PathPrefix, key)
	logrus.Infof("Read 호출: key=%s", key)
	resp, err := s.Client.Get(ctx, key)
	if err != nil {
		logrus.Infof("키 읽기 실패: %v", err)
		return err
	}
	if len(resp.Kvs) == 0 {
		log.Println("키를 찾을 수 없음")
		return fmt.Errorf("key not found")
	}
	err = obj.Decode(resp.Kvs[0].Value)
	if err != nil {
		logrus.Infof("디코딩 실패: %v", err)
		return err
	}
	logrus.Infof("키 읽기 성공: object=%+v", obj)
	return nil
}

// Update 메서드 구현 (동시성 보장, 추적 추가, 재시도 횟수 제한)
func (s *Store) Update(ctx context.Context, key string, obj runtime.Object) error {
	key = path.Join(s.PathPrefix, key)
	encodedValue, err := obj.Encode()
	if err != nil {
		logrus.Infof("값 인코딩 실패: %v", err)
		return err
	}
	logrus.Infof("Update 호출: key=%s, value=%s", key, encodedValue)
	retryLimit := 5 // 재시도 횟수 제한
	for attempt := 0; attempt < retryLimit; attempt++ {
		resp, err := s.Client.Get(ctx, key)
		if err != nil {
			logrus.Infof("키 업데이트 중 에러: %v", err)
			return err
		}
		if len(resp.Kvs) == 0 {
			logrus.Println("업데이트할 키를 찾을 수 없음")
			return fmt.Errorf("key not found")
		}
		kv := resp.Kvs[0]
		txn := s.Client.Txn(ctx)
		txn.If(clientv3.Compare(clientv3.ModRevision(key), "=", kv.ModRevision)).
			Then(clientv3.OpPut(key, string(encodedValue)))

		txnResp, err := txn.Commit()
		if err != nil {
			logrus.Infof("업데이트 트랜잭션 실패: %v", err)
			return err
		}
		if txnResp.Succeeded {
			log.Println("업데이트 트랜잭션 성공")
			return nil
		}
		logrus.Infof("업데이트 트랜잭션 실패, 재시도 %d번", attempt+1)
		time.Sleep(time.Duration(attempt+1) * 10 * time.Millisecond) // 지수 백오프
	}
	logrus.Println("업데이트 최대 재시도 횟수 초과")
	return fmt.Errorf("update failed after retries")
}

// Delete 메서드 구현 (동시성 보장, 추적 추가)
func (s *Store) Delete(ctx context.Context, key string) error {
	key = path.Join(s.PathPrefix, key)
	logrus.Infof("Delete 호출: key=%s", key)
	for {
		resp, err := s.Client.Get(ctx, key)
		if err != nil {
			logrus.Infof("삭제 중 에러: %v", err)
			return err
		}
		if len(resp.Kvs) == 0 {
			logrus.Println("삭제할 키를 찾을 수 없음")
			return fmt.Errorf("key not found")
		}
		kv := resp.Kvs[0]
		txn := s.Client.Txn(ctx)
		txn.If(clientv3.Compare(clientv3.ModRevision(key), "=", kv.ModRevision)).
			Then(clientv3.OpDelete(key, clientv3.WithPrevKV()))

		txnResp, err := txn.Commit()
		if err != nil {
			logrus.Infof("삭제 트랜잭션 실패: %v", err)
			return err
		}
		if txnResp.Succeeded {
			logrus.Println("삭제 트랜잭션 성공")
			return nil
		}
		log.Println("삭제 트랜잭션 재시도")
		time.Sleep(10 * time.Millisecond)
	}
}

// Watch 메서드 구현
func (s *Store) Watch(ctx context.Context, key string, refobj runtime.Object) (watch.Interface, error) {
	fullKey := path.Join(s.PathPrefix, key)
	watchChan := s.Client.Watch(ctx, fullKey, clientv3.WithPrefix(), clientv3.WithPrevKV())

	eventChan := make(chan watch.Event, watch.DefaultChanSize)
	stopChan := make(chan struct{})

	go func() {
		defer close(eventChan)
		defer logrus.Warnln("Watch 고루틴 종료됨")
		for {
			select {
			case <-ctx.Done():
				logrus.Warnln("취소됨")
				return
			case <-stopChan:
				logrus.Warnln("중지됨")
				return
			case resp := <-watchChan:
				if resp.Canceled {
					logrus.Warnf("취소됨: %v", resp.Err())
					eventChan <- watch.Event{Type: watch.Error, Object: nil}
					return
				}
				for _, ev := range resp.Events {
					logrus.Infof("이벤트: %v", ev.Type)
					logrus.Infof("이벤트: key %s %s", string(ev.Kv.Key), string(ev.Kv.Value))
					event := watch.Event{}
					switch ev.Type {
					case clientv3.EventTypePut:
						if ev.IsCreate() {
							event.Type = watch.Added
						} else {
							event.Type = watch.Modified
						}
						obj := refobj.Clone()
						if err := obj.Decode(ev.Kv.Value); err != nil {
							logrus.Errorf("디코딩 실패: %v", err)
							eventChan <- watch.Event{Type: watch.Error, Object: nil}
							continue
						}
						event.Object = obj
					case clientv3.EventTypeDelete:
						event.Type = watch.Deleted
						obj := refobj.Clone()
						if err := obj.Decode(ev.PrevKv.Value); err != nil {
							logrus.Errorf("디코딩 실패: %v", err)
							eventChan <- watch.Event{Type: watch.Error, Object: nil}
							continue
						}
						logrus.Infof("DELETED KEY %s VAL: %s: ", string(ev.Kv.Key), string(ev.PrevKv.Value))
						event.Object = obj
					default:
						logrus.Infof("알 수 없는 이벤트 타입: %v", ev.Type)
						event.Type = watch.Error
					}
					logrus.Infoln("이벤트 타입: ", string(event.Type))
					eventChan <- event
				}
			}
		}
	}()

	return &etcdWatch{
		resultChan: eventChan,
		stopChan:   stopChan,
	}, nil
}

type etcdWatch struct {
	resultChan chan watch.Event
	stopChan   chan struct{}
}

func (e *etcdWatch) Stop() {
	logrus.Infoln("Stop 호출")
	select {
	case <-e.stopChan:
		// 이미 중지됨
	default:
		close(e.stopChan)
	}
}

func (e *etcdWatch) ResultChan() <-chan watch.Event {
	logrus.Infoln("ResultChan 호출")
	return e.resultChan
}
