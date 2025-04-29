package main

import (
	"bufio"
	"fmt"
	"log"
	"nosql-db/internal/orm/models"
	"os"
	"strconv"
	"strings"
)

func main() {
	// Initialisation du gestionnaire de base de données
	fmt.Println("Initialisation de la base de données...")
	dbManager, err := models.NewDatabaseManager()
	if err != nil {
		log.Fatalf("Erreur: %v", err)
	}

	// Démarrer le gestionnaire
	dbManager.Start()
	log.Println("Base de données prête !")

	// Boucle interactive
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Println("\n=== NoSQL CLI ===")
		fmt.Println("Que souhaitez-vous faire ?")
		fmt.Println("1. Créer une nouvelle collection")
		fmt.Println("2. Lister les collections existantes")
		fmt.Println("3. Manipuler une collection")
		fmt.Println("4. Quitter")
		fmt.Print("\nVotre choix (1-4): ")

		scanner.Scan()
		choice := scanner.Text()

		switch choice {
		case "1":
			createNewCollection(scanner, dbManager)
		case "2":
			listCollections(dbManager)
		case "3":
			manipulateCollection(scanner, dbManager)
		case "4":
			fmt.Println("Au revoir !")
			return
		default:
			fmt.Println("Choix invalide. Veuillez choisir un numéro entre 1 et 4.")
		}
	}
}

func createNewCollection(scanner *bufio.Scanner, dbManager *models.DatabaseManager) {
	fmt.Print("Nom de la nouvelle collection: ")
	scanner.Scan()
	collectionName := strings.TrimSpace(scanner.Text())

	// Créer un nouveau modèle dynamique
	model, err := models.NewDynamicModel(dbManager.GetDatabase(), collectionName)
	if err != nil {
		fmt.Printf("Erreur lors de la création de la collection: %v\n", err)
		return
	}

	// Ajouter le modèle à la base de données
	dbManager.AddModel(collectionName, model)

	fmt.Printf("\nCollection '%s' créée avec succès !\n", collectionName)
	fmt.Println("Voulez-vous ajouter des champs à cette collection ? (oui/non)")

	scanner.Scan()
	if strings.ToLower(scanner.Text()) == "oui" {
		addFieldsToCollection(scanner, model)
	}
}

func addFieldsToCollection(scanner *bufio.Scanner, model *models.DynamicModel) {
	for {
		fmt.Println("\nAjouter un nouveau champ ? (oui/non)")
		scanner.Scan()
		if strings.ToLower(scanner.Text()) != "oui" {
			break
		}

		fmt.Print("Nom du champ: ")
		scanner.Scan()
		fieldName := strings.TrimSpace(scanner.Text())

		fmt.Println("Type du champ:")
		fmt.Println("1. Texte")
		fmt.Println("2. Nombre")
		fmt.Println("3. Booléen")
		fmt.Print("Votre choix (1-3): ")

		scanner.Scan()
		fieldType := scanner.Text()

		var fieldTypeStr string
		switch fieldType {
		case "1":
			fieldTypeStr = "string"
		case "2":
			fieldTypeStr = "number"
		case "3":
			fieldTypeStr = "boolean"
		default:
			fmt.Println("Type invalide, champ ignoré.")
			continue
		}

		model.AddField(fieldName, fieldTypeStr)
		fmt.Printf("Champ '%s' ajouté avec succès !\n", fieldName)
	}
}

func listCollections(dbManager *models.DatabaseManager) {
	fmt.Println("\nCollections disponibles:")
	collections := []string{"users", "products"}
	for _, name := range collections {
		if _, exists := dbManager.GetModel(name); exists {
			fmt.Printf("  - %s\n", name)
		}
	}
}

func manipulateCollection(scanner *bufio.Scanner, dbManager *models.DatabaseManager) {
	fmt.Print("Nom de la collection à manipuler: ")
	scanner.Scan()
	collectionName := strings.TrimSpace(scanner.Text())

	model, exists := dbManager.GetModel(collectionName)
	if !exists {
		fmt.Printf("Collection '%s' non trouvée\n", collectionName)
		return
	}

	collection := model.GetCollection()
	fmt.Printf("\nCollection '%s' sélectionnée\n", collectionName)

	for {
		fmt.Println("\nQue souhaitez-vous faire ?")
		fmt.Println("1. Ajouter un document")
		fmt.Println("2. Lister les documents")
		fmt.Println("3. Rechercher un document")
		fmt.Println("4. Modifier un document")
		fmt.Println("5. Supprimer un document")
		fmt.Println("6. Retour au menu principal")
		fmt.Print("\nVotre choix (1-6): ")

		scanner.Scan()
		choice := scanner.Text()

		switch choice {
		case "1":
			addDocument(scanner, collection)
		case "2":
			listDocuments(collection)
		case "3":
			searchDocument(scanner, collection)
		case "4":
			updateDocument(scanner, collection)
		case "5":
			deleteDocument(scanner, collection)
		case "6":
			return
		default:
			fmt.Println("Choix invalide. Veuillez choisir un numéro entre 1 et 6.")
		}
	}
}

func addDocument(scanner *bufio.Scanner, collection interface{}) {
	doc := make(map[string]interface{})

	// Récupérer les champs de la collection
	fields := getCollectionFields(collection)
	for fieldName, fieldType := range fields {
		fmt.Printf("Valeur pour le champ '%s' (%s): ", fieldName, fieldType)
		scanner.Scan()
		value := scanner.Text()

		// Convertir la valeur selon le type
		switch fieldType {
		case "number":
			num, err := strconv.ParseFloat(value, 64)
			if err != nil {
				fmt.Printf("Erreur: la valeur doit être un nombre\n")
				return
			}
			doc[fieldName] = num
		case "boolean":
			doc[fieldName] = strings.ToLower(value) == "true"
		default:
			doc[fieldName] = value
		}
	}

	if c, ok := collection.(interface {
		Insert(interface{}) (string, error)
	}); ok {
		id, err := c.Insert(doc)
		if err != nil {
			fmt.Printf("Erreur lors de l'insertion: %v\n", err)
		} else {
			fmt.Printf("Document inséré avec l'ID: %s\n", id)
		}
	}
}

func getCollectionFields(collection interface{}) map[string]string {
	if dc, ok := collection.(*models.DynamicCollection); ok {
		fields := make(map[string]string)
		for _, field := range dc.GetFields() {
			fields[field.Name] = field.Type
		}
		return fields
	}
	// Pour les collections prédéfinies, retourner les champs par défaut
	return map[string]string{
		"name":   "string",
		"age":    "number",
		"active": "boolean",
	}
}

func listDocuments(collection interface{}) {
	if c, ok := collection.(interface{ GetAll() ([]interface{}, error) }); ok {
		docs, err := c.GetAll()
		if err != nil {
			fmt.Printf("Erreur lors de la récupération des documents: %v\n", err)
			return
		}

		fmt.Println("\nDocuments dans la collection:")
		for i, doc := range docs {
			fmt.Printf("%d. %v\n", i+1, doc)
		}
	}
}

func searchDocument(scanner *bufio.Scanner, collection interface{}) {
	fmt.Print("Champ à rechercher: ")
	scanner.Scan()
	field := scanner.Text()

	fmt.Print("Valeur à rechercher: ")
	scanner.Scan()
	value := scanner.Text()

	if c, ok := collection.(interface {
		FindByField(string, interface{}) ([]interface{}, error)
	}); ok {
		docs, err := c.FindByField(field, value)
		if err != nil {
			fmt.Printf("Erreur lors de la recherche: %v\n", err)
			return
		}

		fmt.Printf("\nDocuments trouvés: %d\n", len(docs))
		for i, doc := range docs {
			fmt.Printf("%d. %v\n", i+1, doc)
		}
	}
}

func updateDocument(scanner *bufio.Scanner, collection interface{}) {
	fmt.Print("ID du document à modifier: ")
	scanner.Scan()
	id := scanner.Text()

	doc := make(map[string]interface{})
	fields := getCollectionFields(collection)
	for fieldName, fieldType := range fields {
		fmt.Printf("Nouvelle valeur pour '%s' (%s) (laissez vide pour ne pas modifier): ", fieldName, fieldType)
		scanner.Scan()
		value := scanner.Text()
		if value != "" {
			switch fieldType {
			case "number":
				num, err := strconv.ParseFloat(value, 64)
				if err != nil {
					fmt.Printf("Erreur: la valeur doit être un nombre\n")
					return
				}
				doc[fieldName] = num
			case "boolean":
				doc[fieldName] = strings.ToLower(value) == "true"
			default:
				doc[fieldName] = value
			}
		}
	}

	if c, ok := collection.(interface {
		Update(string, interface{}) error
	}); ok {
		if err := c.Update(id, doc); err != nil {
			fmt.Printf("Erreur lors de la mise à jour: %v\n", err)
		} else {
			fmt.Println("Document mis à jour avec succès")
		}
	}
}

func deleteDocument(scanner *bufio.Scanner, collection interface{}) {
	fmt.Print("ID du document à supprimer: ")
	scanner.Scan()
	id := scanner.Text()

	if c, ok := collection.(interface{ Delete(string) error }); ok {
		if err := c.Delete(id); err != nil {
			fmt.Printf("Erreur lors de la suppression: %v\n", err)
		} else {
			fmt.Println("Document supprimé avec succès")
		}
	}
}
