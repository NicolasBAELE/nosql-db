package database

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Document représente un document dans la base de données
type Document map[string]interface{}

// Use existing types from transaction.go

// Index représente un index sur un champ
type Index struct {
	field  string
	values map[interface{}][]string // valeur -> liste d'IDs
	unique bool                     // indique si l'index est unique
	mu     sync.RWMutex
}

// Collection représente une collection de documents
type Collection struct {
	name    string
	path    string
	indexes map[string]*Index
	mu      sync.RWMutex
}

// Database représente la base de données
type Database struct {
	path         string
	collections  map[string]*Collection
	transactions map[string]*Transaction
	txManager    *TransactionManager
	mu           sync.RWMutex
}

// NewDatabase crée une nouvelle instance de base de données
func NewDatabase(path string) (*Database, error) {
	if err := os.MkdirAll(path, 0755); err != nil {
		return nil, fmt.Errorf("erreur création répertoire: %v", err)
	}

	// Create WAL directory
	walPath := filepath.Join(path, "wal")
	if err := os.MkdirAll(walPath, 0755); err != nil {
		return nil, fmt.Errorf("erreur création répertoire WAL: %v", err)
	}

	// Initialize transaction manager
	txManager, err := NewTransactionManager(walPath)
	if err != nil {
		return nil, fmt.Errorf("erreur initialisation gestionnaire transactions: %v", err)
	}

	return &Database{
		path:         path,
		collections:  make(map[string]*Collection),
		transactions: make(map[string]*Transaction),
		txManager:    txManager,
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

// GetCollection returns a collection by name
func (db *Database) GetCollection(name string) (*Collection, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	collection, exists := db.collections[name]
	if !exists {
		return nil, fmt.Errorf("collection %s not found", name)
	}
	return collection, nil
}

// GetCollections returns all collection names
func (db *Database) GetCollections() map[string]*Collection {
	db.mu.RLock()
	defer db.mu.RUnlock()
	collections := make(map[string]*Collection)
	for name, collection := range db.collections {
		collections[name] = collection
	}
	return collections
}

// CreateIndex crée un index sur un champ
func (c *Collection) CreateIndex(field string, unique bool) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.indexes[field]; exists {
		return fmt.Errorf("index %s existe déjà", field)
	}

	index := &Index{
		field:  field,
		values: make(map[interface{}][]string),
		unique: unique,
	}

	// Construire l'index à partir des documents existants
	files, err := os.ReadDir(c.path)
	if err != nil {
		return err
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) == ".json" {
			docID := file.Name()[:len(file.Name())-5] // Remove .json extension
			path := filepath.Join(c.path, file.Name())
			data, err := os.ReadFile(path)
			if err != nil {
				continue
			}

			var doc Document
			if err := json.Unmarshal(data, &doc); err != nil {
				continue
			}

			if value, exists := doc[field]; exists {
				index.values[value] = append(index.values[value], docID)
			}
		}
	}

	c.indexes[field] = index
	return nil
}

// Insert inserts a document into a collection
func (c *Collection) Insert(doc Document) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	docID := generateID()

	// Unique index check
	for _, index := range c.indexes {
		if index.unique {
			if value, exists := doc[index.field]; exists {
				if ids, ok := index.values[value]; ok && len(ids) > 0 {
					return "", fmt.Errorf("violation de contrainte unique sur %s: %v", index.field, value)
				}
			}
		}
	}

	path := filepath.Join(c.path, docID+".json")
	data, err := json.Marshal(doc)
	if err != nil {
		return "", err
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return "", err
	}

	// Update indexes
	for _, index := range c.indexes {
		if value, exists := doc[index.field]; exists {
			index.values[value] = append(index.values[value], docID)
		}
	}

	return docID, nil
}

// InsertWithTransaction logs an insert operation during a transaction (deferred writing)
func (db *Database) InsertWithTransaction(tx *Transaction, collectionName string, doc Document) (string, error) {
	_, err := db.GetCollection(collectionName)
	if err != nil {
		return "", fmt.Errorf("collection %s not found", collectionName)
	}

	docID := generateID()

	entry := LogEntry{
		TransactionID: tx.ID,
		Timestamp:     time.Now().UnixNano(),
		Operation:     OpInsert,
		Collection:    collectionName,
		DocumentID:    docID,
		Data:          doc,
	}

	tx.AddLogEntry(entry)
	return docID, nil
}

// UpdateWithTransaction logs an update operation during a transaction (deferred writing)
func (db *Database) UpdateWithTransaction(tx *Transaction, collectionName string, docID string, doc Document) error {
	collection, err := db.GetCollection(collectionName)
	if err != nil {
		return fmt.Errorf("collection %s not found", collectionName)
	}

	// Read existing document to get old data
	path := filepath.Join(collection.path, docID+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var oldDoc Document
	if err := json.Unmarshal(data, &oldDoc); err != nil {
		return err
	}

	entry := LogEntry{
		TransactionID: tx.ID,
		Timestamp:     time.Now().UnixNano(),
		Operation:     OpUpdate,
		Collection:    collectionName,
		DocumentID:    docID,
		Data:          doc,
		OldData:       oldDoc,
	}

	tx.AddLogEntry(entry)
	return nil
}

// DeleteWithTransaction logs a delete operation during a transaction (deferred writing)
func (db *Database) DeleteWithTransaction(tx *Transaction, collectionName string, docID string) error {
	collection, err := db.GetCollection(collectionName)
	if err != nil {
		return fmt.Errorf("collection %s not found", collectionName)
	}

	// Read existing document to get old data
	path := filepath.Join(collection.path, docID+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var oldDoc Document
	if err := json.Unmarshal(data, &oldDoc); err != nil {
		return err
	}

	entry := LogEntry{
		TransactionID: tx.ID,
		Timestamp:     time.Now().UnixNano(),
		Operation:     OpDelete,
		Collection:    collectionName,
		DocumentID:    docID,
		OldData:       oldDoc,
	}

	tx.AddLogEntry(entry)
	return nil
}

// StartTransaction starts a new transaction
func (db *Database) StartTransaction() *Transaction {
	return db.txManager.BeginTransaction()
}

// BeginTransaction is an alias for StartTransaction
func (db *Database) BeginTransaction() *Transaction {
	return db.StartTransaction()
}

// GetTransaction returns a transaction by ID
func (db *Database) GetTransaction(id string) (*Transaction, bool) {
	return db.txManager.GetTransaction(id)
}

// Commit commits a transaction by applying all WAL entries
func (db *Database) Commit(tx *Transaction) error {
	if err := db.ApplyTransactionLog(tx); err != nil {
		return err
	}

	return db.txManager.Commit(db, tx)
}

// Rollback rolls back a transaction by removing WAL files (no data files to delete since we use deferred writing)
func (db *Database) Rollback(tx *Transaction) error {
	return db.txManager.Rollback(tx)
}

// ApplyLogEntry applies a single WAL log entry to the database
func (db *Database) ApplyLogEntry(entry LogEntry) error {
	collection, err := db.GetCollection(entry.Collection)
	if err != nil {
		return fmt.Errorf("collection %s not found", entry.Collection)
	}

	switch entry.Operation {
	case OpInsert:
		// Unique index check
		for _, index := range collection.indexes {
			if index.unique {
				if value, exists := entry.Data[index.field]; exists {
					index.mu.RLock()
					if ids, ok := index.values[value]; ok && len(ids) > 0 {
						index.mu.RUnlock()
						return fmt.Errorf("valeur '%v' du champ '%s' déjà utilisée (index unique)", value, index.field)
					}
					index.mu.RUnlock()
				}
			}
		}
		// Write the document to disk
		path := filepath.Join(collection.path, entry.DocumentID+".json")
		data, err := json.Marshal(entry.Data)
		if err != nil {
			return err
		}
		if err := os.WriteFile(path, data, 0644); err != nil {
			return err
		}
		// Update indexes
		for _, index := range collection.indexes {
			if value, exists := entry.Data[index.field]; exists {
				index.mu.Lock()
				index.values[value] = append(index.values[value], entry.DocumentID)
				index.mu.Unlock()
			}
		}
	case OpUpdate:
		// Unique index check
		for field, index := range collection.indexes {
			if index.unique {
				if value, exists := entry.Data[field]; exists {
					index.mu.RLock()
					if ids, ok := index.values[value]; ok {
						for _, id := range ids {
							if id != entry.DocumentID {
								index.mu.RUnlock()
								return fmt.Errorf("valeur '%v' du champ '%s' déjà utilisée (index unique)", value, field)
							}
						}
					}
					index.mu.RUnlock()
				}
			}
		}
		// Update the document on disk
		path := filepath.Join(collection.path, entry.DocumentID+".json")
		data, err := json.Marshal(entry.Data)
		if err != nil {
			return err
		}
		if err := os.WriteFile(path, data, 0644); err != nil {
			return err
		}
		// Update indexes
		for field, index := range collection.indexes {
			if value, exists := entry.Data[field]; exists {
				index.mu.Lock()
				if oldValue, oldExists := entry.OldData[field]; oldExists {
					if ids, exists := index.values[oldValue]; exists {
						for i, docID := range ids {
							if docID == entry.DocumentID {
								index.values[oldValue] = append(ids[:i], ids[i+1:]...)
								break
							}
						}
					}
				}
				index.values[value] = append(index.values[value], entry.DocumentID)
				index.mu.Unlock()
			}
		}
	case OpDelete:
		// Remove the document from disk
		path := filepath.Join(collection.path, entry.DocumentID+".json")
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return err
		}
		// Update indexes
		for _, index := range collection.indexes {
			if value, exists := entry.OldData[index.field]; exists {
				index.mu.Lock()
				if ids, exists := index.values[value]; exists {
					for i, docID := range ids {
						if docID == entry.DocumentID {
							index.values[value] = append(ids[:i], ids[i+1:]...)
							break
						}
					}
				}
				index.mu.Unlock()
			}
		}
	}
	return nil
}

// ApplyTransactionLog applies all WAL log entries for a transaction (in order)
func (db *Database) ApplyTransactionLog(tx *Transaction) error {
	tx.mu.Lock()
	defer tx.mu.Unlock()
	for _, entry := range tx.Log {
		if err := db.ApplyLogEntry(entry); err != nil {
			return err
		}
	}
	return nil
}

// GetDocument retrieves a document by ID
func (c *Collection) GetDocument(docID string) (Document, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	path := filepath.Join(c.path, docID+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var doc Document
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, err
	}

	return doc, nil
}

// GetAllDocuments retrieves all documents in a collection
func (c *Collection) GetAllDocuments() ([]Document, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	files, err := os.ReadDir(c.path)
	if err != nil {
		return nil, err
	}

	var documents []Document
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".json" {
			path := filepath.Join(c.path, file.Name())
			data, err := os.ReadFile(path)
			if err != nil {
				continue
			}

			var doc Document
			if err := json.Unmarshal(data, &doc); err != nil {
				continue
			}

			documents = append(documents, doc)
		}
	}

	return documents, nil
}

// FindByID retrieves a document by ID
func (c *Collection) FindByID(docID string) (Document, error) {
	return c.GetDocument(docID)
}

// Update updates a document by ID
func (c *Collection) Update(docID string, doc Document) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	path := filepath.Join(c.path, docID+".json")

	// Read existing document to get old data for index updates
	var oldDoc Document
	if data, err := os.ReadFile(path); err == nil {
		json.Unmarshal(data, &oldDoc)
	}

	data, err := json.Marshal(doc)
	if err != nil {
		return err
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return err
	}

	// Update indexes
	for field, index := range c.indexes {
		if value, exists := doc[field]; exists {
			index.mu.Lock()
			// Remove from old value
			if oldValue, oldExists := oldDoc[field]; oldExists {
				if ids, exists := index.values[oldValue]; exists {
					for i, id := range ids {
						if id == docID {
							index.values[oldValue] = append(ids[:i], ids[i+1:]...)
							break
						}
					}
				}
			}
			// Add to new value
			index.values[value] = append(index.values[value], docID)
			index.mu.Unlock()
		}
	}

	return nil
}

// Delete deletes a document by ID
func (c *Collection) Delete(docID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	path := filepath.Join(c.path, docID+".json")

	// Read existing document to get old data for index updates
	var oldDoc Document
	if data, err := os.ReadFile(path); err == nil {
		json.Unmarshal(data, &oldDoc)
	}

	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}

	// Update indexes
	for _, index := range c.indexes {
		if value, exists := oldDoc[index.field]; exists {
			index.mu.Lock()
			if ids, exists := index.values[value]; exists {
				for i, id := range ids {
					if id == docID {
						index.values[value] = append(ids[:i], ids[i+1:]...)
						break
					}
				}
			}
			index.mu.Unlock()
		}
	}

	return nil
}

// FindByField finds documents by field value
func (c *Collection) FindByField(field string, value interface{}) ([]Document, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Check if there's an index on this field
	if index, exists := c.indexes[field]; exists {
		index.mu.RLock()
		if docIDs, exists := index.values[value]; exists {
			var documents []Document
			for _, docID := range docIDs {
				if doc, err := c.GetDocument(docID); err == nil {
					documents = append(documents, doc)
				}
			}
			index.mu.RUnlock()
			return documents, nil
		}
		index.mu.RUnlock()
	}

	// Fallback to scanning all documents
	var documents []Document
	files, err := os.ReadDir(c.path)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) == ".json" {
			path := filepath.Join(c.path, file.Name())
			data, err := os.ReadFile(path)
			if err != nil {
				continue
			}

			var doc Document
			if err := json.Unmarshal(data, &doc); err != nil {
				continue
			}

			if docValue, exists := doc[field]; exists && docValue == value {
				documents = append(documents, doc)
			}
		}
	}

	return documents, nil
}

// FindByIndex finds documents by index (alias for FindByField)
func (c *Collection) FindByIndex(field string, value interface{}) ([]Document, error) {
	return c.FindByField(field, value)
}

// InsertWithTransaction logs an insert operation during a transaction (deferred writing)
func (c *Collection) InsertWithTransaction(tx *Transaction, doc Document) (string, error) {
	// This is a wrapper around the database method
	// We need access to the database instance, so this method should be called on the database
	return "", fmt.Errorf("use Database.InsertWithTransaction instead")
}

// UpdateWithTransaction logs an update operation during a transaction (deferred writing)
func (c *Collection) UpdateWithTransaction(tx *Transaction, docID string, doc Document) error {
	// This is a wrapper around the database method
	// We need access to the database instance, so this method should be called on the database
	return fmt.Errorf("use Database.UpdateWithTransaction instead")
}

// DeleteWithTransaction logs a delete operation during a transaction (deferred writing)
func (c *Collection) DeleteWithTransaction(tx *Transaction, docID string) error {
	// This is a wrapper around the database method
	// We need access to the database instance, so this method should be called on the database
	return fmt.Errorf("use Database.DeleteWithTransaction instead")
}

// generateID generates a unique document ID
func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
