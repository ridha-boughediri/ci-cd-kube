package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	AccessKeyID     string
	SecretAccessKey string
	Region          string
}

// Fonction principale
func main() {
	cfg := Config{
		AccessKeyID:     "admin1234",
		SecretAccessKey: "adminsecretkey12345678",
		Region:          "us-east-1",
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handleRequest(w, r, cfg)
	})

	log.Println("Serveur démarré sur le port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// handleRequest gère les différentes requêtes HTTP
func handleRequest(w http.ResponseWriter, r *http.Request, cfg Config) {
	// Normaliser le chemin
	path := strings.TrimSuffix(r.URL.Path, "/")
	splitPath := strings.Split(path, "/")

	// Gestion des buckets
	if len(splitPath) == 2 {
		bucketName := splitPath[1]
		switch r.Method {
		case http.MethodPut:
			createBucket(w, r, bucketName, cfg)
		case http.MethodGet:
			listBuckets(w, r, cfg)
		case http.MethodHead:
			w.WriteHeader(http.StatusOK)
		case http.MethodOptions:
			w.Header().Set("Allow", "PUT, GET, HEAD, OPTIONS")
			w.WriteHeader(http.StatusOK)
		default:
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
		return
	}

	// Gestion des objets dans les buckets
	if len(splitPath) == 3 {
		bucketName := splitPath[1]
		objectName := splitPath[2]

		switch r.Method {
		case http.MethodPut:
			uploadObject(w, r, bucketName, objectName, cfg)
		case http.MethodGet:
			downloadObject(w, r, bucketName, objectName, cfg)
		case http.MethodDelete:
			deleteObject(w, r, bucketName, objectName, cfg)
		default:
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
		return
	}

	// Gestion des listes d'objets dans un bucket
	if len(splitPath) == 3 && splitPath[2] == "objects" {
		bucketName := splitPath[1]
		if r.Method == http.MethodGet {
			listObjects(w, r, bucketName, cfg)
		} else {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
		return
	}
}

// Fonction pour créer un bucket
func createBucket(w http.ResponseWriter, r *http.Request, bucketName string, cfg Config) {
	path := "./data/" + bucketName
	err := os.Mkdir(path, 0755)
	if err != nil {
		if os.IsExist(err) {
			http.Error(w, "Bucket Already Exists", http.StatusConflict)
			return
		}
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Printf("Erreur lors de la création du bucket %s : %v", bucketName, err)
		return
	}

	log.Printf("Bucket %s créé avec succès à l'emplacement : %s", bucketName, path)
	w.WriteHeader(http.StatusOK)
}

// Fonction pour lister les buckets
func listBuckets(w http.ResponseWriter, r *http.Request, cfg Config) {
	dataDir := "./data/"
	entries, err := os.ReadDir(dataDir)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Printf("Erreur lors de la lecture du répertoire data : %v", err)
		return
	}

	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("<ListAllMyBucketsResult xmlns=\"http://s3.amazonaws.com/doc/2006-03-01/\">"))
	w.Write([]byte("<Buckets>"))
	for _, entry := range entries {
		if entry.IsDir() {
			w.Write([]byte("<Bucket><Name>" + entry.Name() + "</Name></Bucket>"))
		}
	}
	w.Write([]byte("</Buckets></ListAllMyBucketsResult>"))
}

// Fonction pour uploader un objet
func uploadObject(w http.ResponseWriter, r *http.Request, bucketName, objectName string, cfg Config) {
	bucketPath := "./data/" + bucketName
	if _, err := os.Stat(bucketPath); os.IsNotExist(err) {
		http.Error(w, "Bucket Not Found", http.StatusNotFound)
		return
	}

	objectPath := filepath.Join(bucketPath, objectName)
	file, err := os.Create(objectPath)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	_, err = io.Copy(file, r.Body)
	if err != nil {
		http.Error(w, "Error Writing File", http.StatusInternalServerError)
		return
	}

	log.Printf("Objet %s uploadé dans le bucket %s", objectName, bucketName)
	w.WriteHeader(http.StatusOK)
}

// Fonction pour lister les objets dans un bucket
func listObjects(w http.ResponseWriter, r *http.Request, bucketName string, cfg Config) {
	bucketPath := "./data/" + bucketName
	entries, err := os.ReadDir(bucketPath)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Printf("Erreur lors de la lecture du répertoire bucket %s : %v", bucketName, err)
		return
	}

	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("<ListBucketResult xmlns=\"http://s3.amazonaws.com/doc/2006-03-01/\">"))
	w.Write([]byte("<Objects>"))
	for _, entry := range entries {
		if !entry.IsDir() {
			w.Write([]byte("<Object><Name>" + entry.Name() + "</Name></Object>"))
		}
	}
	w.Write([]byte("</Objects></ListBucketResult>"))
}

// Fonction pour télécharger un objet
func downloadObject(w http.ResponseWriter, r *http.Request, bucketName, objectName string, cfg Config) {
	objectPath := filepath.Join("./data/", bucketName, objectName)
	file, err := os.Open(objectPath)
	if err != nil {
		if os.IsNotExist(err) {
			http.Error(w, "Object Not Found", http.StatusNotFound)
			return
		}
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	w.Header().Set("Content-Type", "application/octet-stream")
	_, err = io.Copy(w, file)
	if err != nil {
		http.Error(w, "Error Sending File", http.StatusInternalServerError)
		return
	}

	log.Printf("Objet %s téléchargé depuis le bucket %s", objectName, bucketName)
}

// Fonction pour supprimer un objet
func deleteObject(w http.ResponseWriter, r *http.Request, bucketName, objectName string, cfg Config) {
	objectPath := filepath.Join("./data/", bucketName, objectName)
	err := os.Remove(objectPath)
	if err != nil {
		if os.IsNotExist(err) {
			http.Error(w, "Object Not Found", http.StatusNotFound)
			return
		}
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	log.Printf("Objet %s supprimé du bucket %s", objectName, bucketName)
	w.WriteHeader(http.StatusOK)
}
