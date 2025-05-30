package database

import (
	"fmt"
	"os"
	"path/filepath"
)

// NewDatabase crée une nouvelle instance de base de données
func NewDatabase(path string) (*Database, error) {
	if err := os.MkdirAll(path, 0755); err != nil {
		return nil, fmt.Errorf("erreur création répertoire: %v", err)
	}

	return &Database{
		path:        path,
		collections: make(map[string]*Collection),
	}, nil
}

// CreateCollection crée une nouvelle collection
func (db *Database) CreateCollection(name string) (*Collection, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	if _, exists := db.collections[name]; exists {
		return nil, fmt.Errorf("collection %s existe déjà", name)
	}

	path := filepath.Join(db.path, name)
	if err := os.MkdirAll(path, 0755); err != nil {
		return nil, fmt.Errorf("erreur création répertoire collection: %v", err)
	}

	collection := &Collection{
		name:    name,
		path:    path,
		indexes: make(map[string]*Index),
	}

	db.collections[name] = collection
	return collection, nil
}

// GetCollections retourne toutes les collections de la base de données
func (db *Database) GetCollections() map[string]*Collection {
	db.mu.RLock()
	defer db.mu.RUnlock()
	return db.collections
}

// GetCollection retourne une collection par son nom
func (db *Database) GetCollection(name string) (*Collection, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	collection, exists := db.collections[name]
	if !exists {
		return nil, fmt.Errorf("collection %s n'existe pas", name)
	}

	return collection, nil
}
