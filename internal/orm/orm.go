package orm

import (
	"encoding/json"
	"errors"
	"nosql-db/internal/db"
	"nosql-db/internal/index"
	"reflect"
)

// Collection représente une collection typée
type Collection[T any] struct {
	db         *db.Database
	name       string
	collection *db.Collection
	indexes    map[string]*index.Index
}

// NewCollection crée une nouvelle collection typée
func NewCollection[T any](database *db.Database, name string) (*Collection[T], error) {
	if err := database.CreateCollection(name); err != nil && !errors.Is(err, db.ErrCollectionExists) {
		return nil, err
	}

	collection := &Collection[T]{
		db:      database,
		name:    name,
		indexes: make(map[string]*index.Index),
	}

	// Créer automatiquement des index pour tous les champs
	t := reflect.TypeOf((*T)(nil)).Elem()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		indexName := field.Name + "_index"
		collection.indexes[indexName] = index.NewIndex(indexName, field.Name)
	}

	return collection, nil
}

// Insert insère un document dans la collection
func (c *Collection[T]) Insert(document T) (string, error) {
	data, err := json.Marshal(document)
	if err != nil {
		return "", err
	}

	id, err := c.db.InsertDocument(c.name, data)
	if err != nil {
		return "", err
	}

	// Mettre à jour les index
	for _, idx := range c.indexes {
		if err := idx.AddDocument(id, data); err != nil {
			return id, err
		}
	}

	return id, nil
}

// Get récupère un document par son ID
func (c *Collection[T]) Get(id string) (T, error) {
	var result T

	doc, err := c.db.GetDocument(c.name, id)
	if err != nil {
		return result, err
	}

	if err := json.Unmarshal(doc.Data, &result); err != nil {
		return result, err
	}

	return result, nil
}

// Update met à jour un document existant
func (c *Collection[T]) Update(id string, document T) error {
	data, err := json.Marshal(document)
	if err != nil {
		return err
	}

	// Mettre à jour les index
	for _, idx := range c.indexes {
		if err := idx.RemoveDocument(id, data); err != nil {
			return err
		}
		if err := idx.AddDocument(id, data); err != nil {
			return err
		}
	}

	return c.db.UpdateDocument(c.name, id, data)
}

// Delete supprime un document
func (c *Collection[T]) Delete(id string) error {
	doc, err := c.db.GetDocument(c.name, id)
	if err != nil {
		return err
	}

	// Supprimer des index
	for _, idx := range c.indexes {
		if err := idx.RemoveDocument(id, doc.Data); err != nil {
			return err
		}
	}

	return c.db.DeleteDocument(c.name, id)
}

// Find retourne tous les documents de la collection
func (c *Collection[T]) Find() ([]T, error) {
	var results []T

	if coll, exists := c.db.Collections[c.name]; exists {
		for _, doc := range coll.Documents {
			var result T
			if err := json.Unmarshal(doc.Data, &result); err != nil {
				return nil, err
			}
			results = append(results, result)
		}
	}

	return results, nil
}

// FindByField recherche des documents par un champ spécifique
func (c *Collection[T]) FindByField(field string, value interface{}) ([]T, error) {
	var results []T

	// Utiliser l'index correspondant
	indexName := field + "_index"
	if idx, exists := c.indexes[indexName]; exists {
		ids := idx.FindDocuments(value)
		for _, id := range ids {
			doc, err := c.db.GetDocument(c.name, id)
			if err != nil {
				return nil, err
			}
			var result T
			if err := json.Unmarshal(doc.Data, &result); err != nil {
				return nil, err
			}
			results = append(results, result)
		}
		return results, nil
	}

	return nil, errors.New("index not found")
}

// Count retourne le nombre de documents dans la collection
func (c *Collection[T]) Count() int {
	if coll, exists := c.db.Collections[c.name]; exists {
		return len(coll.Documents)
	}
	return 0
}
