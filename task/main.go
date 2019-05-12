package main

import (
	"context"
	"net"

	"os"

	"log"

	"time"

	"os/signal"
	"syscall"

	pbTask "sample-grpc-micro/proto/task"
	"sample-grpc-micro/shared/interceptor"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"google.golang.org/grpc"
)

const port = ":60000"

func main() {
	// インタセプタの追加
	chain := grpc_middleware.ChainUnaryServer(
		interceptor.XTraceID(),
		interceptor.Logging(),
		interceptor.XUserID(),
	)
	srvOpt := grpc.UnaryInterceptor(chain)
	srv := grpc.NewServer(srvOpt)
	// サービスの登録
	pbTask.RegisterTaskServiceServer(srv, &TaskService{
		store: NewStoreOnMemory(),
	})
	// gRPC接続の待ち受け
	go func() {
		listener, err := net.Listen("tcp", port)
		if err != nil {
			log.Fatalf("failed to create listener: %s",
				err)
		}
		log.Println("start server on port", port)
		if err := srv.Serve(listener); err != nil {
			log.Println("failed to exit serve: ", err)
		}
	}()
	// グレースフルストップ
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, syscall.SIGTERM)
	<-sigint
	log.Println("received a signal of graceful shutdown")
	stopped := make(chan struct{})
	go func() {
		srv.GracefulStop()
		close(stopped)
	}()
	ctx, cancel := context.WithTimeout(
		context.Background(), 1*time.Minute)
	select {
	case <-ctx.Done():
		srv.Stop()
	case <-stopped:
		cancel()
	}
	log.Println("completed graceful shutdown")
}
