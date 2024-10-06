package handlers

import (
	"log"
	"net/http"
	"os"
	"plateformes3/config"

	"github.com/gorilla/mux"
)

func GetBucketLocation(w http.ResponseWriter, r *http.Request) {
	// Simplement renvoyer un statut 200 OK avec une réponse XML basique
	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`<LocationConstraint>us-east-1</LocationConstraint>`))
	log.Println("Bucket location returned")
}

func HeadBucketHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bucketName := vars["bucketName"]
	cfg := config.LoadConfig()

	bucketPath := cfg.StoragePath + "/" + bucketName

	if _, err := os.Stat(bucketPath); os.IsNotExist(err) {
		log.Printf("Bucket %s does not exist", bucketName)
		http.Error(w, "Bucket does not exist", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	log.Printf("Bucket %s exists", bucketName)
}
