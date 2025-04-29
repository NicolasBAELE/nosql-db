package orm

import (
	"nosql-db/internal/db"
	"reflect"
	"strings"
)

// CollectionInterface définit l'interface de base pour toutes les collections
type CollectionInterface[T any] interface {
	Insert(document T) (string, error)
	Get(id string) (T, error)
	Update(id string, document T) error
	Delete(id string) error
	FindByField(field string, value interface{}) ([]T, error)
	Count() int
}

// BaseCollection implémente les opérations CRUD de base pour une collection
type BaseCollection[T any] struct {
	collection *Collection[interface{}]
}

// NewBaseCollection crée une nouvelle collection de base
func NewBaseCollection[T any](database *db.Database, name string) (*BaseCollection[T], error) {
	collection, err := NewCollection[interface{}](database, name)
	if err != nil {
		return nil, err
	}

	return &BaseCollection[T]{
		collection: collection,
	}, nil
}

// Insert insère un document dans la collection
func (bc *BaseCollection[T]) Insert(document T) (string, error) {
	return bc.collection.Insert(document)
}

// Get récupère un document par son ID
func (bc *BaseCollection[T]) Get(id string) (T, error) {
	var result T
	doc, err := bc.collection.Get(id)
	if err != nil {
		return result, err
	}

	// Convertir le résultat en T
	docMap, ok := doc.(map[string]interface{})
	if !ok {
		return result, err
	}

	// Utiliser la réflexion pour convertir les champs
	value := reflect.ValueOf(&result).Elem()
	for i := 0; i < value.NumField(); i++ {
		field := value.Type().Field(i)
		jsonTag := field.Tag.Get("json")
		if jsonTag == "" {
			jsonTag = strings.ToLower(field.Name)
		}

		if val, exists := docMap[jsonTag]; exists {
			value.Field(i).Set(reflect.ValueOf(val))
		}
	}

	return result, nil
}

// Update met à jour un document existant
func (bc *BaseCollection[T]) Update(id string, document T) error {
	return bc.collection.Update(id, document)
}

// Delete supprime un document
func (bc *BaseCollection[T]) Delete(id string) error {
	return bc.collection.Delete(id)
}

// FindByField recherche des documents par un champ spécifique
func (bc *BaseCollection[T]) FindByField(field string, value interface{}) ([]T, error) {
	results, err := bc.collection.FindByField(field, value)
	if err != nil {
		return nil, err
	}

	documents := make([]T, len(results))
	for i, result := range results {
		doc, err := bc.Get(result.(string))
		if err != nil {
			continue
		}
		documents[i] = doc
	}

	return documents, nil
}

// Count retourne le nombre de documents dans la collection
func (bc *BaseCollection[T]) Count() int {
	return bc.collection.Count()
}
