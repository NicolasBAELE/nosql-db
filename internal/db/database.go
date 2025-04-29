package db

import (
	"encoding/json"
	"errors"
	"sync"

	"github.com/google/uuid"
)

var (
	ErrCollectionExists   = errors.New("collection already exists")
	ErrCollectionNotFound = errors.New("collection not found")
	ErrDocumentNotFound   = errors.New("document not found")
)

// Document représente un document dans la base de données
type Document struct {
	ID   string          `json:"id"`
	Data json.RawMessage `json:"data"`
}

// Collection représente une collection de documents
type Collection struct {
	Name      string
	Documents map[string]*Document
	Indexes   map[string]map[string][]string // field -> value -> document IDs
	mu        sync.RWMutex
}

// Database représente la base de données principale
type Database struct {
	Collections map[string]*Collection
	mu          sync.RWMutex
}

// NewDatabase crée une nouvelle instance de la base de données
func NewDatabase() *Database {
	return &Database{
		Collections: make(map[string]*Collection),
	}
}

// CreateCollection crée une nouvelle collection
func (db *Database) CreateCollection(name string) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if _, exists := db.Collections[name]; exists {
		return ErrCollectionExists
	}

	db.Collections[name] = &Collection{
		Name:      name,
		Documents: make(map[string]*Document),
		Indexes:   make(map[string]map[string][]string),
	}

	return nil
}

// InsertDocument insère un document dans une collection
func (db *Database) InsertDocument(collectionName string, data json.RawMessage) (string, error) {
	db.mu.RLock()
	collection, exists := db.Collections[collectionName]
	db.mu.RUnlock()

	if !exists {
		return "", errors.New("collection not found")
	}

	id := uuid.New().String()
	doc := &Document{
		ID:   id,
		Data: data,
	}

	collection.mu.Lock()
	collection.Documents[id] = doc
	collection.mu.Unlock()

	return id, nil
}

// GetDocument récupère un document par son ID
func (db *Database) GetDocument(collectionName, id string) (*Document, error) {
	db.mu.RLock()
	collection, exists := db.Collections[collectionName]
	db.mu.RUnlock()

	if !exists {
		return nil, errors.New("collection not found")
	}

	collection.mu.RLock()
	doc, exists := collection.Documents[id]
	collection.mu.RUnlock()

	if !exists {
		return nil, errors.New("document not found")
	}

	return doc, nil
}

// UpdateDocument met à jour un document existant
func (db *Database) UpdateDocument(collectionName, id string, data json.RawMessage) error {
	db.mu.RLock()
	collection, exists := db.Collections[collectionName]
	db.mu.RUnlock()

	if !exists {
		return errors.New("collection not found")
	}

	collection.mu.Lock()
	defer collection.mu.Unlock()

	if _, exists := collection.Documents[id]; !exists {
		return errors.New("document not found")
	}

	collection.Documents[id].Data = data
	return nil
}

// DeleteDocument supprime un document
func (db *Database) DeleteDocument(collectionName, id string) error {
	db.mu.RLock()
	collection, exists := db.Collections[collectionName]
	db.mu.RUnlock()

	if !exists {
		return errors.New("collection not found")
	}

	collection.mu.Lock()
	defer collection.mu.Unlock()

	if _, exists := collection.Documents[id]; !exists {
		return errors.New("document not found")
	}

	delete(collection.Documents, id)
	return nil
}
