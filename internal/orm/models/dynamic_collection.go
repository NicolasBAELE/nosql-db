package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"nosql-db/internal/db"

	"github.com/google/uuid"
)

// DynamicField représente un champ dynamique
type DynamicField struct {
	Name string
	Type string
}

// DynamicCollection représente une collection avec des champs définis à l'exécution
type DynamicCollection struct {
	db        *db.Database
	name      string
	fields    []DynamicField
	Indexes   map[string]string
	documents map[string][]byte
}

// NewDynamicCollection crée une nouvelle collection dynamique
func NewDynamicCollection(database *db.Database, name string) (*DynamicCollection, error) {
	if err := database.CreateCollection(name); err != nil && !errors.Is(err, db.ErrCollectionExists) {
		return nil, err
	}

	return &DynamicCollection{
		db:        database,
		name:      name,
		fields:    make([]DynamicField, 0),
		Indexes:   make(map[string]string),
		documents: make(map[string][]byte),
	}, nil
}

// AddField ajoute un nouveau champ à la collection
func (dc *DynamicCollection) AddField(name, fieldType string) {
	dc.fields = append(dc.fields, DynamicField{
		Name: name,
		Type: fieldType,
	})
}

// GetFields retourne tous les champs de la collection
func (dc *DynamicCollection) GetFields() []DynamicField {
	return dc.fields
}

// ValidateDocument vérifie si un document est valide selon les champs définis
func (dc *DynamicCollection) ValidateDocument(doc map[string]interface{}) error {
	for _, field := range dc.fields {
		value, exists := doc[field.Name]
		if !exists {
			continue // Les champs sont optionnels
		}

		switch field.Type {
		case "string":
			if _, ok := value.(string); !ok {
				return fmt.Errorf("le champ '%s' doit être une chaîne de caractères", field.Name)
			}
		case "number":
			switch value.(type) {
			case float64, int, int64:
				// Valide
			default:
				return fmt.Errorf("le champ '%s' doit être un nombre", field.Name)
			}
		case "boolean":
			if _, ok := value.(bool); !ok {
				return fmt.Errorf("le champ '%s' doit être un booléen", field.Name)
			}
		}
	}
	return nil
}

// Insert insère un document en vérifiant sa validité
func (dc *DynamicCollection) Insert(doc map[string]interface{}) (string, error) {
	if err := dc.ValidateDocument(doc); err != nil {
		return "", err
	}

	// Vérifier les contraintes uniques
	for field, indexType := range dc.Indexes {
		if indexType == "unique" {
			if value, exists := doc[field]; exists {
				// Rechercher les documents existants avec la même valeur
				existing, err := dc.FindByField(field, value)
				if err != nil {
					return "", err
				}
				if len(existing) > 0 {
					return "", fmt.Errorf("unique constraint violation for field '%s'", field)
				}
			}
		}
	}

	data, err := json.Marshal(doc)
	if err != nil {
		return "", err
	}
	id := uuid.New().String()
	dc.documents[id] = data
	return id, nil
}

// Get récupère un document par son ID
func (dc *DynamicCollection) Get(id string) (map[string]interface{}, error) {
	data, exists := dc.documents[id]
	if !exists {
		return nil, fmt.Errorf("document not found")
	}
	var doc map[string]interface{}
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, err
	}
	return doc, nil
}

// Update met à jour un document en vérifiant sa validité
func (dc *DynamicCollection) Update(id string, doc map[string]interface{}) error {
	if err := dc.ValidateDocument(doc); err != nil {
		return err
	}
	data, err := json.Marshal(doc)
	if err != nil {
		return err
	}
	if _, exists := dc.documents[id]; !exists {
		return fmt.Errorf("document not found")
	}
	dc.documents[id] = data
	return nil
}

// Delete supprime un document
func (dc *DynamicCollection) Delete(id string) error {
	if _, exists := dc.documents[id]; !exists {
		return fmt.Errorf("document not found")
	}
	delete(dc.documents, id)
	return nil
}

// GetAll retourne tous les documents de la collection
func (dc *DynamicCollection) GetAll() ([]map[string]interface{}, error) {
	results := make([]map[string]interface{}, 0, len(dc.documents))
	for _, data := range dc.documents {
		var doc map[string]interface{}
		if err := json.Unmarshal(data, &doc); err != nil {
			return nil, err
		}
		results = append(results, doc)
	}
	return results, nil
}

// FindByField recherche des documents par un champ spécifique
func (dc *DynamicCollection) FindByField(field string, value interface{}) ([]map[string]interface{}, error) {
	results := make([]map[string]interface{}, 0)
	for _, data := range dc.documents {
		var doc map[string]interface{}
		if err := json.Unmarshal(data, &doc); err != nil {
			return nil, err
		}

		// Trouver le type du champ
		var fieldType string
		for _, f := range dc.fields {
			if f.Name == field {
				fieldType = f.Type
				break
			}
		}

		// Comparer les valeurs en fonction du type
		docValue := doc[field]
		var matches bool

		switch fieldType {
		case "number":
			// Convertir les deux valeurs en float64 pour la comparaison
			var docNum, searchNum float64
			switch v := docValue.(type) {
			case float64:
				docNum = v
			case int:
				docNum = float64(v)
			case int64:
				docNum = float64(v)
			}
			switch v := value.(type) {
			case float64:
				searchNum = v
			case int:
				searchNum = float64(v)
			case int64:
				searchNum = float64(v)
			}
			matches = docNum == searchNum
		default:
			matches = docValue == value
		}

		if matches {
			results = append(results, doc)
		}
	}
	return results, nil
}

// Count retourne le nombre de documents dans la collection
func (dc *DynamicCollection) Count() int {
	return len(dc.documents)
}

// AddIndex ajoute un nouvel index à la collection
func (dc *DynamicCollection) AddIndex(field string, indexType string) error {
	// Vérifier si le champ existe
	fieldExists := false
	for _, f := range dc.fields {
		if f.Name == field {
			fieldExists = true
			break
		}
	}
	if !fieldExists {
		return fmt.Errorf("field %s does not exist", field)
	}

	// Ajouter l'index
	if dc.Indexes == nil {
		dc.Indexes = make(map[string]string)
	}
	dc.Indexes[field] = indexType
	return nil
}

// GetIndexes retourne tous les index de la collection
func (dc *DynamicCollection) GetIndexes() map[string]string {
	if dc.Indexes == nil {
		dc.Indexes = make(map[string]string)
	}
	return dc.Indexes
}

// Clear supprime tous les documents de la collection
func (dc *DynamicCollection) Clear() {
	dc.documents = make(map[string][]byte)
}
