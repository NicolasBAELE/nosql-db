package models

import (
	"nosql-db/internal/db"
)

// UserModel représente le modèle de données pour les utilisateurs
type UserModel struct {
	*BaseModel
	collection *UserCollection
}

// Initialize initialise le modèle User
func (um *UserModel) Initialize(database *db.Database) error {
	if err := um.BaseModel.Initialize(database); err != nil {
		return err
	}

	collection, err := NewUserCollection(database)
	if err != nil {
		return err
	}

	um.collection = collection
	return nil
}

// GetCollection retourne la collection d'utilisateurs
func (um *UserModel) GetCollection() interface{} {
	return um.collection
}
