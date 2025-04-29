package nosql

import (
	"testing"
)

func TestIndexes(t *testing.T) {
	// Setup
	db, _ := NewDatabase()
	db.Start()
	defer db.Stop()

	collection, _ := db.CreateCollection("test_collection")

	// Test d'ajout de champs
	collection.AddField("name", FieldTypeString)
	collection.AddField("price", FieldTypeNumber)
	collection.AddField("stock", FieldTypeNumber)

	// Test d'ajout d'index
	t.Run("AddIndex", func(t *testing.T) {
		// Test d'ajout d'index unique
		err := collection.AddIndex("name", "unique")
		if err != nil {
			t.Errorf("AddIndex() error = %v", err)
		}

		// Test d'ajout d'index standard
		err = collection.AddIndex("price", "standard")
		if err != nil {
			t.Errorf("AddIndex() error = %v", err)
		}

		// Test d'ajout d'index sur un champ inexistant
		err = collection.AddIndex("non_existent", "unique")
		if err == nil {
			t.Error("AddIndex() succeeded with non-existent field")
		}
	})

	// Test de récupération des index
	t.Run("GetIndexes", func(t *testing.T) {
		indexes := collection.GetIndexes()
		if len(indexes) != 2 {
			t.Errorf("GetIndexes() returned %d indexes, expected 2", len(indexes))
		}

		// Vérifier les types d'index
		if indexes["name"] != "unique" {
			t.Errorf("GetIndexes() returned wrong type for name index: %v", indexes["name"])
		}
		if indexes["price"] != "standard" {
			t.Errorf("GetIndexes() returned wrong type for price index: %v", indexes["price"])
		}
	})

	// Test de contrainte d'unicité
	t.Run("UniqueConstraint", func(t *testing.T) {
		// Insérer un premier document
		doc1 := map[string]interface{}{
			"name":  "Product1",
			"price": 100,
			"stock": 10,
		}
		_, err := collection.Insert(doc1)
		if err != nil {
			t.Errorf("Insert() error = %v", err)
		}

		// Essayer d'insérer un document avec le même nom
		doc2 := map[string]interface{}{
			"name":  "Product1", // Même nom
			"price": 200,
			"stock": 20,
		}
		_, err = collection.Insert(doc2)
		if err == nil {
			t.Error("Insert() succeeded with duplicate unique field")
		}

		// Insérer un document avec un nom différent
		doc3 := map[string]interface{}{
			"name":  "Product2", // Nom différent
			"price": 300,
			"stock": 30,
		}
		_, err = collection.Insert(doc3)
		if err != nil {
			t.Errorf("Insert() error = %v", err)
		}
	})

	// Test de recherche avec index
	t.Run("SearchWithIndex", func(t *testing.T) {
		// Rechercher par nom (index unique)
		results, err := collection.FindByField("name", "Product1")
		if err != nil {
			t.Errorf("FindByField() error = %v", err)
		}
		if len(results) != 1 {
			t.Errorf("FindByField() returned %d results, expected 1", len(results))
		}

		// Rechercher par prix (index standard)
		results, err = collection.FindByField("price", 100)
		if err != nil {
			t.Errorf("FindByField() error = %v", err)
		}
		if len(results) != 1 {
			t.Errorf("FindByField() returned %d results, expected 1", len(results))
		}
	})
}
