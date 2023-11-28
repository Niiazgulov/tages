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

	log.Printf("image %s uploaded at: %s", filename, res.GetCreatedAt())
}

func (imgClient *imgClient) InformImage() (*pb.InformResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stream, err := imgClient.service.InformImage(ctx)
	if err != nil {
		log.Fatal("cannot call inform_image method: ", err)
	}

	req := &pb.InformRequest{}
	err = stream.Send(req)
	if err != nil {
		log.Fatal("cannot send request to server: ", err)
	}

	resp, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatal("I cannot receive response: ", err)
	}

	return resp, nil
}

func (imgClient *imgClient) DownloadImage(filename string) (*pb.DownloadResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stream, err := imgClient.service.DownloadImage(ctx)
	if err != nil {
		log.Fatal("cannot call download_image method: ", err)
	}

	req := &pb.DownloadRequest{Filename: filename}
	err = stream.Send(req)
	if err != nil {
		log.Fatal("cannot send request to server: ", err)
	}

	resp, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatal("I cannot receive response: ", err)
	}

	return resp, nil
}
