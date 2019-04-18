package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"time"

	"github.com/gogo/protobuf/types"
	"github.com/thatique/snowman/api/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/grpclog"
)

var (
	addr     = flag.String("addr", "localhost:6996", "The address and port of the server to connect to")
	clientCa = flag.String("client_ca", "", "The TLS client CA")
)

var log grpclog.LoggerV2

func init() {
	log = grpclog.NewLoggerV2(os.Stdout, ioutil.Discard, ioutil.Discard)
}

func newSnowmanClient(hostAndPort, caPath string) (v1.SnowflakeServiceClient, error) {
	var opts []grpc.DialOption

	if caPath != "" {
		creds, err := credentials.NewClientTLSFromFile(caPath, "")
		if err != nil {
			return nil, fmt.Errorf("load cert: %v", err)
		}
		opts = append(opts, grpc.WithTransportCredentials(creds))
		log.Infoln("create connection with TSL transport creds")
	} else {
		opts = append(opts, grpc.WithInsecure())
	}
	conn, err := grpc.Dial(hostAndPort, opts...)
	if err != nil {
		return nil, fmt.Errorf("dial: %v", err)
	}
	return v1.NewSnowflakeServiceClient(conn), nil
}

func main() {
	flag.Parse()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client, err := newSnowmanClient(*addr, *clientCa)
	if err != nil {
		log.Fatalln("Failed create connection to:", *addr)
	}
	snowflake, err := client.NextID(ctx, &types.Empty{})
	if err != nil {
		log.Fatalln("Failed to get next ID 🤔:", err)
	}
	log.Infoln("Read ID:", snowflake.ID)

	srv, err := client.BatchNextID(ctx, &v1.BatchIDsRequest{Length: 10})
	if err != nil {
		log.Fatalln("Failed to get batch of ID:", err)
	}
	for {
		snowflake, err = srv.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatalln("Failed to receive:", err)
		}
		log.Infoln("Read ID:", snowflake.ID)
	}
	log.Infoln("Success!")
}
