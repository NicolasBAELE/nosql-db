package main

import (
	"fmt"
	"log"

	"nosql-db/pkg/nosql"
)

func main() {
	// Créer une nouvelle instance de base de données
	db, err := nosql.NewDatabase()
	if err != nil {
		log.Fatal(err)
	}

	// Démarrer la base de données
	db.Start()
	defer db.Stop()

	// Créer une collection "products"
	products, err := db.CreateCollection("products")
	if err != nil {
		log.Fatal(err)
	}

	// Ajouter des champs avec différents types
	products.AddField("name", nosql.FieldTypeString)         // Champ string
	products.AddField("price", nosql.FieldTypeNumber)        // Champ number
	products.AddField("stock", nosql.FieldTypeNumber)        // Champ number
	products.AddField("description", nosql.FieldTypeString)  // Champ string
	products.AddField("isAvailable", nosql.FieldTypeBoolean) // Champ boolean

	// Ajouter des index pour améliorer les performances de recherche
	products.AddIndex("name", "unique")    // Index unique sur le nom
	products.AddIndex("price", "standard") // Index standard sur le prix
	products.AddIndex("stock", "standard") // Index standard sur le stock

	// Afficher les index de la collection
	fmt.Println("Indexes:", products.GetIndexes())

	// Insérer des documents
	doc1 := map[string]interface{}{
		"name":        "Laptop",
		"price":       999.99,
		"stock":       50,
		"description": "High-performance laptop",
		"isAvailable": true,
	}
	id1, err := products.Insert(doc1)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Inserted document with ID: %s\n", id1)

	doc2 := map[string]interface{}{
		"name":        "Smartphone",
		"price":       599.99,
		"stock":       100,
		"description": "Latest model smartphone",
		"isAvailable": true,
	}
	id2, err := products.Insert(doc2)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Inserted document with ID: %s\n", id2)

	// Récupérer un document par ID
	product, err := products.Get(id1)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Retrieved product: %+v\n", product)

	// Mettre à jour un document
	updates := map[string]interface{}{
		"price": 899.99,
		"stock": 45,
	}
	if err := products.Update(id1, updates); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Document updated successfully")

	// Rechercher des produits par prix
	results, err := products.FindByField("price", 599.99)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Products with price 599.99: %+v\n", results)

	// Lister tous les produits
	allProducts, err := products.GetAll()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("All products: %+v\n", allProducts)

	// Compter le nombre de documents
	count := products.Count()
	fmt.Printf("Total number of products: %d\n", count)

	// Supprimer un document
	if err := products.Delete(id2); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Document deleted successfully")

	// Vérifier que le document a été supprimé
	count = products.Count()
	fmt.Printf("Total number of products after deletion: %d\n", count)
}
