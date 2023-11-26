package storage

import (
	"bytes"
	"context"
	"fmt"
	"os"

	"strings"
	"sync"

	"github.com/google/uuid"

	pb "github.com/Niiazgulov/tages.git/protos/gen/go"
)

type ImageProcessor interface {
	SaveNewImage(img bytes.Buffer, filename string) (string, error)
	ImagesView(ctx context.Context) ([]pb.ImageInfo, error)           // TODO: IMPLEMENT
	SendBack(ctx context.Context, imgId string) (bytes.Buffer, error) // TODO: IMPLEMENT
}

type DiskImageStore struct {
	mutex       sync.RWMutex
	imageFolder string
	images      map[string]*pb.ImageInfo
}

func NewDiskImageStore(imageFolder string) *DiskImageStore {
	return &DiskImageStore{
		imageFolder: imageFolder,
		images:      make(map[string]*pb.ImageInfo),
	}
}

func (store *DiskImageStore) ImagesView(ctx context.Context) ([]pb.ImageInfo, error) {
	return nil, nil
}

func (store *DiskImageStore) SendBack(ctx context.Context, imgId string) (bytes.Buffer, error) {
	return bytes.Buffer{}, nil
}

func (store *DiskImageStore) SaveNewImage(img bytes.Buffer, filename string) (string, error) {
	imageID, err := uuid.NewRandom()
	if err != nil {
		return "", fmt.Errorf("cannot generate image id: %w", err)
	}
	files := filesInFolder(store.imageFolder)
	_, ok := files[filename]
	if ok {
		return "", fmt.Errorf("image already exists on the server! %w", err)
	}

	imagePath := strings.Join([]string{store.imageFolder, filename}, "/")

	file, err := os.Create(imagePath)
	if err != nil {
		return "", fmt.Errorf("ich cannot create image file: %w", err)
	}

	_, err = img.WriteTo(file)
	if err != nil {
		return "", fmt.Errorf("cannot write image to file: %w", err)
	}

	store.mutex.Lock()
	defer store.mutex.Unlock()

	store.images[imageID.String()] = &pb.ImageInfo{
		Path: imagePath,
	}

	return imageID.String(), nil
}

func filesInFolder(imagesFolderPath string) map[string]int {
	dir, err := os.Open(imagesFolderPath)
	if err != nil {
		return nil
	}
	defer dir.Close()

	files, err := dir.Readdir(-1)
	if err != nil {
		return nil
	}
	filenames := make(map[string]int)
	for i, file := range files {
		filenames[file.Name()] = i
	}

	return filenames
}
