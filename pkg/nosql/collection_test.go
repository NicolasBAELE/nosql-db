package nosql

import (
	"testing"
)

func TestCollection(t *testing.T) {
	// Setup
	db, _ := NewDatabase()
	db.Start()
	defer db.Stop()

	collection, _ := db.CreateCollection("test_collection")

	// Test d'ajout de champs
	t.Run("AddField", func(t *testing.T) {
		collection.AddField("name", FieldTypeString, true, false)
		collection.AddField("age", FieldTypeNumber, false, false)
		collection.AddField("active", FieldTypeBoolean, false, false)

		fields := collection.GetFields()
		if len(fields) != 3 {
			t.Errorf("Expected 3 fields, got %d", len(fields))
		}

		// Vérifier les types des champs
		fieldTypes := map[string]FieldType{
			"name":   FieldTypeString,
			"age":    FieldTypeNumber,
			"active": FieldTypeBoolean,
		}

		for _, field := range fields {
			expectedType, exists := fieldTypes[field.Name]
			if !exists {
				t.Errorf("Unexpected field name: %s", field.Name)
			}
			if field.Type != expectedType {
				t.Errorf("Field %s has type %v, expected %v", field.Name, field.Type, expectedType)
			}
		}
	})

	// Test d'insertion de document
	t.Run("Insert", func(t *testing.T) {
		collection.Clear()
		doc := map[string]interface{}{
			"name":   "John Doe",
			"age":    30,
			"active": true,
		}

		id, err := collection.Insert(doc)
		if err != nil {
			t.Errorf("Insert() error = %v", err)
		}
		if id == "" {
			t.Error("Insert() returned empty ID")
		}
	})

	// Test de récupération de document
	t.Run("Get", func(t *testing.T) {
		collection.Clear()
		// Insérer un document pour le test
		doc := map[string]interface{}{
			"name":   "Jane Doe",
			"age":    25,
			"active": true,
		}
		id, _ := collection.Insert(doc)

		// Récupérer le document
		retrieved, err := collection.Get(id)
		if err != nil {
			t.Errorf("Get() error = %v", err)
		}
		if retrieved == nil {
			t.Error("Get() returned nil document")
		}

		// Vérifier les valeurs
		if retrieved["name"] != "Jane Doe" {
			t.Errorf("Get() returned wrong name: %v", retrieved["name"])
		}
		if retrieved["age"].(float64) != 25 {
			t.Errorf("Get() returned wrong age: %v", retrieved["age"])
		}
		if retrieved["active"] != true {
			t.Errorf("Get() returned wrong active status: %v", retrieved["active"])
		}
	})

	// Test de mise à jour de document
	t.Run("Update", func(t *testing.T) {
		collection.Clear()
		// Insérer un document pour le test
		doc := map[string]interface{}{
			"name":   "Bob Smith",
			"age":    40,
			"active": false,
		}
		id, _ := collection.Insert(doc)

		// Mettre à jour le document
		update := map[string]interface{}{
			"name":   "Bob Smith",
			"age":    41,
			"active": true,
		}
		err := collection.Update(id, update)
		if err != nil {
			t.Errorf("Update() error = %v", err)
		}

		// Vérifier la mise à jour
		updated, _ := collection.Get(id)
		if updated["age"].(float64) != 41 {
			t.Errorf("Update() did not update age correctly: %v", updated["age"])
		}
		if updated["active"] != true {
			t.Errorf("Update() did not update active status correctly: %v", updated["active"])
		}
	})

	// Test de suppression de document
	t.Run("Delete", func(t *testing.T) {
		collection.Clear()
		// Insérer un document pour le test
		doc := map[string]interface{}{
			"name":   "Alice Brown",
			"age":    35,
			"active": true,
		}
		id, _ := collection.Insert(doc)

		// Supprimer le document
		err := collection.Delete(id)
		if err != nil {
			t.Errorf("Delete() error = %v", err)
		}

		// Vérifier que le document n'existe plus
		_, err = collection.Get(id)
		if err == nil {
			t.Error("Delete() did not remove the document")
		}
	})

	// Test de recherche par champ
	t.Run("FindByField", func(t *testing.T) {
		collection.Clear()
		// Insérer plusieurs documents pour le test
		docs := []map[string]interface{}{
			{"name": "Test1", "age": 20, "active": true},
			{"name": "Test2", "age": 20, "active": false},
			{"name": "Test3", "age": 30, "active": true},
		}

		for _, doc := range docs {
			collection.Insert(doc)
		}

		// Rechercher par âge
		results, err := collection.FindByField("age", 20)
		if err != nil {
			t.Errorf("FindByField() error = %v", err)
		}
		if len(results) != 2 {
			t.Errorf("FindByField() returned %d results, expected 2", len(results))
		}

		// Rechercher par statut actif
		results, err = collection.FindByField("active", true)
		if err != nil {
			t.Errorf("FindByField() error = %v", err)
		}
		if len(results) != 2 {
			t.Errorf("FindByField() returned %d results, expected 2", len(results))
		}
	})

	// Test de récupération de tous les documents
	t.Run("GetAll", func(t *testing.T) {
		collection.Clear()
		// Insérer quelques documents pour le test
		docs := []map[string]interface{}{
			{"name": "All1", "age": 25, "active": true},
			{"name": "All2", "age": 35, "active": false},
		}

		for _, doc := range docs {
			collection.Insert(doc)
		}

		// Récupérer tous les documents
		results, err := collection.GetAll()
		if err != nil {
			t.Errorf("GetAll() error = %v", err)
		}
		if len(results) < 2 {
			t.Errorf("GetAll() returned %d results, expected at least 2", len(results))
		}
	})

	// Test de comptage des documents
	t.Run("Count", func(t *testing.T) {
		collection.Clear()
		// Insérer quelques documents pour le test
		docs := []map[string]interface{}{
			{"name": "Count1", "age": 40, "active": true},
			{"name": "Count2", "age": 45, "active": true},
		}

		initialCount := collection.Count()
		for _, doc := range docs {
			collection.Insert(doc)
		}

		finalCount := collection.Count()
		if finalCount != initialCount+2 {
			t.Errorf("Count() returned %d, expected %d", finalCount, initialCount+2)
		}
	})
}
