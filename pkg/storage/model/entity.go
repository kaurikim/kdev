package model

import (
	"log"

	"kdev/pkg/runtime"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// Entity 객체는 MongoDB 데이터베이스에 접근하는 데이터를 정의합니다. proto 파일에 정의한 인터페이스를 임베딩하여 사용합니다.
type Entity[M proto.Message] struct {
	Entry M `bson:",inline"`
}

func NewEntity[M proto.Message](m M) *Entity[M] {
	entry := proto.Clone(m).(M)
	return &Entity[M]{Entry: entry}
}

// Set 매서드는 Entity 객체 내의 Entry 필드를 설정합니다.
func (entity *Entity[M]) Set(entry M) {
	entity.Entry = proto.Clone(entry).(M)
}

func (entity *Entity[M]) Clone() runtime.Object {
	return NewEntity(entity.Entry)
}

// Encode는 Entity 객체를 JSON 형식의 문자열로 직렬화합니다.
func (entity *Entity[M]) Encode() (string, error) {
	// 프로토콜 버퍼 메시지를 JSON으로 직렬화
	entryData, err := protojson.Marshal(entity.Entry)
	if err != nil {
		return "", err
	}

	return string(entryData), nil
}

func (entity *Entity[M]) Decode(data []byte) error {
	// Unmarshal 함수에 적절한 메시지 타입을 제공하기 위해 새로운 인스턴스를 생성합니다.
	entry := proto.Clone(entity.Entry).(M)
	if data == nil {
		log.Fatalln("data is nil")
		return nil
	}
	if err := protojson.Unmarshal(data, entry); err != nil {
		return err
	}
	entity.Entry = entry
	return nil
}
