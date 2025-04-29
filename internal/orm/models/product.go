package models

import (
	"nosql-db/internal/db"
	"nosql-db/internal/orm"
)

// Product représente un produit dans la base de données
type Product struct {
	Name        string  `json:"name"`
	Price       float64 `json:"price"`
	Description string  `json:"description"`
	InStock     bool    `json:"in_stock"`
}

// ProductCollection représente une collection de produits
type ProductCollection struct {
	*orm.Collection[Product]
}

// NewProductCollection crée une nouvelle collection de produits
func NewProductCollection(database *db.Database) (*ProductCollection, error) {
	// Créer la collection
	collection, err := orm.NewCollection[Product](database, "products")
	if err != nil {
		return nil, err
	}

	return &ProductCollection{
		Collection: collection,
	}, nil
}

// FindByPrice récupère tous les produits d'un prix spécifique
func (pc *ProductCollection) FindByPrice(price float64) ([]Product, error) {
	return pc.FindByField("Price", price)
}

// FindInStock récupère tous les produits en stock
func (pc *ProductCollection) FindInStock() ([]Product, error) {
	return pc.FindByField("InStock", true)
}
