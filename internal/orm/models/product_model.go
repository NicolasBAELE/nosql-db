package models

import (
	"nosql-db/internal/db"
)

// ProductModel représente le modèle de données pour les produits
type ProductModel struct {
	*BaseModel
	collection *ProductCollection
}

// Initialize initialise le modèle Product
func (pm *ProductModel) Initialize(database *db.Database) error {
	if err := pm.BaseModel.Initialize(database); err != nil {
		return err
	}

	collection, err := NewProductCollection(database)
	if err != nil {
		return err
	}

	pm.collection = collection
	return nil
}

// GetCollection retourne la collection de produits
func (pm *ProductModel) GetCollection() interface{} {
	return pm.collection
}
