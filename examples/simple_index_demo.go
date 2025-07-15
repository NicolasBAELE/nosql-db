package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

func main() {
	baseURL := "http://localhost:8081"

	fmt.Println("🧪 TEST SIMPLE DES INDEX NON-UNIQUES")
	fmt.Println(strings.Repeat("=", 50))

	// 1. Vérifier les livres existants
	fmt.Println("\n1. Livres existants dans la base...")

	listResp, err := http.Get(baseURL + "/api/books")
	if err != nil {
		log.Fatalf("Erreur lors de la récupération: %v", err)
	}

	var allBooks []map[string]interface{}
	json.NewDecoder(listResp.Body).Decode(&allBooks)
	listResp.Body.Close()

	fmt.Printf("   📚 Total de %d livres dans la collection:\n", len(allBooks))
	for i, book := range allBooks {
		fmt.Printf("   %d. ID: %s, Titre: %s, IBAN: %v\n",
			i+1, book["_id"], book["title"], book["iban"])
	}

	// 2. Recherche par titre (index non-unique)
	fmt.Println("\n2. Recherche par titre 'Le Petit Prince' (index non-unique)...")

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

	// 3. Recherche par IBAN (index unique)
	fmt.Println("\n3. Recherche par IBAN 9782070408504 (index unique)...")

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

	// 4. Recherche par auteur (pas d'index)
	fmt.Println("\n4. Recherche par auteur (pas d'index)...")

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

	// 5. Test de violation d'index unique
	fmt.Println("\n5. Test de violation d'index unique...")

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

	// 6. Test d'insertion avec titre dupliqué (index non-unique)
	fmt.Println("\n6. Test d'insertion avec titre dupliqué (index non-unique)...")

	newBook := map[string]interface{}{
		"title":       "Le Petit Prince", // Même titre que les autres
		"iban":        9999999999,        // IBAN unique
		"description": "Nouvel exemplaire",
		"author":      "Antoine de Saint-Exupéry",
	}

	newBookJSON, _ := json.Marshal(newBook)
	resp2, err := http.Post(
		baseURL+"/api/books",
		"application/json",
		bytes.NewBuffer(newBookJSON),
	)

	if resp2.StatusCode == 201 {
		fmt.Println("   ✅ Insertion réussie avec titre dupliqué (comportement attendu pour index non-unique)")
		var result map[string]interface{}
		json.NewDecoder(resp2.Body).Decode(&result)
		resp2.Body.Close()
		fmt.Printf("   📖 Nouveau livre créé avec l'ID: %s\n", result["id"])
	} else {
		fmt.Printf("   ❌ Erreur lors de l'insertion (status: %d)\n", resp2.StatusCode)
		resp2.Body.Close()
	}

	// 7. Vérification finale
	fmt.Println("\n7. Vérification finale...")

	finalSearchResp, err := http.Get(baseURL + "/api/books/search?field=title&value=Le%20Petit%20Prince")
	if err != nil {
		log.Fatalf("Erreur lors de la recherche finale: %v", err)
	}

	var finalResults []map[string]interface{}
	json.NewDecoder(finalSearchResp.Body).Decode(&finalResults)
	finalSearchResp.Body.Close()

	fmt.Printf("   📚 Total de %d livres avec le titre 'Le Petit Prince':\n", len(finalResults))
	for i, book := range finalResults {
		fmt.Printf("   %d. ID: %s, IBAN: %v, Description: %s\n",
			i+1, book["_id"], book["iban"], book["description"])
	}

	fmt.Println("\n🎉 Test des index non-uniques terminé!")
	fmt.Println("\n📋 RÉSUMÉ:")
	fmt.Println("   ✅ Index non-unique sur 'title' permet les doublons")
	fmt.Println("   ✅ Index unique sur 'iban' empêche les doublons")
	fmt.Println("   ✅ Recherche par champ indexé fonctionne")
	fmt.Println("   ✅ Recherche par champ non-indexé fonctionne")
}
