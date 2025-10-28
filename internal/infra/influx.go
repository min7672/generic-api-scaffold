/*
 * InfluxRepo : InfluxDB 1.x 저장소
 *  - 역할 : 수집된 데이터를 InfluxDB에 저장하는 역할
 *  - 구성 :
 *      - log : 로깅 도구 (zap.Logger)
 *      - cfg : InfluxDB 연결 및 설정 정보 (Config)
 *      - client : InfluxDB 클라이언트 (client.Client)
 *  - 기능 :
 *      - 데이터 수집 이벤트를 EventBus에서 구독
 *      - 이벤트가 발생하면 데이터를 InfluxDB에 저장
 *      - 이벤트 구독은 비동기로 처리
 */
package infra

import (
	"context"
	"generic-api-scaffold/internal/bus"  // 이벤트 처리 (DataCollectedEvent)
	
	"time"
	"os"
	"github.com/influxdata/influxdb1-client/v2" // InfluxDB 1.x 클라이언트
	"go.uber.org/fx"  // Fx 프레임워크
	"go.uber.org/zap" // 로깅 도구
)

// InfluxRepo : InfluxDB에 데이터를 쓰는 저장소
type InfluxRepo struct {
	log    *zap.Logger      // 로깅 도구
	
	client client.Client    // InfluxDB 클라이언트
}

/*
 * NewInfluxRepo : InfluxRepo 생성자
 *  - fx 프레임워크에 의해 호출되는 생성자 함수
 *  - InfluxDB 클라이언트 설정, EventBus 구독 등록, OnStop 시 client.Close 호출을 설정
 *  - 반환값 : *InfluxRepo (InfluxRepo 객체)
 */
func NewInfluxRepo(lc fx.Lifecycle, log *zap.Logger, eb *bus.EventBus) *InfluxRepo {
	// 환경변수로부터 읽은 InfluxDB 관련 값들
	influxURL := os.Getenv("APP_INFLUX_URL")       // InfluxDB URL
	influxUsername := os.Getenv("APP_INFLUX_USERNAME") // InfluxDB 사용자 이름
	influxPassword := os.Getenv("APP_INFLUX_PASSWORD") // InfluxDB 비밀번호
	influxDatabase := os.Getenv("APP_INFLUX_DATABASE") // InfluxDB 데이터베이스 이름
	influxPrecision := os.Getenv("APP_INFLUX_PRECISION") // InfluxDB 시간 정밀도
	influxTimeout := os.Getenv("APP_INFLUX_TIMEOUT") // InfluxDB 타임아웃 설정

	// 기본값 설정 (환경변수로 설정되지 않으면 기본값을 사용)
	if influxURL == "" {
		influxURL = "http://localhost:8086" // 기본 InfluxDB URL
	}
	if influxUsername == "" {
		influxUsername = "admin" // 기본 사용자 이름
	}
	if influxPassword == "" {
		influxPassword = "" // 기본 비밀번호 (비어 있을 수 있음)
	}
	if influxDatabase == "" {
		log.Fatal("influx database is required") // 데이터베이스는 필수
	}
	if influxPrecision == "" {
		influxPrecision = "s" // 기본 정밀도는 초 단위(s)
	}
	if influxTimeout == "" {
		influxTimeout = "5s" // 기본 타임아웃 5초
	}

	// influxTimeout을 string에서 time.Duration으로 변환
	timeout, err := time.ParseDuration(influxTimeout)
	if err != nil {
		log.Fatal("failed to parse influx timeout", zap.Error(err)) // 변환 실패 시 애플리케이션 종료
	}

	// InfluxDB 클라이언트 생성
	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     influxURL,  // InfluxDB 서버 URL
		Username: influxUsername, // 사용자 이름
		Password: influxPassword, // 비밀번호
		Timeout:  timeout,  // 연결 타임아웃
	})
	if err != nil {
		log.Fatal("failed to connect influxdb", zap.Error(err)) // 연결 실패 시 애플리케이션 종료
	}

	// InfluxRepo 객체 생성
	repo := &InfluxRepo{
		log:    log,
		
		client: c,
	}

	// EventBus의 구독자 함수 등록
	// 수집된 데이터 이벤트가 발생하면 InfluxDB에 데이터를 기록
	eb.Subscribe(func(e bus.DataCollectedEvent) {
		// 배치 포인트 생성 (InfluxDB에 데이터를 한 번에 전송)
		bp, _ := client.NewBatchPoints(client.BatchPointsConfig{
			Database:  influxDatabase,  // 사용할 데이터베이스
			Precision: influxPrecision, // 시간 정밀도
		})

		// 데이터 포인트에 태그 추가 (예: 장치 ID)
		tags := map[string]string{
			"device": e.DeviceID,
		}

		// 수집된 데이터를 필드에 추가 (예: temperature, humidity)
		fields := make(map[string]interface{}, len(e.Values))
		for k, v := range e.Values {
			fields[k] = v
		}

		// 데이터 포인트 생성
		pt, err := client.NewPoint("device_data", tags, fields, time.Now())
		if err != nil {
			repo.log.Error("influx point create failed", zap.Error(err)) // 포인트 생성 실패 시 로그
			return
		}

		// 배치 포인트에 데이터 포인트 추가
		bp.AddPoint(pt)

		// 배치 포인트를 InfluxDB에 기록
		if err := repo.client.Write(bp); err != nil {
			repo.log.Error("influx write failed", zap.Error(err)) // 쓰기 실패 시 로그
			return
		}

		// 성공적인 데이터 기록 로그
		repo.log.Info("influx write success", zap.String("device", e.DeviceID))
	})

	// 애플리케이션 종료 시 클라이언트 연결을 종료하는 후크 등록
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			repo.client.Close()  // InfluxDB 클라이언트 연결 종료
			return nil
		},
	})

	// 생성된 InfluxRepo 객체 반환
	return repo
}
