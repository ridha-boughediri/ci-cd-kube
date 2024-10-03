package main

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// setup : Crée le répertoire ./data/ avant les tests
func setup() {
	err := os.Mkdir("./data", 0755)
	if err != nil && !os.IsExist(err) {
		panic(err)
	}
}

// teardown : Supprime le répertoire ./data/ après les tests
func teardown() {
	err := os.RemoveAll("./data")
	if err != nil {
		panic(err)
	}
}

// TestMain est utilisé pour configurer et nettoyer l'environnement des tests
func TestMain(m *testing.M) {
	// Setup : Créer le répertoire ./data/
	setup()

	// Exécuter les tests
	code := m.Run()

	// Teardown : Supprimer le répertoire ./data/
	teardown()

	// Sortir avec le code de retour des tests
	os.Exit(code)
}

// Test de la fonction uploadObject
func TestUploadObject(t *testing.T) {
	// Créer un bucket pour le test
	os.Mkdir("./data/testbucket", 0755)

	// Simuler l'upload d'un objet
	req := httptest.NewRequest(http.MethodPut, "/testbucket/testobject.txt", bytes.NewReader([]byte("Contenu de l'objet")))
	w := httptest.NewRecorder()

	uploadObject(w, req, "testbucket", "testobject.txt", Config{})

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d but got %d", http.StatusOK, resp.StatusCode)
	}

	// Vérifier que l'objet a bien été créé dans le répertoire du bucket
	objectPath := filepath.Join("./data/testbucket", "testobject.txt")
	if _, err := os.Stat(objectPath); os.IsNotExist(err) {
		t.Errorf("Expected object to be created, but it was not")
	}
}

// Test de la fonction listObjects
func TestListObjects(t *testing.T) {
	// Créer un bucket et des objets fictifs pour le test
	os.Mkdir("./data/testbucket", 0755)
	os.WriteFile("./data/testbucket/object1.txt", []byte("contenu1"), 0644)
	os.WriteFile("./data/testbucket/object2.txt", []byte("contenu2"), 0644)

	// Simuler la requête GET pour lister les objets dans le bucket
	req := httptest.NewRequest(http.MethodGet, "/testbucket/objects", nil)
	w := httptest.NewRecorder()

	listObjects(w, req, "testbucket", Config{})

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d but got %d", http.StatusOK, resp.StatusCode)
	}

	// Vérifier que les objets sont listés dans la réponse
	body := w.Body.String()
	if !strings.Contains(body, "<Name>object1.txt</Name>") || !strings.Contains(body, "<Name>object2.txt</Name>") {
		t.Errorf("Expected objects to be listed in response")
	}
}

// Test de la fonction downloadObject
func TestDownloadObject(t *testing.T) {
	// Créer un bucket et un objet pour le test
	os.Mkdir("./data/testbucket", 0755)
	os.WriteFile("./data/testbucket/testobject.txt", []byte("Contenu de l'objet"), 0644)

	// Simuler la requête GET pour télécharger l'objet
	req := httptest.NewRequest(http.MethodGet, "/testbucket/testobject.txt", nil)
	w := httptest.NewRecorder()

	downloadObject(w, req, "testbucket", "testobject.txt", Config{})

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d but got %d", http.StatusOK, resp.StatusCode)
	}

	// Vérifier que le contenu téléchargé est correct
	body, _ := io.ReadAll(resp.Body)
	if string(body) != "Contenu de l'objet" {
		t.Errorf("Expected content %q but got %q", "Contenu de l'objet", string(body))
	}
}

// Test de la fonction deleteObject
func TestDeleteObject(t *testing.T) {
	// Créer un bucket et un objet pour le test
	os.Mkdir("./data/testbucket", 0755)
	os.WriteFile("./data/testbucket/testobject.txt", []byte("Contenu de l'objet"), 0644)

	// Simuler la requête DELETE pour supprimer l'objet
	req := httptest.NewRequest(http.MethodDelete, "/testbucket/testobject.txt", nil)
	w := httptest.NewRecorder()

	deleteObject(w, req, "testbucket", "testobject.txt", Config{})

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d but got %d", http.StatusOK, resp.StatusCode)
	}

	// Vérifier que l'objet a bien été supprimé
	objectPath := filepath.Join("./data/testbucket", "testobject.txt")
	if _, err := os.Stat(objectPath); !os.IsNotExist(err) {
		t.Errorf("Expected object to be deleted, but it still exists")
	}
}
