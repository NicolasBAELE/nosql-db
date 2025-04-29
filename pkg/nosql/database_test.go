package nosql

import (
	"testing"
)

func TestDatabase(t *testing.T) {
	// Test de création de la base de données
	t.Run("NewDatabase", func(t *testing.T) {
		db, err := NewDatabase()
		if err != nil {
			t.Errorf("NewDatabase() error = %v", err)
			return
		}
		if db == nil {
			t.Error("NewDatabase() returned nil database")
		}
	})

	// Test de démarrage et arrêt
	t.Run("StartStop", func(t *testing.T) {
		db, _ := NewDatabase()
		db.Start()
		db.Stop()
		// Si on arrive ici sans erreur, c'est que Start/Stop fonctionnent
	})

	// Test de création de collection
	t.Run("CreateCollection", func(t *testing.T) {
		db, _ := NewDatabase()
		db.Start()
		defer db.Stop()

		collection, err := db.CreateCollection("test_collection")
		if err != nil {
			t.Errorf("CreateCollection() error = %v", err)
			return
		}
		if collection == nil {
			t.Error("CreateCollection() returned nil collection")
		}
	})

	// Test de récupération de collection
	t.Run("GetCollection", func(t *testing.T) {
		db, _ := NewDatabase()
		db.Start()
		defer db.Stop()

		// Créer une collection
		db.CreateCollection("test_collection")

		// Tester la récupération d'une collection existante
		collection, exists := db.GetCollection("test_collection")
		if !exists {
			t.Error("GetCollection() returned false for existing collection")
		}
		if collection == nil {
			t.Error("GetCollection() returned nil for existing collection")
		}

		// Tester la récupération d'une collection inexistante
		collection, exists = db.GetCollection("non_existent")
		if exists {
			t.Error("GetCollection() returned true for non-existent collection")
		}
	})

	// Test de liste des collections
	t.Run("ListCollections", func(t *testing.T) {
		db, _ := NewDatabase()
		db.Start()
		defer db.Stop()

		// Créer quelques collections
		db.CreateCollection("collection1")
		db.CreateCollection("collection2")

		collections := db.ListCollections()
		if len(collections) < 2 {
			t.Errorf("ListCollections() returned %d collections, expected at least 2", len(collections))
		}

		// Vérifier que les collections créées sont dans la liste
		found1, found2 := false, false
		for _, name := range collections {
			if name == "collection1" {
				found1 = true
			}
			if name == "collection2" {
				found2 = true
			}
		}
		if !found1 || !found2 {
			t.Error("ListCollections() did not return all created collections")
		}
	})
}
