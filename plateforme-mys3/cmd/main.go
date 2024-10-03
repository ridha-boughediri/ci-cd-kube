package main

import (
	"io"
	"log"
	"net/http"
	"os"
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
		// Normaliser le chemin
		path := strings.TrimPrefix(strings.TrimSuffix(r.URL.Path, "/"), "/")

		log.Printf("Requête reçue: %s %s", r.Method, path)

		// Gérer les différentes méthodes HTTP et les authentifications S3
		if !authenticateRequest(r, cfg) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		switch {
		case path == "":
			// Si le chemin est vide, liste des buckets
			if r.Method == http.MethodGet {
				listBuckets(w, r, cfg)
			} else {
				http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			}
		case path == "favicon.ico":
			handleFavicon(w, r)
		default:
			// Gérer les opérations sur les buckets et les objets S3
			bucketName := strings.Split(path, "/")[0]
			objectName := strings.TrimPrefix(path, bucketName+"/")

			if r.Method == http.MethodPut {
				if objectName == "" {
					createBucket(w, r, bucketName, cfg)
				} else {
					uploadObject(w, r, bucketName, objectName, cfg)
				}
			} else if r.Method == http.MethodGet {
				if objectName == "" {
					listObjectsInBucket(w, r, bucketName, cfg)
				} else {
					downloadObject(w, r, bucketName, objectName, cfg)
				}
			} else if r.Method == http.MethodDelete {
				if objectName == "" {
					deleteBucket(w, r, bucketName, cfg)
				} else {
					deleteObject(w, r, bucketName, objectName, cfg)
				}
			} else {
				http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			}
		}
	})

	log.Println("Serveur démarré sur le port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// Fonction pour authentifier les requêtes S3
func authenticateRequest(r *http.Request, cfg Config) bool {
	accessKey := r.Header.Get("Authorization")
	// Normalement, ici, tu décomposerais l'en-tête Authorization et vérifierais la signature
	// mais pour cette version simple, on vérifie juste si l'accès est correct.
	return accessKey == "Bearer "+cfg.AccessKeyID // Simule une autorisation basée sur l'en-tête
}

// Fonction pour gérer l'upload d'objets
func uploadObject(w http.ResponseWriter, r *http.Request, bucketName, objectName string, cfg Config) {
	bucketPath := "./data/" + bucketName
	objectPath := bucketPath + "/" + objectName

	// Créer le répertoire du bucket si nécessaire
	if _, err := os.Stat(bucketPath); os.IsNotExist(err) {
		http.Error(w, "Bucket Not Found", http.StatusNotFound)
		return
	}

	// Créer ou écraser l'objet
	file, err := os.Create(objectPath)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	// Copier le contenu du corps de la requête vers le fichier
	_, err = io.Copy(file, r.Body)
	if err != nil {
		http.Error(w, "Failed to upload object", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	log.Printf("Objet %s téléchargé dans le bucket %s", objectName, bucketName)
}

// Fonction pour télécharger un objet
func downloadObject(w http.ResponseWriter, r *http.Request, bucketName, objectName string, cfg Config) {
	objectPath := "./data/" + bucketName + "/" + objectName

	// Ouvrir le fichier de l'objet
	file, err := os.Open(objectPath)
	if err != nil {
		http.Error(w, "Object Not Found", http.StatusNotFound)
		return
	}
	defer file.Close()

	// Envoyer le contenu de l'objet en réponse
	http.ServeFile(w, r, objectPath)
}

// Fonction pour supprimer un objet
func deleteObject(w http.ResponseWriter, r *http.Request, bucketName, objectName string, cfg Config) {
	objectPath := "./data/" + bucketName + "/" + objectName

	// Supprimer l'objet
	err := os.Remove(objectPath)
	if err != nil {
		http.Error(w, "Object Not Found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	log.Printf("Objet %s supprimé du bucket %s", objectName, bucketName)
}

// Fonction pour supprimer un bucket
func deleteBucket(w http.ResponseWriter, r *http.Request, bucketName string, cfg Config) {
	bucketPath := "./data/" + bucketName

	// Vérifier si le bucket est vide
	entries, err := os.ReadDir(bucketPath)
	if err != nil {
		http.Error(w, "Bucket Not Found", http.StatusNotFound)
		return
	}
	if len(entries) > 0 {
		http.Error(w, "Bucket Not Empty", http.StatusConflict)
		return
	}

	// Supprimer le bucket
	err = os.Remove(bucketPath)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	log.Printf("Bucket %s supprimé", bucketName)
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

// Fonction pour lister les objets dans un bucket
func listObjectsInBucket(w http.ResponseWriter, r *http.Request, bucketName string, cfg Config) {
	bucketPath := "./data/" + bucketName
	entries, err := os.ReadDir(bucketPath)
	if err != nil {
		if os.IsNotExist(err) {
			http.Error(w, "Bucket Not Found", http.StatusNotFound)
			return
		}
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Printf("Erreur lors de la lecture du bucket %s : %v", bucketName, err)
		return
	}

	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("<ListBucketResult xmlns=\"http://s3.amazonaws.com/doc/2006-03-01/\">"))
	w.Write([]byte("<Name>" + bucketName + "</Name>"))
	w.Write([]byte("<Contents>"))
	for _, entry := range entries {
		if !entry.IsDir() {
			w.Write([]byte("<Object><Key>" + entry.Name() + "</Key></Object>"))
		}
	}
	w.Write([]byte("</Contents></ListBucketResult>"))
}

// Exemple de fonction pour /favicon.ico
func handleFavicon(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Favicon handled."))
}
