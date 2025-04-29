package nosql

import (
	"testing"
)

func TestValidation(t *testing.T) {
	// Setup
	db, _ := NewDatabase()
	db.Start()
	defer db.Stop()

	collection, _ := db.CreateCollection("test_collection")

	// Test de validation des types de champs
	t.Run("FieldTypeValidation", func(t *testing.T) {
		// Ajouter des champs avec différents types
		collection.AddField("string_field", FieldTypeString)
		collection.AddField("number_field", FieldTypeNumber)
		collection.AddField("boolean_field", FieldTypeBoolean)

		// Test avec des valeurs valides
		validDoc := map[string]interface{}{
			"string_field":  "test",
			"number_field":  42,
			"boolean_field": true,
		}
		_, err := collection.Insert(validDoc)
		if err != nil {
			t.Errorf("Insert() failed with valid document: %v", err)
		}

		// Test avec une chaîne au lieu d'un nombre
		invalidNumberDoc := map[string]interface{}{
			"string_field":  "test",
			"number_field":  "not a number",
			"boolean_field": true,
		}
		_, err = collection.Insert(invalidNumberDoc)
		if err == nil {
			t.Error("Insert() succeeded with invalid number field")
		}

		// Test avec un nombre au lieu d'un booléen
		invalidBooleanDoc := map[string]interface{}{
			"string_field":  "test",
			"number_field":  42,
			"boolean_field": 1,
		}
		_, err = collection.Insert(invalidBooleanDoc)
		if err == nil {
			t.Error("Insert() succeeded with invalid boolean field")
		}
	})

	// Test de validation des champs manquants
	t.Run("MissingFields", func(t *testing.T) {
		// Créer une nouvelle collection pour ce test
		collection, _ = db.CreateCollection("test_collection_2")
		collection.AddField("required_field", FieldTypeString)
		collection.AddField("optional_field", FieldTypeString)

		// Test avec tous les champs
		completeDoc := map[string]interface{}{
			"required_field": "value",
			"optional_field": "value",
		}
		_, err := collection.Insert(completeDoc)
		if err != nil {
			t.Errorf("Insert() failed with complete document: %v", err)
		}

		// Test avec uniquement le champ requis
		requiredOnlyDoc := map[string]interface{}{
			"required_field": "value",
		}
		_, err = collection.Insert(requiredOnlyDoc)
		if err != nil {
			t.Errorf("Insert() failed with required fields only: %v", err)
		}

		// Test sans le champ requis
		missingRequiredDoc := map[string]interface{}{
			"optional_field": "value",
		}
		_, err = collection.Insert(missingRequiredDoc)
		if err == nil {
			t.Error("Insert() succeeded with missing required field")
		}
	})

	// Test de validation des types de nombres
	t.Run("NumberTypeValidation", func(t *testing.T) {
		// Créer une nouvelle collection pour ce test
		collection, _ = db.CreateCollection("test_collection_3")
		collection.AddField("integer_field", FieldTypeNumber)
		collection.AddField("float_field", FieldTypeNumber)

		// Test avec des entiers
		integerDoc := map[string]interface{}{
			"integer_field": 42,
			"float_field":   3.14,
		}
		_, err := collection.Insert(integerDoc)
		if err != nil {
			t.Errorf("Insert() failed with valid number types: %v", err)
		}

		// Test avec des flottants
		floatDoc := map[string]interface{}{
			"integer_field": 42.0,
			"float_field":   3.14,
		}
		_, err = collection.Insert(floatDoc)
		if err != nil {
			t.Errorf("Insert() failed with float values: %v", err)
		}

		// Test avec des nombres négatifs
		negativeDoc := map[string]interface{}{
			"integer_field": -42,
			"float_field":   -3.14,
		}
		_, err = collection.Insert(negativeDoc)
		if err != nil {
			t.Errorf("Insert() failed with negative numbers: %v", err)
		}
	})

	// Test de validation des chaînes de caractères
	t.Run("StringValidation", func(t *testing.T) {
		// Créer une nouvelle collection pour ce test
		collection, _ = db.CreateCollection("test_collection_4")
		collection.AddField("string_field", FieldTypeString)

		// Test avec une chaîne vide
		emptyStringDoc := map[string]interface{}{
			"string_field": "",
		}
		_, err := collection.Insert(emptyStringDoc)
		if err != nil {
			t.Errorf("Insert() failed with empty string: %v", err)
		}

		// Test avec une chaîne contenant des caractères spéciaux
		specialCharsDoc := map[string]interface{}{
			"string_field": "test@#$%^&*()",
		}
		_, err = collection.Insert(specialCharsDoc)
		if err != nil {
			t.Errorf("Insert() failed with special characters: %v", err)
		}

		// Test avec une chaîne contenant des espaces
		spacesDoc := map[string]interface{}{
			"string_field": "test with spaces",
		}
		_, err = collection.Insert(spacesDoc)
		if err != nil {
			t.Errorf("Insert() failed with spaces: %v", err)
		}
	})

	// Test de validation des booléens
	t.Run("BooleanValidation", func(t *testing.T) {
		// Créer une nouvelle collection pour ce test
		collection, _ = db.CreateCollection("test_collection_5")
		collection.AddField("boolean_field", FieldTypeBoolean)

		// Test avec true
		trueDoc := map[string]interface{}{
			"boolean_field": true,
		}
		_, err := collection.Insert(trueDoc)
		if err != nil {
			t.Errorf("Insert() failed with true value: %v", err)
		}

		// Test avec false
		falseDoc := map[string]interface{}{
			"boolean_field": false,
		}
		_, err = collection.Insert(falseDoc)
		if err != nil {
			t.Errorf("Insert() failed with false value: %v", err)
		}

		// Test avec une valeur invalide
		invalidDoc := map[string]interface{}{
			"boolean_field": "not a boolean",
		}
		_, err = collection.Insert(invalidDoc)
		if err == nil {
			t.Error("Insert() succeeded with invalid boolean value")
		}
	})
}
