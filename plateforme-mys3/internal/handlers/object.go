// internal/handlers/object.go
package handlers

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/xml"
	"io"
	"net/http"
	"os"
	"plateforme-mys3/internal/dto"
	"plateforme-mys3/internal/storage"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

// ObjectHandler gère les opérations sur les objets (PUT, GET, DELETE)
func ObjectHandler(s *storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		bucketName := vars["bucket"]
		objectName := vars["object"]

		switch r.Method {
		case http.MethodPut:
			bodyBytes, err := io.ReadAll(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			r.Body = io.NopCloser(strings.NewReader(string(bodyBytes)))

			err = s.PutObject(bucketName, objectName, r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			// Calculer l'ETag (MD5 du contenu)
			etag := calculateETag(bodyBytes)
			w.Header().Set("ETag", "\""+etag+"\"")
			w.WriteHeader(http.StatusOK)
		case http.MethodGet:
			file, err := s.GetObject(bucketName, objectName)
			if err != nil {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			defer file.Close()
			io.Copy(w, file)
		case http.MethodDelete:
			err := s.DeleteObject(bucketName, objectName)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}
}

// ListObjectsHandler gère la liste des objets dans un bucket
func ListObjectsHandler(s *storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		bucketName := vars["bucket"]

		objectsInfo, err := s.ListObjects(bucketName)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		var objects []dto.Object
		for _, info := range objectsInfo {
			if info.IsDir() {
				continue
			}
			objects = append(objects, dto.Object{
				Key:          info.Name(),
				LastModified: info.ModTime().UTC().Format(time.RFC3339),
				ETag:         "\"" + calculateETagFromFile(s.ObjectPath(bucketName, info.Name())) + "\"",
				Size:         info.Size(),
				StorageClass: "STANDARD",
			})
		}

		response := dto.ListBucketResult{
			XMLNS:       "http://s3.amazonaws.com/doc/2006-03-01/",
			Name:        bucketName,
			Prefix:      "",
			Marker:      "",
			MaxKeys:     1000,
			IsTruncated: false,
			Contents:    objects,
		}

		w.Header().Set("Content-Type", "application/xml")
		xml.NewEncoder(w).Encode(response)
	}
}

// calculateETag calcule l'ETag (MD5) du contenu des données
func calculateETag(data []byte) string {
	hash := md5.Sum(data)
	return hex.EncodeToString(hash[:])
}

// calculateETagFromFile calcule l'ETag (MD5) à partir du contenu d'un fichier
func calculateETagFromFile(filePath string) string {
	file, err := os.Open(filePath)
	if err != nil {
		return ""
	}
	defer file.Close()
	hash := md5.New()
	io.Copy(hash, file)
	return hex.EncodeToString(hash.Sum(nil))
}
