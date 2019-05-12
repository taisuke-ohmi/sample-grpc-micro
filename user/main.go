package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"sample-grpc-micro/shared/interceptor"
	"syscall"
	"time"

	pb "sample-grpc-micro/proto/user"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"google.golang.org/grpc"
)

const port = ":60000"

func main() {
	srv := grpc.NewServer(grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
		interceptor.XTraceID(),
		interceptor.Logging(),
	)))
	pb.RegisterUserServiceServer(srv, &UserService{store: NewStoreOnMemory()})
	go func() {
		listener, err := net.Listen("tcp", port)
		if err != nil {
			log.Fatalf("failed to create listener: %s", err)
		}
		log.Println("start server on port", port)
		if err := srv.Serve(listener); err != nil {
			log.Println("failed to exit serve: ", err)
		}
	}()
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, syscall.SIGTERM)
	// SIGTERMを受信するまでブロックする
	<-sigint
	log.Println("received a signal of graceful shutdown")
	// grpcの接続を途中で止めないように、グレースフルストップで全てのやりとりが
	// 完了するまでチャネルをcloseしないようにする
	stopped := make(chan struct{})
	go func() {
		srv.GracefulStop()
		close(stopped)
	}()
	// 1分を限度に残りのgrpc接続を待って終了する
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	select {
	case <-ctx.Done():
		srv.Stop()
	case <-stopped:
		cancel()
	}
	log.Println("completed graceful shutdown")
}
