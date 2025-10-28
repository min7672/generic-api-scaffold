/*
 * Server : HTTP 서버 컨테이너
 *  - HTTP 서버를 실행하고 REST API 엔드포인트를 정의하는 구조체
 *  - Spring의 @RestController와 유사
 */
package infra

import (
	"os"
	"context"
	"fmt"
	"net/http"
	"time"
	"strconv"
	
	"github.com/gorilla/mux" // HTTP 라우팅을 위한 Gorilla Mux
	"go.uber.org/fx"         // Fx 프레임워크를 통한 라이프사이클 관리
	"go.uber.org/zap"        // 로깅 도구
)

// Server : HTTP 서버 컨테이너
//  - HTTP 서버, 라우터, 서버 설정을 관리하는 구조체
type Server struct {
	log    *zap.Logger    // 로그를 기록하는 로깅 도구
	router *mux.Router    // HTTP 라우터 (요청을 라우팅할 때 사용)
	srv    *http.Server   // 실제 HTTP 서버
	port   int            // 서버가 리스닝할 포트 번호
}

/*
 * NewHTTPServer : HTTP 서버를 생성하는 생성자 함수
 *  - 기본 포트는 8080으로 설정 (필요시 환경변수나 설정 파일을 통해 변경 가능)
 *  - HTTP 라우터를 초기화하고, 각 엔드포인트를 등록합니다.
 *  - 반환값 : *Server (HTTP 서버 객체)
 */
func NewHTTPServer(log *zap.Logger) *Server {
	portStr := os.Getenv("APP_PORT")
	if portStr == "" {
		portStr = "8080" // 기본값 8080
	}
	// string을 int로 변환
	port, err := strconv.Atoi(portStr)
	if err != nil {
		log.Fatal("Invalid port value, unable to convert to int", zap.Error(err))
	}
	r := mux.NewRouter() // Gorilla Mux 라우터 생성

	// Server 구조체 초기화
	s := &Server{
		log:    log,    // 로깅 도구
		router: r,      // 라우터
		port:   port,   // 기본 포트 8080
	}

	// === 라우팅 등록 ===
	// 헬스 체크 API: 서버 상태 확인용
	r.HandleFunc("/healthz", s.handleHealth).Methods(http.MethodGet)

	// 간단한 Ping API: 응답에 "pong"을 반환
	r.HandleFunc("/api/ping", s.handlePing).Methods(http.MethodGet)

	// 제어 명령 API: /api/control?action=charge&kw10=50와 같은 형태로 제어 명령을 처리
	r.HandleFunc("/api/control", s.handleControl).Methods(http.MethodPost)

	// 생성된 Server 객체 반환
	return s
}

/*
 * RegisterHooks : 앱 라이프사이클에 HTTP 서버 시작 및 종료를 위한 후크 등록
 *  - fx.Lifecycle을 사용하여 애플리케이션 시작 시 서버 시작, 종료 시 서버 종료 처리
 */
func RegisterHooks(lc fx.Lifecycle, s *Server) {
	// 서버 시작 및 종료 시 동작을 관리하는 후크 등록
	lc.Append(fx.Hook{
		// 애플리케이션 시작 시 서버 시작
		OnStart: func(ctx context.Context) error {
			// 서버 주소 구성
			addr := fmt.Sprintf(":%d", s.port)

			// HTTP 서버 설정
			s.srv = &http.Server{
				Addr:              addr,             // 서버 주소
				Handler:           s.router,          // 요청을 처리할 라우터
				ReadHeaderTimeout: 5 * time.Second,   // HTTP 헤더 읽기 타임아웃
				ReadTimeout:       10 * time.Second,  // HTTP 요청 읽기 타임아웃
				WriteTimeout:      10 * time.Second,  // HTTP 응답 쓰기 타임아웃
				IdleTimeout:       60 * time.Second,  // 유휴 상태의 타임아웃
			}

			// 서버를 고루틴에서 실행 (비동기 실행)
			go func() {
				s.log.Info("http server starting", zap.String("addr", addr))
				// 서버 실행 (서버가 종료되면 에러 로그 출력)
				if err := s.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					s.log.Error("http server error", zap.Error(err))
				}
			}()
			return nil
		},
		// 애플리케이션 종료 시 서버 종료
		OnStop: func(ctx context.Context) error {
			s.log.Info("http server stopping")
			// 그레이스풀 셧다운 (5초 타임아웃)
			shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()
			return s.srv.Shutdown(shutdownCtx) // 서버 종료
		},
	})
}

// ===== Handlers =====

/*
 * handleHealth : 헬스 체크 엔드포인트
 *  - 서버 상태를 확인하는 간단한 핑 요청을 처리합니다.
 */
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)          // HTTP 상태 코드 200 OK 반환
	_, _ = w.Write([]byte("ok"))         // 응답 본문에 "ok" 메시지 반환
}

/*
 * handlePing : 간단한 Ping 엔드포인트
 *  - 서버가 정상적으로 작동하는지 확인하는 데 사용됩니다.
 */
func (s *Server) handlePing(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)          // HTTP 상태 코드 200 OK 반환
	_, _ = w.Write([]byte(`{"pong":true}`)) // JSON 응답: {"pong": true}
}

/*
 * controlReq : 제어 명령 요청을 처리하기 위한 구조체
 *  - Action : 수행할 액션 (예: "charge", "discharge", "on", "off" 등)
 *  - KW10 : kW 단위로 10배수로 지정된 값 (예: 50은 5.0kW)
 */
type controlReq struct {
	Action string `json:"action"` // 예: charge|discharge|on|off
	KW10   int    `json:"kw10"`   // kW*10 (예: 50 => 5.0kW)
}

/*
 * handleControl : 제어 명령을 처리하는 엔드포인트
 *  - 요청: /api/control?action=charge&kw10=50 형태의 쿼리 파라미터로 전달
 *  - 실제 제어는 나중에 연결될 수 있음 (현재는 단순한 응답을 보냄)
 */
func (s *Server) handleControl(w http.ResponseWriter, r *http.Request) {
	// 요청에서 쿼리 파라미터 받기
	q := r.URL.Query()
	action := q.Get("action") // action: charge|discharge|ready|on|off
	kw10 := q.Get("kw10")     // kw10: kW 단위 (예: 50 => 5.0kW)

	// 요청 로그 출력
	s.log.Info("control request received", zap.String("action", action), zap.String("kw10", kw10))

	// 응답 반환: 명령이 큐에 추가되었음을 나타내는 상태 코드 202 (Accepted)
	w.WriteHeader(http.StatusAccepted)
	_, _ = w.Write([]byte(`{"status":"queued"}`)) // {"status": "queued"} 메시지 응답
}
