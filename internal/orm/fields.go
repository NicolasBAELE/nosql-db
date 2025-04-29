package orm

import (
	"encoding/json"
	"nosql-db/internal/db"
	"nosql-db/internal/index"
	"reflect"
)

// FieldType définit les types de champs disponibles
type FieldType int

const (
	// FieldTypeString pour les chaînes de caractères
	FieldTypeString FieldType = iota
	// FieldTypeInt pour les entiers
	FieldTypeInt
	// FieldTypeFloat pour les nombres à virgule flottante
	FieldTypeFloat
	// FieldTypeBool pour les booléens
	FieldTypeBool
	// FieldTypeJSON pour les objets JSON
	FieldTypeJSON
)

// FieldOptions définit les options pour un champ
type FieldOptions struct {
	Indexed bool
	Unique  bool
	Null    bool
}

// Field représente un champ dans un modèle
type Field struct {
	Name    string
	Type    FieldType
	Options FieldOptions
}

// StringField crée un champ de type chaîne de caractères
func StringField(options ...FieldOptions) Field {
	opts := FieldOptions{Indexed: true}
	if len(options) > 0 {
		opts = options[0]
	}
	return Field{
		Name:    "",
		Type:    FieldTypeString,
		Options: opts,
	}
}

// IntField crée un champ de type entier
func IntField(options ...FieldOptions) Field {
	opts := FieldOptions{Indexed: true}
	if len(options) > 0 {
		opts = options[0]
	}
	return Field{
		Name:    "",
		Type:    FieldTypeInt,
		Options: opts,
	}
}

// FloatField crée un champ de type nombre à virgule flottante
func FloatField(options ...FieldOptions) Field {
	opts := FieldOptions{Indexed: true}
	if len(options) > 0 {
		opts = options[0]
	}
	return Field{
		Name:    "",
		Type:    FieldTypeFloat,
		Options: opts,
	}
}

// BoolField crée un champ de type booléen
func BoolField(options ...FieldOptions) Field {
	opts := FieldOptions{Indexed: true}
	if len(options) > 0 {
		opts = options[0]
	}
	return Field{
		Name:    "",
		Type:    FieldTypeBool,
		Options: opts,
	}
}

// JSONField crée un champ de type JSON
func JSONField(options ...FieldOptions) Field {
	opts := FieldOptions{Indexed: false}
	if len(options) > 0 {
		opts = options[0]
	}
	return Field{
		Name:    "",
		Type:    FieldTypeJSON,
		Options: opts,
	}
}

// Model définit un modèle de données
type Model struct {
	Name   string
	Fields map[string]Field
}

// NewModel crée un nouveau modèle
func NewModel(name string) *Model {
	return &Model{
		Name:   name,
		Fields: make(map[string]Field),
	}
}

// AddField ajoute un champ au modèle
func (m *Model) AddField(name string, field Field) {
	field.Name = name
	m.Fields[name] = field
}

// CreateStructType crée un type de structure Go à partir du modèle
func (m *Model) CreateStructType() reflect.Type {
	fields := make([]reflect.StructField, 0, len(m.Fields))

	for name, field := range m.Fields {
		var fieldType reflect.Type

		switch field.Type {
		case FieldTypeString:
			fieldType = reflect.TypeOf("")
		case FieldTypeInt:
			fieldType = reflect.TypeOf(0)
		case FieldTypeFloat:
			fieldType = reflect.TypeOf(0.0)
		case FieldTypeBool:
			fieldType = reflect.TypeOf(false)
		case FieldTypeJSON:
			fieldType = reflect.TypeOf(json.RawMessage{})
		}

		fields = append(fields, reflect.StructField{
			Name: name,
			Type: fieldType,
			Tag:  reflect.StructTag(`json:"` + name + `"`),
		})
	}

	return reflect.StructOf(fields)
}

// CreateCollection crée une collection typée à partir du modèle
func (m *Model) CreateCollection(database *db.Database) (*Collection[interface{}], error) {
	// Créer une collection typée
	collection, err := NewCollection[interface{}](database, m.Name)
	if err != nil {
		return nil, err
	}

	// Créer les index pour les champs indexés
	for name, field := range m.Fields {
		if field.Options.Indexed {
			indexName := name + "_index"
			collection.indexes[indexName] = index.NewIndex(indexName, name)
		}
	}

	return collection, nil
}
