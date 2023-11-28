package storage

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

var (
	ErrImgNotFound = errors.New("image not found")
)

type ImageProcessor interface {
	SaveNewImage(img bytes.Buffer, newImage ImagesInfo, repo ImageDB) (string, error)
	ImagesView(repo ImageDB) ([]ImagesInfo, error)
	GetImage(filename string) ([]byte, error)
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

func (store *DiskImageStore) SaveNewImage(img bytes.Buffer, newImage ImagesInfo, repo ImageDB) (string, error) {
	imageID, err := uuid.NewRandom()
	if err != nil {
		return "", fmt.Errorf("cannot generate image id: %w", err)
	}
	newImage.ImageId = imageID.String()

	imagePath := strings.Join([]string{store.imageFolder, newImage.Filename}, "/")

	files := filesInFolderMap(store.imageFolder)
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

func (store *DiskImageStore) ImagesView(repo ImageDB) ([]ImagesInfo, error) {
	files := filesInFolderSlice(store.imageFolder)
	if files == nil {
		return nil, fmt.Errorf("storage is empty - no files here: %v", ErrImgNotFound)
	}

	records, err := repo.GetAllInfo(files)
	if err != nil {
		return nil, fmt.Errorf("cannot download images info from db: %w", err)
	}

	return records, nil
}

func (store *DiskImageStore) GetImage(filename string) ([]byte, error) {
	files := filesInFolderMap(store.imageFolder)
	if files == nil {
		return nil, fmt.Errorf("storage is empty - no files here: %v", ErrImgNotFound)
	}

	_, ok := files[filename]
	if !ok {
		return nil, fmt.Errorf("cannot find such image in storage: %v", ErrImgNotFound)
	}

	imagePath := strings.Join([]string{store.imageFolder, filename}, "/")

	file, err := os.Open(imagePath)
	if err != nil {
		return nil, fmt.Errorf("cannot open file: %v", err)
	}
	defer file.Close()

	byteImg, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("cannot io.readall file to byte: %v", err)
	}

	return byteImg, nil
}

func filesInFolderMap(imagesFolderPath string) map[string]int {
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

func filesInFolderSlice(imagesFolderPath string) []string {
	dir, err := os.Open(imagesFolderPath)
	if err != nil {
		return nil
	}
	defer dir.Close()

	files, err := dir.Readdir(-1)
	if err != nil {
		return nil
	}

	filenames := []string{}
	for _, file := range files {
		filenames = append(filenames, file.Name())
	}
	return filenames
}
