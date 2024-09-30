// internal/handlers/bucket.go
package handlers

import (
	"encoding/xml"
	"log" // Importer le package log
	"net/http"
	"plateforme-mys3/internal/dto"
	"plateforme-mys3/internal/storage"
	"time"

	"github.com/gorilla/mux"
)

// ListBucketsHandler gère la liste de tous les buckets
func ListBucketsHandler(s *storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("ListBucketsHandler appelé") // Log ajouté

		if r.Method != http.MethodGet {
			log.Printf("Méthode non autorisée: %s", r.Method)
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		bucketsInfo, err := s.ListBuckets()
		if err != nil {
			log.Printf("Erreur lors de la liste des buckets: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		var buckets []dto.Bucket
		for _, info := range bucketsInfo {
			buckets = append(buckets, dto.Bucket{
				Name:         info.Name(),
				CreationDate: info.ModTime().Format(time.RFC3339),
			})
		}

		response := dto.ListAllMyBucketsResult{
			XMLNS: "http://s3.amazonaws.com/doc/2006-03-01/",
			Owner: dto.Owner{
				ID:          "1234567890",
				DisplayName: "owner",
			},
			Buckets: dto.Buckets{
				Bucket: buckets,
			},
		}

		w.Header().Set("Content-Type", "application/xml")
		xml.NewEncoder(w).Encode(response)
	}
}

// BucketHandler gère les opérations sur un bucket spécifique (PUT, DELETE)
func BucketHandler(s *storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		bucketName := vars["bucket"]

		log.Printf("BucketHandler appelé pour le bucket: %s avec la méthode: %s", bucketName, r.Method) // Log ajouté

		switch r.Method {
		case http.MethodPut:
			err := s.CreateBucket(bucketName)
			if err != nil {
				log.Printf("Erreur lors de la création du bucket %s: %v", bucketName, err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			log.Printf("Bucket %s créé avec succès", bucketName)
			w.WriteHeader(http.StatusOK)
		case http.MethodDelete:
			err := s.DeleteBucket(bucketName)
			if err != nil {
				log.Printf("Erreur lors de la suppression du bucket %s: %v", bucketName, err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			log.Printf("Bucket %s supprimé avec succès", bucketName)
			w.WriteHeader(http.StatusNoContent)
		default:
			log.Printf("Méthode non autorisée: %s", r.Method)
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}
}
