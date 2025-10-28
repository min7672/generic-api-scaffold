/*
 * EventBus : 단순한 이벤트 발행/구독 시스템
 *  - 역할 : Spring의 ApplicationEventPublisher / Observer 패턴과 유사
 *  - Publish(발행) 시, 등록된 모든 구독자 함수가 비동기로 호출됩니다.
 */
package bus

import (
	"go.uber.org/zap" // 로깅(디버깅 및 오류 추적용)
)

/*
 * DataCollectedEvent 구조체
 *  - 의미 : "데이터가 수집되었다"는 사실을 표현하는 이벤트 객체
 *  - 필드 :
 *      DeviceID : 이벤트 발생 장치 식별자
 *      Values   : 수집된 데이터 (key-value 형태)
 *  - Java 대응 : ApplicationEvent 하위 클래스 또는 DTO
 */
type DataCollectedEvent struct {
	DeviceID string
	Values   map[string]float64
}

/*
 * EventBus 구조체
 *  - 역할 : 이벤트를 전달할 "버스" 객체 (Spring의 ApplicationEventPublisher 유사)
 *  - 필드 :
 *      log         : 로깅 도구 (*zap.Logger)
 *      subscribers : 구독자(Subscriber) 함수 목록
 */
type EventBus struct {
	log         *zap.Logger
	subscribers []func(DataCollectedEvent)
}

/*
 * NewEventBus : fx가 호출하는 EventBus 생성자
 *  - Java 대응 : @Bean ApplicationEventPublisher
 *  - 반환 : *EventBus
 */
func NewEventBus(log *zap.Logger) *EventBus {
	return &EventBus{log: log}
}

/*
 * Subscribe : 이벤트 수신 함수를 등록하는 메서드
 *  - 인자 : func(DataCollectedEvent)
 *  - 동작 : 이벤트가 발행될 때마다 해당 함수를 호출
 *  - Java 대응 : @EventListener 또는 addObserver()
 */
func (b *EventBus) Subscribe(fn func(DataCollectedEvent)) {
	b.subscribers = append(b.subscribers, fn)
}

/*
 * Publish : 이벤트를 실제로 발행하는 메서드
 *  - 인자 : DataCollectedEvent (발행할 이벤트)
 *  - 동작 :
 *      ① 등록된 모든 구독자 함수(subscribers)를 순회
 *      ② 각 함수를 별도의 고루틴으로 비동기 실행
 *  - 효과 : 빠른 반응, 비동기 이벤트 처리
 *  - Java 대응 : ApplicationEventPublisher.publishEvent() 또는 Observer.notifyObservers()
 */
func (b *EventBus) Publish(e DataCollectedEvent) {
	for _, sub := range b.subscribers {
		go sub(e) // 비동기 실행(별도 고루틴)
	}
}
