package grpc

import (
	"context"
	"net"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	"github.com/senyabanana/pvz-service/internal/service"
	pbv1 "github.com/senyabanana/pvz-service/pkg/pb/pvz_v1"
)

const tcpNetwork = "tcp"

type GRPCServer struct {
	server   *grpc.Server
	listener net.Listener
	log      *logrus.Logger
}

func NewGRPCServer(port string, pvzService service.PVZOperations, log *logrus.Logger) (*GRPCServer, error) {
	listener, err := net.Listen(tcpNetwork, ":"+port)
	if err != nil {
		log.Errorf("failed to listen on port %s: %v", port, err)
		return nil, err
	}

	grpcServer := grpc.NewServer()
	pbv1.RegisterPVZServiceServer(grpcServer, NewPVZGRPCHandler(pvzService))

	return &GRPCServer{
		server:   grpcServer,
		listener: listener,
		log:      log,
	}, nil
}

func (s *GRPCServer) Run() error {
	s.log.Infof("Starting gRPC server on %s", s.listener.Addr().String())
	return s.server.Serve(s.listener)
}

func (s *GRPCServer) Shutdown(ctx context.Context) {
	s.log.Info("Gracefully stopping gRPC server...")
	s.server.GracefulStop()
}
