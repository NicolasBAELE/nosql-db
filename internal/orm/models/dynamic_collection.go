package models

import (
	"fmt"
	"nosql-db/internal/db"
	"nosql-db/internal/orm"
)

// DynamicField représente un champ dynamique
type DynamicField struct {
	Name string
	Type string
}

// DynamicCollection représente une collection avec des champs définis à l'exécution
type DynamicCollection struct {
	*orm.Collection[interface{}]
	fields []DynamicField
}

// NewDynamicCollection crée une nouvelle collection dynamique
func NewDynamicCollection(database *db.Database, name string) (*DynamicCollection, error) {
	collection, err := orm.NewCollection[interface{}](database, name)
	if err != nil {
		return nil, err
	}

	return &DynamicCollection{
		Collection: collection,
		fields:     make([]DynamicField, 0),
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
	return dc.Collection.Insert(doc)
}

// Update met à jour un document en vérifiant sa validité
func (dc *DynamicCollection) Update(id string, doc map[string]interface{}) error {
	if err := dc.ValidateDocument(doc); err != nil {
		return err
	}
	return dc.Collection.Update(id, doc)
}

// GetAll retourne tous les documents de la collection
func (dc *DynamicCollection) GetAll() ([]interface{}, error) {
	return dc.Collection.Find()
}
