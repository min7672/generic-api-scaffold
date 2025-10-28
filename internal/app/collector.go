/*
 * Collector : 주기적으로 데이터를 수집하고, 그 결과를 이벤트로 발행하는 컴포넌트입니다.
 */
package app

import (
	"context"
	"time"

	"go.uber.org/fx"  // 애플리케이션 생명주기(Lifecycle) 훅 제공
	"go.uber.org/zap" // 구조화 로그 출력 라이브러리

	"generic-api-scaffold/internal/bus"   // 이벤트 정의 및 전달
	"generic-api-scaffold/internal/infra" // 저장소(Infrastructure) 계층
)

/*
 * Collector 구조체
 *  - 역할 : Spring의 @Service 또는 Bean 개념에 해당
 *  - 필드 : 의존성 주입 대상 (Logger, EventBus, InfluxRepo)
 */
type Collector struct {
	log  *zap.Logger
	bus  *bus.EventBus
	repo *infra.InfluxRepo
}

/*
 * NewCollector : fx가 호출하는 Collector 생성자
 *  - Java Lombok의 @RequiredArgsConstructor 또는 Spring의 @Autowired 생성자와 동일한 개념
 *  - 반환 : *Collector
 */
func NewCollector(log *zap.Logger, b *bus.EventBus, r *infra.InfluxRepo) *Collector {
	return &Collector{log: log, bus: b, repo: r}
}
/*
 * registerHandlers : Collector의 시작(Start)·정지(Stop) 시점을 fx.Lifecycle에 등록
 *  - fx.Invoke(registerHandlers)로 실행되며, 애플리케이션 구동 시 자동으로 훅(Append) 추가
 *  - OnStart : Collector의 주기적 수집 루프를 고루틴으로 시작
 *  - OnStop  : 컨텍스트 종료 시 루프를 정리하고 로그 출력
 */
func registerHandlers(lc fx.Lifecycle, c *Collector) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go c.Start(ctx)
			return nil
		},
		OnStop: func(ctx context.Context) error {
			c.log.Info("collector stopped")
			return nil
		},
	})
}

/*
 * Start : Collector의 메인 루프
 *  - 3초 주기로 데이터 수집을 시뮬레이션하고, 이벤트 버스에 발행
 *  - ctx.Done() 신호가 오면 루프를 종료하고 리소스를 정리
 *  - 내부 동작 :
 *     ① time.Ticker 생성 (3초 주기)
 *     ② 매 주기마다 임의의 데이터(temp=23.5)를 생성
 *     ③ bus.Publish()를 통해 DataCollectedEvent 발행
 */
func (c *Collector) Start(ctx context.Context) {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			c.log.Info("collector exit")
			return
		case <-ticker.C:
			c.log.Info("collecting data...")

			data := map[string]float64{"temp": 23.5} // 샘플 데이터
			c.bus.Publish(bus.DataCollectedEvent{
				DeviceID: "A1",
				Values:   data,
			})
		}
	}
}
