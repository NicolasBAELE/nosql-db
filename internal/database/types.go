package database

import "sync"

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
