package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"io/ioutil"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/grpclog"

	"github.com/thatique/snowman/server"
	"github.com/thatique/snowman/api/v1"
)

var (
	certFile    = flag.String("cert_file", "", "The TLS cert file")
	keyFile     = flag.String("key_file", "", "The TLS key file")
	gRPCPort    = flag.Int("grpc-port", 6996, "The gRPC server port")
	gatewayPort = flag.Int("gateway-port", 6997, "The gRPC-Gateway server port")
)

var (
	machineID int
	log grpclog.LoggerV2
	// this channel gets notified when process receives signal. It is global to ease unit testing
	quit = make(chan os.Signal, 1)
)

func init() {
	log = grpclog.NewLoggerV2(os.Stdout, ioutil.Discard, ioutil.Discard)
	grpclog.SetLoggerV2(log)
	rand.Seed(time.Now().UnixNano())
	machineID = rand.Intn(1023)
}

func main() {
	flag.Parse()
	addr := fmt.Sprintf(":%d", *gRPCPort)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalln("Failed to listen:", err)
	}
	var opts []grpc.ServerOption
	if *certFile == "" || *keyFile == "" {
		log.Info("serving without tls file")
	} else {
		creds, err := credentials.NewServerTLSFromFile(*certFile, *keyFile)
		if err != nil {
			log.Fatalf("Failed to generate credentials %v", err)
		}
		opts = []grpc.ServerOption{grpc.Creds(creds)}
	}
	s := grpc.NewServer(opts...)
	v1.RegisterSnowflakeServiceServer(s, server.New(machineID))
	// Serve gRPC Server
	log.Info("Serving gRPC on", addr)

	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	serveErr := make(chan error)

	go func() {
		serveErr <- s.Serve(lis)
	}()
	select {
	case err := <-serveErr:
		log.Fatalf("Failed to start gRPC server: %v", err)

	case <-quit:
		// shutdown the server with a grace period of configured timeout
		log.Info("stopping gRPC server ")
	}
}