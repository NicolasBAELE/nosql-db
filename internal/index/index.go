package index

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"nosql-db/internal/db"
)

// IndexType définit le type d'index
type IndexType int

const (
	// IndexTypeSingle index simple sur un champ
	IndexTypeSingle IndexType = iota
	// IndexTypeCompound index composé sur plusieurs champs
	IndexTypeCompound
)

// Index représente un index sur un ou plusieurs champs
type Index struct {
	Name   string
	Fields []string
	Type   IndexType
	Values map[string][]string // value -> document IDs
	mu     sync.RWMutex
}

// NewIndex crée un nouvel index
func NewIndex(name string, fields ...string) *Index {
	indexType := IndexTypeSingle
	if len(fields) > 1 {
		indexType = IndexTypeCompound
	}

	return &Index{
		Name:   name,
		Fields: fields,
		Type:   indexType,
		Values: make(map[string][]string),
	}
}

// AddDocument ajoute un document à l'index
func (idx *Index) AddDocument(id string, data json.RawMessage) error {
	var doc map[string]interface{}
	if err := json.Unmarshal(data, &doc); err != nil {
		return err
	}

	// Construire la clé d'index
	key := idx.buildIndexKey(doc)
	if key == "" {
		return nil
	}

	idx.mu.Lock()
	defer idx.mu.Unlock()

	if _, exists := idx.Values[key]; !exists {
		idx.Values[key] = make([]string, 0)
	}
	idx.Values[key] = append(idx.Values[key], id)

	return nil
}

// RemoveDocument supprime un document de l'index
func (idx *Index) RemoveDocument(id string, data json.RawMessage) error {
	var doc map[string]interface{}
	if err := json.Unmarshal(data, &doc); err != nil {
		return err
	}

	key := idx.buildIndexKey(doc)
	if key == "" {
		return nil
	}

	idx.mu.Lock()
	defer idx.mu.Unlock()

	if ids, exists := idx.Values[key]; exists {
		newIDs := make([]string, 0, len(ids)-1)
		for _, existingID := range ids {
			if existingID != id {
				newIDs = append(newIDs, existingID)
			}
		}
		if len(newIDs) == 0 {
			delete(idx.Values, key)
		} else {
			idx.Values[key] = newIDs
		}
	}

	return nil
}

// FindDocuments trouve tous les documents correspondant à une valeur
func (idx *Index) FindDocuments(value interface{}) []string {
	key := toString(value)

	idx.mu.RLock()
	defer idx.mu.RUnlock()

	if ids, exists := idx.Values[key]; exists {
		result := make([]string, len(ids))
		copy(result, ids)
		return result
	}

	return []string{}
}

// FindDocumentsByFields trouve tous les documents correspondant à plusieurs valeurs
func (idx *Index) FindDocumentsByFields(values map[string]interface{}) []string {
	if idx.Type != IndexTypeCompound {
		return []string{}
	}

	key := idx.buildCompoundKey(values)

	idx.mu.RLock()
	defer idx.mu.RUnlock()

	if ids, exists := idx.Values[key]; exists {
		result := make([]string, len(ids))
		copy(result, ids)
		return result
	}

	return []string{}
}

// GetStats retourne les statistiques de l'index
func (idx *Index) GetStats() map[string]interface{} {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	totalDocs := 0
	for _, ids := range idx.Values {
		totalDocs += len(ids)
	}

	return map[string]interface{}{
		"name":      idx.Name,
		"type":      idx.Type,
		"fields":    idx.Fields,
		"keys":      len(idx.Values),
		"documents": totalDocs,
		"avg_docs":  float64(totalDocs) / float64(len(idx.Values)),
	}
}

// buildIndexKey construit la clé d'index pour un document
func (idx *Index) buildIndexKey(doc map[string]interface{}) string {
	if idx.Type == IndexTypeSingle {
		value, exists := doc[idx.Fields[0]]
		if !exists {
			return ""
		}
		return toString(value)
	}
	return ""
}

// buildCompoundKey construit une clé composée pour plusieurs champs
func (idx *Index) buildCompoundKey(values map[string]interface{}) string {
	parts := make([]string, len(idx.Fields))
	for i, field := range idx.Fields {
		value, exists := values[field]
		if !exists {
			return ""
		}
		parts[i] = toString(value)
	}
	return strings.Join(parts, "|")
}

// toString convertit une valeur en string pour l'indexation
func toString(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case bool:
		if v {
			return "true"
		}
		return "false"
	default:
		return fmt.Sprintf("%v", v)
	}
}

// RebuildIndex reindexe tous les documents d'une collection
func (idx *Index) RebuildIndex(collection *db.Collection) error {
	// Vider l'index existant
	idx.Values = make(map[string][]string)

	// Reindexer tous les documents
	for id, doc := range collection.Documents {
		if err := idx.AddDocument(id, doc.Data); err != nil {
			return fmt.Errorf("erreur lors de l'indexation du document %s: %v", id, err)
		}
	}

	return nil
}
