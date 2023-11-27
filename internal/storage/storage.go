package storage

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

type ImageProcessor interface {
	SaveNewImage(img bytes.Buffer, newImage ImagesInfo, repo ImageDB) (string, error)
	ImagesView(repo ImageDB) ([]ImagesInfo, error)
	SendBack(ctx context.Context, imgId string) (bytes.Buffer, error) // TODO: IMPLEMENT
}

type DiskImageStore struct {
	mutex       sync.RWMutex
	imageFolder string
}

type ImagesInfo struct {
	ImageId   string
	Filename  string
	CreatedAt string
	ChangedAt string
}

func NewDiskImageStore(imageFolder string) *DiskImageStore {
	return &DiskImageStore{
		imageFolder: imageFolder,
	}
}

func (store *DiskImageStore) ImagesView(repo ImageDB) ([]ImagesInfo, error) {
	records, err := repo.GetAllInfo()
	if err != nil {
		return nil, fmt.Errorf("cannot download images info from db: %w", err)
	}

	return records, nil
}

func (store *DiskImageStore) SendBack(ctx context.Context, imgId string) (bytes.Buffer, error) {

	return bytes.Buffer{}, nil
}

func (store *DiskImageStore) SaveNewImage(img bytes.Buffer, newImage ImagesInfo, repo ImageDB) (string, error) {
	imageID, err := uuid.NewRandom()
	if err != nil {
		return "", fmt.Errorf("cannot generate image id: %w", err)
	}
	newImage.ImageId = imageID.String()

	imagePath := strings.Join([]string{store.imageFolder, newImage.Filename}, "/")

	files := filesInFolder(store.imageFolder)
	_, ok := files[newImage.Filename]
	if ok || files == nil {
		err := os.Remove(imagePath)
		if err != nil {
			return "", fmt.Errorf("cannot delete old image: %v", err)
		}

		file, err := os.Create(imagePath)
		if err != nil {
			return "", fmt.Errorf("cannot create image file: %w", err)
		}

		_, err = img.WriteTo(file)
		if err != nil {
			return "", fmt.Errorf("cannot write image to file: %w", err)
		}

		store.mutex.Lock()
		defer store.mutex.Unlock()
		newImage.ChangedAt = time.Now().Format(time.RFC850)
		oldImageId, err := repo.UpdateInfo(newImage)
		if err != nil {
			return "", fmt.Errorf("cannot save image info to the DB: %v", err)
		}

		return oldImageId, nil
	}

	file, err := os.Create(imagePath)
	if err != nil {
		return "", fmt.Errorf("cannot create image file: %w", err)
	}

	_, err = img.WriteTo(file)
	if err != nil {
		return "", fmt.Errorf("cannot write image to file: %w", err)
	}

	store.mutex.Lock()
	defer store.mutex.Unlock()
	newImage.ChangedAt = newImage.CreatedAt
	err = repo.SaveNewInfo(newImage)
	if err != nil {
		return "", fmt.Errorf("save image info to the DB: %w", err)
	}

	return newImage.ImageId, nil
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
