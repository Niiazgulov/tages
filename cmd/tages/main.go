package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Niiazgulov/tages.git/internal/app"
	"github.com/Niiazgulov/tages.git/internal/config"
	"github.com/Niiazgulov/tages.git/internal/storage"
)

func main() {
	cfg := config.MustLoad()
	imageStore := storage.NewDiskImageStore(cfg.StoragePath)
	repo, err := storage.NewDB(cfg.DBPath)
	if err != nil {
		log.Fatal(err)
	}
	appl := app.New(cfg.GRPC.Port, cfg.StoragePath, imageStore, repo)
	go appl.GRPCServ.Run()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	<-stop
	appl.GRPCServ.Stop()
}
