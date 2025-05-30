package database

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

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
