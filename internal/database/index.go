package database

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

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

// GetIndexes retourne tous les indexs de la collection
func (c *Collection) GetIndexes() map[string]*Index {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.indexes
}

// IsUnique retourne si l'index est unique
func (i *Index) IsUnique() bool {
	return i.unique
}

// GetValues retourne les valeurs de l'index
func (i *Index) GetValues() map[interface{}][]string {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return i.values
}

// PrintIndexMappings affiche les correspondances des indexs
func (c *Collection) PrintIndexMappings() {
	c.mu.RLock()
	defer c.mu.RUnlock()

	fmt.Printf("\nIndex mappings for collection '%s':\n", c.name)
	for field, index := range c.indexes {
		fmt.Printf("\nField: %s (Unique: %v)\n", field, index.unique)
		fmt.Println("Value -> Document IDs:")
		for value, ids := range index.values {
			fmt.Printf("  %v -> %v\n", value, ids)
		}
	}
}
