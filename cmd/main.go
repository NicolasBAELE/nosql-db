package main

import (
	"fmt"
	"log"

	"nosql-db/internal/database"
)

func main() {
	// Créer une nouvelle base de données
	db, err := database.NewDatabase("data")
	if err != nil {
		log.Fatal(err)
	}

	// Créer une collection
	users, err := db.CreateCollection("users")
	if err != nil {
		log.Fatal(err)
	}

	// Créer un index unique sur le champ "email"
	err = users.CreateIndex("email", true)
	if err != nil {
		log.Fatal(err)
	}

	// Créer un index non-unique sur le champ "age"
	err = users.CreateIndex("age", false)
	if err != nil {
		log.Fatal(err)
	}

	// Insérer un document
	doc := database.Document{
		"name":  "John Doe",
		"email": "john@example.com",
		"age":   30,
	}

	id, err := users.Insert(doc)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Document inséré avec l'ID: %s\n", id)

	// Essayer d'insérer un document avec le même email (devrait échouer)
	duplicateDoc := database.Document{
		"name":  "John Smith",
		"email": "john@example.com", // Même email
		"age":   35,
	}
	_, err = users.Insert(duplicateDoc)
	if err != nil {
		fmt.Printf("Erreur attendue (violation d'index unique): %v\n", err)
	}

	// Insérer un autre document avec un âge différent
	doc2 := database.Document{
		"name":  "Jane Doe",
		"email": "jane@example.com",
		"age":   30, // Même âge que John
	}
	id2, err := users.Insert(doc2)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Deuxième document inséré avec l'ID: %s\n", id2)

	// Trouver un document par ID
	foundDoc, err := users.FindByID(id)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Document trouvé par ID: %+v\n", foundDoc)

	// Trouver des documents par email (index unique)
	results, err := users.FindByIndex("email", "john@example.com")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Documents trouvés par email: %+v\n", results)

	// Trouver des documents par âge (index non-unique)
	results, err = users.FindByIndex("age", 30)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Documents trouvés par âge: %+v\n", results)

	// Mettre à jour un document
	updateDoc := database.Document{
		"name":  "John Doe",
		"email": "john.doe@example.com", // Changement d'email
		"age":   31,                     // Changement d'âge
	}

	err = users.Update(id, updateDoc)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Document mis à jour")

	// Vérifier la mise à jour
	updatedDoc, err := users.FindByID(id)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Document mis à jour: %+v\n", updatedDoc)

	// Supprimer un document
	err = users.Delete(id)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Document supprimé")

	// Vérifier la suppression
	_, err = users.FindByID(id)
	if err != nil {
		fmt.Printf("Erreur attendue car le document a été supprimé: %v\n", err)
	}
}
