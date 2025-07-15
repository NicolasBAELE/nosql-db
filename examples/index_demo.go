package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

func main() {
	baseURL := "http://localhost:8081"

	fmt.Println("🧪 TEST DES INDEX NON-UNIQUES")
	fmt.Println(strings.Repeat("=", 50))

	// 1. Insérer plusieurs livres avec le même titre (index non-unique)
	fmt.Println("\n1. Insertion de plusieurs livres avec le même titre...")

	books := []map[string]interface{}{
		{
			"title":       "Le Petit Prince",
			"iban":        9782070408504,
			"description": "Premier exemplaire",
			"author":      "Antoine de Saint-Exupéry",
		},
		{
			"title":       "Le Petit Prince",
			"iban":        9782070408505,
			"description": "Deuxième exemplaire",
			"author":      "Antoine de Saint-Exupéry",
		},
		{
			"title":       "Le Petit Prince",
			"iban":        9782070408506,
			"description": "Troisième exemplaire",
			"author":      "Antoine de Saint-Exupéry",
		},
	}

	var bookIDs []string
	for i, book := range books {
		fmt.Printf("   Insertion du livre %d...\n", i+1)

		bookJSON, _ := json.Marshal(book)
		resp, err := http.Post(
			baseURL+"/api/books",
			"application/json",
			bytes.NewBuffer(bookJSON),
		)
		if err != nil {
			log.Fatalf("Erreur lors de l'insertion du livre %d: %v", i+1, err)
		}

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		resp.Body.Close()

		bookID := result["id"].(string)
		bookIDs = append(bookIDs, bookID)
		fmt.Printf("   ✅ Livre %d inséré avec l'ID: %s\n", i+1, bookID)
	}

	// 2. Tester la recherche par titre (index non-unique)
	fmt.Println("\n2. Recherche par titre (index non-unique)...")

	searchResp, err := http.Get(baseURL + "/api/books/search?field=title&value=Le%20Petit%20Prince")
	if err != nil {
		log.Fatalf("Erreur lors de la recherche: %v", err)
	}

	var searchResults []map[string]interface{}
	json.NewDecoder(searchResp.Body).Decode(&searchResults)
	searchResp.Body.Close()

	fmt.Printf("   📚 Trouvé %d livres avec le titre 'Le Petit Prince':\n", len(searchResults))
	for i, book := range searchResults {
		fmt.Printf("   %d. ID: %s, IBAN: %v, Description: %s\n",
			i+1, book["_id"], book["iban"], book["description"])
	}

	// 3. Tester la recherche par IBAN (index unique)
	fmt.Println("\n3. Recherche par IBAN (index unique)...")

	searchResp2, err := http.Get(baseURL + "/api/books/search?field=iban&value=9782070408504")
	if err != nil {
		log.Fatalf("Erreur lors de la recherche par IBAN: %v", err)
	}

	var searchResults2 []map[string]interface{}
	json.NewDecoder(searchResp2.Body).Decode(&searchResults2)
	searchResp2.Body.Close()

	fmt.Printf("   📖 Trouvé %d livre(s) avec l'IBAN 9782070408504:\n", len(searchResults2))
	for i, book := range searchResults2 {
		fmt.Printf("   %d. ID: %s, Titre: %s\n",
			i+1, book["_id"], book["title"])
	}

	// 4. Tester la violation d'index unique
	fmt.Println("\n4. Test de violation d'index unique...")

	duplicateBook := map[string]interface{}{
		"title":       "Nouveau Livre",
		"iban":        9782070408504, // Même IBAN que le premier livre
		"description": "Ceci devrait échouer",
		"author":      "Test Author",
	}

	duplicateJSON, _ := json.Marshal(duplicateBook)
	resp, err := http.Post(
		baseURL+"/api/books",
		"application/json",
		bytes.NewBuffer(duplicateJSON),
	)

	if resp.StatusCode == 500 {
		fmt.Println("   ✅ Violation d'index unique détectée (comportement attendu)")
	} else {
		fmt.Printf("   ❌ Erreur: La violation d'index unique n'a pas été détectée (status: %d)\n", resp.StatusCode)
	}
	resp.Body.Close()

	// 5. Tester la recherche par auteur (pas d'index)
	fmt.Println("\n5. Recherche par auteur (pas d'index)...")

	searchResp3, err := http.Get(baseURL + "/api/books/search?field=author&value=Antoine%20de%20Saint-Exupéry")
	if err != nil {
		log.Fatalf("Erreur lors de la recherche par auteur: %v", err)
	}

	var searchResults3 []map[string]interface{}
	json.NewDecoder(searchResp3.Body).Decode(&searchResults3)
	searchResp3.Body.Close()

	fmt.Printf("   👨‍💼 Trouvé %d livre(s) de 'Antoine de Saint-Exupéry':\n", len(searchResults3))
	for i, book := range searchResults3 {
		fmt.Printf("   %d. ID: %s, Titre: %s\n",
			i+1, book["_id"], book["title"])
	}

	// 6. Lister tous les livres
	fmt.Println("\n6. Liste de tous les livres...")

	listResp, err := http.Get(baseURL + "/api/books")
	if err != nil {
		log.Fatalf("Erreur lors de la récupération de la liste: %v", err)
	}

	var allBooks []map[string]interface{}
	json.NewDecoder(listResp.Body).Decode(&allBooks)
	listResp.Body.Close()

	fmt.Printf("   📚 Total de %d livres dans la collection:\n", len(allBooks))
	for i, book := range allBooks {
		fmt.Printf("   %d. ID: %s, Titre: %s, IBAN: %v\n",
			i+1, book["_id"], book["title"], book["iban"])
	}

	// 7. Test de performance : recherche avec et sans index
	fmt.Println("\n7. Test de performance...")

	// Recherche avec index (titre)
	start := time.Now()
	_, err = http.Get(baseURL + "/api/books/search?field=title&value=Le%20Petit%20Prince")
	if err != nil {
		log.Fatalf("Erreur lors du test de performance: %v", err)
	}
	indexedTime := time.Since(start)

	// Recherche sans index (auteur)
	start = time.Now()
	_, err = http.Get(baseURL + "/api/books/search?field=author&value=Antoine%20de%20Saint-Exupéry")
	if err != nil {
		log.Fatalf("Erreur lors du test de performance: %v", err)
	}
	nonIndexedTime := time.Since(start)

	fmt.Printf("   ⚡ Recherche avec index (titre): %v\n", indexedTime)
	fmt.Printf("   🐌 Recherche sans index (auteur): %v\n", nonIndexedTime)
	fmt.Printf("   📊 Ratio: %.2fx plus rapide avec index\n", float64(nonIndexedTime)/float64(indexedTime))

	fmt.Println("\n🎉 Tests des index non-uniques terminés!")
}
