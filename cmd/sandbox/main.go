package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"nosql-db/internal/orm/models"
	"os"
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
		fmt.Println("\n=== Sandbox NoSQL ===")
		fmt.Println("Commandes disponibles:")
		fmt.Println("  help                    - Afficher l'aide")
		fmt.Println("  collections             - Lister les collections disponibles")
		fmt.Println("  use <collection>        - Sélectionner une collection")
		fmt.Println("  insert <json>           - Insérer un document")
		fmt.Println("  get <id>                - Récupérer un document")
		fmt.Println("  update <id> <json>      - Mettre à jour un document")
		fmt.Println("  delete <id>             - Supprimer un document")
		fmt.Println("  find <field> <value>    - Rechercher des documents")
		fmt.Println("  count                   - Compter les documents")
		fmt.Println("  clear                   - Vider la collection")
		fmt.Println("  quit                    - Quitter")
		fmt.Print("\n> ")

		scanner.Scan()
		input := scanner.Text()
		parts := strings.Fields(input)
		if len(parts) == 0 {
			continue
		}

		command := parts[0]
		args := parts[1:]

		switch command {
		case "help":
			// L'aide est déjà affichée au début de la boucle

		case "collections":
			fmt.Println("Collections disponibles:")
			// Liste des collections prédéfinies
			collections := []string{"users", "products"}
			for _, name := range collections {
				if _, exists := dbManager.GetModel(name); exists {
					fmt.Printf("  - %s\n", name)
				}
			}

		case "use":
			if len(args) < 1 {
				fmt.Println("Usage: use <collection>")
				continue
			}
			collectionName := args[0]
			model, exists := dbManager.GetModel(collectionName)
			if !exists {
				fmt.Printf("Collection '%s' non trouvée\n", collectionName)
				continue
			}
			collection := model.GetCollection()
			fmt.Printf("Collection '%s' sélectionnée\n", collectionName)
			handleCollectionCommands(collection, scanner)

		case "quit":
			fmt.Println("Au revoir !")
			return

		default:
			fmt.Println("Commande non reconnue. Tapez 'help' pour voir les commandes disponibles.")
		}
	}
}

func handleCollectionCommands(collection interface{}, scanner *bufio.Scanner) {
	for {
		fmt.Print("collection> ")
		scanner.Scan()
		input := scanner.Text()
		parts := strings.Fields(input)
		if len(parts) == 0 {
			continue
		}

		command := parts[0]
		args := parts[1:]

		switch command {
		case "help":
			fmt.Println("Commandes de collection:")
			fmt.Println("  insert <json>           - Insérer un document")
			fmt.Println("  get <id>                - Récupérer un document")
			fmt.Println("  update <id> <json>      - Mettre à jour un document")
			fmt.Println("  delete <id>             - Supprimer un document")
			fmt.Println("  find <field> <value>    - Rechercher des documents")
			fmt.Println("  count                   - Compter les documents")
			fmt.Println("  clear                   - Vider la collection")
			fmt.Println("  back                    - Retourner au menu principal")

		case "insert":
			if len(args) < 1 {
				fmt.Println("Usage: insert <json>")
				continue
			}
			jsonStr := strings.Join(args, " ")
			var doc interface{}
			if err := json.Unmarshal([]byte(jsonStr), &doc); err != nil {
				fmt.Printf("Erreur JSON: %v\n", err)
				continue
			}
			if c, ok := collection.(interface {
				Insert(interface{}) (string, error)
			}); ok {
				id, err := c.Insert(doc)
				if err != nil {
					fmt.Printf("Erreur: %v\n", err)
				} else {
					fmt.Printf("Document inséré avec l'ID: %s\n", id)
				}
			}

		case "get":
			if len(args) < 1 {
				fmt.Println("Usage: get <id>")
				continue
			}
			id := args[0]
			if c, ok := collection.(interface {
				Get(string) (interface{}, error)
			}); ok {
				doc, err := c.Get(id)
				if err != nil {
					fmt.Printf("Erreur: %v\n", err)
				} else {
					jsonBytes, _ := json.MarshalIndent(doc, "", "  ")
					fmt.Println(string(jsonBytes))
				}
			}

		case "update":
			if len(args) < 2 {
				fmt.Println("Usage: update <id> <json>")
				continue
			}
			id := args[0]
			jsonStr := strings.Join(args[1:], " ")
			var doc interface{}
			if err := json.Unmarshal([]byte(jsonStr), &doc); err != nil {
				fmt.Printf("Erreur JSON: %v\n", err)
				continue
			}
			if c, ok := collection.(interface {
				Update(string, interface{}) error
			}); ok {
				if err := c.Update(id, doc); err != nil {
					fmt.Printf("Erreur: %v\n", err)
				} else {
					fmt.Println("Document mis à jour")
				}
			}

		case "delete":
			if len(args) < 1 {
				fmt.Println("Usage: delete <id>")
				continue
			}
			id := args[0]
			if c, ok := collection.(interface{ Delete(string) error }); ok {
				if err := c.Delete(id); err != nil {
					fmt.Printf("Erreur: %v\n", err)
				} else {
					fmt.Println("Document supprimé")
				}
			}

		case "find":
			if len(args) < 2 {
				fmt.Println("Usage: find <field> <value>")
				continue
			}
			field := args[0]
			value := args[1]
			if c, ok := collection.(interface {
				FindByField(string, interface{}) ([]interface{}, error)
			}); ok {
				docs, err := c.FindByField(field, value)
				if err != nil {
					fmt.Printf("Erreur: %v\n", err)
				} else {
					fmt.Printf("Documents trouvés: %d\n", len(docs))
					for i, doc := range docs {
						jsonBytes, _ := json.MarshalIndent(doc, "", "  ")
						fmt.Printf("%d. %s\n", i+1, string(jsonBytes))
					}
				}
			}

		case "count":
			if c, ok := collection.(interface{ Count() int }); ok {
				count := c.Count()
				fmt.Printf("Nombre de documents: %d\n", count)
			}

		case "clear":
			if c, ok := collection.(interface{ Clear() error }); ok {
				if err := c.Clear(); err != nil {
					fmt.Printf("Erreur: %v\n", err)
				} else {
					fmt.Println("Collection vidée")
				}
			}

		case "back":
			return

		default:
			fmt.Println("Commande non reconnue. Tapez 'help' pour voir les commandes disponibles.")
		}
	}
}
