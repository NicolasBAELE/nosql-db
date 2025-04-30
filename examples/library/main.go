package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"nosql-db/internal/database"
)

type Book struct {
	Title     string
	Author    string
	ISBN      string
	Available bool
}

func main() {
	// Créer une nouvelle base de données
	db, err := database.NewDatabase("library_data")
	if err != nil {
		log.Fatal(err)
	}

	// Créer une collection pour les livres
	books, err := db.CreateCollection("books")
	if err != nil {
		log.Fatal(err)
	}

	// Créer un index unique sur l'ISBN
	err = books.CreateIndex("isbn", true)
	if err != nil {
		log.Fatal(err)
	}

	// Créer un index non-unique sur l'auteur
	err = books.CreateIndex("author", false)
	if err != nil {
		log.Fatal(err)
	}

	// Menu interactif
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Println("\n=== Gestion de Bibliothèque ===")
		fmt.Println("1. Ajouter un livre")
		fmt.Println("2. Rechercher un livre par ISBN")
		fmt.Println("3. Rechercher des livres par auteur")
		fmt.Println("4. Mettre à jour la disponibilité d'un livre")
		fmt.Println("5. Supprimer un livre")
		fmt.Println("6. Lister tous les livres")
		fmt.Println("7. Quitter")
		fmt.Print("Choix: ")

		scanner.Scan()
		choice := scanner.Text()

		switch choice {
		case "1":
			addBook(books, scanner)
		case "2":
			findBookByISBN(books, scanner)
		case "3":
			findBooksByAuthor(books, scanner)
		case "4":
			updateBookAvailability(books, scanner)
		case "5":
			deleteBook(books, scanner)
		case "6":
			listAllBooks(books)
		case "7":
			fmt.Println("Au revoir!")
			return
		default:
			fmt.Println("Choix invalide")
		}
	}
}

func addBook(books *database.Collection, scanner *bufio.Scanner) {
	fmt.Print("Titre: ")
	scanner.Scan()
	title := scanner.Text()

	fmt.Print("Auteur: ")
	scanner.Scan()
	author := scanner.Text()

	fmt.Print("ISBN: ")
	scanner.Scan()
	isbn := scanner.Text()

	book := database.Document{
		"title":     title,
		"author":    author,
		"isbn":      isbn,
		"available": true,
	}

	id, err := books.Insert(book)
	if err != nil {
		fmt.Printf("Erreur: %v\n", err)
		return
	}

	fmt.Printf("Livre ajouté avec l'ID: %s\n", id)
}

func findBookByISBN(books *database.Collection, scanner *bufio.Scanner) {
	fmt.Print("ISBN: ")
	scanner.Scan()
	isbn := scanner.Text()

	results, err := books.FindByIndex("isbn", isbn)
	if err != nil {
		fmt.Printf("Erreur: %v\n", err)
		return
	}

	if len(results) == 0 {
		fmt.Println("Aucun livre trouvé avec cet ISBN")
		return
	}

	printBook(results[0])
}

func findBooksByAuthor(books *database.Collection, scanner *bufio.Scanner) {
	fmt.Print("Auteur: ")
	scanner.Scan()
	author := scanner.Text()

	results, err := books.FindByIndex("author", author)
	if err != nil {
		fmt.Printf("Erreur: %v\n", err)
		return
	}

	if len(results) == 0 {
		fmt.Println("Aucun livre trouvé pour cet auteur")
		return
	}

	fmt.Printf("\nLivres trouvés (%d):\n", len(results))
	for _, book := range results {
		printBook(book)
	}
}

func updateBookAvailability(books *database.Collection, scanner *bufio.Scanner) {
	fmt.Print("ISBN du livre à mettre à jour: ")
	scanner.Scan()
	isbn := scanner.Text()

	results, err := books.FindByIndex("isbn", isbn)
	if err != nil {
		fmt.Printf("Erreur: %v\n", err)
		return
	}

	if len(results) == 0 {
		fmt.Println("Aucun livre trouvé avec cet ISBN")
		return
	}

	book := results[0]
	fmt.Print("Disponible (true/false): ")
	scanner.Scan()
	available, err := strconv.ParseBool(scanner.Text())
	if err != nil {
		fmt.Println("Valeur invalide")
		return
	}

	book["available"] = available
	err = books.Update(book["_id"].(string), book)
	if err != nil {
		fmt.Printf("Erreur: %v\n", err)
		return
	}

	fmt.Println("Livre mis à jour avec succès")
}

func deleteBook(books *database.Collection, scanner *bufio.Scanner) {
	fmt.Print("ISBN du livre à supprimer: ")
	scanner.Scan()
	isbn := scanner.Text()

	results, err := books.FindByIndex("isbn", isbn)
	if err != nil {
		fmt.Printf("Erreur: %v\n", err)
		return
	}

	if len(results) == 0 {
		fmt.Println("Aucun livre trouvé avec cet ISBN")
		return
	}

	err = books.Delete(results[0]["_id"].(string))
	if err != nil {
		fmt.Printf("Erreur: %v\n", err)
		return
	}

	fmt.Println("Livre supprimé avec succès")
}

func printBook(book database.Document) {
	fmt.Printf("\nID: %s\n", book["_id"])
	fmt.Printf("Titre: %s\n", book["title"])
	fmt.Printf("Auteur: %s\n", book["author"])
	fmt.Printf("ISBN: %s\n", book["isbn"])
	fmt.Printf("Disponible: %v\n", book["available"])
	fmt.Println(strings.Repeat("-", 30))
}

// listAllBooks liste tous les livres de la collection
func listAllBooks(books *database.Collection) {
	// Lire tous les fichiers JSON dans le répertoire de la collection
	files, err := os.ReadDir(books.GetPath())
	if err != nil {
		fmt.Printf("Erreur lors de la lecture des fichiers: %v\n", err)
		return
	}

	if len(files) == 0 {
		fmt.Println("Aucun livre dans la base de données")
		return
	}

	fmt.Printf("\nListe de tous les livres (%d):\n", len(files))
	for _, file := range files {
		if filepath.Ext(file.Name()) != ".json" {
			continue
		}

		// Lire le fichier
		data, err := os.ReadFile(filepath.Join(books.GetPath(), file.Name()))
		if err != nil {
			fmt.Printf("Erreur lors de la lecture du fichier %s: %v\n", file.Name(), err)
			continue
		}

		// Désérialiser le document
		var book database.Document
		if err := json.Unmarshal(data, &book); err != nil {
			fmt.Printf("Erreur lors de la désérialisation du fichier %s: %v\n", file.Name(), err)
			continue
		}

		printBook(book)
	}
}
