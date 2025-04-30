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
	path        string
	collections map[string]*Collection
	mu          sync.RWMutex
}

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
		return fmt.Errorf("erreur lecture répertoire: %v", err)
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) != ".json" {
			continue
		}

		data, err := os.ReadFile(filepath.Join(c.path, file.Name()))
		if err != nil {
			return fmt.Errorf("erreur lecture fichier: %v", err)
		}

		var doc Document
		if err := json.Unmarshal(data, &doc); err != nil {
			return fmt.Errorf("erreur désérialisation: %v", err)
		}

		if value, exists := doc[field]; exists {
			if unique && len(index.values[value]) > 0 {
				return fmt.Errorf("valeur dupliquée pour l'index unique %s: %v", field, value)
			}
			id := file.Name()[:len(file.Name())-5] // Retirer .json
			index.values[value] = append(index.values[value], id)
		}
	}

	c.indexes[field] = index
	return nil
}

// Insert insère un document dans la collection
func (c *Collection) Insert(doc Document) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Vérifier les contraintes d'index unique
	for _, index := range c.indexes {
		if index.unique {
			if value, exists := doc[index.field]; exists {
				if len(index.values[value]) > 0 {
					return "", fmt.Errorf("violation de contrainte unique sur %s: %v", index.field, value)
				}
			}
		}
	}

	// Générer un ID unique
	id := generateID()
	doc["_id"] = id

	// Sauvegarder le document
	path := filepath.Join(c.path, id+".json")
	data, err := json.Marshal(doc)
	if err != nil {
		return "", fmt.Errorf("erreur sérialisation: %v", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return "", fmt.Errorf("erreur écriture fichier: %v", err)
	}

	// Mettre à jour les indexs
	for _, index := range c.indexes {
		if value, exists := doc[index.field]; exists {
			index.mu.Lock()
			index.values[value] = append(index.values[value], id)
			index.mu.Unlock()
		}
	}

	return id, nil
}

// FindByID trouve un document par son ID
func (c *Collection) FindByID(id string) (Document, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	path := filepath.Join(c.path, id+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("erreur lecture fichier: %v", err)
	}

	var doc Document
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("erreur désérialisation: %v", err)
	}

	return doc, nil
}

// FindByIndex trouve des documents par valeur d'index
func (c *Collection) FindByIndex(field string, value interface{}) ([]Document, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	index, exists := c.indexes[field]
	if !exists {
		return nil, fmt.Errorf("index %s n'existe pas", field)
	}

	index.mu.RLock()
	ids, exists := index.values[value]
	index.mu.RUnlock()

	if !exists {
		return nil, nil
	}

	var results []Document
	for _, id := range ids {
		doc, err := c.FindByID(id)
		if err != nil {
			return nil, err
		}
		results = append(results, doc)
	}

	return results, nil
}

// Update met à jour un document existant
func (c *Collection) Update(id string, updates Document) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Vérifier si le document existe
	path := filepath.Join(c.path, id+".json")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("document %s n'existe pas", id)
	}

	// Lire le document existant
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var existingDoc Document
	if err := json.Unmarshal(content, &existingDoc); err != nil {
		return err
	}

	// Mettre à jour uniquement les champs fournis
	for key, value := range updates {
		existingDoc[key] = value
	}

	// Vérifier les contraintes d'index unique
	for field, index := range c.indexes {
		if index.unique {
			if value, exists := existingDoc[field]; exists {
				// Vérifier si la nouvelle valeur est unique
				if ids, exists := index.values[value]; exists {
					// Si l'ID n'est pas dans la liste ou s'il y a plus d'un ID, c'est une violation
					if len(ids) > 1 || (len(ids) == 1 && ids[0] != id) {
						return fmt.Errorf("violation de contrainte unique sur %s: %v", field, value)
					}
				}
			}
		}
	}

	// Sauvegarder le document mis à jour
	data, err := json.Marshal(existingDoc)
	if err != nil {
		return err
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return err
	}

	// Mettre à jour les indexs
	for field, index := range c.indexes {
		if value, exists := existingDoc[field]; exists {
			index.mu.Lock()
			// Supprimer l'ancienne valeur de l'index
			if oldValue, oldExists := existingDoc[field]; oldExists {
				if ids, exists := index.values[oldValue]; exists {
					for i, docID := range ids {
						if docID == id {
							index.values[oldValue] = append(ids[:i], ids[i+1:]...)
							break
						}
					}
				}
			}
			// Ajouter la nouvelle valeur
			index.values[value] = append(index.values[value], id)
			index.mu.Unlock()
		}
	}

	return nil
}

// Delete supprime un document
func (c *Collection) Delete(id string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Vérifier si le document existe
	path := filepath.Join(c.path, id+".json")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("document %s n'existe pas", id)
	}

	// Lire le document pour mettre à jour les indexs
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("erreur lecture fichier: %v", err)
	}

	var doc Document
	if err := json.Unmarshal(data, &doc); err != nil {
		return fmt.Errorf("erreur désérialisation: %v", err)
	}

	// Mettre à jour les indexs
	for _, index := range c.indexes {
		if value, exists := doc[index.field]; exists {
			index.mu.Lock()
			if ids, exists := index.values[value]; exists {
				for i, docID := range ids {
					if docID == id {
						index.values[value] = append(ids[:i], ids[i+1:]...)
						break
					}
				}
			}
			index.mu.Unlock()
		}
	}

	// Supprimer le fichier
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("erreur suppression fichier: %v", err)
	}

	return nil
}

// generateID génère un ID unique
func generateID() string {
	return fmt.Sprintf("%x", time.Now().UnixNano())
}

// GetPath retourne le chemin de la collection
func (c *Collection) GetPath() string {
	return c.path
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

// GetAllDocuments retourne tous les documents d'une collection
func (c *Collection) GetAllDocuments() ([]Document, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	files, err := os.ReadDir(c.path)
	if err != nil {
		return nil, err
	}

	var documents []Document
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		content, err := os.ReadFile(filepath.Join(c.path, file.Name()))
		if err != nil {
			continue
		}

		var doc Document
		if err := json.Unmarshal(content, &doc); err != nil {
			continue
		}

		documents = append(documents, doc)
	}

	return documents, nil
}

// FindByField trouve des documents par valeur d'un champ
func (c *Collection) FindByField(field string, value interface{}) ([]Document, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Si le champ est indexé, utiliser l'index
	if _, exists := c.indexes[field]; exists {
		return c.FindByIndex(field, value)
	}

	// Sinon, chercher dans tous les documents
	files, err := os.ReadDir(c.path)
	if err != nil {
		return nil, err
	}

	var results []Document
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		content, err := os.ReadFile(filepath.Join(c.path, file.Name()))
		if err != nil {
			continue
		}

		var doc Document
		if err := json.Unmarshal(content, &doc); err != nil {
			continue
		}

		// Comparer la valeur du champ
		if docValue, exists := doc[field]; exists {
			// Comparaison basique pour les types simples
			if docValue == value {
				results = append(results, doc)
			}
		}
	}

	return results, nil
}
