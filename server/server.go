package server

import (
	"errors"
	"context"

	"github.com/thatique/snowman/api/v1"
	"github.com/gogo/protobuf/types"
)

var _ v1.SnowflakeServiceServer = (*Server)(nil)

type Server struct {
	gen *Generator
}

func New(machineID int) *Server {
	return &Server{gen: NewGenerator(machineID)}
}

func (s *Server) NextID(ctx context.Context, _ *types.Empty) (*v1.Snowflake, error) {
	id := s.gen.Next()
	return &v1.Snowflake{ID: v1.ID(id)}, nil
}

func (s *Server) BatchNextID(req *v1.BatchIDsRequest, srv v1.SnowflakeService_BatchNextIDServer) error {
	len := int(req.GetLength())
	if len <= 0 {
		return errors.New("length can't be zero or negative")
	}
	var (
		id uint64
		snowflake *v1.Snowflake
	)
	for i := 0; i < len; i++ {
		id = s.gen.Next()
		snowflake = &v1.Snowflake{ID: v1.ID(id)}
		err := srv.Send(snowflake)
		if err != nil {
			return err
		}
	}
	return nil
}