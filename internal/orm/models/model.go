package models

import "nosql-db/internal/db"

// ModelInterface définit l'interface commune pour tous les modèles
type ModelInterface interface {
	// GetCollection retourne la collection associée au modèle
	GetCollection() interface{}
	// Initialize initialise le modèle avec la base de données
	Initialize(database *db.Database) error
}

// BaseModel fournit une implémentation de base pour ModelInterface
type BaseModel struct {
	database *db.Database
}

// NewBaseModel crée un nouveau modèle de base
func NewBaseModel(database *db.Database) *BaseModel {
	return &BaseModel{
		database: database,
	}
}

// GetDatabase retourne la base de données associée au modèle
func (bm *BaseModel) GetDatabase() *db.Database {
	return bm.database
}

// Initialize initialise le modèle avec la base de données
func (bm *BaseModel) Initialize(database *db.Database) error {
	bm.database = database
	return nil
}
