//

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
}

// Fonction principale
func main() {
	dataDir := "./data/"
	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		err := os.Mkdir(dataDir, 0755)
		if err != nil {
			log.Fatalf("Erreur lors de la création du répertoire %s : %v", dataDir, err)
		}
		log.Printf("Répertoire %s créé avec succès.", dataDir)
	}

	cfg := Config{
		AccessKeyID:     "admin1234",
		SecretAccessKey: "adminsecretkey12345678",
		Region:          "eu-west-1",
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Normaliser le chemin
		path := strings.TrimSuffix(r.URL.Path, "/")
		bucketName := strings.Split(strings.TrimPrefix(path, "/"), "/")[0]
		objectName := strings.TrimPrefix(path, bucketName+"/")

		log.Printf("Requête reçue: %s %s", r.Method, path)

		// Authentification simplifiée
		if !authenticateRequest(r, cfg) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		switch r.Method {
		case http.MethodPut:
			if objectName == "" {
				createBucket(w, r, bucketName, cfg)
			}
		case http.MethodGet:
			if objectName == "" {
				listBuckets(w, r, cfg)
			} else {
				downloadObject(w, r, bucketName, objectName, cfg)
			}
		case http.MethodHead:
			handleHeadRequest(w, r, bucketName, objectName, cfg)
		case http.MethodPost:
			// Ajoutez le traitement des requêtes POST si nécessaire
			w.WriteHeader(http.StatusOK)
		case http.MethodOptions:
			// Répondre correctement aux requêtes OPTIONS
			w.WriteHeader(http.StatusOK)
		default:
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	})

	log.Println("Serveur démarré sur le port 9000")
	log.Fatal(http.ListenAndServe(":9000", nil))
}

// Fonction pour authentifier les requêtes
func authenticateRequest(r *http.Request, cfg Config) bool {
	return true // Pour simplifier, nous acceptons toutes les requêtes
}

func createBucket(w http.ResponseWriter, r *http.Request, bucketName string, cfg Config) {
	path := "./data/" + bucketName
	log.Printf("Tentative de création du bucket : %s", bucketName)

	// Vérifiez si le répertoire existe déjà
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		log.Printf("Le bucket %s existe déjà", bucketName)
		http.Error(w, "Bucket Already Exists", http.StatusConflict)
		return
	}

	// Créer le répertoire
	err := os.Mkdir(path, 0755)
	if err != nil {
		log.Printf("Erreur lors de la création du bucket %s : %v", bucketName, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Vérifiez si le répertoire a bien été créé
	if _, err := os.Stat(path); err == nil {
		log.Printf("Bucket %s créé avec succès à l'emplacement : %s", bucketName, path)
	} else {
		log.Printf("Échec de la création du bucket %s", bucketName)
	}

	w.WriteHeader(http.StatusOK)
}

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

func downloadObject(w http.ResponseWriter, r *http.Request, bucketName, objectName string, cfg Config) {
	objectPath := "./data/" + bucketName + "/" + objectName

	// Vérifier que l'objet existe
	if _, err := os.Stat(objectPath); os.IsNotExist(err) {
		http.Error(w, "Object Not Found", http.StatusNotFound)
		log.Printf("Objet non trouvé : %s dans le bucket %s", objectName, bucketName)
		return
	}

	// Envoyer le fichier en réponse
	log.Printf("Téléchargement de l'objet %s depuis le bucket %s", objectName, bucketName)
	http.ServeFile(w, r, objectPath)
}

func handleHeadRequest(w http.ResponseWriter, r *http.Request, bucketName, objectName string, cfg Config) {
	if objectName == "" {
		// Vérifier si le bucket existe
		bucketPath := "./data/" + bucketName
		if _, err := os.Stat(bucketPath); os.IsNotExist(err) {
			http.Error(w, "Bucket Not Found", http.StatusNotFound)
			log.Printf("Bucket non trouvé : %s", bucketName)
			return
		}
		w.WriteHeader(http.StatusOK)
		log.Printf("Bucket %s existe", bucketName)
	} else {
		// Vérifier si l'objet existe
		objectPath := "./data/" + bucketName + "/" + objectName
		if _, err := os.Stat(objectPath); os.IsNotExist(err) {
			http.Error(w, "Object Not Found", http.StatusNotFound)
			log.Printf("Objet non trouvé : %s dans le bucket %s", objectName, bucketName)
			return
		}
		w.WriteHeader(http.StatusOK)
		log.Printf("Objet %s existe dans le bucket %s", objectName, bucketName)
	}
}
