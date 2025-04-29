package nosql

import (
	"nosql-db/internal/orm/models"
)

// FieldType définit les types de champs disponibles
type FieldType string

const (
	// FieldTypeString pour les chaînes de caractères
	FieldTypeString FieldType = "string"
	// FieldTypeNumber pour les nombres
	FieldTypeNumber FieldType = "number"
	// FieldTypeBoolean pour les booléens
	FieldTypeBoolean FieldType = "boolean"
)

// Field représente un champ dans une collection
type Field struct {
	Name string
	Type FieldType
}

// Collection représente une collection dans la base de données
type Collection struct {
	model *models.DynamicModel
}

// AddField ajoute un nouveau champ à la collection
func (c *Collection) AddField(name string, fieldType FieldType) {
	c.model.AddField(name, string(fieldType))
}

// GetFields retourne tous les champs de la collection
func (c *Collection) GetFields() []Field {
	fields := make([]Field, 0)
	for _, f := range c.model.GetFields() {
		fields = append(fields, Field{
			Name: f.Name,
			Type: FieldType(f.Type),
		})
	}
	return fields
}

// Insert insère un nouveau document dans la collection
func (c *Collection) Insert(document map[string]interface{}) (string, error) {
	return c.model.GetCollection().(*models.DynamicCollection).Insert(document)
}

// Get récupère un document par son ID
func (c *Collection) Get(id string) (map[string]interface{}, error) {
	doc, err := c.model.GetCollection().(*models.DynamicCollection).Get(id)
	if err != nil {
		return nil, err
	}
	return doc.(map[string]interface{}), nil
}

// Update met à jour un document existant
func (c *Collection) Update(id string, document map[string]interface{}) error {
	return c.model.GetCollection().(*models.DynamicCollection).Update(id, document)
}

// Delete supprime un document
func (c *Collection) Delete(id string) error {
	return c.model.GetCollection().(*models.DynamicCollection).Delete(id)
}

// FindByField recherche des documents par un champ spécifique
func (c *Collection) FindByField(field string, value interface{}) ([]map[string]interface{}, error) {
	docs, err := c.model.GetCollection().(*models.DynamicCollection).FindByField(field, value)
	if err != nil {
		return nil, err
	}

	results := make([]map[string]interface{}, len(docs))
	for i, doc := range docs {
		results[i] = doc.(map[string]interface{})
	}
	return results, nil
}

// GetAll retourne tous les documents de la collection
func (c *Collection) GetAll() ([]map[string]interface{}, error) {
	docs, err := c.model.GetCollection().(*models.DynamicCollection).GetAll()
	if err != nil {
		return nil, err
	}

	results := make([]map[string]interface{}, len(docs))
	for i, doc := range docs {
		results[i] = doc.(map[string]interface{})
	}
	return results, nil
}

// Count retourne le nombre de documents dans la collection
func (c *Collection) Count() int {
	return c.model.GetCollection().(*models.DynamicCollection).Count()
}
