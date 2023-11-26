package client

import (
	"bufio"
	"context"
	"io"
	"log"
	"os"
	"time"

	"google.golang.org/grpc"

	pb "github.com/Niiazgulov/tages.git/protos/gen/go"
)

type imgClient struct {
	service pb.ImageWorkerClient
}

func NewImgWorkerClient(cc *grpc.ClientConn) *imgClient {
	service := pb.NewImageWorkerClient(cc)
	return &imgClient{service}
}

func (imgClient *imgClient) UploadImage(imagePath string, filename string) {

	file, err := os.Open(imagePath)
	if err != nil {
		log.Fatal("cannot open image file: ", err)
	}
	defer file.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stream, err := imgClient.service.UploadImage(ctx)
	if err != nil {
		log.Fatal("cannot upload image: ", err)
	}

	reader := bufio.NewReader(file)
	buffer := make([]byte, 1024)

	for {
		n, err := reader.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal("cannot read chunk to buffer: ", err)
		}

		req := &pb.UploadRequest{
			ImageData: buffer[:n],
			Filename:  filename,
		}
		err = stream.Send(req)
		if err != nil {
			log.Fatal("cannot send chunk to server: ", err)
		}
	}

	res, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatal("cannot receive response: ", err)
	}

	log.Printf("image %s uploaded with id: %s, at: %s", filename, res.GetImageId(), res.GetCreatedAt())
}

func (imgClient *imgClient) InformImage(imagePath string) {
	// TODO: IMPLEMENT
}

func (imgClient *imgClient) DownloadImage(imagePath string) {
	// TODO: IMPLEMENT
}
