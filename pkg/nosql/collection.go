package nosql

import (
	"fmt"
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
	Name     string
	Type     FieldType
	Required bool
	Unique   bool
}

// Collection représente une collection dans la base de données
type Collection struct {
	model   *models.DynamicModel
	fields  []Field
	indexes map[string]string
}

// AddField ajoute un nouveau champ à la collection
func (c *Collection) AddField(name string, fieldType FieldType, required bool, unique bool) {
	c.model.AddField(name, string(fieldType))
	c.fields = append(c.fields, Field{
		Name:     name,
		Type:     fieldType,
		Required: required,
		Unique:   unique,
	})
}

// GetFields retourne tous les champs de la collection
func (c *Collection) GetFields() []Field {
	return c.fields
}

// validateDocument vérifie si un document est valide selon les contraintes
func (c *Collection) validateDocument(document map[string]interface{}) error {
	// Vérifier les champs requis
	for _, field := range c.fields {
		if field.Required {
			if _, exists := document[field.Name]; !exists {
				return fmt.Errorf("required field '%s' is missing", field.Name)
			}
		}
	}

	// Vérifier les contraintes uniques
	for _, field := range c.fields {
		if field.Unique {
			if value, exists := document[field.Name]; exists {
				docs, err := c.FindByField(field.Name, value)
				if err != nil {
					return err
				}
				if len(docs) > 0 {
					return fmt.Errorf("unique constraint violation for field '%s'", field.Name)
				}
			}
		}
	}

	return nil
}

// Insert insère un nouveau document dans la collection
func (c *Collection) Insert(document map[string]interface{}) (string, error) {
	if err := c.validateDocument(document); err != nil {
		return "", err
	}
	return c.model.GetCollection().(*models.DynamicCollection).Insert(document)
}

// Get récupère un document par son ID
func (c *Collection) Get(id string) (map[string]interface{}, error) {
	return c.model.GetCollection().(*models.DynamicCollection).Get(id)
}

// Update met à jour un document existant
func (c *Collection) Update(id string, document map[string]interface{}) error {
	// Pour la mise à jour, nous devons vérifier les contraintes uniques en excluant le document actuel
	oldDoc, err := c.Get(id)
	if err != nil {
		return err
	}

	// Vérifier les contraintes uniques
	for _, field := range c.fields {
		if field.Unique {
			if newValue, exists := document[field.Name]; exists {
				if oldValue, oldExists := oldDoc[field.Name]; !oldExists || newValue != oldValue {
					docs, err := c.FindByField(field.Name, newValue)
					if err != nil {
						return err
					}
					if len(docs) > 0 {
						return fmt.Errorf("unique constraint violation for field '%s'", field.Name)
					}
				}
			}
		}
	}

	return c.model.GetCollection().(*models.DynamicCollection).Update(id, document)
}

// Delete supprime un document
func (c *Collection) Delete(id string) error {
	return c.model.GetCollection().(*models.DynamicCollection).Delete(id)
}

// FindByField recherche des documents par un champ spécifique
func (c *Collection) FindByField(field string, value interface{}) ([]map[string]interface{}, error) {
	return c.model.GetCollection().(*models.DynamicCollection).FindByField(field, value)
}

// GetAll retourne tous les documents de la collection
func (c *Collection) GetAll() ([]map[string]interface{}, error) {
	return c.model.GetCollection().(*models.DynamicCollection).GetAll()
}

// Count retourne le nombre de documents dans la collection
func (c *Collection) Count() int {
	return c.model.GetCollection().(*models.DynamicCollection).Count()
}

// Index représente un index sur un champ
type Index struct {
	Field string
	Type  string
}

// AddIndex ajoute un nouvel index à la collection
func (c *Collection) AddIndex(field string, indexType string) error {
	// Vérifier si le champ existe
	fieldExists := false
	for _, f := range c.fields {
		if f.Name == field {
			fieldExists = true
			break
		}
	}
	if !fieldExists {
		return fmt.Errorf("field %s does not exist", field)
	}

	// Ajouter l'index
	if c.indexes == nil {
		c.indexes = make(map[string]string)
	}
	c.indexes[field] = indexType
	return nil
}

// GetIndexes retourne tous les index de la collection
func (c *Collection) GetIndexes() map[string]string {
	if c.indexes == nil {
		c.indexes = make(map[string]string)
	}
	return c.indexes
}

// Clear supprime tous les documents de la collection
func (c *Collection) Clear() {
	c.model.GetCollection().(*models.DynamicCollection).Clear()
}
