package imageworker

import (
	"bytes"
	"context"
	"io"
	"log"
	"time"

	"github.com/Niiazgulov/tages.git/internal/storage"
	pb "github.com/Niiazgulov/tages.git/protos/gen/go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type serverAPI struct {
	pb.UnimplementedImageWorkerServer
	imgProcessor storage.ImageProcessor
	repo         storage.ImageDB
}

func Register(gRPCServer *grpc.Server, imgProcessor storage.ImageProcessor, repo storage.ImageDB) {
	pb.RegisterImageWorkerServer(gRPCServer, &serverAPI{imgProcessor: imgProcessor, repo: repo})
}

func logError(err error) error {
	if err != nil {
		log.Print(err)
	}
	return err
}

const maxImageSize = 1 << 20

func (server *serverAPI) UploadImage(stream pb.ImageWorker_UploadImageServer) error {
	limiter := make(chan struct{}, 10)
	limiter <- struct{}{}

	var newImage storage.ImagesInfo
	imageData := bytes.Buffer{}
	imageSize := 0

	for {
		err := contextError(stream.Context())
		if err != nil {
			return err
		}

		req, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return logError(status.Errorf(codes.Unknown, "cannot receive chunk data: %v", err))
		}

		chunk := req.GetImageData()
		size := len(chunk)
		newImage.Filename = req.GetFilename()

		imageSize += size
		if imageSize > maxImageSize {
			return logError(status.Errorf(codes.InvalidArgument, "image is too large: %d > %d", imageSize, maxImageSize))
		}

		_, err = imageData.Write(chunk)
		if err != nil {
			return logError(status.Errorf(codes.Internal, "cannot write chunk data: %v", err))
		}
	}
	var err error
	newImage.CreatedAt = time.Now().Format(time.RFC850)
	newImage.ImageId, err = server.imgProcessor.SaveNewImage(imageData, newImage, server.repo)
	if err != nil {
		return logError(status.Errorf(codes.Internal, "cannot save image to the store: %v", err))
	}

	<-limiter

	res := &pb.UploadResponse{
		ImageId:   newImage.ImageId,
		Filename:  newImage.Filename,
		CreatedAt: newImage.CreatedAt,
	}

	err = stream.SendAndClose(res)
	if err != nil {
		return logError(status.Errorf(codes.Unknown, "cannot send response: %v", err))
	}

	log.Printf("saved image %s at: %s", newImage.Filename, newImage.CreatedAt)
	return nil
}

func contextError(ctx context.Context) error {
	switch ctx.Err() {
	case context.Canceled:
		return logError(status.Error(codes.Canceled, "request is canceled"))
	case context.DeadlineExceeded:
		return logError(status.Error(codes.DeadlineExceeded, "deadline is exceeded"))
	default:
		return nil
	}
}

func (s *serverAPI) InformImage(stream pb.ImageWorker_InformImageServer) error {
	limiter := make(chan struct{}, 100)
	limiter <- struct{}{}

	err := contextError(stream.Context())
	if err != nil {
		return err
	}

	_, err = stream.Recv()
	if err != nil {
		return logError(status.Errorf(codes.Unknown, "cannot receive request: %v", err))
	}

	records, err := s.imgProcessor.ImagesView(s.repo)
	if err != nil {
		return logError(status.Errorf(codes.Unknown, "cannot send resquest to get image info (server): %v", err))
	}

	<-limiter

	var resp []*pb.InfoSlice
	for _, v := range records {
		entry := &pb.InfoSlice{Value: []string{v.Filename, v.CreatedAt, v.ChangedAt}}
		resp = append(resp, entry)
	}

	res := &pb.InformResponse{
		Response: resp,
	}

	err = stream.SendAndClose(res)
	if err != nil {
		return logError(status.Errorf(codes.Unknown, "cannot send response: %v", err))
	}

	log.Println("all image's info successfully sended to client")

	return nil
}

func (s *serverAPI) DownloadImage(stream pb.ImageWorker_DownloadImageServer) error {
	limiter := make(chan struct{}, 10)
	limiter <- struct{}{}

	err := contextError(stream.Context())
	if err != nil {
		return err
	}

	req, err := stream.Recv()
	if err != nil {
		return logError(status.Errorf(codes.Unknown, "cannot receive chunk data: %v", err))
	}

	filename := req.GetFilename()
	img, err := s.imgProcessor.GetImage(filename)
	if err != nil {
		return logError(status.Errorf(codes.Unknown, "cannot send resquest to get image (server): %v", err))
	}

	<-limiter

	res := &pb.DownloadResponse{
		ImageData: img,
	}

	err = stream.SendAndClose(res)
	if err != nil {
		return logError(status.Errorf(codes.Unknown, "cannot send response: %v", err))
	}

	log.Println("image successfully sended to client")

	return nil
}
