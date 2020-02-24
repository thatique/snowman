package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"time"

	"github.com/gogo/protobuf/types"
	v1 "github.com/thatique/snowman/api/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/grpclog"
)

var (
	addr      = flag.String("addr", "localhost:6996", "The address and port of the server to connect to")
	caCrt     = flag.String("ca-crt", "", "CA certificate")
	clientCrt = flag.String("client-crt", "", "Client certificate")
	clientKey = flag.String("client-key", "", "Client key")
)

var log grpclog.LoggerV2

func init() {
	log = grpclog.NewLoggerV2(os.Stdout, ioutil.Discard, ioutil.Discard)
}

func newSnowmanClient(hostAndPort, caPath, clientCrt, clientKey string) (v1.SnowflakeServiceClient, error) {
	var opts []grpc.DialOption

	if caPath != "" {
		cPool := x509.NewCertPool()
		caCert, err := ioutil.ReadFile(caPath)
		if err != nil {
			return nil, fmt.Errorf("invalid CA crt file: %s", caPath)
		}
		if cPool.AppendCertsFromPEM(caCert) != true {
			return nil, fmt.Errorf("failed to parse CA crt")
		}

		clientCert, err := tls.LoadX509KeyPair(clientCrt, clientKey)
		if err != nil {
			return nil, fmt.Errorf("invalid client crt file: %s", caPath)
		}

		clientTLSConfig := &tls.Config{
			RootCAs:      cPool,
			Certificates: []tls.Certificate{clientCert},
		}
		creds := credentials.NewTLS(clientTLSConfig)
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

	client, err := newSnowmanClient(*addr, *caCrt, *clientCrt, *clientKey)
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
