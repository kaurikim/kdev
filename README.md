# kdev

## API Gateway의 개념

API Gateway는 클라이언트 요청을 받아 적절한 백엔드 서비스로 라우팅하는 역할을 합니다. API Gateway는 하나의 진입점으로서 여러 개의 서비스에 대한 접근을 중앙에서 관리합니다. 이를 통해 클라이언트는 다양한 백엔드 서비스에 직접 접근하지 않고, API Gateway를 통해 간접적으로 접근할 수 있습니다.

## API Gateway 주요 기능

* 요청 라우팅 (Request Routing):
  * 클라이언트 요청을 적절한 백엔드 서비스로 라우팅합니다. 이 라우팅은 URL 경로나 요청 메서드, 헤더 등의 정보를 기반으로 수행됩니다.

* 프로토콜 변환 (Protocol Translation):
  * 클라이언트가 사용하는 프로토콜과 백엔드 서비스가 사용하는 프로토콜 간의 변환을 수행합니다. 예를 들어, RESTful HTTP 요청을 gRPC 호출로 변환하는 등의 역할을 합니다.

* 로드 밸런싱 (Load Balancing):
  * 여러 백엔드 서비스 인스턴스 간에 요청을 균등하게 분배하여, 서비스의 가용성과 성능을 최적화합니다.

* 인증 및 권한 부여 (Authentication and Authorization):
  * 요청에 대해 사용자 인증과 권한 검사를 수행합니다. JWT, OAuth, API Key 등의 인증 방식을 지원할 수 있습니다.

* 요청/응답 변환 (Request/Response Transformation):
  * 클라이언트 요청이나 백엔드 서비스의 응답을 변환합니다. 예를 들어, 응답 포맷을 JSON에서 XML로 변환하거나, 요청 데이터에 추가적인 정보를 주입하는 등의 작업을 수행합니다.

* API 조합 (API Composition):
  * 단일 클라이언트 요청에 대해 여러 백엔드 서비스로부터 데이터를 수집하고, 이를 결합하여 응답을 생성합니다. 이를 통해 클라이언트는 하나의 요청으로 필요한 모든 데이터를 받을 수 있습니다.

* 캐싱 (Caching):
  * 빈번히 요청되는 데이터를 캐싱하여 성능을 향상시키고, 백엔드 서비스의 부하를 줄입니다.

* 모니터링 및 로깅 (Monitoring and Logging):
  * 요청과 응답에 대한 로그를 기록하고, 성능 모니터링을 수행하여 시스템 상태를 추적합니다. 이를 통해 문제 발생 시 신속하게 대응할 수 있습니다.

* 속도 제한 및 스로틀링 (Rate Limiting and Throttling):
  * 서비스의 과도한 사용을 방지하기 위해 클라이언트 요청 수를 제한합니다. 예를 들어, 일정 시간 내에 특정 클라이언트가 보낼 수 있는 최대 요청 수를 제한합니다.

* 보안 (Security):
  * SSL/TLS를 사용하여 데이터 전송을 암호화하고, 공격 방어를 위한 다양한 보안 메커니즘을 구현합니다. 또한, IP 화이트리스트/블랙리스트와 같은 접근 제어 기능을 제공합니다.

* 서비스 디스커버리 (Service Discovery):
  * 동적으로 백엔드 서비스의 위치를 확인하여, 클라이언트 요청을 적절한 서비스 인스턴스로 라우팅합니다.


## TEST

```bash 
curl -X POST -d '{"username": "user1", "password": "password1"}' http://localhost:8080/v1/example/createuser

curl -X POST -d '{"username": "user1", "password": "password1"}' http://localhost:8080/v1/example/login

TOKEN="로그인에서 받은 토큰"
curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/v1/example/name -d '{"name": "World"}'

TOKEN="로그인에서 받은 토큰"
curl -X DELETE -H "Authorization: Bearer $TOKEN" -d '{"username": "user1"}' http://localhost:8080/v1/example/deleteuser
```