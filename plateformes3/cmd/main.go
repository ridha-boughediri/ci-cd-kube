package main

import (
	"log"
	"net/http"
	"plateformes3/internal/handlers"

	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()

	// Serve static files for the frontend
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))

	// Define API routes for bucket operations
	r.HandleFunc("/buckets", handlers.ListBucketsHandler).Methods("GET")
	r.HandleFunc("/bucket", handlers.CreateBucketHandler).Methods("POST")
	r.HandleFunc("/bucket", handlers.DeleteBucketHandler).Methods("DELETE")
	r.HandleFunc("/{bucketName}/rename", handlers.UpdateBucketNameHandler).Methods("POST")

	// Define API routes for object operations
	r.HandleFunc("/{bucketName}/object", handlers.UploadObjectHandler).Methods("POST")                // Upload an object
	r.HandleFunc("/{bucketName}/object/{objectName}", handlers.DeleteObjectHandler).Methods("DELETE") // Delete an object

	// Serve the index.html as the main page
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/index.html")
	})

	// Start the server
	log.Println("Server is running on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", r))
}
