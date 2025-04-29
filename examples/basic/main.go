package main

import (
	"fmt"
	"log"
	"nosql-db/pkg/nosql"
)

func main() {
	// Créer une nouvelle instance de la base de données
	db, err := nosql.NewDatabase()
	if err != nil {
		log.Fatalf("Erreur lors de la création de la base de données: %v", err)
	}

	// Démarrer la base de données
	db.Start()
	defer db.Stop()

	// Créer une nouvelle collection "clients"
	clients, err := db.CreateCollection("clients")
	if err != nil {
		log.Fatalf("Erreur lors de la création de la collection: %v", err)
	}

	// Ajouter des champs à la collection
	clients.AddField("nom", nosql.FieldTypeString)
	clients.AddField("age", nosql.FieldTypeNumber)
	clients.AddField("actif", nosql.FieldTypeBoolean)

	// Insérer un document
	client := map[string]interface{}{
		"nom":   "John Doe",
		"age":   30,
		"actif": true,
	}

	id, err := clients.Insert(client)
	if err != nil {
		log.Printf("Erreur lors de l'insertion: %v", err)
	} else {
		fmt.Printf("Client inséré avec l'ID: %s\n", id)
	}

	// Récupérer le document
	doc, err := clients.Get(id)
	if err != nil {
		log.Printf("Erreur lors de la récupération: %v", err)
	} else {
		fmt.Printf("Client récupéré: %v\n", doc)
	}

	// Mettre à jour le document
	doc["age"] = 31
	if err := clients.Update(id, doc); err != nil {
		log.Printf("Erreur lors de la mise à jour: %v", err)
	} else {
		fmt.Println("Client mis à jour avec succès")
	}

	// Rechercher des clients par âge
	results, err := clients.FindByField("age", 31)
	if err != nil {
		log.Printf("Erreur lors de la recherche: %v", err)
	} else {
		fmt.Printf("Clients trouvés: %v\n", results)
	}

	// Lister tous les clients
	allClients, err := clients.GetAll()
	if err != nil {
		log.Printf("Erreur lors de la récupération de tous les clients: %v", err)
	} else {
		fmt.Printf("Nombre total de clients: %d\n", len(allClients))
		for i, client := range allClients {
			fmt.Printf("%d. %v\n", i+1, client)
		}
	}

	// Supprimer le client
	if err := clients.Delete(id); err != nil {
		log.Printf("Erreur lors de la suppression: %v", err)
	} else {
		fmt.Println("Client supprimé avec succès")
	}
}
