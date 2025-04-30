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
func (c *Collection) Update(id string, doc Document) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Vérifier si le document existe
	path := filepath.Join(c.path, id+".json")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("document %s n'existe pas", id)
	}

	// Lire le document existant pour vérifier les contraintes d'index unique
	existingData, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("erreur lecture fichier: %v", err)
	}

	var existingDoc Document
	if err := json.Unmarshal(existingData, &existingDoc); err != nil {
		return fmt.Errorf("erreur désérialisation: %v", err)
	}

	// Vérifier les contraintes d'index unique pour les champs modifiés
	for _, index := range c.indexes {
		if index.unique {
			if newValue, exists := doc[index.field]; exists {
				if oldValue, oldExists := existingDoc[index.field]; oldExists && oldValue != newValue {
					if len(index.values[newValue]) > 0 {
						return fmt.Errorf("violation de contrainte unique sur %s: %v", index.field, newValue)
					}
				}
			}
		}
	}

	// Conserver l'ID existant
	doc["_id"] = id

	// Sauvegarder le document mis à jour
	data, err := json.Marshal(doc)
	if err != nil {
		return fmt.Errorf("erreur sérialisation: %v", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("erreur écriture fichier: %v", err)
	}

	// Mettre à jour les indexs
	for _, index := range c.indexes {
		if value, exists := doc[index.field]; exists {
			index.mu.Lock()
			// Supprimer l'ancienne valeur de l'index
			if oldValue, oldExists := existingDoc[index.field]; oldExists {
				for i, docID := range index.values[oldValue] {
					if docID == id {
						index.values[oldValue] = append(index.values[oldValue][:i], index.values[oldValue][i+1:]...)
						break
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
