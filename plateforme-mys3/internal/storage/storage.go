// internal/storage/storage.go
package storage

import (
	"io"
	"log"
	"os"
	"path/filepath"
)

// Storage représente le stockage des buckets
type Storage struct {
	BasePath string
}

// NewStorage initialise le stockage avec le chemin de base spécifié
func NewStorage(basePath string) *Storage {
	return &Storage{BasePath: basePath}
}

// BucketPath retourne le chemin complet d'un bucket
func (s *Storage) BucketPath(bucketName string) string {
	return filepath.Join(s.BasePath, bucketName)
}

// ObjectPath retourne le chemin complet d'un objet dans un bucket
func (s *Storage) ObjectPath(bucketName, objectName string) string {
	return filepath.Join(s.BasePath, bucketName, objectName)
}

// CreateBucket crée un nouveau bucket en créant un dossier
func (s *Storage) CreateBucket(bucketName string) error {
	bucketPath := s.BucketPath(bucketName)
	log.Printf("Tentative de création du bucket à l'emplacement : %s", bucketPath)

	// Vérifier si le bucket existe déjà
	if _, err := os.Stat(bucketPath); !os.IsNotExist(err) {
		log.Printf("Le bucket %s existe déjà ou une erreur est survenue : %v", bucketName, err)
		return err
	}

	// Créer le dossier du bucket
	err := os.MkdirAll(bucketPath, 0755)
	if err != nil {
		log.Printf("Erreur lors de la création du dossier du bucket %s : %v", bucketName, err)
		return err
	}

	log.Printf("Bucket %s créé avec succès à l'emplacement : %s", bucketName, bucketPath)
	return nil
}

// DeleteBucket supprime un bucket en supprimant son dossier
func (s *Storage) DeleteBucket(bucketName string) error {
	return os.RemoveAll(s.BucketPath(bucketName))
}

// ListBuckets liste tous les buckets existants
func (s *Storage) ListBuckets() ([]os.FileInfo, error) {
	dirEntries, err := os.ReadDir(s.BasePath)
	if err != nil {
		return nil, err
	}

	var fileInfos []os.FileInfo
	for _, entry := range dirEntries {
		if entry.IsDir() {
			info, err := entry.Info()
			if err != nil {
				continue
			}
			fileInfos = append(fileInfos, info)
		}
	}
	return fileInfos, nil
}

// PutObject ajoute un objet dans un bucket
func (s *Storage) PutObject(bucketName, objectName string, data io.Reader) error {
	objectPath := s.ObjectPath(bucketName, objectName)
	dir := filepath.Dir(objectPath)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}

	file, err := os.Create(objectPath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, data)
	return err
}

// GetObject récupère un objet depuis un bucket
func (s *Storage) GetObject(bucketName, objectName string) (*os.File, error) {
	return os.Open(s.ObjectPath(bucketName, objectName))
}

// DeleteObject supprime un objet depuis un bucket
func (s *Storage) DeleteObject(bucketName, objectName string) error {
	return os.Remove(s.ObjectPath(bucketName, objectName))
}

// ListObjects liste tous les objets dans un bucket
func (s *Storage) ListObjects(bucketName string) ([]os.FileInfo, error) {
	dirEntries, err := os.ReadDir(s.BucketPath(bucketName))
	if err != nil {
		return nil, err
	}

	var fileInfos []os.FileInfo
	for _, entry := range dirEntries {
		if !entry.IsDir() {
			info, err := entry.Info()
			if err != nil {
				continue
			}
			fileInfos = append(fileInfos, info)
		}
	}
	return fileInfos, nil
}
