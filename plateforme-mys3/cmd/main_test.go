package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

// setup : Crée le répertoire ./data/ avant les tests
func setup() {
	os.Mkdir("./data", 0755)
}

// teardown : Supprime le répertoire ./data/ après les tests
func teardown() {
	os.RemoveAll("./data")
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

// Test de la fonction createBucket
func TestCreateBucket(t *testing.T) {
	req := httptest.NewRequest(http.MethodPut, "/testbucket", nil)
	w := httptest.NewRecorder()

	createBucket(w, req, "testbucket", Config{})

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d but got %d", http.StatusOK, resp.StatusCode)
	}

	if _, err := os.Stat("./data/testbucket"); os.IsNotExist(err) {
		t.Errorf("Expected bucket to be created, but it was not")
	}
}

// Test de la fonction listBuckets
func TestListBuckets(t *testing.T) {
	os.Mkdir("./data/bucket1", 0755)
	os.Mkdir("./data/bucket2", 0755)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	listBuckets(w, req, Config{})

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d but got %d", http.StatusOK, resp.StatusCode)
	}

	body := w.Body.String()
	if !strings.Contains(body, "<Name>bucket1</Name>") || !strings.Contains(body, "<Name>bucket2</Name>") {
		t.Errorf("Expected buckets to be listed in response")
	}
}
