package main

import (
	"log"
	"net/http"
	"os"
	"strings"
)

type Config struct {
	AccessKeyID     string
	SecretAccessKey string
	Region          string
	// autres configurations
}

// Fonction principale
func main() {
	cfg := Config{
		AccessKeyID:     "admin1234",
		SecretAccessKey: "adminsecretkey12345678",
		Region:          "us-east-1",
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Normaliser le chemin
		path := strings.TrimSuffix(r.URL.Path, "/")
		bucketName := strings.TrimPrefix(path, "/")

		log.Printf("Requête reçue: %s %s", r.Method, bucketName)

		switch r.Method {
		case http.MethodPut:
			createBucket(w, r, bucketName, cfg)
		case http.MethodGet:
			listBuckets(w, r, cfg)
		default:
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	})

	log.Println("Serveur démarré sur le port 8000S")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// Fonction pour créer un bucket
func createBucket(w http.ResponseWriter, r *http.Request, bucketName string, cfg Config) {
	// Vérifier l'authentification si nécessaire
	// Ici, on suppose que l'authentification est désactivée

	// Créer le répertoire pour le bucket
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
	// Exemple simple sans authentification
	dataDir := "./data/"
	entries, err := os.ReadDir(dataDir)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Printf("Erreur lors de la lecture du répertoire data : %v", err)
		return
	}

	// Construire une réponse XML simple
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
