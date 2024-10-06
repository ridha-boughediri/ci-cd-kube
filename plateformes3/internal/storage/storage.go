package storage

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"time"
)

func CreateBucket(basePath, bucketName string) error {
	path := filepath.Join(basePath, bucketName)
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		return errors.New("bucket already exists")
	}
	return os.MkdirAll(path, 0755)
}

func DeleteBucket(basePath, bucketName string) error {
	path := filepath.Join(basePath, bucketName)
	return os.RemoveAll(path)
}

type Bucket struct {
	Name         string
	CreationDate string
}

func ListBuckets(basePath string) ([]Bucket, error) {
	var buckets []Bucket

	// Lire le répertoire des buckets
	entries, err := os.ReadDir(basePath)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			info, err := entry.Info()
			if err != nil {
				continue
			}
			buckets = append(buckets, Bucket{
				Name:         entry.Name(),
				CreationDate: info.ModTime().Format(time.RFC3339),
			})
		}
	}
	return buckets, nil
}

func SaveObject(bucketName, objectName string, file io.Reader) error {
	bucketPath := filepath.Join("./data/buckets", bucketName)

	objectPath := filepath.Join(bucketPath, objectName) // Utiliser le nom de l'objet passé

	outFile, err := os.Create(objectPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, file)
	return err
}

// GetObject récupère un fichier d'un bucket
func GetObject(bucketName, objectName string) (io.ReadCloser, error) {
	objectPath := filepath.Join("./data/buckets", bucketName, objectName)
	return os.Open(objectPath)
}

// DeleteObject supprime un fichier d'un bucket
func DeleteObject(bucketName, objectName string) error {
	objectPath := filepath.Join("./data/buckets", bucketName, objectName)
	return os.Remove(objectPath)
}
