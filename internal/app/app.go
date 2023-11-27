package app

import (
	grpcapp "github.com/Niiazgulov/tages.git/internal/app/grpc"
	"github.com/Niiazgulov/tages.git/internal/storage"
)

type App struct {
	GRPCServ *grpcapp.App
}

func New(grpcPort int, storagePath string, imgProcessor storage.ImageProcessor, repo storage.ImageDB) *App {
	grpcApp := grpcapp.New(grpcPort, imgProcessor, repo)
	return &App{GRPCServ: grpcApp}
}
