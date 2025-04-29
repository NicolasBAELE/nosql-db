package models

import (
	"nosql-db/internal/db"
)

// DynamicModel représente un modèle dynamique
type DynamicModel struct {
	*BaseModel
	collection *DynamicCollection
}

// NewDynamicModel crée un nouveau modèle dynamique
func NewDynamicModel(database *db.Database, name string) (*DynamicModel, error) {
	collection, err := NewDynamicCollection(database, name)
	if err != nil {
		return nil, err
	}

	return &DynamicModel{
		BaseModel:  NewBaseModel(database),
		collection: collection,
	}, nil
}

// Initialize initialise le modèle dynamique
func (dm *DynamicModel) Initialize(database *db.Database) error {
	if err := dm.BaseModel.Initialize(database); err != nil {
		return err
	}
	return nil
}

// GetCollection retourne la collection dynamique
func (dm *DynamicModel) GetCollection() interface{} {
	return dm.collection
}

// AddField ajoute un champ au modèle
func (dm *DynamicModel) AddField(name, fieldType string) {
	dm.collection.AddField(name, fieldType)
}

// GetFields retourne tous les champs du modèle
func (dm *DynamicModel) GetFields() []DynamicField {
	return dm.collection.GetFields()
}
