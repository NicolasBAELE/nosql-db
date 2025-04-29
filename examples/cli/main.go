package main

import (
	"fmt"
	"log"
	"nosql-db/pkg/nosql"
	"os"
	"strconv"
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

	fmt.Println("=== NoSQL Database CLI ===")
	fmt.Println("1. Créer une nouvelle collection")
	fmt.Println("2. Lister les collections existantes")
	fmt.Println("3. Manipuler une collection")
	fmt.Println("4. Quitter")

	for {
		fmt.Print("\nChoisissez une option (1-4): ")
		var choice int
		fmt.Scanf("%d", &choice)

		switch choice {
		case 1:
			createCollection(db)
		case 2:
			listCollections(db)
		case 3:
			manipulateCollection(db)
		case 4:
			fmt.Println("Au revoir!")
			os.Exit(0)
		default:
			fmt.Println("Option invalide. Veuillez choisir entre 1 et 4.")
		}
	}
}

func createCollection(db *nosql.Database) {
	fmt.Print("Nom de la collection: ")
	var name string
	fmt.Scanf("%s", &name)

	collection, err := db.CreateCollection(name)
	if err != nil {
		fmt.Printf("Erreur lors de la création de la collection: %v\n", err)
		return
	}

	fmt.Println("Collection créée avec succès!")
	fmt.Println("Ajoutons des champs à la collection.")

	for {
		fmt.Print("Nom du champ (ou 'fin' pour terminer): ")
		var fieldName string
		fmt.Scanf("%s", &fieldName)

		if fieldName == "fin" {
			break
		}

		fmt.Println("Type de champ:")
		fmt.Println("1. Texte")
		fmt.Println("2. Nombre")
		fmt.Println("3. Booléen")

		fmt.Print("Choisissez un type (1-3): ")
		var typeChoice int
		fmt.Scanf("%d", &typeChoice)

		var fieldType nosql.FieldType
		switch typeChoice {
		case 1:
			fieldType = nosql.FieldTypeString
		case 2:
			fieldType = nosql.FieldTypeNumber
		case 3:
			fieldType = nosql.FieldTypeBoolean
		default:
			fmt.Println("Type invalide. Veuillez choisir entre 1 et 3.")
			continue
		}

		collection.AddField(fieldName, fieldType)
		fmt.Printf("Champ '%s' ajouté avec succès!\n", fieldName)
	}

	fmt.Println("Collection configurée avec succès!")
}

func listCollections(db *nosql.Database) {
	collections := db.ListCollections()
	if len(collections) == 0 {
		fmt.Println("Aucune collection n'existe.")
		return
	}

	fmt.Println("Collections disponibles:")
	for i, name := range collections {
		fmt.Printf("%d. %s\n", i+1, name)
	}
}

func manipulateCollection(db *nosql.Database) {
	collections := db.ListCollections()
	if len(collections) == 0 {
		fmt.Println("Aucune collection n'existe. Veuillez d'abord en créer une.")
		return
	}

	fmt.Println("Collections disponibles:")
	for i, name := range collections {
		fmt.Printf("%d. %s\n", i+1, name)
	}

	fmt.Print("Choisissez une collection (numéro): ")
	var choice int
	fmt.Scanf("%d", &choice)

	if choice < 1 || choice > len(collections) {
		fmt.Println("Choix invalide.")
		return
	}

	collectionName := collections[choice-1]
	collection, exists := db.GetCollection(collectionName)
	if !exists {
		fmt.Printf("Collection '%s' non trouvée.\n", collectionName)
		return
	}

	fmt.Println("\nOpérations disponibles:")
	fmt.Println("1. Ajouter un document")
	fmt.Println("2. Lister les documents")
	fmt.Println("3. Rechercher des documents")
	fmt.Println("4. Mettre à jour un document")
	fmt.Println("5. Supprimer un document")
	fmt.Println("6. Retour au menu principal")

	fmt.Print("Choisissez une opération (1-6): ")
	var opChoice int
	fmt.Scanf("%d", &opChoice)

	switch opChoice {
	case 1:
		addDocument(collection)
	case 2:
		listDocuments(collection)
	case 3:
		searchDocuments(collection)
	case 4:
		updateDocument(collection)
	case 5:
		deleteDocument(collection)
	case 6:
		return
	default:
		fmt.Println("Option invalide.")
	}
}

func addDocument(collection *nosql.Collection) {
	fields := collection.GetFields()
	if len(fields) == 0 {
		fmt.Println("Cette collection n'a pas de champs définis.")
		return
	}

	doc := make(map[string]interface{})

	for _, field := range fields {
		fmt.Printf("Valeur pour '%s' (%s): ", field.Name, field.Type)
		var value string
		fmt.Scanf("%s", &value)

		switch field.Type {
		case nosql.FieldTypeString:
			doc[field.Name] = value
		case nosql.FieldTypeNumber:
			num, err := strconv.ParseFloat(value, 64)
			if err != nil {
				fmt.Printf("Erreur: '%s' n'est pas un nombre valide.\n", value)
				return
			}
			doc[field.Name] = num
		case nosql.FieldTypeBoolean:
			if value == "true" || value == "1" {
				doc[field.Name] = true
			} else if value == "false" || value == "0" {
				doc[field.Name] = false
			} else {
				fmt.Printf("Erreur: '%s' n'est pas un booléen valide (utilisez 'true' ou 'false').\n", value)
				return
			}
		}
	}

	id, err := collection.Insert(doc)
	if err != nil {
		fmt.Printf("Erreur lors de l'insertion: %v\n", err)
		return
	}

	fmt.Printf("Document inséré avec succès! ID: %s\n", id)
}

func listDocuments(collection *nosql.Collection) {
	docs, err := collection.GetAll()
	if err != nil {
		fmt.Printf("Erreur lors de la récupération des documents: %v\n", err)
		return
	}

	if len(docs) == 0 {
		fmt.Println("Aucun document dans cette collection.")
		return
	}

	fmt.Printf("Documents dans la collection (%d):\n", len(docs))
	for i, doc := range docs {
		fmt.Printf("%d. %v\n", i+1, doc)
	}
}

func searchDocuments(collection *nosql.Collection) {
	fields := collection.GetFields()
	if len(fields) == 0 {
		fmt.Println("Cette collection n'a pas de champs définis.")
		return
	}

	fmt.Println("Champs disponibles:")
	for i, field := range fields {
		fmt.Printf("%d. %s (%s)\n", i+1, field.Name, field.Type)
	}

	fmt.Print("Choisissez un champ (numéro): ")
	var fieldChoice int
	fmt.Scanf("%d", &fieldChoice)

	if fieldChoice < 1 || fieldChoice > len(fields) {
		fmt.Println("Choix invalide.")
		return
	}

	field := fields[fieldChoice-1]
	fmt.Printf("Valeur à rechercher pour '%s': ", field.Name)
	var value string
	fmt.Scanf("%s", &value)

	var searchValue interface{}
	switch field.Type {
	case nosql.FieldTypeString:
		searchValue = value
	case nosql.FieldTypeNumber:
		num, err := strconv.ParseFloat(value, 64)
		if err != nil {
			fmt.Printf("Erreur: '%s' n'est pas un nombre valide.\n", value)
			return
		}
		searchValue = num
	case nosql.FieldTypeBoolean:
		if value == "true" || value == "1" {
			searchValue = true
		} else if value == "false" || value == "0" {
			searchValue = false
		} else {
			fmt.Printf("Erreur: '%s' n'est pas un booléen valide (utilisez 'true' ou 'false').\n", value)
			return
		}
	}

	results, err := collection.FindByField(field.Name, searchValue)
	if err != nil {
		fmt.Printf("Erreur lors de la recherche: %v\n", err)
		return
	}

	if len(results) == 0 {
		fmt.Println("Aucun document trouvé.")
		return
	}

	fmt.Printf("Documents trouvés (%d):\n", len(results))
	for i, doc := range results {
		fmt.Printf("%d. %v\n", i+1, doc)
	}
}

func updateDocument(collection *nosql.Collection) {
	docs, err := collection.GetAll()
	if err != nil {
		fmt.Printf("Erreur lors de la récupération des documents: %v\n", err)
		return
	}

	if len(docs) == 0 {
		fmt.Println("Aucun document dans cette collection.")
		return
	}

	fmt.Printf("Documents dans la collection (%d):\n", len(docs))
	for i, doc := range docs {
		fmt.Printf("%d. %v\n", i+1, doc)
	}

	fmt.Print("Choisissez un document à mettre à jour (numéro): ")
	var docChoice int
	fmt.Scanf("%d", &docChoice)

	if docChoice < 1 || docChoice > len(docs) {
		fmt.Println("Choix invalide.")
		return
	}

	// Pour simplifier, nous allons juste afficher les champs disponibles
	// et demander à l'utilisateur de saisir les nouvelles valeurs
	fields := collection.GetFields()
	fmt.Println("Entrez les nouvelles valeurs (laissez vide pour conserver la valeur actuelle):")

	updateDoc := make(map[string]interface{})
	for _, field := range fields {
		fmt.Printf("Nouvelle valeur pour '%s' (%s): ", field.Name, field.Type)
		var value string
		fmt.Scanf("%s", &value)

		if value == "" {
			continue // Conserver la valeur actuelle
		}

		switch field.Type {
		case nosql.FieldTypeString:
			updateDoc[field.Name] = value
		case nosql.FieldTypeNumber:
			num, err := strconv.ParseFloat(value, 64)
			if err != nil {
				fmt.Printf("Erreur: '%s' n'est pas un nombre valide.\n", value)
				return
			}
			updateDoc[field.Name] = num
		case nosql.FieldTypeBoolean:
			if value == "true" || value == "1" {
				updateDoc[field.Name] = true
			} else if value == "false" || value == "0" {
				updateDoc[field.Name] = false
			} else {
				fmt.Printf("Erreur: '%s' n'est pas un booléen valide (utilisez 'true' ou 'false').\n", value)
				return
			}
		}
	}

	// Pour simplifier, nous allons utiliser l'ID du document comme clé
	// Dans une implémentation réelle, il faudrait extraire l'ID du document
	docID := fmt.Sprintf("%d", docChoice) // Simplification

	err = collection.Update(docID, updateDoc)
	if err != nil {
		fmt.Printf("Erreur lors de la mise à jour: %v\n", err)
		return
	}

	fmt.Println("Document mis à jour avec succès!")
}

func deleteDocument(collection *nosql.Collection) {
	docs, err := collection.GetAll()
	if err != nil {
		fmt.Printf("Erreur lors de la récupération des documents: %v\n", err)
		return
	}

	if len(docs) == 0 {
		fmt.Println("Aucun document dans cette collection.")
		return
	}

	fmt.Printf("Documents dans la collection (%d):\n", len(docs))
	for i, doc := range docs {
		fmt.Printf("%d. %v\n", i+1, doc)
	}

	fmt.Print("Choisissez un document à supprimer (numéro): ")
	var docChoice int
	fmt.Scanf("%d", &docChoice)

	if docChoice < 1 || docChoice > len(docs) {
		fmt.Println("Choix invalide.")
		return
	}

	// Pour simplifier, nous allons utiliser l'index comme ID
	// Dans une implémentation réelle, il faudrait extraire l'ID du document
	docID := fmt.Sprintf("%d", docChoice) // Simplification

	err = collection.Delete(docID)
	if err != nil {
		fmt.Printf("Erreur lors de la suppression: %v\n", err)
		return
	}

	fmt.Println("Document supprimé avec succès!")
}
