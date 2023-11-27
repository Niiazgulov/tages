package grpcapp

import (
	"fmt"
	"net"

	imageworkergrpc "github.com/Niiazgulov/tages.git/internal/grpc/imageworker"
	"github.com/Niiazgulov/tages.git/internal/storage"
	"google.golang.org/grpc"
)

type App struct {
	gRPCServer   *grpc.Server
	port         int
	imgProcessor storage.ImageProcessor
	repo         storage.ImageDB
}

func New(port int, imgProcessor storage.ImageProcessor, repo storage.ImageDB) *App {
	gRPCServer := grpc.NewServer()
	imageworkergrpc.Register(gRPCServer, imgProcessor, repo)

	return &App{gRPCServer: gRPCServer, port: port, imgProcessor: imgProcessor, repo: repo}
}

func (a *App) Run() error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		return fmt.Errorf("%w", err)
	}
	if err := a.gRPCServer.Serve(listener); err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
}

func (a *App) Stop() {
	a.gRPCServer.GracefulStop()
}
