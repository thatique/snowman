package main

import (
	"context"
	"flag"
	"io"
	"io/ioutil"
	"os"
	"time"

	"github.com/thatique/snowman/client/v1"
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

func main() {
	flag.Parse()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	c, err := client.NewSnowmanClient(*addr, *caCrt, *clientCrt, *clientKey)
	if err != nil {
		log.Fatalln("Failed create connection to:", *addr)
	}
	id, err := c.NextID(ctx)
	if err != nil {
		log.Fatalln("Failed to get next ID 🤔:", err)
	}
	log.Infoln("Read ID:", id)

	cursor, err := c.NextBatchIDs(ctx, 10)
	if err != nil {
		log.Fatalln("Failed to get batch of ID:", err)
	}
	for {
		id, err = cursor.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatalln("Failed to receive:", err)
		}
		log.Infoln("Read ID:", id)
	}
	log.Infoln("Success!")
}
