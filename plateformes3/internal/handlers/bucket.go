package handlers

import (
	"encoding/xml"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"plateformes3/config"
	"plateformes3/internal/storage"
	"sync"

	"github.com/gorilla/mux"
)

type Bucket struct {
	Name         string `xml:"Name"`
	CreationDate string `xml:"CreationDate"`
}

type ListAllMyBucketsResult struct {
	XMLName xml.Name `xml:"ListAllMyBucketsResult"`
	Buckets []Bucket `xml:"Buckets>Bucket"`
}

var mu sync.Mutex

func ListBucketsHandler(w http.ResponseWriter, r *http.Request) {
	cfg := config.LoadConfig()

	buckets, err := storage.ListBuckets(cfg.StoragePath)
	if err != nil {
		log.Printf("Failed to list buckets: %s", err)
		http.Error(w, "Unable to list buckets", http.StatusInternalServerError)
		return
	}

	var bucketList []Bucket
	for _, b := range buckets {
		bucketList = append(bucketList, Bucket{
			Name:         b.Name,
			CreationDate: b.CreationDate,
		})
	}

	result := ListAllMyBucketsResult{
		Buckets: bucketList,
	}

	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(http.StatusOK)
	xml.NewEncoder(w).Encode(result)
	log.Println("Listed all buckets successfully")
}

func CreateBucketHandler(w http.ResponseWriter, r *http.Request) {
	cfg := config.LoadConfig()
	vars := mux.Vars(r)
	bucketName := vars["bucketName"]

	mu.Lock()
	defer mu.Unlock()

	log.Printf("Received request to create bucket: %s", bucketName)

	err := storage.CreateBucket(cfg.StoragePath, bucketName)
	if err != nil {
		if os.IsExist(err) {
			log.Printf("Bucket %s already exists", bucketName)
			http.Error(w, "Bucket already exists", http.StatusConflict)
		} else {
			log.Printf("Failed to create bucket: %s", err)
			http.Error(w, "Unable to create bucket", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`<CreateBucketConfiguration><LocationConstraint>us-east-1</LocationConstraint></CreateBucketConfiguration>`))
	log.Printf("Bucket %s created successfully", bucketName)
}

func DeleteBucketHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bucketName := vars["bucketName"]
	cfg := config.LoadConfig()

	log.Printf("Received request to delete bucket: %s", bucketName)

	bucketPath := cfg.StoragePath + "/" + bucketName

	if _, err := os.Stat(bucketPath); os.IsNotExist(err) {
		log.Printf("Bucket %s does not exist", bucketName)
		http.Error(w, "Bucket does not exist", http.StatusNotFound)
		return
	}

	err := os.RemoveAll(bucketPath)
	if err != nil {
		log.Printf("Failed to delete bucket: %s", err)
		http.Error(w, "Unable to delete bucket", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
	log.Printf("Bucket %s deleted successfully", bucketName)
}

func UpdateBucketNameHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	oldBucketName := vars["bucketName"]
	newBucketName := r.URL.Query().Get("newName")

	if newBucketName == "" {
		http.Error(w, "New bucket name is required", http.StatusBadRequest)
		return
	}

	oldBucketPath := "./data/buckets/" + oldBucketName
	newBucketPath := "./data/buckets/" + newBucketName

	if _, err := os.Stat(oldBucketPath); os.IsNotExist(err) {
		http.Error(w, "Bucket does not exist", http.StatusNotFound)
		return
	}

	if _, err := os.Stat(newBucketPath); !os.IsNotExist(err) {
		http.Error(w, "Bucket with new name already exists", http.StatusConflict)
		return
	}

	err := os.Rename(oldBucketPath, newBucketPath)
	if err != nil {
		log.Printf("Failed to rename bucket: %v", err)
		http.Error(w, "Unable to rename bucket", http.StatusInternalServerError)
		return
	}

	log.Printf("Bucket renamed from %s to %s", oldBucketName, newBucketName)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Bucket renamed successfully"))
}

// Object handlers

func DeleteObjectHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bucketName := vars["bucketName"]
	objectName := vars["objectName"]

	err := storage.DeleteObject(bucketName, objectName)
	if err != nil {
		http.Error(w, "Unable to delete object", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func UploadObjectHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bucketName := vars["bucketName"]

	// Récupérer le fichier depuis la requête
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Unable to get file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Enregistrer l'objet
	err = storage.SaveObject(bucketName, header.Filename, file)
	if err != nil {
		http.Error(w, "Unable to upload object", http.StatusInternalServerError)
		return
	}

	w.Write([]byte("Object uploaded successfully"))
}

// GetObject récupère un fichier d'un bucket
func GetObject(bucketName, objectName string) (io.ReadCloser, error) {
	objectPath := filepath.Join("./data/buckets", bucketName, objectName)
	return os.Open(objectPath)
}
