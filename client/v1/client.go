package client

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"

	"github.com/gogo/protobuf/types"
	v1 "github.com/thatique/snowman/api/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// SnowmanClient is a client to Snowflake ID generator server
type SnowmanClient struct {
	c v1.SnowflakeServiceClient
}

// SnowmanCursor is cursor for iterating batch ID request
type SnowmanCursor struct {
	c v1.SnowflakeService_BatchNextIDClient
}

// Next get the next available ID
func (client *SnowmanCursor) Next() (v1.ID, error) {
	snowflake, err := client.c.Recv()
	if err != nil {
		return v1.ID(0), err
	}

	return snowflake.ID, nil
}

// NewSnowmanClient create snowflake client
func NewSnowmanClient(hostAndPort, caPath, clientCrt, clientKey string) (*SnowmanClient, error) {
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
	} else {
		opts = append(opts, grpc.WithInsecure())
	}
	conn, err := grpc.Dial(hostAndPort, opts...)
	if err != nil {
		return nil, fmt.Errorf("dial: %v", err)
	}
	return &SnowmanClient{c: v1.NewSnowflakeServiceClient(conn)}, nil
}

// NextID get the nextID
func (client *SnowmanClient) NextID(ctx context.Context) (v1.ID, error) {
	snowflake, err := client.c.NextID(ctx, &types.Empty{})
	if err != nil {
		return v1.ID(0), err
	}

	return snowflake.ID, nil
}

// NextBatchIDs get many batch at once
func (client *SnowmanClient) NextBatchIDs(ctx context.Context, length int) (*SnowmanCursor, error) {
	srv, err := client.c.BatchNextID(ctx, &v1.BatchIDsRequest{Length: int32(length)})
	if err != nil {
		return nil, err
	}

	return &SnowmanCursor{c: srv}, nil
}
