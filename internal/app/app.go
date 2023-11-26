package app

import (
	grpcapp "github.com/Niiazgulov/tages.git/internal/app/grpc"
	"github.com/Niiazgulov/tages.git/internal/storage"
)

type App struct {
	GRPCServ *grpcapp.App
}

func New(grpcPort int, storagePath string, imgProcessor storage.ImageProcessor) *App {
	grpcApp := grpcapp.New(grpcPort, imgProcessor)
	return &App{GRPCServ: grpcApp}
}
