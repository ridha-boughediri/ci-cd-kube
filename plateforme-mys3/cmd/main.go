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
	// autres configurations
}

// Fonction principale
func main() {
	cfg := Config{
		AccessKeyID:     "admin1234",
		SecretAccessKey: "adminsecretkey12345678",
		Region:          "eu-west-1", // Région européenne (Irlande)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Normaliser le chemin
		path := strings.TrimSuffix(r.URL.Path, "/")
		bucketName := strings.Split(strings.TrimPrefix(path, "/"), "/")[0]
		objectName := strings.TrimPrefix(path, bucketName+"/")

		log.Printf("Requête reçue: %s %s", r.Method, path)

		switch r.Method {
		case http.MethodPut:
			if objectName == "" {
				createBucket(w, r, bucketName, cfg)
			} else {
				uploadObject(w, r, bucketName, objectName, cfg)
			}
		case http.MethodGet:
			if objectName == "" {
				listBuckets(w, r, cfg)
			} else {
				downloadObject(w, r, bucketName, objectName, cfg)
			}
		case http.MethodDelete:
			if objectName == "" {
				deleteBucket(w, r, bucketName, cfg)
			} else {
				deleteObject(w, r, bucketName, objectName, cfg)
			}
		default:
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	})

	log.Println("Serveur démarré sur le port 9000")
	log.Fatal(http.ListenAndServe(":9000", nil))
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

// Fonction pour téléverser un objet dans un bucket
func uploadObject(w http.ResponseWriter, r *http.Request, bucketName, objectName string, cfg Config) {
	bucketPath := "./data/" + bucketName
	objectPath := bucketPath + "/" + objectName

	// Vérifier que le bucket existe
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

	// Copier le contenu du corps de la requête dans l'objet
	_, err = io.Copy(file, r.Body)
	if err != nil {
		http.Error(w, "Failed to upload object", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	log.Printf("Objet %s téléchargé dans le bucket %s", objectName, bucketName)
}

// Fonction pour télécharger un objet depuis un bucket
func downloadObject(w http.ResponseWriter, r *http.Request, bucketName, objectName string, cfg Config) {
	objectPath := "./data/" + bucketName + "/" + objectName

	// Vérifier que l'objet existe
	file, err := os.Open(objectPath)
	if err != nil {
		http.Error(w, "Object Not Found", http.StatusNotFound)
		return
	}
	defer file.Close()

	// Envoyer le fichier en réponse
	http.ServeFile(w, r, objectPath)
}

// Fonction pour supprimer un objet dans un bucket
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

// Fonction pour supprimer un bucket (uniquement s'il est vide)
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
