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
}

func Register(gRPCServer *grpc.Server, imgProcessor storage.ImageProcessor) {
	pb.RegisterImageWorkerServer(gRPCServer, &serverAPI{imgProcessor: imgProcessor})
}

func logError(err error) error {
	if err != nil {
		log.Print(err)
	}
	return err
}

const maxImageSize = 1 << 20

func (server *serverAPI) UploadImage(stream pb.ImageWorker_UploadImageServer) error {
	imageData := bytes.Buffer{}
	imageSize := 0
	var filename string

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
		filename = req.GetFilename()

		imageSize += size
		if imageSize > maxImageSize {
			return logError(status.Errorf(codes.InvalidArgument, "image is too large: %d > %d", imageSize, maxImageSize))
		}

		_, err = imageData.Write(chunk)
		if err != nil {
			return logError(status.Errorf(codes.Internal, "cannot write chunk data: %v", err))
		}
	}

	imageID, err := server.imgProcessor.SaveNewImage(imageData, filename)
	if err != nil {
		return logError(status.Errorf(codes.Internal, "cannot save image to the store: %v", err))
	}

	createdTime := time.Now().Format(time.RFC850)
	res := &pb.UploadResponse{
		ImageId:   imageID,
		CreatedAt: createdTime,
	}

	err = stream.SendAndClose(res)
	if err != nil {
		return logError(status.Errorf(codes.Unknown, "cannot send response: %v", err))
	}

	log.Printf("saved image %s with id: %s, at: %s", filename, imageID, createdTime)
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

func (s *serverAPI) Inform(
	ctx context.Context,
	in *pb.InformRequest,
) (*pb.InformResponse, error) {
	panic("implement me")
}

func (s *serverAPI) Download(
	ctx context.Context,
	in *pb.DownloadRequest,
) (*pb.DownloadResponse, error) {
	panic("implement me")
}
