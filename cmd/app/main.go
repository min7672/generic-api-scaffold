package main

import (
	"log"
	"context"
	"os/signal"
	"syscall"      // 실제 신호 상수들을 제공
	"github.com/joho/godotenv"
	"generic-api-scaffold/internal/app" 
)

func main() {
		// .env 파일 로드
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	
	/* func NotifyContext(parent context.Context, signals ...os.Signal) : OS 신호를 감지하는 새로운 컨텍스트 생성 */
	 
	ctx, stop := signal.NotifyContext( context.Background(), syscall.SIGINT, syscall.SIGTERM )

	/*
	 * defer : 25개 예약어중 지정한 함수를 현재 함수의 실해잉 끝날때까지 지연 시키는 문법
	 * stop : go 표준라이브러리 
	*/
	defer stop()

	app.Run(ctx)
}
