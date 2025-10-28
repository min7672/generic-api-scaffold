/*
 * 애플리케이션을 조립(의존성 주입)하고 실행/종료를 관리하는 파일입니다.
 * Uber Fx 프레임워크를 통해 DI(Dependency Injection)와 생명주기(Start/Stop)를 자동 처리합니다.
 */
package app

import (
	"context"

	"go.uber.org/fx"  // DI 컨테이너 및 라이프사이클 관리
	"go.uber.org/zap" // 고성능 구조화 로깅 패키지
	
	"generic-api-scaffold/internal/bus"   // 이벤트 버스(내부 컴포넌트 간 이벤트 전달)
	"generic-api-scaffold/internal/infra" // 외부 연동(Infrastructure) 예: Influx 저장 시뮬
)

/*
 * Run : main 함수에서 호출되는 애플리케이션 구동 함수
 * Fx 컨테이너(fx.New)를 통해 모든 구성요소를 등록(Provide) 및 실행(Invoke)합니다.
 */
func Run(ctx context.Context) {
	app := fx.New(

		/* 
		 * Provide : fx에 객체 생성자(의존성 주입용)를 등록
		 * - 생성자 - (func 키워드 : 함수 )
		 * 코드 포인터(Code pointer) : 해당 함수의 실제 기계 코드 주소 (C/C++의 함수 포인터와 유사)
		 * 클로저 환경(Environment) : 함수가 참조 중인 외부 변수들의 주소나 복사본 (closure capture)
		 * 내가 이해한 표현 : 함수 원형에서 복사한 값을 통으로 들고다닌다. 함수 원형 스냅샷
		*/
		fx.Provide(
			NewLogger,
			
			bus.NewEventBus,
			infra.NewHTTPServer,
			infra.NewInfluxRepo, // ★ 추가: *infra.InfluxRepo 제공
			NewCollector,
    	),
		
		
		/* Invoke : 앱 시작 시 실행할 초기 함수 등록 */
		fx.Invoke(registerHandlers, infra.RegisterHooks),
		
		
	)

	/* 앱 시작 : 내부적으로 모든 OnStart 훅을 실행 */
	_ = app.Start(ctx)

	/* ctx.Done() : OS 종료 신호(SIGINT, SIGTERM) 수신 시까지 대기 */
	<-ctx.Done()

	/* 앱 종료 : 내부적으로 모든 OnStop 훅을 실행하여 자원 정리 */
	_ = app.Stop(context.Background())
}

/*
 * NewLogger : 개발용 로거(Logger) 생성 함수
 * zap.NewDevelopment() → 사람이 보기 쉬운 포맷으로 로그를 출력
 * fx.Provide(NewLogger)를 통해 자동 주입 가능
 */
func NewLogger() (*zap.Logger, error) {
	return zap.NewDevelopment()
}
