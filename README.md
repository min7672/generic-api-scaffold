# go-rest-api-scaffold

**go-rest-api-scaffold**는 Go 언어로 **RESTful API 서버**를 설정할 수 있는 스캐폴드 프로젝트입니다. 
이 프로젝트는 **의존성 주입(DI)** 및 **서비스 레이어**를 포함한 구조로, **fx 라이프사이클 관리**와 **Gorilla Mux 라우팅**을 활용하여 API 서버를 구축할 수 있는 기본적인 뼈대를 제공합니다.

---

## 주요 기능

- **Go 언어** 기반으로 RESTful API 서버 구축
- **Gorilla Mux** 라우터를 사용하여 HTTP 요청 처리
- **Uber Fx**를 이용한 의존성 주입(DI) 및 라이프사이클 관리
- 서비스 및 컨트롤러의 모듈화 및 확장 가능
- godotenv 사용한 환경변수 주입
- 간단한 핑 및 헬스 체크 API 엔드포인트 제공

---

## 설치

1. 이 레포지토리를 클론합니다.

```bash
git clone https://github.com/<your-username>/go-rest-api-scaffold.git
cd go-rest-api-scaffold
```
2. 이 레포지토리를 클론합니다.

```bash
go mod tidy
```

---

## 사용법 

1. 애플리케이션을 실행합니다.
```bash
go run ./cmd/app
```

2. 서버가 실행되면, 다음 엔드포인트에서 API를 사용할 수 있습니다.
- /healthz: 헬스 체크
- /api/ping: 핑 확인
- /api/control: 제어 명령 처리