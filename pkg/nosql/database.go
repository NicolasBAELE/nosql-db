package nosql

import (
	"nosql-db/internal/orm/models"
)

// Database représente l'interface principale de la base de données NoSQL
type Database struct {
	manager *models.DatabaseManager
}

// NewDatabase crée une nouvelle instance de la base de données
func NewDatabase() (*Database, error) {
	manager, err := models.NewDatabaseManager()
	if err != nil {
		return nil, err
	}

	return &Database{
		manager: manager,
	}, nil
}

// Start démarre la base de données
func (db *Database) Start() {
	db.manager.Start()
}

// Stop arrête la base de données
func (db *Database) Stop() {
	db.manager.Stop()
}

// CreateCollection crée une nouvelle collection dynamique
func (db *Database) CreateCollection(name string) (*Collection, error) {
	model, err := models.NewDynamicModel(db.manager.GetDatabase(), name)
	if err != nil {
		return nil, err
	}

	db.manager.AddModel(name, model)
	return &Collection{
		model: model,
	}, nil
}

// GetCollection récupère une collection existante
func (db *Database) GetCollection(name string) (*Collection, bool) {
	model, exists := db.manager.GetModel(name)
	if !exists {
		return nil, false
	}

	return &Collection{
		model: model.(*models.DynamicModel),
	}, true
}

// ListCollections retourne la liste des collections disponibles
func (db *Database) ListCollections() []string {
	return db.manager.GetModelNames()
}
